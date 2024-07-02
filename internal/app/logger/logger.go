package logger

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/kirillgashkov/assignment-youthumb/internal/app/config"
)

func New(cfg *config.Config) (*slog.Logger, error) {
	var log *slog.Logger

	switch cfg.Mode {
	case config.ModeDevelopment:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case config.ModeProduction:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		return nil, fmt.Errorf("invalid mode: %s", cfg.Mode)
	}

	return log, nil
}
