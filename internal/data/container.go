package data

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type container struct {
	dbpool *pgxpool.Pool
}

func NewContainer(ctx context.Context, dbURI string) (*container, error) {
	err := doMigrations(dbURI)
	if err != nil {
		return nil, fmt.Errorf("do migrations error: %w", err)
	}

	dbpool, err := pgxpool.New(ctx, dbURI)
	if err != nil {
		return nil, fmt.Errorf("DB pool creation error: %w", err)
	}

	return &container{dbpool: dbpool}, nil
}

func (c *container) Close() {
	c.dbpool.Close()
}
