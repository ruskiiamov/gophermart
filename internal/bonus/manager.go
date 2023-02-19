package bonus

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/ferdypruis/go-luhn"
)

const (
	registered = "REGISTERED"
	invalid    = "INVALID"
	processing = "PROCESSING"
	processed  = "PROCESSED"
)

var (
	ErrUserHasOrder    = errors.New("user already has that order")
	ErrOrderExists     = errors.New("order exists")
	ErrLuhnAlgo        = errors.New("luhn check fail")
	ErrNotEnough       = errors.New("not enough balance")
	ErrOrderNotFound   = errors.New("order not found")
	ErrWrongSum        = errors.New("wrong sum")
	ErrAccrualNotReady = errors.New("accrual not ready")
)

type BonusDataContainer interface {
	CreateOrder(ctx context.Context, userID string, orderID int) (*Order, error)
	UpdateOrder(ctx context.Context, orderID, accrual int, status string) error
	GetOrder(ctx context.Context, orderID int) (*Order, error)
	GetOrders(ctx context.Context, userID string) ([]*Order, error)
	GetNotFinalOrders(ctx context.Context) ([]*Order, error)
	GetBalance(ctx context.Context, userID string) (current, withdrawn int, err error)
	CreateWithdraw(ctx context.Context, userID string, orderID, sum int) error
	GetWithdrawals(ctx context.Context, userID string) ([]*Withdrawal, error)
}

type AccrualProvider interface {
	GetAccrual(ctx context.Context, orderID int) (status string, accrual int, err error)
}

//TODO delete
type ManagerI interface {
	AddOrder(ctx context.Context, userID string, orderID int) error
	GetOrders(ctx context.Context, userID string) ([]*Order, error)
	GetNotFinalOrders(ctx context.Context) ([]*Order, error)
	SetOrderAccrual(ctx context.Context, orderID int) error
	SetOrderInvalid(ctx context.Context, orderID int) error
	GetBalance(ctx context.Context, userID string) (current, withdrawn int, err error)
	Withdraw(ctx context.Context, userID string, order int, sum int) error
	GetWithdrawals(ctx context.Context, userID string) ([]*Withdrawal, error)
}

type Order struct {
	ID        int
	UserID    string
	Status    string
	Accrual   int
	CreatedAt time.Time
}

type Withdrawal struct {
	ID        int
	UserID    string
	Sum       int
	CreatedAt time.Time
}

type Manager struct {
	dc              BonusDataContainer
	accrualProvider AccrualProvider
}

func NewManager(dc BonusDataContainer, accrualProvider AccrualProvider) ManagerI {
	return &Manager{
		dc:              dc,
		accrualProvider: accrualProvider,
	}
}

func (m *Manager) AddOrder(ctx context.Context, userID string, orderID int) error {
	if !luhn.Valid(strconv.Itoa(orderID)) {
		return ErrLuhnAlgo
	}

	order, err := m.dc.CreateOrder(ctx, userID, orderID)
	if err == nil {
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

func (m *Manager) GetOrders(ctx context.Context, userID string) ([]*Order, error) {
	orders, err := m.dc.GetOrders(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get orders error: %w", err)
	}

	return orders, nil
}

func (m *Manager) GetNotFinalOrders(ctx context.Context) ([]*Order, error) {
	orders, err := m.dc.GetNotFinalOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("get not final orders error: %w", err)
	}

	return orders, nil
}

func (m *Manager) SetOrderAccrual(ctx context.Context, orderID int) error {
	status, accrual, err := m.accrualProvider.GetAccrual(ctx, orderID)
	if errors.Is(err, ErrAccrualNotReady) {
		return err
	}
	if err != nil {
		return fmt.Errorf("get accrual error: %w", err)
	}

	if status == registered {
		status = processing
	}

	err = m.dc.UpdateOrder(ctx, orderID, accrual, status)
	if err != nil {
		return fmt.Errorf("update order error: %w", err)
	}

	if status != invalid && status != processed {
		return ErrAccrualNotReady
	}

	return nil
}

func (m *Manager) SetOrderInvalid(ctx context.Context, orderID int) error {
	err := m.dc.UpdateOrder(ctx, orderID, 0, invalid)
	if err != nil {
		return fmt.Errorf("update order error: %w", err)
	}

	return nil
}

func (m *Manager) GetBalance(ctx context.Context, userID string) (current, withdrawn int, err error) {
	current, withdrawn, err = m.dc.GetBalance(ctx, userID)
	if err != nil {
		return 0, 0, fmt.Errorf("get balance error: %w", err)
	}

	return current, withdrawn, nil
}

func (m *Manager) Withdraw(ctx context.Context, userID string, order int, sum int) error {
	if !luhn.Valid(strconv.Itoa(order)) {
		return ErrLuhnAlgo
	}

	if sum <= 0 {
		return ErrWrongSum
	}

	_, err := m.dc.GetOrder(ctx, order)
	if err == nil {
		return ErrOrderExists
	}

	current, _, err := m.dc.GetBalance(ctx, userID)
	if err != nil {
		return fmt.Errorf("get balance error: %w", err)
	}

	if current < sum {
		return ErrNotEnough
	}

	err = m.dc.CreateWithdraw(context.Background(), userID, order, sum)
	if errors.Is(err, ErrOrderExists) {
		return err
	}
	if err != nil {
		return fmt.Errorf("create withdraw error: %w", err)
	}

	return nil
}

func (m *Manager) GetWithdrawals(ctx context.Context, userID string) ([]*Withdrawal, error) {
	withdrawals, err := m.dc.GetWithdrawals(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get withdrawals error: %w", err)
	}

	return withdrawals, nil
}
