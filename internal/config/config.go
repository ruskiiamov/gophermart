package config

import (
	"flag"

	"github.com/caarlos0/env/v7"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS" envDefault:":8080"`
	DatabaseURI          string `env:"DATABASE_URI" envDefault:"postgres://root:root@localhost:54320/gophermart?sslmode=disable"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"http://localhost:8081"`
	SignSecret           string `env:"SIGN_SECRET" envDefault:"1sBKAOv8uCDrEJU7LDS9RFqRiSN7DN3s"`
}

func Load() *Config {
	cfg := new(Config)

	env.Parse(cfg)

	flag.StringVar(&(cfg.RunAddress), "a", cfg.RunAddress, "Server address")
	flag.StringVar(&(cfg.DatabaseURI), "d", cfg.DatabaseURI, "DB URI")
	flag.StringVar(&(cfg.AccrualSystemAddress), "r", cfg.AccrualSystemAddress, "Accrual system address")
	flag.StringVar(&(cfg.SignSecret), "s", cfg.SignSecret, "Sign secret for JWT")

	flag.Parse()

	return cfg
}
