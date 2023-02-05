package httpserver

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/ferdypruis/go-luhn"
	"github.com/rs/zerolog/log"
)

func postOrders(bm BonusManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(authUserIDContextKey).(string)
		if !ok || userID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		orderID, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Msgf("body read error: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		intOrderID, err := strconv.Atoi(string(orderID))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !luhn.Valid(string(orderID)) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		err = bm.AddOrder(ctx, userID, intOrderID)
		if errors.Is(err, ErrUserHasOrder) {
			w.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, ErrOrderExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		if err != nil {
			log.Error().Msgf("add order error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	})
}
