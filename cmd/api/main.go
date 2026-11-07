package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-kanna/config/source"

	"github.com/go-kanna/showcase/config"
	"github.com/go-kanna/showcase/internal/di"
	"github.com/go-kanna/showcase/internal/server"
)

func main() {
	config.DefineAPIFlags(flag.CommandLine)
	config.DefineDatabaseFlags(flag.CommandLine)
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})))
	if err := run(); err != nil {
		slog.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	src, err := source.Open(source.Sources{Flags: flag.CommandLine})
	if err != nil {
		return fmt.Errorf("open config sources: %w", err)
	}
	cfg, err := di.NewConfig(src)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	infra, cu, err := di.NewInfra(ctx, cfg)
	if err != nil {
		return fmt.Errorf("build infrastructure: %w", err)
	}
	defer func() { cu() }()

	srv := server.New(":"+strconv.Itoa(cfg.API.Port), infra)

	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "tcp", srv.Addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", srv.Addr, err)
	}
	slog.Info("server listening", "addr", ln.Addr().String())

	errCh := make(chan error, 1)
	go func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("serve: %w", err)
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}
