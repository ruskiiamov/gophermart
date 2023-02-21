package task

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"
	"github.com/ruskiiamov/gophermart/internal/bonus"
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
			if errors.Is(err, bonus.ErrAccrualNotReady) {
				w.taskDispatcher.Push(t)
				continue
			}
			if err != nil {
				log.Error().Msgf("set order accrual error: %s", err)
				t.tries = t.tries + 1
				if t.tries < retries {
					w.taskDispatcher.Push(t)
					continue
				}
				err = w.bonusManager.SetOrderInvalid(ctx, t.orderID)
				if err != nil {
					log.Error().Msgf("set order invalid error: %s", err)
				}
			}
		}
	}
}
