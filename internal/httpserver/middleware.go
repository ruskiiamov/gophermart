package httpserver

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/ruskiiamov/gophermart/internal/user"
)

const (
	allowHeader                 = "Allow"
	authUserIDContextKey ctxKey = "auth_user_id"
)

type ctxKey string

type middleware func(http.Handler) http.Handler

func withMiddleware(h http.Handler, ms ...middleware) http.Handler {
	for _, m := range ms {
		h = m(h)
	}
	return h
}

func post(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Add(allowHeader, http.MethodPost)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func get(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Add(allowHeader, http.MethodGet)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getOrPost(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			w.Header().Add(allowHeader, strings.Join([]string{http.MethodGet, http.MethodPost}, ", "))
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func contType(value string) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contTypeValue := r.Header.Get(contTypeHeader)
			if contTypeValue == "" || contTypeValue != value {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func auth(ua user.AuthorizerI) middleware {
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

			rCtx := context.WithValue(r.Context(), authUserIDContextKey, userID)
			req := r.Clone(rCtx)

			next.ServeHTTP(w, req)
		})
	}
}
