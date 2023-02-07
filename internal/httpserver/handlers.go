package httpserver

import (
	"net/http"

	"github.com/ruskiiamov/gophermart/internal/bonus"
)

func balance(bm bonus.Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("balance..."))
	})
}

func withdraw(bm bonus.Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("withdraw..."))
	})
}

func withdrawals(bm bonus.Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("withdrawals..."))
	})
}
