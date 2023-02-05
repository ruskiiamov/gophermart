package httpserver

import (
	"context"
	"errors"
)

var ErrLoginExists = errors.New("user with that login exists")

type UserAuthorizer interface {
	Register(ctx context.Context, login, password string) (string, error)
}
