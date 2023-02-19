package user

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const bearer = "Bearer "

var (
	ErrLoginExists   = errors.New("user with that login exists")
	ErrLoginNotFound = errors.New("login not found")
	ErrLoginPassword = errors.New("wrong login-password pair")
	ErrTokenNotValid = errors.New("wrong access token")
)

type User struct {
	ID       string
	Login    string
	PassHash string
}

type AuthDataContainer interface {
	CreateUser(ctx context.Context, login, passHash string) error
	GetUser(ctx context.Context, login string) (*User, error)
}

//TODO delete
type AuthorizerI interface {
	Register(ctx context.Context, login, password string) error
	Login(ctx context.Context, login, password string) (accessToken string, err error)
	AuthByToken(ctx context.Context, accessToken string) (userID string, err error)
}

type Authorizer struct {
	dc     AuthDataContainer
	secret string
}

func NewAuthorizer(dc AuthDataContainer, secret string) AuthorizerI {
	return &Authorizer{
		dc:     dc,
		secret: secret,
	}
}

func (a *Authorizer) Register(ctx context.Context, login, password string) error {
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return fmt.Errorf("password hashing error: %w", err)
	}

	err = a.dc.CreateUser(ctx, login, string(passHash))
	if errors.Is(err, ErrLoginExists) {
		return err
	}
	if err != nil {
		return fmt.Errorf("create user error: %w", err)
	}

	return nil
}

func (a *Authorizer) Login(ctx context.Context, login, password string) (string, error) {
	user, err := a.dc.GetUser(ctx, login)
	if errors.Is(err, ErrLoginNotFound) {
		return "", ErrLoginPassword
	}
	if err != nil {
		return "", fmt.Errorf("get user error: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PassHash), []byte(password))
	if err != nil {
		return "", ErrLoginPassword
	}

	accessToken, err := createJWT(a.secret, user.ID)
	if err != nil {
		return "", fmt.Errorf("create JWT error: %w", err)
	}

	return bearer + accessToken, nil
}

func (a *Authorizer) AuthByToken(ctx context.Context, accessToken string) (userID string, err error) {
	if len(accessToken) <= len(bearer) {
		return "", ErrTokenNotValid
	}

	if accessToken[:len(bearer)] != bearer {
		return "", ErrTokenNotValid
	}

	userID, err = getUserIDFromJWT(a.secret, accessToken[len(bearer):])

	if errors.Is(err, ErrTokenNotValid) {
		return "", err
	}
	if err != nil {
		return "", fmt.Errorf("get user from JWT error: %w", err)
	}

	return userID, nil
}
