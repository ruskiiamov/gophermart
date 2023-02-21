package access

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type mockedUserProvider struct {
	mock.Mock
}

func (m *mockedUserProvider) CreateUser(ctx context.Context, login, passHash string) error {
	args := m.Called(ctx, login, passHash)
	return args.Error(0)
}

func (m *mockedUserProvider) GetUser(ctx context.Context, login string) (*User, error) {
	args := m.Called(ctx, login)
	return args.Get(0).(*User), args.Error(1)
}
