package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/ruskiiamov/gophermart/internal/access"
	"github.com/ruskiiamov/gophermart/internal/bonus"
	"github.com/ruskiiamov/gophermart/internal/logger"
	"github.com/ruskiiamov/gophermart/internal/task"
)

type request struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func registerHnadler(accessManager *access.Manager) http.HandlerFunc {
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

		err = accessManager.Register(ctx, req.Login, req.Password)
		if errors.Is(err, access.ErrLoginExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		if err != nil {
			logger.Error(fmt.Sprintf("register error: %s", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		accessToken, err := accessManager.Login(ctx, req.Login, req.Password)
		if err != nil {
			logger.Error(fmt.Sprintf("login error: %s", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add(authHeader, accessToken)
		w.WriteHeader(http.StatusOK)
	})
}

func loginHandler(accessManager *access.Manager) http.HandlerFunc {
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

		accessToken, err := accessManager.Login(ctx, req.Login, req.Password)
		if errors.Is(err, access.ErrLoginPassword) {
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

func postOrderHandler(bonusManager *bonus.Manager, taskDispatcher *task.Dispatcher) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(userIDKey).(string)
		if !ok || userID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		orderID, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error(fmt.Sprintf("body read error: %s", err))
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

		err = bonusManager.AddOrder(ctx, userID, intOrderID)
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
			logger.Error(fmt.Sprintf("add order error: %s", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		taskDispatcher.Push(task.NewTask(intOrderID))

		w.WriteHeader(http.StatusAccepted)
	})
}

type orderRes struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

func getOrdersHandler(bonusManager *bonus.Manager) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(userIDKey).(string)
		if !ok || userID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		orders, err := bonusManager.GetOrders(ctx, userID)
		if err != nil {
			logger.Error(fmt.Sprintf("get orders error: %s", err))
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
			logger.Error(fmt.Sprintf("json marshall error: %s", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add(contTypeHeader, appJSON)
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	})
}

type balanceRes struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func balanceHandler(bonusManager *bonus.Manager) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(userIDKey).(string)
		if !ok || userID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		current, withdrawn, err := bonusManager.GetBalance(ctx, userID)
		if err != nil {
			logger.Error(fmt.Sprintf("get balance error: %s", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		res := balanceRes{
			Current:   float64(current) / 100,
			Withdrawn: float64(withdrawn) / 100,
		}

		content, err := json.Marshal(res)
		if err != nil {
			logger.Error(fmt.Sprintf("json marshall error: %s", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add(contTypeHeader, appJSON)
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	})
}

type withdrawReq struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func withdrawHandler(bonusManager *bonus.Manager) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(userIDKey).(string)
		if !ok || userID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var withdraw withdrawReq

		content, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error(fmt.Sprintf("body read error: %s", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = json.Unmarshal(content, &withdraw)
		if err != nil {
			logger.Error(fmt.Sprintf("json unmarshall error: %s", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if withdraw.Order == "" || withdraw.Sum == 0 {
			logger.Error("wrong json structure")
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

		err = bonusManager.Withdraw(ctx, userID, intOrder, int(withdraw.Sum*100))
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
			logger.Error(fmt.Sprintf("withdraw error: %s", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}

type withdrawRes struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

func withdrawalsHandler(bonusManager *bonus.Manager) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(userIDKey).(string)
		if !ok || userID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		withdrawals, err := bonusManager.GetWithdrawals(ctx, userID)
		if err != nil {
			logger.Error(fmt.Sprintf("get orders error: %s", err))
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
			logger.Error(fmt.Sprintf("json marshall error: %s", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add(contTypeHeader, appJSON)
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	})
}
