package httpserver

import (
	"context"

	"github.com/ruskiiamov/gophermart/internal/bonus"
	"github.com/stretchr/testify/mock"
)

type mockedBonusManager struct {
	mock.Mock
}

func (m *mockedBonusManager) AddOrder(ctx context.Context, userID string, orderID int) error {
	args := m.Called(ctx, userID, orderID)
	return args.Error(0)
}

func (m *mockedBonusManager) GetOrders(ctx context.Context, userID string) ([]*bonus.Order, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*bonus.Order), args.Error(1)
}

func (m *mockedBonusManager) GetBalance(ctx context.Context, userID string) (current, withdrawn int, err error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Int(1), args.Error(2)
}

func (m *mockedBonusManager) Withdraw(ctx context.Context, userID string, order int, sum int) error {
	args := m.Called(ctx, userID, order, sum)
	return args.Error(0)
}

func (m *mockedBonusManager) GetWithdrawals(ctx context.Context, userID string) ([]*bonus.Withdrawal, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*bonus.Withdrawal), args.Error(1)
}
