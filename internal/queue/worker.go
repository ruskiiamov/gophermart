package queue

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ruskiiamov/gophermart/internal/logger"
)

type ErrRetryAfter struct {
	Duration time.Duration
}

func (e *ErrRetryAfter) Error() string {
	return fmt.Sprintf("retry after %s", e.Duration.String())
}

type Worker struct {
	taskDispatcher *Dispatcher
}

func NewWorker(taskDispatcher *Dispatcher) *Worker {
	return &Worker{taskDispatcher: taskDispatcher}
}

func (w *Worker) Loop(ctx context.Context) {
	for {
		t := w.taskDispatcher.PopWait()

		select {
		case <-ctx.Done():
			return
		default:
		}

		err := t.Handle()

		var errRetryAfter *ErrRetryAfter
		if errors.As(err, &errRetryAfter) {
			if errRetryAfter.Duration > 0 {
				w.taskDispatcher.SetLockedUntil(time.Now().Add(errRetryAfter.Duration))
			}
			w.taskDispatcher.Push(t)
			continue
		}

		if err != nil {
			logger.Error(fmt.Sprintf("task failed: %s", err))
		}
	}
}
