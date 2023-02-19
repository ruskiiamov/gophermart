package httpserver

import (
	"github.com/ruskiiamov/gophermart/internal/task"
	"github.com/stretchr/testify/mock"
)

type mockedQueueController struct {
	mock.Mock
}

func (m *mockedQueueController) Push(t *task.Task) {
	m.Called(t)
}

func (m *mockedQueueController) PopWait() *task.Task {
	args := m.Called()
	return args.Get(0).(*task.Task)
}
