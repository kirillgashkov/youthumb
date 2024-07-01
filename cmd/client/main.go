package main

import (
	"flag"
	"fmt"
	"github.com/kirillgashkov/assignment-youthumb/internal/config"
	"github.com/kirillgashkov/assignment-youthumb/internal/logger"
	"log/slog"
	"os"
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
	if err := mainErr(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func mainErr() error {
	flag.Usage = usage
	flag.Parse()

	cfg, err := config.New()
	if err != nil {
		return err
	}

	log, err := logger.New(cfg)
	if err != nil {
		return err
	}
	slog.SetDefault(log)

	return nil
}
