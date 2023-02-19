package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/ruskiiamov/gophermart/internal/user"
)

const (
	authHeader        = "Authorization"
	userIDKey  ctxKey = "auth_user_id"
)

type ctxKey string

func auth(ua *user.Authorizer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			accessToken := r.Header.Get(authHeader)
			if accessToken == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), 100*time.Millisecond)
			defer cancel()

			userID, err := ua.AuthByToken(ctx, accessToken)
			if errors.Is(err, user.ErrTokenNotValid) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if err != nil {
				log.Error().Msgf("auth by token error: %s", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			rCtx := context.WithValue(r.Context(), userIDKey, userID)
			req := r.Clone(rCtx)

			next.ServeHTTP(w, req)
		})
	}
}
