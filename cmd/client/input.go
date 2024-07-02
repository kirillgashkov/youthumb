package main

import (
	"bufio"
	"io"
	"log/slog"
	"os"
)

func readVideoURLs(r io.Reader) ([]string, error) {
	urls := make([]string, 0)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}

	return urls, nil
}

func readVideoURLsFromFiles(files []string) ([]string, error) {
	urls := make([]string, 0)

	for _, file := range files {
		err := func() error {
			f, err := os.Open(file)
			if err != nil {
				return err
			}
			defer func(f *os.File) {
				if err := f.Close(); err != nil {
					slog.Error("failed to close file", "error", err)
				}
			}(f)

			fileURLs, err := readVideoURLs(f)
			if err != nil {
				return err
			}

			urls = append(urls, fileURLs...)
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}

	return urls, nil
}
