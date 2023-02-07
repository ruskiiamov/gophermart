package httpserver

// import (
// 	"context"
// 	"errors"
// )

// var (
// 	ErrLoginExists   = errors.New("user with that login exists")
// 	ErrLoginPassword = errors.New("wrong login-password pair")
// 	ErrTokenNotValid = errors.New("wrong access token")
// )

// type UserAuthorizer interface {
// 	Register(ctx context.Context, login, password string) (accessToken string, err error)
// 	Login(ctx context.Context, login, password string) (accessToken string, err error)
// 	AuthByToken(ctx context.Context, accessToken string) (userID string, err error)
// }
