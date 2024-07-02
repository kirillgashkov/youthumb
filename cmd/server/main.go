package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/kirillgashkov/assignment-youthumb/internal/thumbnail"

	"github.com/kirillgashkov/assignment-youthumb/internal/app/log"

	"github.com/kirillgashkov/assignment-youthumb/internal/app/config"

	"github.com/kirillgashkov/assignment-youthumb/internal/rpc"
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

	logger, err := log.NewLogger(cfg)
	if err != nil {
		return err
	}
	slog.SetDefault(logger)

	cch, err := thumbnail.OpenCache(*cachePath)
	if err != nil {
		return err
	}
	defer func(cch *thumbnail.Cache) {
		if err := cch.Close(); err != nil {
			slog.Error("failed to close cache", "error", err)
		}
	}(cch)

	srv := rpc.NewServer(cch, cfg)

	addr := &net.TCPAddr{IP: net.ParseIP(cfg.GRPC.Host), Port: cfg.GRPC.Port}
	lis, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}

	slog.Info("starting server", "addr", addr, "mode", cfg.Mode)
	err = srv.Serve(lis)
	return err
}
