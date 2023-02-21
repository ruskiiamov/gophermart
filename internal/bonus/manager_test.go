package bonus

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAddOrder(t *testing.T) {
	bonusProvider := new(mockedBonusProvider)
	accrualProvider := new(mockedAccrualProvider)
	bonusManager := NewManager(bonusProvider, accrualProvider)

	t.Run("luhn", func(t *testing.T) {
		err := bonusManager.AddOrder(context.Background(), "aaa-bbb-ccc", 79927398714)
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

			bonusProvider.On("CreateOrder", mock.Anything, userID, tt.orderID).Return(o, tt.dErr).Once()
			bonusProvider.On("GetOrder", mock.Anything, tt.orderID).Return(tt.order, nil).Once()

			err := bonusManager.AddOrder(context.Background(), userID, tt.orderID)

			bonusProvider.AssertExpectations(t)

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

		bonusProvider.On("CreateOrder", mock.Anything, userID, orderID).Return(order, nil).Once()

		err := bonusManager.AddOrder(context.Background(), userID, orderID)

		bonusProvider.AssertExpectations(t)

		assert.NoError(t, err)
	})
}

func TestGetOrders(t *testing.T) {
	bonusProvider := new(mockedBonusProvider)
	accrualProvider := new(mockedAccrualProvider)
	bonusManager := NewManager(bonusProvider, accrualProvider)

	tests := []struct {
		orders []*Order
	}{
		{
			orders: []*Order{},
		},
		{
			orders: []*Order{
				{
					ID:      9278923470,
					UserID:  "aaaa-bbbb-cccc-dddd",
					Status:  "PROCESSED",
					Accrual: 500,
					CreatedAt: func() time.Time {
						t, _ := time.Parse(time.RFC3339, "2020-12-10T15:15:45+03:00")
						return t
					}(),
				},
				{
					ID:      12345678903,
					UserID:  "aaaa-bbbb-cccc-dddd",
					Status:  "PROCESSING",
					Accrual: 0,
					CreatedAt: func() time.Time {
						t, _ := time.Parse(time.RFC3339, "2020-12-10T15:12:01+03:00")
						return t
					}(),
				},
				{
					ID:      346436439,
					UserID:  "aaaa-bbbb-cccc-dddd",
					Status:  "INVALID",
					Accrual: 0,
					CreatedAt: func() time.Time {
						t, _ := time.Parse(time.RFC3339, "2020-12-09T16:09:53+03:00")
						return t
					}(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run("name", func(t *testing.T) {
			userID := "aaaa-bbbb-cccc-dddd"

			bonusProvider.On("GetOrders", mock.Anything, userID).Return(tt.orders, nil).Once()

			orders, err := bonusManager.GetOrders(context.Background(), userID)

			bonusProvider.AssertExpectations(t)

			assert.NoError(t, err)
			assert.ElementsMatch(t, orders, tt.orders)
		})
	}
}

func TestGetBalance(t *testing.T) {
	bonusProvider := new(mockedBonusProvider)
	accrualProvider := new(mockedAccrualProvider)
	bonusManager := NewManager(bonusProvider, accrualProvider)

	t.Run("balance", func(t *testing.T) {
		userID := "aaaa-bbbb-cccc-dddd"
		current := 50050
		withdrawn := 4200

		bonusProvider.On("GetBalance", mock.Anything, userID).Return(current, withdrawn, nil)

		c, w, err := bonusManager.GetBalance(context.Background(), userID)

		bonusProvider.AssertExpectations(t)

		assert.NoError(t, err)
		assert.Equal(t, current, c)
		assert.Equal(t, withdrawn, w)
	})
}

func TestWithdraw(t *testing.T) {
	bonusProvider := new(mockedBonusProvider)
	accrualProvider := new(mockedAccrualProvider)
	bonusManager := NewManager(bonusProvider, accrualProvider)

	t.Run("luhn", func(t *testing.T) {
		err := bonusManager.Withdraw(context.Background(), "aaa-bbb-ccc", 79927398714, 5000)
		assert.ErrorIs(t, err, ErrLuhnAlgo)
	})

	t.Run("negative sum", func(t *testing.T) {
		err := bonusManager.Withdraw(context.Background(), "aaa-bbb-ccc", 79927398713, -5000)
		assert.ErrorIs(t, err, ErrWrongSum)
	})

	t.Run("order exists", func(t *testing.T) {
		orderID := 79927398713
		order := &Order{
			ID:        orderID,
			UserID:    "ddd-eee-fff",
			Status:    "NEW",
			CreatedAt: time.Now(),
		}

		bonusProvider.On("GetOrder", mock.Anything, orderID).Return(order, nil).Once()
		err := bonusManager.Withdraw(context.Background(), "aaa-bbb-ccc", orderID, 5000)

		bonusProvider.AssertExpectations(t)
		assert.ErrorIs(t, err, ErrOrderExists)
	})

	t.Run("not enough", func(t *testing.T) {
		var ord *Order
		orderID := 79927398713
		userID := "aaaa-bbbb-cccc-dddd"

		bonusProvider.On("GetOrder", mock.Anything, orderID).Return(ord, ErrOrderNotFound).Once()
		bonusProvider.On("GetBalance", mock.Anything, userID).Return(50050, 4200, nil).Once()

		err := bonusManager.Withdraw(context.Background(), userID, orderID, 70000)

		bonusProvider.AssertExpectations(t)
		assert.ErrorIs(t, err, ErrNotEnough)
	})

	t.Run("withdraw exists", func(t *testing.T) {
		var ord *Order
		orderID := 79927398713
		userID := "aaaa-bbbb-cccc-dddd"
		sum := 40000

		bonusProvider.On("GetOrder", mock.Anything, orderID).Return(ord, ErrOrderNotFound).Once()
		bonusProvider.On("GetBalance", mock.Anything, userID).Return(50050, 4200, nil).Once()
		bonusProvider.On("CreateWithdraw", mock.Anything, userID, orderID, sum).Return(ErrOrderExists).Once()

		err := bonusManager.Withdraw(context.Background(), userID, orderID, sum)

		bonusProvider.AssertExpectations(t)
		assert.ErrorIs(t, err, ErrOrderExists)
	})

	t.Run("ok", func(t *testing.T) {
		var ord *Order
		orderID := 79927398713
		userID := "aaaa-bbbb-cccc-dddd"
		sum := 40000

		bonusProvider.On("GetOrder", mock.Anything, orderID).Return(ord, ErrOrderNotFound).Once()
		bonusProvider.On("GetBalance", mock.Anything, userID).Return(50050, 4200, nil).Once()
		bonusProvider.On("CreateWithdraw", mock.Anything, userID, orderID, sum).Return(nil).Once()

		err := bonusManager.Withdraw(context.Background(), userID, orderID, sum)

		bonusProvider.AssertExpectations(t)
		assert.NoError(t, err)
	})
}

func TestGetWithdrawals(t *testing.T) {
	bonusProvider := new(mockedBonusProvider)
	accrualProvider := new(mockedAccrualProvider)
	bonusManager := NewManager(bonusProvider, accrualProvider)

	tests := []struct {
		withdrawals []*Withdrawal
	}{
		{
			withdrawals: []*Withdrawal{},
		},
		{
			withdrawals: []*Withdrawal{
				{
					ID:        2377225624,
					UserID:    "aaaa-bbbb-cccc-dddd",
					Sum:       15200,
					CreatedAt: time.Now(),
				},
				{
					ID:        2377225615,
					UserID:    "aaaa-bbbb-cccc-dddd",
					Sum:       2800,
					CreatedAt: time.Now(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			userID := "aaaa-bbbb-cccc-dddd"

			bonusProvider.On("GetWithdrawals", mock.Anything, userID).Return(tt.withdrawals, nil).Once()
			withdrawals, err := bonusManager.GetWithdrawals(context.Background(), userID)

			bonusProvider.AssertExpectations(t)
			assert.NoError(t, err)
			assert.ElementsMatch(t, withdrawals, tt.withdrawals)
		})
	}
}
