package server

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type mockedUserAuthorizer struct {
	mock.Mock
}

func (m *mockedUserAuthorizer) Register(ctx context.Context, login, password string) error {
	args := m.Called(ctx, login, password)
	return args.Error(0)
}

func (m *mockedUserAuthorizer) Login(ctx context.Context, login, password string) (string, error) {
	args := m.Called(ctx, login, password)
	return args.String(0), args.Error(1)
}

func (m *mockedUserAuthorizer) AuthByToken(ctx context.Context, accessToken string) (userID string, err error) {
	args := m.Called(ctx, accessToken)
	return args.String(0), args.Error(1)
}
