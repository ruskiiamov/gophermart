package task

import (
	"context"
	"errors"
	"fmt"

	"github.com/ruskiiamov/gophermart/internal/accrualsystem"
	"github.com/ruskiiamov/gophermart/internal/bonus"
	"github.com/ruskiiamov/gophermart/internal/logger"
)

const retries = 3

type Worker struct {
	taskDispatcher *Dispatcher
	bonusManager   *bonus.Manager
}

func NewWorker(taskDispatcher *Dispatcher, bonusManager *bonus.Manager) *Worker {
	return &Worker{
		taskDispatcher: taskDispatcher,
		bonusManager:   bonusManager,
	}
}

func (w *Worker) Loop(ctx context.Context) {
	for {
		t := w.taskDispatcher.PopWait()

		select {
		case <-ctx.Done():
			return
		default:
			err := w.bonusManager.SetOrderAccrual(ctx, t.orderID)
			if err == nil {
				continue
			}

			if errors.Is(err, bonus.ErrAccrualNotReady) {
				w.taskDispatcher.Push(t)
				continue
			}

			var errNotAvailable *accrualsystem.ErrNotAvailable
			if errors.As(err, &errNotAvailable) {
				w.taskDispatcher.SetLockedUntil(errNotAvailable.AvailableSince)
				w.taskDispatcher.Push(t)
				continue
			}

			logger.Error(fmt.Sprintf("set order accrual error: %s", err))
			t.tries = t.tries + 1
			if t.tries < retries {
				w.taskDispatcher.Push(t)
				continue
			}

			err = w.bonusManager.SetOrderInvalid(ctx, t.orderID)
			if err != nil {
				logger.Error(fmt.Sprintf("set order invalid error: %s", err))
			}
		}
	}
}
