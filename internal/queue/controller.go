package queue

import (
	"context"
	"fmt"
	"sync"

	"github.com/ruskiiamov/gophermart/internal/bonus"
)

type Task struct {
	orderID int
	tries   int
}

func NewTask(orderID int) *Task {
	return &Task{orderID: orderID}
}

type Controller interface {
	Push(t *Task)
	PopWait() *Task
}

type controller struct {
	arr  []*Task
	mu   sync.Mutex
	cond *sync.Cond
}

func NewController(ctx context.Context, bm bonus.Manager) (*controller, error) {
	c := &controller{}
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

func (c *controller) Push(t *Task) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.arr = append(c.arr, t)

	c.cond.Signal()
}

func (c *controller) PopWait() *Task {
	c.mu.Lock()
	defer c.mu.Unlock()

	for len(c.arr) == 0 {
		c.cond.Wait()
	}

	t := c.arr[0]
	c.arr = c.arr[1:]

	return t
}
