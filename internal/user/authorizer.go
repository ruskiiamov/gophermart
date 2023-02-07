package user

import (
	"context"
	"errors"
)

var (
	ErrLoginExists   = errors.New("user with that login exists")
	ErrLoginPassword = errors.New("wrong login-password pair")
	ErrTokenNotValid = errors.New("wrong access token")
)

type AuthDataContainer interface {
	//TODO
}

type Authorizer interface {
	Register(ctx context.Context, login, password string) (accessToken string, err error)
	Login(ctx context.Context, login, password string) (accessToken string, err error)
	AuthByToken(ctx context.Context, accessToken string) (userID string, err error)
}

type authorizer struct {
	dataContainer AuthDataContainer
}

func NewAuthorizer(dc AuthDataContainer) Authorizer {
	return &authorizer{dataContainer: dc}
}

func (a *authorizer) Register(ctx context.Context, login, password string) (string, error) {
	//TODO
	return "Bearer 1234abcd", nil
}

func (a *authorizer) Login(ctx context.Context, login, password string) (string, error) {
	//TODO
	return "Bearer 1234abcd", nil
}

func (a *authorizer) AuthByToken(ctx context.Context, accessToken string) (userID string, err error) {
	//TODO
	return "aaaa-bbbb-cccc-dddd", nil
}
