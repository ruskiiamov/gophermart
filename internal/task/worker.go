package task

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"
	"github.com/ruskiiamov/gophermart/internal/bonus"
)

const retries = 3

type Worker struct {
	qc DispatcherI
	bm bonus.ManagerI
}

func NewWorker(qc DispatcherI, bm bonus.ManagerI) *Worker {
	return &Worker{
		qc: qc,
		bm: bm,
	}
}

func (w *Worker) Loop(ctx context.Context) {
	for {
		t := w.qc.PopWait()

		select {
		case <-ctx.Done():
			return
		default:
			err := w.bm.SetOrderAccrual(ctx, t.orderID)
			if errors.Is(err, bonus.ErrAccrualNotReady) {
				w.qc.Push(t)
				continue
			}
			if err != nil {
				log.Error().Msgf("set order accrual error: %s", err)
				t.tries = t.tries + 1
				if t.tries < retries {
					w.qc.Push(t)
					continue
				}
				err = w.bm.SetOrderInvalid(ctx, t.orderID)
				if err != nil {
					log.Error().Msgf("set order invalid error: %s", err)
				}
			}
		}
	}
}
