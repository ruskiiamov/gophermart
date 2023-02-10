package bonus

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type mockedBonusDataContainer struct {
	mock.Mock
}

func (m *mockedBonusDataContainer) CreateOrder(ctx context.Context, userID string, orderID int) (*Order, error) {
	args := m.Called(ctx, userID, orderID)
	return args.Get(0).(*Order), args.Error(1)
}

func (m *mockedBonusDataContainer) GetOrder(ctx context.Context, orderID int) (*Order, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).(*Order), args.Error(1)
}

func (m *mockedBonusDataContainer) GetOrders(ctx context.Context, userID string) ([]*Order, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*Order), args.Error(1)
}

func (m *mockedBonusDataContainer) GetBalance(ctx context.Context, userID string) (current, withdrawn int, err error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Int(1), args.Error(2)
}

func (m *mockedBonusDataContainer) CreateWithdraw(ctx context.Context, userID string, orderID, sum int) error {
	args := m.Called(ctx, userID, orderID, sum)
	return args.Error(0)
}

func (m *mockedBonusDataContainer) GetWithdrawals(ctx context.Context, userID string) ([]*Withdrawal, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*Withdrawal), args.Error(1)
}
