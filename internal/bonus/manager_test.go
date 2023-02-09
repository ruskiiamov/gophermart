package bonus

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAddOrder(t *testing.T) {
	dc := new(mockerBonusDataContainer)
	m := NewManager(dc)

	t.Run("luhn", func(t *testing.T) {
		err := m.AddOrder(context.Background(), "aaa-bbb-ccc", 79927398714)
		assert.ErrorIs(t, err, ErrLuhnAlgo)
	})

	tests := []struct {
		orderID int
		dErr    error
		order   *Order
		fErr    error
	}{
		{
			orderID: 79927398713,
			dErr:    ErrOrderExists,
			order: &Order{
				ID:        79927398713,
				UserID:    "aaaa-bbbb-cccc-dddd",
				Status:    "PROCESSED",
				Accrual:   100,
				CreatedAt: time.Date(2023, 1, 15, 20, 25, 45, 0, time.Local),
			},
			fErr: ErrUserHasOrder,
		},
		{
			orderID: 79927398713,
			dErr:    ErrOrderExists,
			order: &Order{
				ID:        79927398713,
				UserID:    "nnnnn-llll-eeee-ssss",
				Status:    "INVALID",
				CreatedAt: time.Date(2023, 2, 5, 22, 5, 15, 0, time.Local),
			},
			fErr: ErrOrderExists,
		},
	}
	for _, tt := range tests {
		t.Run("errors", func(t *testing.T) {
			var o *Order
			userID := "aaaa-bbbb-cccc-dddd"

			dc.On("CreateOrder", mock.Anything, userID, tt.orderID).Return(o, tt.dErr).Once()
			dc.On("GetOrder", mock.Anything, tt.orderID).Return(tt.order, nil).Once()

			err := m.AddOrder(context.Background(), userID, tt.orderID)

			dc.AssertExpectations(t)

			assert.ErrorIs(t, err, tt.fErr)
		})
	}

	t.Run("ok", func(t *testing.T) {
		orderID := 79927398713
		userID := "aaaa-bbbb-cccc-dddd"

		order := &Order{
			ID:        orderID,
			UserID:    userID,
			Status:    "NEW",
			CreatedAt: time.Now(),
		}

		dc.On("CreateOrder", mock.Anything, userID, orderID).Return(order, nil).Once()

		err := m.AddOrder(context.Background(), userID, orderID)

		dc.AssertExpectations(t)

		assert.NoError(t, err)
	})
}
