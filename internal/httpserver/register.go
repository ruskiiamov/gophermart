package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"
)

type registerReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func register(ua UserAuthorizer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var regReq registerReq
		
		content, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("body read error: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = json.Unmarshal(content, &regReq)
		if err != nil {
			log.Printf("json unmarshall error: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if regReq.Login == "" || regReq.Password == "" {
			log.Println("wrong json structure")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		accessToken, err := ua.Register(ctx, regReq.Login, regReq.Password)
		if errors.Is(err, ErrLoginExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		if err != nil {
			log.Printf("register error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add(authHeader, accessToken)
		w.WriteHeader(http.StatusOK)
	})
}
