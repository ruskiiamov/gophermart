package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/caarlos0/env/v7"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/ruskiiamov/gophermart/internal/accrualsystem"
	"github.com/ruskiiamov/gophermart/internal/bonus"
	"github.com/ruskiiamov/gophermart/internal/data"
	"github.com/ruskiiamov/gophermart/internal/httpserver"
	"github.com/ruskiiamov/gophermart/internal/queue"
	"github.com/ruskiiamov/gophermart/internal/user"
	"golang.org/x/sync/errgroup"
)

type config struct {
	RunAddress           string `env:"RUN_ADDRESS" envDefault:":8080"`
	DatabaseURI          string `env:"DATABASE_URI" envDefault:"postgres://root:root@localhost:54320/gophermart?sslmode=disable"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"http://localhost:8081"`
	SignSecret           string `env:"SIGN_SECRET" envDefault:"1sBKAOv8uCDrEJU7LDS9RFqRiSN7DN3s"`
}

var cfg config

func initConfig() {
	env.Parse(&cfg)

	flag.StringVar(&(cfg.RunAddress), "a", cfg.RunAddress, "Server address")
	flag.StringVar(&(cfg.DatabaseURI), "d", cfg.DatabaseURI, "DB URI")
	flag.StringVar(&(cfg.AccrualSystemAddress), "r", cfg.AccrualSystemAddress, "Accrual system address")
	flag.StringVar(&(cfg.SignSecret), "s", cfg.SignSecret, "Sign secret for JWT")
}

func initLogger() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "2006-01-02 15:04:05",
	})
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	initConfig()
	initLogger()

	dataContainer, err := data.NewContainer(ctx, cfg.DatabaseURI)
	if err != nil {
		panic(err)
	}

	accrualSystemConnector := accrualsystem.NewConnector(cfg.AccrualSystemAddress)
	userAuthorizer := user.NewAuthorizer(dataContainer, cfg.SignSecret)
	bonusManager := bonus.NewManager(dataContainer, accrualSystemConnector)

	queueController, err := queue.NewController(ctx, bonusManager)
	if err != nil {
		panic(err)
	}

	workers := make([]*queue.Worker, 0, runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		workers = append(workers, queue.NewWorker(queueController, bonusManager))
	}
	for _, w := range workers {
		go w.Loop(ctx)
	}

	server := httpserver.NewServer(ctx, cfg.RunAddress, userAuthorizer, bonusManager)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return server.ListenAndServe()
	})

	g.Go(func() error {
		<-gCtx.Done()

		sdCtx, sdCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer sdCancel()

		server.Shutdown(sdCtx)
		dataContainer.Close()

		return nil
	})

	log.Info().Msgf("started at %s", cfg.RunAddress)
	if err := g.Wait(); err != nil {
		log.Info().Msgf("stopped: %s", err)
	}

	os.Exit(0)
}
