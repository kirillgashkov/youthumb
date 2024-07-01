package main

import (
	"github.com/kirillgashkov/assignment-youthumb/internal/api"
	"github.com/kirillgashkov/assignment-youthumb/internal/config"
	"github.com/kirillgashkov/assignment-youthumb/internal/logger"
	"log/slog"
	"net"
)

func main() {
	if err := mainErr(); err != nil {
		panic(err)
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

	srv := api.NewServer(cfg)

	addr := &net.TCPAddr{IP: net.ParseIP(cfg.GRPC.Host), Port: cfg.GRPC.Port}
	lis, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}

	slog.Info("starting server", "addr", addr, "mode", cfg.Mode)
	err = srv.Serve(lis)
	return err
}
