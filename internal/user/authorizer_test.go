package user

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister(t *testing.T) {
	dc := new(mockedAuthDataContainer)
	a := NewAuthorizer(dc, "secret")

	tests := []error{nil, ErrLoginExists, errors.New("test")}
	for _, tt := range tests {
		t.Run("reg", func(t *testing.T) {
			login := "test_login"
			password := "test_pass"

			dc.On("CreateUser", mock.Anything, login, mock.AnythingOfType("string")).Return(tt).Once()
			err := a.Register(context.Background(), login, password)
			dc.AssertExpectations(t)

			if tt != nil {
				assert.ErrorAs(t, err, &tt)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	dc := new(mockedAuthDataContainer)
	a := NewAuthorizer(dc, "secret")

	tests := []struct {
		user *User
		uErr error
		fErr error
	}{
		{
			user: nil,
			uErr: ErrLoginNotFound,
			fErr: ErrLoginPassword,
		},
		{
			user: &User{
				ID:       "aaa-bbb-ccc",
				Login:    "test_login",
				PassHash: "$2a$04$Ku2EONfIuZDBmMYPyShiU.GVZmGSD3rRxXErhp9igDgx9hu/XGdEq",
			},
			uErr: nil,
			fErr: nil,
		},
		{
			user: &User{
				ID:       "aaa-bbb-ccc",
				Login:    "test_login",
				PassHash: "$2a04$u2EONfIuZDBmMYPyShiU.GVZmSDrRxXErhp9igDgx9u/XGdEq",
			},
			uErr: nil,
			fErr: ErrLoginPassword,
		},
		{
			user: nil,
			uErr: errors.New("test"),
			fErr: errors.New("test"),
		},
	}
	for _, tt := range tests {
		t.Run("login", func(t *testing.T) {
			login := "test_login"
			password := "test_pass"

			dc.On("GetUser", mock.Anything, login).Return(tt.user, tt.uErr).Once()
			accessToken, err := a.Login(context.Background(), login, password)
			dc.AssertExpectations(t)

			if tt.fErr != nil {
				assert.Error(t, err)
				assert.Empty(t, accessToken)
				assert.ErrorAs(t, err, &tt.fErr)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, accessToken)
				assert.Regexp(t, "^Bearer\\s[\\w-.]+\\w$", accessToken)
			}
		})
	}
}

func TestAuthByToken(t *testing.T) {
	dc := new(mockedAuthDataContainer)
	a := NewAuthorizer(dc, "secret")

	tests := []struct {
		token string
		err   error
	}{
		{
			token: "",
			err:   ErrTokenNotValid,
		},
		{
			token: "34tt45g8735hg9737hg97h97h",
			err:   ErrTokenNotValid,
		},
		{
			token: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjcmVhdGVkX2F0IjoiMTYwNzk3OTYwMCIsInVzZXJfaWQiOiJhYWEtYmJiLWNjYyJ9.Uxf2h_Feu1K74RhjS7xY7K6baZ5RMhwIWMioObxIvjw",
			err:   ErrTokenNotValid,
		},
		{
			token: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjcmVhdGVkX2F0IjoiMjAxODIwNjgwMCIsInVzZXJfaWQiOiJhYWEtYmJiLWNjYyJ9.6eoX-mJghrm2lSDqiiuwMY5rJnksEM4ZgxRdAydYlRY",
			err:   nil,
		},
	}
	for _, tt := range tests {
		t.Run("auth", func(t *testing.T) {
			userID, err := a.AuthByToken(context.Background(), tt.token)
			if tt.err != nil {
				assert.Empty(t, userID)
				assert.ErrorIs(t, err, tt.err)
			} else {
				assert.NotEmpty(t, userID)
				assert.NoError(t, err)
			}
		})
	}
}
