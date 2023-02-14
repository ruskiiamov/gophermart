package httpserver

import (
	"github.com/ruskiiamov/gophermart/internal/queue"
	"github.com/stretchr/testify/mock"
)

type mockedQueueController struct {
	mock.Mock
}

func (m *mockedQueueController) Push(t *queue.Task) {
	m.Called(t)
}

func (m *mockedQueueController) PopWait() *queue.Task {
	args := m.Called()
	return args.Get(0).(*queue.Task)
}
