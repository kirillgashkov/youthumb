package main

import (
	"github.com/kirillgashkov/assignment-youthumb/internal/config"
	"github.com/kirillgashkov/assignment-youthumb/internal/logger"
	"log/slog"
	"os"
)

func main() {
	if err := mainErr(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func mainErr() error {
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
