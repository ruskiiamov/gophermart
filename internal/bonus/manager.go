package bonus

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/ferdypruis/go-luhn"
)

var (
	ErrUserHasOrder = errors.New("user already has that order")
	ErrOrderExists  = errors.New("order exists")
	ErrLuhnAlgo     = errors.New("luhn check fail")
	ErrNotEnough    = errors.New("not enough balance")
)

type BonusDataContainer interface {
	//TODO
}

type Manager interface {
	AddOrder(ctx context.Context, userID string, orderID int) error
	GetOrders(ctx context.Context, userID string) ([]Order, error)
	GetBalance(ctx context.Context, userID string) (current, withdrawn float64, err error)
	Withdraw(ctx context.Context, userID string, order int, sum float64) error
	GetWithdrawals(ctx context.Context, userID string) ([]Withdrawal, error)
}

type Order struct {
	ID        int
	UserID    string
	Status    string
	Accrual   float64
	CreatedAt time.Time
}

type Withdrawal struct {
	ID        int
	UserID    string
	Sum       float64
	CreatedAt time.Time
}

type manager struct {
	dataContainer BonusDataContainer
}

func NewManager(dc BonusDataContainer) Manager {
	return &manager{dataContainer: dc}
}

func (m *manager) AddOrder(ctx context.Context, userID string, orderID int) error {
	//TODO

	if !luhn.Valid(strconv.Itoa(orderID)) {
		return ErrLuhnAlgo
	}

	return nil
}

func (m *manager) GetOrders(ctx context.Context, userID string) ([]Order, error) {
	//TODO
	return []Order{}, nil
}

func (m *manager) GetBalance(ctx context.Context, userID string) (current, withdrawn float64, err error) {
	//TODO
	return 500.5, 42, nil
}

func (m *manager) Withdraw(ctx context.Context, userID string, order int, sum float64) error {
	//TODO
	return nil
}

func (m *manager) GetWithdrawals(ctx context.Context, userID string) ([]Withdrawal, error) {
	//TODO
	return []Withdrawal{}, nil
}
