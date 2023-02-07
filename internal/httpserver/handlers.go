package httpserver

import (
	"net/http"

	"github.com/ruskiiamov/gophermart/internal/bonus"
)

func withdrawals(bm bonus.Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("withdrawals..."))
	})
}
