package httpserver

import (
	"context"
	"errors"
)

var (
	ErrUserHasOrder = errors.New("user already has that order")
	ErrOrderExists  = errors.New("order exists")
)

type BonusManager interface {
	AddOrder(ctx context.Context, userID string, orderID int) error
}
