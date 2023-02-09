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
