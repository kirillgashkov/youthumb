package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/kirillgashkov/assignment-youthumb/internal/app/config"
	"github.com/kirillgashkov/assignment-youthumb/internal/app/log"
	"github.com/kirillgashkov/assignment-youthumb/internal/rpc"
	"github.com/kirillgashkov/assignment-youthumb/internal/thumbnail"
)

var (
	dsn = flag.String("d", ":memory:", "Path to the SQLite database.")
)

func main() {
	flag.Usage = usage
	flag.Parse()

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

	// Prepare cache.

	c, err := thumbnail.OpenCache(*dsn)
	if err != nil {
		return err
	}
	defer func(c *thumbnail.Cache) {
		if err := c.Close(); err != nil {
			slog.Error("failed to close cache", "error", err)
		}
	}(c)

	// Create and start the server.

	srv := rpc.NewServer(c, cfg)

	addr := &net.TCPAddr{IP: net.ParseIP(cfg.GRPC.Host), Port: cfg.GRPC.Port}
	lis, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}

	slog.Info("starting server", "addr", addr, "mode", cfg.Mode)
	err = srv.Serve(lis)
	return err
}

func usage() {
	u := fmt.Sprintf(`Usage: %s [OPTIONS]

A server for proxying YouTube video thumbnails. It downloads thumbnails from
YouTube, caches them in a SQLite database and serves them via gRPC.

Options:
`, os.Args[0])

	if _, err := fmt.Fprint(flag.CommandLine.Output(), u); err != nil {
		panic(err)
	}
	flag.PrintDefaults()
}
