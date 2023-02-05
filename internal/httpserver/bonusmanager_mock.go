package httpserver

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type mockedBonusManager struct {
	mock.Mock
}

func (m *mockedBonusManager) AddOrder(ctx context.Context, userID string, orderID int) error {
	args := m.Called(ctx, userID, orderID)
	return args.Error(0)
}
