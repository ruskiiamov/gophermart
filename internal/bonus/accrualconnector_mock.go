package bonus

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type mockedAccrualSystemConnector struct {
	mock.Mock
}

func (m *mockedAccrualSystemConnector) GetAccrual(ctx context.Context, orderID int) (status string, accrual int, err error) {
	args := m.Called(ctx, orderID)
	return args.String(0), args.Int(1), args.Error(2)
}
