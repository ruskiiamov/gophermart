package queue

import (
	"context"
	"sync"
	"time"
)

type Task interface {
	Handle() error
}

type Dispatcher struct {
	arr         []Task
	mu          sync.Mutex
	cond        *sync.Cond
	ctx         context.Context
	lockedUntil time.Time
}

func NewDispatcher(ctx context.Context) *Dispatcher {
	c := &Dispatcher{
		ctx:         ctx,
		lockedUntil: time.Now(),
	}
	c.cond = sync.NewCond(&c.mu)

	return c
}

func (c *Dispatcher) Push(t Task) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.arr = append(c.arr, t)

	c.cond.Signal()
}

func (c *Dispatcher) PopWait() Task {
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
	time.AfterFunc(time.Until(lockedUntil), func() {
		c.cond.Broadcast()
	})
}

func (c *Dispatcher) Close() {
	c.cond.Broadcast()
}
