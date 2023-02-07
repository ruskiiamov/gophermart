package bonus

import (
	"context"
	"errors"
	"time"
)

var (
	ErrUserHasOrder = errors.New("user already has that order")
	ErrOrderExists  = errors.New("order exists")
)

type BonusDataContainer interface {
	//TODO
}

type Manager interface {
	AddOrder(ctx context.Context, userID string, orderID int) error
	GetOrders(ctx context.Context, userID string) ([]Order, error)
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
	return nil
}

func (m *manager) GetOrders(ctx context.Context, userID string) ([]Order, error) {
	//TODO
	return []Order{}, nil
}
