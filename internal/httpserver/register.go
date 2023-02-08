package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/ruskiiamov/gophermart/internal/user"
)

type registerReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func register(ua user.Authorizer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var regReq registerReq

		content, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Msgf("body read error: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = json.Unmarshal(content, &regReq)
		if err != nil {
			log.Info().Msgf("json unmarshall error: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if regReq.Login == "" || regReq.Password == "" {
			log.Info().Msg("wrong json structure")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		err = ua.Register(ctx, regReq.Login, regReq.Password)
		if errors.Is(err, user.ErrLoginExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		if err != nil {
			log.Error().Msgf("register error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		accessToken, err := ua.Login(ctx, regReq.Login, regReq.Password)
		if err != nil {
			log.Error().Msgf("login error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add(authHeader, accessToken)
		w.WriteHeader(http.StatusOK)
	})
}
