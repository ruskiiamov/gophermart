package data

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ruskiiamov/gophermart/internal/bonus"
	"github.com/ruskiiamov/gophermart/internal/user"
)

var _ user.AuthDataContainer = (*Container)(nil)
var _ bonus.BonusDataContainer = (*Container)(nil)

type Container struct {
	dbpool *pgxpool.Pool
}

func NewContainer(ctx context.Context, dbURI string) (*Container, error) {
	err := doMigrations(dbURI)
	if err != nil {
		return nil, fmt.Errorf("do migrations error: %w", err)
	}

	dbpool, err := pgxpool.New(ctx, dbURI)
	if err != nil {
		return nil, fmt.Errorf("DB pool creation error: %w", err)
	}

	return &Container{dbpool: dbpool}, nil
}

func (c *Container) Close() {
	c.dbpool.Close()
}

func (c *Container) CreateUser(ctx context.Context, login, passHash string) error {
	//TODO
	return nil
}

func (c *Container) GetUser(ctx context.Context, login string) (*user.User, error) {
	//TODO
	return nil, user.ErrLoginNotFound
}
