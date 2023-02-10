package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/ruskiiamov/gophermart/internal/bonus"
)

type withdrawReq struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func withdraw(bm bonus.Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(authUserIDContextKey).(string)
		if !ok || userID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var withdraw withdrawReq

		content, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Msgf("body read error: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = json.Unmarshal(content, &withdraw)
		if err != nil {
			log.Info().Msgf("json unmarshall error: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if withdraw.Order == "" || withdraw.Sum == 0 {
			log.Info().Msg("wrong json structure")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		intOrder, err := strconv.Atoi(withdraw.Order)
		if err != nil || intOrder <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		err = bm.Withdraw(ctx, userID, intOrder, int(withdraw.Sum*100))
		log.Printf("error %s", err)
		if errors.Is(err, bonus.ErrWrongSum) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if errors.Is(err, bonus.ErrNotEnough) {
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}
		if errors.Is(err, bonus.ErrLuhnAlgo) || errors.Is(err, bonus.ErrOrderExists) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		if err != nil {
			log.Error().Msgf("withdraw error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}
