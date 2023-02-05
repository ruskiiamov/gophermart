package httpserver

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type mockedUserAuthorizer struct {
	mock.Mock
}

func (m *mockedUserAuthorizer) Register(ctx context.Context, login, password string) (string, error) {
	args := m.Called(ctx, login, password)
	return args.String(0), args.Error(1)
}

func (m *mockedUserAuthorizer) Login(ctx context.Context, login, password string) (string, error) {
	args := m.Called(ctx, login, password)
	return args.String(0), args.Error(1)
}