package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/kirillgashkov/assignment-youthumb/internal/app/log"

	"github.com/kirillgashkov/assignment-youthumb/internal/app/config"
)

var (
	isAsync   = flag.Bool("async", false, "Download thumbnails asynchronously.")
	outputDir = flag.String("o", "", "Path to the output directory.")
)

func main() {
	flag.Usage = usage
	flag.Parse()

	if *outputDir == "" {
		s := "output directory is required, use -o flag"
		if _, err := fmt.Fprintln(flag.CommandLine.Output(), s); err != nil {
			panic(err)
		}
		os.Exit(2)
	}

	if err := mainErr(); err != nil {
		s := fmt.Sprintf("fatal error: %v", err)
		if _, err := fmt.Fprintln(flag.CommandLine.Output(), s); err != nil {
			panic(err)
		}
		os.Exit(1)
	}
	os.Exit(0)
}

func mainErr() error {
	// Prepare configuration and logging.

	cfg, err := config.New()
	if err != nil {
		return err
	}

	logger, err := log.NewLogger(cfg)
	if err != nil {
		return err
	}
	slog.SetDefault(logger)

	// Prepare the client.

	clientConn, err := newClient(cfg.GRPC)
	if err != nil {
		return err
	}

	cli, err := newThumbnailServiceClient(clientConn)
	if err != nil {
		return err
	}

	// Read video URLs.

	var videoURLs []string
	if flag.NArg() == 0 {
		videoURLs, err = readVideoURLs(os.Stdin)
		if err != nil {
			return err
		}
	} else {
		videoURLs, err = readVideoURLsFromFiles(flag.Args())
		if err != nil {
			return err
		}
	}

	// Download thumbnails and write them to the output directory.

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	downloader := newThumbnailDownloader(cli, *outputDir)

	if *isAsync {
		func() {
			wg := sync.WaitGroup{}
			defer wg.Wait()

			for _, videoURL := range videoURLs {
				wg.Add(1)
				go func() { // https://go.dev/blog/loopvar-preview
					defer wg.Done()
					if err := downloader.DownloadThumbnailForVideoURL(ctx, videoURL); err != nil {
						slog.Error("failed to download thumbnail", "video_url", videoURL, "error", err)
					}
				}()
			}
		}()
	} else {
		for _, videoURL := range videoURLs {
			if err := downloader.DownloadThumbnailForVideoURL(ctx, videoURL); err != nil {
				slog.Error("failed to download thumbnail", "video_url", videoURL, "error", err)
			}
		}
	}

	return nil
}

func usage() {
	u := fmt.Sprintf(`Usage: %s [OPTIONS] [FILE_WITH_VIDEO_URLS...]

A client for downloading thumbnails for YouTube videos with the specified URLs.

gRPC server address is specified via the environment variables APP_GRPC_HOST and
APP_GRPC_PORT.

Arguments:
  FILE_WITH_VIDEO_URLS
		Path to a file with new-line separated YouTube video URLs. If no files
		are provided, the URLs are read from the standard input.

Options:
`, os.Args[0])

	if _, err := fmt.Fprint(flag.CommandLine.Output(), u); err != nil {
		panic(err)
	}
	flag.PrintDefaults()
}
