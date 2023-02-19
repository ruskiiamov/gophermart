package handler

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

type balanceRes struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type withdrawReq struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type withdrawRes struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

func balance(bm *bonus.Manager) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(userIDKey).(string)
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

func withdraw(bm *bonus.Manager) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(userIDKey).(string)
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

func withdrawals(bm *bonus.Manager) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(userIDKey).(string)
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
