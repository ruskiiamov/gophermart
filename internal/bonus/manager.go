package bonus

import (
	"context"
	"errors"
	"fmt"
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
	CreateOrder(ctx context.Context, userID string, orderID int) (*Order, error)
	GetOrder(ctx context.Context, orderID int) (*Order, error)
	GetOrders(ctx context.Context, userID string) ([]*Order, error)
}

type Manager interface {
	AddOrder(ctx context.Context, userID string, orderID int) error
	GetOrders(ctx context.Context, userID string) ([]*Order, error)
	GetBalance(ctx context.Context, userID string) (current, withdrawn float64, err error)
	Withdraw(ctx context.Context, userID string, order int, sum float64) error
	GetWithdrawals(ctx context.Context, userID string) ([]*Withdrawal, error)
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
	dc BonusDataContainer
}

func NewManager(dc BonusDataContainer) Manager {
	return &manager{dc: dc}
}

func (m *manager) AddOrder(ctx context.Context, userID string, orderID int) error {
	if !luhn.Valid(strconv.Itoa(orderID)) {
		return ErrLuhnAlgo
	}

	order, err := m.dc.CreateOrder(ctx, userID, orderID)
	if err == nil {
		//TODO add Task to Queue
		return nil
	}

	if errors.Is(err, ErrOrderExists) {
		order, err = m.dc.GetOrder(ctx, orderID)
	}
	if err != nil {
		return fmt.Errorf("create order error: %w", err)
	}

	if order.UserID == userID {
		return ErrUserHasOrder
	}

	return ErrOrderExists
}

func (m *manager) GetOrders(ctx context.Context, userID string) ([]*Order, error) {
	orders, err := m.dc.GetOrders(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get orders error: %w", err)
	}

	return orders, nil
}

func (m *manager) GetBalance(ctx context.Context, userID string) (current, withdrawn float64, err error) {
	//TODO
	return 500.5, 42, nil
}

func (m *manager) Withdraw(ctx context.Context, userID string, order int, sum float64) error {
	//TODO
	return nil
}

func (m *manager) GetWithdrawals(ctx context.Context, userID string) ([]*Withdrawal, error) {
	//TODO
	return []*Withdrawal{}, nil
}
