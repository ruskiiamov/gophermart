package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/ruskiiamov/gophermart/internal/bonus"
)

type withdrawRes struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

func withdrawals(bm bonus.ManagerI) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(authUserIDContextKey).(string)
		if !ok || userID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		withdrawals, err := bm.GetWithdrawals(ctx, userID)
		if err != nil {
			log.Error().Msgf("get orders error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(withdrawals) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		resItems := make([]withdrawRes, 0, len(withdrawals))
		for _, withdrawal := range withdrawals {
			resItems = append(resItems, withdrawRes{
				Order:       strconv.Itoa(withdrawal.ID),
				Sum:         float64(withdrawal.Sum) / 100,
				ProcessedAt: withdrawal.CreatedAt.Format(time.RFC3339),
			})
		}

		content, err := json.Marshal(resItems)
		if err != nil {
			log.Error().Msgf("json marshall error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add(contTypeHeader, appJSON)
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	})
}
