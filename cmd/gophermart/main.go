package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ruskiiamov/gophermart/internal/app"
	"github.com/ruskiiamov/gophermart/internal/config"
	"github.com/ruskiiamov/gophermart/internal/logger"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	g, gCtx := errgroup.WithContext(ctx)
	defer cancel()

	cfg := config.Load()

	app.Run(gCtx, g, cfg)

	logger.Info(fmt.Sprintf("started at %s", cfg.RunAddress))
	if err := g.Wait(); err != nil {
		logger.Info(fmt.Sprintf("stopped: %s", err))
	}

	os.Exit(0)
}
