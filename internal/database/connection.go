package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"github.com/ruskiiamov/gophermart/internal/access"
	"github.com/ruskiiamov/gophermart/internal/bonus"
)

const source = "file://internal/database/migrations"

type Connection struct {
	dbpool *pgxpool.Pool
}

func NewConnection(ctx context.Context, dbURI string) (*Connection, error) {
	dbpool, err := pgxpool.New(ctx, dbURI)
	if err != nil {
		return nil, fmt.Errorf("DB pool creation error: %w", err)
	}

	return &Connection{dbpool: dbpool}, nil
}

func (c *Connection) Migrate() error {
	dbURI := c.dbpool.Config().ConnString()
	m, err := migrate.New(source, dbURI)
	if err != nil {
		return fmt.Errorf("migrate instance creation error: %w", err)
	}

	err = m.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		log.Info().Msg("migrations no changed")
		return nil
	}

	if err != nil {
		return fmt.Errorf("applying migrations error: %w", err)
	}

	log.Info().Msg("migrations done")
	return nil
}

func (c *Connection) Close() {
	c.dbpool.Close()
}

func (c *Connection) CreateUser(ctx context.Context, login, passHash string) error {
	var id string

	err := c.dbpool.QueryRow(
		ctx,
		`INSERT INTO users (login, pass_hash) VALUES ($1, $2) 
		ON CONFLICT (login) DO NOTHING RETURNING id;`,
		login,
		passHash,
	).Scan(&id)

	if errors.Is(err, pgx.ErrNoRows) {
		return access.ErrLoginExists
	}
	if err != nil {
		return fmt.Errorf("insert user error: %w", err)
	}

	return nil
}

func (c *Connection) GetUser(ctx context.Context, login string) (*access.User, error) {
	u := &access.User{Login: login}

	err := c.dbpool.QueryRow(
		ctx,
		`SELECT id, pass_hash FROM users WHERE login = $1;`,
		login,
	).Scan(&(u.ID), &(u.PassHash))

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, access.ErrLoginNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select user error: %w", err)
	}

	return u, nil
}

func (c *Connection) CreateOrder(ctx context.Context, userID string, orderID int) (*bonus.Order, error) {
	var createdAt time.Time

	err := c.dbpool.QueryRow(
		ctx,
		`INSERT INTO orders (id, user_id, status) values ($1, $2, 'NEW') 
		ON CONFLICT (id) DO NOTHING RETURNING created_at;`,
		orderID,
		userID,
	).Scan(&createdAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, bonus.ErrOrderExists
	}
	if err != nil {
		return nil, fmt.Errorf("inser order error: %w", err)
	}

	ord := &bonus.Order{
		ID:        orderID,
		UserID:    userID,
		CreatedAt: createdAt,
	}

	return ord, nil
}

func (c *Connection) UpdateOrder(ctx context.Context, orderID, accrual int, status string) error {
	row, err := c.dbpool.Query(
		ctx,
		`UPDATE orders SET accrual = $1, status = $2 WHERE id = $3;`,
		accrual,
		status,
		orderID,
	)
	row.Close()
	if err != nil {
		return fmt.Errorf("update order error: %w", err)
	}

	return nil
}

func (c *Connection) GetOrder(ctx context.Context, orderID int) (*bonus.Order, error) {
	ord := &bonus.Order{ID: orderID}

	err := c.dbpool.QueryRow(
		ctx,
		`SELECT user_id, status, accrual, created_at FROM orders WHERE id = $1;`,
		orderID,
	).Scan(&(ord.UserID), &(ord.Status), &(ord.Accrual), &(ord.CreatedAt))

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, bonus.ErrOrderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select order error: %w", err)
	}

	return ord, nil
}

func (c *Connection) GetOrders(ctx context.Context, userID string) ([]*bonus.Order, error) {
	orders := make([]*bonus.Order, 0)

	rows, err := c.dbpool.Query(
		ctx,
		`SELECT id, status, accrual, created_at FROM orders WHERE user_id = $1;`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("select orders error: %w", err)
	}

	for rows.Next() {
		ord := &bonus.Order{UserID: userID}
		err = rows.Scan(&(ord.ID), &(ord.Status), &(ord.Accrual), &(ord.CreatedAt))
		if err != nil {
			return nil, fmt.Errorf("row scan error: %w", err)
		}
		orders = append(orders, ord)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("db error: %w", err)
	}

	return orders, nil
}

func (c *Connection) GetNotFinalOrders(ctx context.Context) ([]*bonus.Order, error) {
	orders := make([]*bonus.Order, 0)

	rows, err := c.dbpool.Query(
		ctx,
		`SELECT id, user_id, status, created_at FROM orders WHERE status IN ('NEW', 'PROCESSING');`,
	)
	if err != nil {
		return nil, fmt.Errorf("select orders error: %w", err)
	}

	for rows.Next() {
		ord := &bonus.Order{}
		err = rows.Scan(&(ord.ID), &(ord.UserID), &(ord.Status), &(ord.CreatedAt))
		if err != nil {
			return nil, fmt.Errorf("row scan error: %w", err)
		}
		orders = append(orders, ord)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("db error: %w", err)
	}

	return orders, nil
}

func (c *Connection) GetBalance(ctx context.Context, userID string) (current, withdrawn int, err error) {
	tx, err := c.dbpool.Begin(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("transaction begin error: %w", err)
	}
	defer tx.Rollback(ctx)

	var accrualSum int

	err = tx.QueryRow(
		ctx,
		`SELECT COALESCE(sum(accrual), 0) as tmp FROM orders WHERE user_id = $1;`,
		userID,
	).Scan(&accrualSum)
	if err != nil {
		return 0, 0, fmt.Errorf("select accrual sum error: %w", err)
	}

	err = tx.QueryRow(
		ctx,
		`SELECT COALESCE(sum(sum), 0) as tmp FROM withdrawals WHERE user_id = $1;`,
		userID,
	).Scan(&withdrawn)
	if err != nil {
		return 0, 0, fmt.Errorf("select withdrawals sum error: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("transaction commit error: %w", err)
	}

	current = accrualSum - withdrawn

	return current, withdrawn, nil
}

func (c *Connection) CreateWithdraw(ctx context.Context, userID string, orderID, sum int) error {
	var createdAt time.Time

	err := c.dbpool.QueryRow(
		ctx,
		`INSERT INTO withdrawals (id, user_id, sum) values ($1, $2, $3) 
		ON CONFLICT (id) DO NOTHING RETURNING created_at`,
		orderID,
		userID,
		sum,
	).Scan(&createdAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return bonus.ErrOrderExists
	}
	if err != nil {
		return fmt.Errorf("insert withdrawal error: %w", err)
	}

	return nil
}

func (c *Connection) GetWithdrawals(ctx context.Context, userID string) ([]*bonus.Withdrawal, error) {
	withdrawals := make([]*bonus.Withdrawal, 0)

	rows, err := c.dbpool.Query(
		ctx,
		`SELECT id, sum, created_at FROM withdrawals WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("")
	}

	for rows.Next() {
		wdl := &bonus.Withdrawal{UserID: userID}
		err = rows.Scan(&(wdl.ID), &(wdl.Sum), &(wdl.CreatedAt))
		if err != nil {
			return nil, fmt.Errorf("select withdrawals error: %w", err)
		}
		withdrawals = append(withdrawals, wdl)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("db error: %w", err)
	}

	return withdrawals, nil
}
