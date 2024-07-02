package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"mime"
	"os"
	"path/filepath"

	"github.com/kirillgashkov/assignment-youthumb/internal/thumbnail"
	"github.com/kirillgashkov/assignment-youthumb/proto/youthumbpb/v1"
)

type thumbnailDownloader struct {
	cli       youthumbpb.ThumbnailServiceClient
	outputDir string
	muCh      chan struct{}
}

func newThumbnailDownloader(cli youthumbpb.ThumbnailServiceClient, outputDir string) *thumbnailDownloader {
	return &thumbnailDownloader{cli: cli, outputDir: outputDir, muCh: make(chan struct{}, 1)}
}

func (d *thumbnailDownloader) DownloadThumbnail(ctx context.Context, videoURL string) error {
	// Create a temporary file to store the thumbnail content.

	contentFile, err := os.CreateTemp("", "thumbnail-*")
	if err != nil {
		return err
	}
	defer func(contentFile *os.File) {
		if err := contentFile.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
			slog.Error("failed to close file", "error", err)
		}
	}(contentFile)

	// Download the thumbnail content.

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

	// Determine the extension of the thumbnail file.

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

	// Move the temporary file to the output directory.

	select {
	case d.muCh <- struct{}{}:
		func() {
			defer func() {
				<-d.muCh
			}()

			videoID, err := thumbnail.ParseVideoID(videoURL)
			if err != nil {
				slog.Error("failed to parse video ID", "video_url", videoURL, "error", err)
				return
			}

			outputFilePath := filepath.Join(d.outputDir, videoID+extension)

			if err := os.MkdirAll(d.outputDir, 0755); err != nil {
				slog.Error("failed to create directory", "output_dir", d.outputDir, "error", err)
				return
			}

			if err := os.Rename(contentFile.Name(), outputFilePath); err != nil {
				slog.Error("failed to rename file", "error", err)
			}
		}()
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}
