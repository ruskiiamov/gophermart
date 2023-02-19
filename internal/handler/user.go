package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ruskiiamov/gophermart/internal/logger"
	"github.com/ruskiiamov/gophermart/internal/user"
)

type request struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func register(auth *user.Authorizer) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req request

		content, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error(fmt.Sprintf("body read error: %s", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = json.Unmarshal(content, &req)
		if err != nil {
			logger.Error(fmt.Sprintf("json unmarshall error: %s", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if req.Login == "" || req.Password == "" {
			logger.Error("wrong json structure")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		err = auth.Register(ctx, req.Login, req.Password)
		if errors.Is(err, user.ErrLoginExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		if err != nil {
			logger.Error(fmt.Sprintf("register error: %s", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		accessToken, err := auth.Login(ctx, req.Login, req.Password)
		if err != nil {
			logger.Error(fmt.Sprintf("login error: %s", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add(authHeader, accessToken)
		w.WriteHeader(http.StatusOK)
	})
}

func login(auth *user.Authorizer) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req request

		content, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error(fmt.Sprintf("body read error: %s", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = json.Unmarshal(content, &req)
		if err != nil {
			logger.Error(fmt.Sprintf("json unmarshall error: %s", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if req.Login == "" || req.Password == "" {
			logger.Error("wrong json structure")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		accessToken, err := auth.Login(ctx, req.Login, req.Password)
		if errors.Is(err, user.ErrLoginPassword) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if err != nil {
			logger.Error(fmt.Sprintf("login error: %s", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add(authHeader, accessToken)
		w.WriteHeader(http.StatusOK)
	})
}
