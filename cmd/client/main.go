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
	"log/slog"
	"os"
	"sync"
)

const (
	maxRetry = 3
)

var (
	errUsage             = errors.New("usage")
	errThumbnailNotFound = errors.New("thumbnail not found")
	errCanceled          = errors.New("canceled")
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

	wg := sync.WaitGroup{}
	defer wg.Wait()

	for _, videoURL := range flag.Args() {
		if *isAsync {
			wg.Add(1)
			go func() { // https://go.dev/blog/loopvar-preview
				defer wg.Done()
				if err := downloadThumbnail(ctx, cli, videoURL, *outputDir); err != nil {
					slog.Error("failed to download thumbnail", "video_url", videoURL, "error", err)
				}
			}()
		} else {
			if err := downloadThumbnail(ctx, cli, videoURL, *outputDir); err != nil {
				slog.Error("failed to download thumbnail", "video_url", videoURL, "error", err)
			}
		}
	}

	return nil
}

func downloadThumbnail(ctx context.Context, cli youthumbpb.ThumbnailServiceClient, videoURL, outputDir string) error {
	retry := 0
	for {
		if err := downloadThumbnailOnce(ctx, cli, videoURL, outputDir); err != nil {
			if errors.Is(err, errCanceled) {
				return err
			}
			if errors.Is(err, errThumbnailNotFound) {
				return err
			}
			if retry == maxRetry {
				return err
			}
			retry++
			slog.Warn("retrying thumbnail download", "video_url", videoURL, "retry", retry, "error", err)
			continue
		}
		return nil
	}
}

func downloadThumbnailOnce(ctx context.Context, cli youthumbpb.ThumbnailServiceClient, videoURL, outputDir string) error {
	slog.Info("downloading thumbnail", "video_url", videoURL)
	return nil
}
