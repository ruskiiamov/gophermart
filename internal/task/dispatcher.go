package task

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ruskiiamov/gophermart/internal/bonus"
)

type Task struct {
	orderID int
	tries   int
}

func NewTask(orderID int) *Task {
	return &Task{orderID: orderID}
}

type Dispatcher struct {
	arr         []*Task
	mu          sync.Mutex
	cond        *sync.Cond
	ctx         context.Context
	lockedUntil time.Time
}

func NewDispatcher(ctx context.Context, bm *bonus.Manager) (*Dispatcher, error) {
	c := &Dispatcher{
		ctx:         ctx,
		lockedUntil: time.Now(),
	}
	c.cond = sync.NewCond(&c.mu)

	orders, err := bm.GetNotFinalOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("get not final orders error: %w", err)
	}

	for _, order := range orders {
		c.arr = append(c.arr, NewTask(order.ID))
	}

	return c, nil
}

func (c *Dispatcher) Push(t *Task) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.arr = append(c.arr, t)

	c.cond.Signal()
}

func (c *Dispatcher) PopWait() *Task {
	c.mu.Lock()
	defer c.mu.Unlock()

	for len(c.arr) == 0 || time.Now().Before(c.lockedUntil) {
		select {
		case <-c.ctx.Done():
			return nil
		default:
			c.cond.Wait()
		}
	}

	t := c.arr[0]
	c.arr = c.arr[1:]
	return t
}

func (c *Dispatcher) SetLockedUntil(lockedUntil time.Time) {
	if time.Now().After(lockedUntil) {
		return
	}

	c.lockedUntil = lockedUntil
	time.AfterFunc(time.Until(lockedUntil) , func() {
		c.cond.Broadcast()
	})
}

func (c *Dispatcher) Close() {
	c.cond.Broadcast()
}
