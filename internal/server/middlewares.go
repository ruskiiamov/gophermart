package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ruskiiamov/gophermart/internal/access"
	"github.com/ruskiiamov/gophermart/internal/logger"
)

func authMiddleware(accessManager *access.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			accessToken := r.Header.Get(authHeader)
			if accessToken == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), 100*time.Millisecond)
			defer cancel()

			userID, err := accessManager.AuthByToken(ctx, accessToken)
			if errors.Is(err, access.ErrTokenNotValid) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if err != nil {
				logger.Error(fmt.Sprintf("auth by token error: %s", err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			rCtx := context.WithValue(r.Context(), userIDKey, userID)
			req := r.Clone(rCtx)

			next.ServeHTTP(w, req)
		})
	}
}
