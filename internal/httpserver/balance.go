package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/ruskiiamov/gophermart/internal/bonus"
)

type balanceRes struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func balance(bm bonus.Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(authUserIDContextKey).(string)
		if !ok || userID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		current, withdrawn, err := bm.GetBalance(ctx, userID)
		if err != nil {
			log.Error().Msgf("get balance error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		res := balanceRes{
			Current:   float64(current) / 100,
			Withdrawn: float64(withdrawn) / 100,
		}

		content, err := json.Marshal(res)

		w.Header().Add(contTypeHeader, appJSON)
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	})
}
