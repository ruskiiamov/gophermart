package access

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

const (
	bearer = "Bearer "
	ttl    = 30 * time.Minute
)

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

type UserProvider interface {
	CreateUser(ctx context.Context, login, passHash string) error
	GetUser(ctx context.Context, login string) (*User, error)
}

type Manager struct {
	userProvider UserProvider
	secret       string
}

func NewManager(up UserProvider, secret string) *Manager {
	return &Manager{
		userProvider: up,
		secret:       secret,
	}
}

func (a *Manager) Register(ctx context.Context, login, password string) error {
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return fmt.Errorf("password hashing error: %w", err)
	}

	err = a.userProvider.CreateUser(ctx, login, string(passHash))
	if errors.Is(err, ErrLoginExists) {
		return err
	}
	if err != nil {
		return fmt.Errorf("create user error: %w", err)
	}

	return nil
}

func (a *Manager) Login(ctx context.Context, login, password string) (string, error) {
	user, err := a.userProvider.GetUser(ctx, login)
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

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    user.ID,
		"created_at": strconv.FormatInt(time.Now().Unix(), 10),
	})

	accessToken, err := token.SignedString([]byte(a.secret))
	if err != nil {
		return "", fmt.Errorf("JWT signing error: %w", err)
	}

	return bearer + accessToken, nil
}

func (a *Manager) AuthByToken(ctx context.Context, accessToken string) (userID string, err error) {
	if len(accessToken) <= len(bearer) {
		return "", ErrTokenNotValid
	}

	if accessToken[:len(bearer)] != bearer {
		return "", ErrTokenNotValid
	}

	token, _ := jwt.Parse(accessToken[len(bearer):], func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(a.secret), nil
	})

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", ErrTokenNotValid
	}

	createdAt, ok := claims["created_at"].(string)
	if !ok {
		return "", ErrTokenNotValid
	}

	intCreatedAt, err := strconv.ParseInt(createdAt, 10, 64)
	if err != nil {
		return "", ErrTokenNotValid
	}

	if time.Unix(intCreatedAt, 0).Add(ttl).Before(time.Now()) {
		return "", ErrTokenNotValid
	}

	userID, ok = claims["user_id"].(string)
	if !ok {
		return "", ErrTokenNotValid
	}

	return userID, nil
}
