package user

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type mockedAuthDataContainer struct {
	mock.Mock
}

func (m *mockedAuthDataContainer) CreateUser(ctx context.Context, login, passHash string) error {
	args := m.Called(ctx, login, passHash)
	return args.Error(0)
}

func (m *mockedAuthDataContainer) GetUser(ctx context.Context, login string) (*User, error) {
	args := m.Called(ctx, login)
	return args.Get(0).(*User), args.Error(1)
}
