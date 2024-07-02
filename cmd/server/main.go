package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/kirillgashkov/assignment-youthumb/internal/app/logger"

	"github.com/kirillgashkov/assignment-youthumb/internal/app/config"

	"github.com/kirillgashkov/assignment-youthumb/internal/api"
	"github.com/kirillgashkov/assignment-youthumb/internal/cache"
)

var (
	cachePath = flag.String("cache", "", "Path to the cache SQLite database.")
)

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), `Usage: %s [OPTIONS]

Starts the server.

Options:
`, os.Args[0])
	flag.PrintDefaults()
}

func main() {
	if err := mainErr(); err != nil {
		panic(err)
	}
}

func mainErr() error {
	flag.Usage = usage
	flag.Parse()

	if *cachePath == "" {
		return fmt.Errorf("cache path is required, see -help")
	}

	cfg, err := config.New()
	if err != nil {
		return err
	}

	log, err := logger.NewLogger(cfg)
	if err != nil {
		return err
	}
	slog.SetDefault(log)

	cch, err := cache.Open(*cachePath)
	if err != nil {
		return err
	}
	defer func(cch *cache.Cache) {
		if err := cch.Close(); err != nil {
			slog.Error("failed to close cache", "error", err)
		}
	}(cch)

	srv := api.NewServer(cch, cfg)

	addr := &net.TCPAddr{IP: net.ParseIP(cfg.GRPC.Host), Port: cfg.GRPC.Port}
	lis, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}

	slog.Info("starting server", "addr", addr, "mode", cfg.Mode)
	err = srv.Serve(lis)
	return err
}
