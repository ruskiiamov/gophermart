package tasks

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ruskiiamov/gophermart/internal/accrualsystem"
	"github.com/ruskiiamov/gophermart/internal/bonus"
	"github.com/ruskiiamov/gophermart/internal/logger"
	"github.com/ruskiiamov/gophermart/internal/queue"
)

const retries = 3

type AccrualTask struct {
	bonusManager *bonus.Manager
	orderID      int
	tries        int
}

func NewAccrualTask(bonusManager *bonus.Manager, orderID int) *AccrualTask {
	return &AccrualTask{
		bonusManager: bonusManager,
		orderID:      orderID,
	}
}

func (a *AccrualTask) Handle() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := a.bonusManager.SetOrderAccrual(ctx, a.orderID)
	if err == nil {
		return nil
	}

	if errors.Is(err, bonus.ErrAccrualNotReady) {
		return &queue.ErrRetryAfter{}
	}

	var errNotAvailable *accrualsystem.ErrNotAvailable
	if errors.As(err, &errNotAvailable) {
		return &queue.ErrRetryAfter{Duration: time.Until(errNotAvailable.AvailableSince)}
	}

	logger.Error(fmt.Sprintf("set order accrual error: %s", err))
	a.tries = a.tries + 1
	if a.tries < retries {
		return &queue.ErrRetryAfter{}
	}

	err = a.bonusManager.SetOrderInvalid(ctx, a.orderID)
	if err != nil {
		logger.Error(fmt.Sprintf("set order invalid error: %s", err))
	}

	return nil
}
