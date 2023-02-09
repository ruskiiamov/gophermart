package bonus

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type mockerBonusDataContainer struct {
	mock.Mock
}

func (m *mockerBonusDataContainer) CreateOrder(ctx context.Context, userID string, orderID int) (*Order, error) {
	args := m.Called(ctx, userID, orderID)
	return args.Get(0).(*Order), args.Error(1)
}

func (m *mockerBonusDataContainer) GetOrder(ctx context.Context, orderID int) (*Order, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).(*Order), args.Error(1)
}

func (m *mockerBonusDataContainer) GetOrders(ctx context.Context, userID string) ([]*Order, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*Order), args.Error(1)
}

func (m *mockerBonusDataContainer) GetBalance(ctx context.Context, userID string) (current, withdrawn int, err error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Int(1), args.Error(2)
}
