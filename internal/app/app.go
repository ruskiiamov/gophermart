package app

import (
	"context"
	"runtime"
	"time"

	"github.com/ruskiiamov/gophermart/internal/access"
	"github.com/ruskiiamov/gophermart/internal/accrualsystem"
	"github.com/ruskiiamov/gophermart/internal/bonus"
	"github.com/ruskiiamov/gophermart/internal/config"
	"github.com/ruskiiamov/gophermart/internal/database"
	"github.com/ruskiiamov/gophermart/internal/queue"
	"github.com/ruskiiamov/gophermart/internal/server"
	"github.com/ruskiiamov/gophermart/internal/tasks"
	"golang.org/x/sync/errgroup"
)

func Run(ctx context.Context, g *errgroup.Group, cfg *config.Config) {
	dbConnection, err := database.NewConnection(ctx, cfg.DatabaseURI)
	if err != nil {
		panic(err)
	}

	accrualProvider := accrualsystem.NewConnector(cfg.AccrualSystemAddress)
	accessManager := access.NewManager(dbConnection, cfg.SignSecret)
	bonusManager := bonus.NewManager(dbConnection, accrualProvider)
	taskDispatcher := queue.NewDispatcher(ctx)
	server := server.NewServer(ctx, cfg.RunAddress, accessManager, bonusManager, taskDispatcher)

	err = dbConnection.Migrate()
	if err != nil {
		panic(err)
	}

	orders, err := bonusManager.GetNotFinalOrders(ctx)
	if err != nil {
		panic(err)
	}
	for _, order := range orders {
		taskDispatcher.Push(tasks.NewAccrualTask(bonusManager, order.ID))
	}

	workers := make([]*queue.Worker, 0, runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		workers = append(workers, queue.NewWorker(taskDispatcher))
	}
	for _, w := range workers {
		worker := w
		g.Go(func() error {
			worker.Loop(ctx)
			return nil
		})
	}

	g.Go(func() error {
		return server.ListenAndServe()
	})

	g.Go(func() error {
		<-ctx.Done()

		sdCtx, sdCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer sdCancel()

		server.Shutdown(sdCtx)
		taskDispatcher.Close()
		dbConnection.Close()

		return nil
	})
}
