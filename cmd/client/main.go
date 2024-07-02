package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/kirillgashkov/assignment-youthumb/internal/api/client"
	"github.com/kirillgashkov/assignment-youthumb/internal/config"
	"github.com/kirillgashkov/assignment-youthumb/internal/logger"
	"github.com/kirillgashkov/assignment-youthumb/proto/youthumbpb/v1"
	"io"
	"log/slog"
	"mime"
	"os"
	"path/filepath"
	"sync"
)

const (
	maxRetry = 3
)

var (
	errUsage             = errors.New("usage")
	errThumbnailNotFound = errors.New("thumbnail not found")
)

var (
	isAsync   = flag.Bool("async", false, "Download thumbnails asynchronously.")
	outputDir = flag.String("output", "", "Download thumbnails to the specified directory.")
)

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), `Usage: %s [OPTIONS] VIDEO_URL...

Download thumbnails for YouTube videos with the specified URLs.

Arguments:
  VIDEO_URL
        The URL of the video to download the thumbnail for.

Options:
`, os.Args[0])
	flag.PrintDefaults()
}

func main() {
	err := mainErr()
	if errors.Is(err, errUsage) {
		flag.Usage()
		os.Exit(2)
	}
	if err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func mainErr() error {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() == 0 {
		return errUsage
	}

	cfg, err := config.New()
	if err != nil {
		return err
	}

	log, err := logger.New(cfg)
	if err != nil {
		return err
	}
	slog.SetDefault(log)

	clientConn, err := client.NewClientConn(cfg.GRPC)
	if err != nil {
		return err
	}

	cli, err := client.NewClient(clientConn)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	downloader := newThumbnailDownloader(cli, *outputDir)

	if *isAsync {
		func() {
			wg := sync.WaitGroup{}
			defer wg.Wait()

			for _, videoURL := range flag.Args() {
				wg.Add(1)
				go func() { // https://go.dev/blog/loopvar-preview
					defer wg.Done()
					if err := downloader.DownloadThumbnail(ctx, videoURL); err != nil {
						slog.Error("failed to download thumbnail", "video_url", videoURL, "error", err)
					}
				}()
			}
		}()
	} else {
		for _, videoURL := range flag.Args() {
			if err := downloader.DownloadThumbnail(ctx, videoURL); err != nil {
				slog.Error("failed to download thumbnail", "video_url", videoURL, "error", err)
			}
		}
	}

	return nil
}

func newThumbnailDownloader(cli youthumbpb.ThumbnailServiceClient, outputDir string) *thumbnailDownloader {
	return &thumbnailDownloader{cli: cli, outputDir: outputDir, muCh: make(chan struct{}, 1)}
}

type thumbnailDownloader struct {
	cli       youthumbpb.ThumbnailServiceClient
	outputDir string
	muCh      chan struct{}
}

func (d *thumbnailDownloader) DownloadThumbnail(ctx context.Context, videoURL string) error {
	contentFile, err := os.CreateTemp("", "thumbnail-*")
	if err != nil {
		return err
	}
	defer func(contentFile *os.File) {
		if err := contentFile.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
			slog.Error("failed to close file", "error", err)
		}
	}(contentFile)

	stream, err := d.cli.GetThumbnail(ctx, &youthumbpb.GetThumbnailRequest{VideoUrl: videoURL})
	if err != nil {
		return err
	}

	contentType := ""
	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		if chunk.ContentType != "" {
			contentType = chunk.ContentType
		}

		if _, err := contentFile.Write(chunk.Data); err != nil {
			return err
		}
	}

	if err := contentFile.Close(); err != nil {
		return err
	}

	extension := ""
	if contentType != "" {
		extensions, err := mime.ExtensionsByType(contentType)
		if err != nil {
			slog.Error("failed to get extensions by type", "content_type", contentType, "error", err)
		}

		if len(extensions) != 0 {
			// Last extension appears to be the most common one.
			extension = extensions[len(extensions)-1]
		}
	}

	select {
	case d.muCh <- struct{}{}:
		func() {
			defer func() {
				<-d.muCh
			}()

			outputFilePath := filepath.Join(d.outputDir, ""+extension)

			if err := os.Rename(contentFile.Name(), outputFilePath); err != nil {
				slog.Error("failed to rename file", "error", err)
			}
		}()
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}
