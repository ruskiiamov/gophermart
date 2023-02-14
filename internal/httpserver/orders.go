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
	"github.com/ruskiiamov/gophermart/internal/queue"
)

type orderRes struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

func postOrders(bm bonus.Manager, qc queue.Controller) http.Handler {
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
		if err != nil || intOrderID <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		err = bm.AddOrder(ctx, userID, intOrderID)
		if errors.Is(err, bonus.ErrUserHasOrder) {
			w.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, bonus.ErrOrderExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		if errors.Is(err, bonus.ErrLuhnAlgo) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		if err != nil {
			log.Error().Msgf("add order error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		qc.Push(queue.NewTask(intOrderID))

		w.WriteHeader(http.StatusAccepted)
	})
}

func getOrders(bm bonus.Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(authUserIDContextKey).(string)
		if !ok || userID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		orders, err := bm.GetOrders(ctx, userID)
		if err != nil {
			log.Error().Msgf("get orders error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(orders) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		resItems := make([]orderRes, 0, len(orders))
		for _, order := range orders {
			resItems = append(resItems, orderRes{
				Number:     strconv.Itoa(order.ID),
				Status:     order.Status,
				Accrual:    float64(order.Accrual) / 100,
				UploadedAt: order.CreatedAt.Format(time.RFC3339),
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
