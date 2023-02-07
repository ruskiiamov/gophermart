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

func (m *mockedBonusManager) GetOrders(ctx context.Context, userID string) ([]bonus.Order, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]bonus.Order), args.Error(1)
}
