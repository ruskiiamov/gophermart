package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ruskiiamov/gophermart/internal/access"
	"github.com/ruskiiamov/gophermart/internal/accrualsystem"
	"github.com/ruskiiamov/gophermart/internal/bonus"
	"github.com/ruskiiamov/gophermart/internal/config"
	"github.com/ruskiiamov/gophermart/internal/database"
	"github.com/ruskiiamov/gophermart/internal/logger"
	"github.com/ruskiiamov/gophermart/internal/server"
	"github.com/ruskiiamov/gophermart/internal/task"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	g, gCtx := errgroup.WithContext(ctx)
	defer cancel()

	cfg := config.Load()

	dbConnection, err := database.NewConnection(gCtx, cfg.DatabaseURI)
	if err != nil {
		panic(err)
	}
	err = dbConnection.Migrate()
	if err != nil {
		panic(err)
	}

	accrualProvider := accrualsystem.NewConnector(cfg.AccrualSystemAddress)
	accessManager := access.NewManager(dbConnection, cfg.SignSecret)
	bonusManager := bonus.NewManager(dbConnection, accrualProvider)

	taskDispatcher, err := task.NewDispatcher(gCtx, bonusManager)
	if err != nil {
		panic(err)
	}

	workers := make([]*task.Worker, 0, runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		workers = append(workers, task.NewWorker(taskDispatcher, bonusManager))
	}
	for _, w := range workers {
		worker := w
		g.Go(func() error {
			worker.Loop(gCtx)
			return nil
		})
	}

	server := server.NewServer(gCtx, cfg.RunAddress, accessManager, bonusManager, taskDispatcher)

	g.Go(func() error {
		return server.ListenAndServe()
	})

	g.Go(func() error {
		<-gCtx.Done()

		sdCtx, sdCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer sdCancel()

		server.Shutdown(sdCtx)
		taskDispatcher.Close()
		dbConnection.Close()

		return nil
	})

	logger.Info(fmt.Sprintf("started at %s", cfg.RunAddress))
	if err := g.Wait(); err != nil {
		logger.Info(fmt.Sprintf("stopped: %s", err))
	}

	os.Exit(0)
}
