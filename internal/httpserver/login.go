package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type loginReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func login(ua UserAuthorizer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var logReq loginReq

		content, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Msgf("body read error: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = json.Unmarshal(content, &logReq)
		if err != nil {
			log.Info().Msgf("json unmarshall error: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if logReq.Login == "" || logReq.Password == "" {
			log.Info().Msg("wrong json structure")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		accessToken, err := ua.Login(ctx, logReq.Login, logReq.Password)
		if errors.Is(err, ErrLoginPassword) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if err != nil {
			log.Error().Msgf("login error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add(authHeader, accessToken)
		w.WriteHeader(http.StatusOK)
	})  
}