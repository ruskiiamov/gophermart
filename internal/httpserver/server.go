package httpserver

import (
	"context"
	"net"
	"net/http"

	"github.com/ruskiiamov/gophermart/internal/bonus"
	"github.com/ruskiiamov/gophermart/internal/task"
	"github.com/ruskiiamov/gophermart/internal/user"
)

const (
	contTypeHeader = "Content-Type"
	authHeader     = "Authorization"
	appJSON        = "application/json"
)

func NewServer(ctx context.Context, address string, ua user.AuthorizerI, bm bonus.ManagerI, qc task.DispatcherI) *http.Server {
	return &http.Server{
		Addr:    address,
		Handler: createHandler(ua, bm, qc),
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}
}

func createHandler(ua user.AuthorizerI, bm bonus.ManagerI, qc task.DispatcherI) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	mux.Handle("/api/user/register", withMiddleware(
		register(ua),
		contType(appJSON),
		post,
	))

	mux.Handle("/api/user/login", withMiddleware(
		login(ua),
		contType(appJSON),
		post,
	))

	mux.Handle("/api/user/orders", withMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				withMiddleware(postOrders(bm, qc), auth(ua)).ServeHTTP(w, r)
			case http.MethodGet:
				withMiddleware(getOrders(bm), auth(ua)).ServeHTTP(w, r)
			}
		}),
		getOrPost,
	))

	mux.Handle("/api/user/balance", withMiddleware(
		balance(bm),
		auth(ua),
		get,
	))

	mux.Handle("/api/user/balance/withdraw", withMiddleware(
		withdraw(bm),
		contType(appJSON),
		auth(ua),
		post,
	))

	mux.Handle("/api/user/withdrawals", withMiddleware(
		withdrawals(bm),
		auth(ua),
		get,
	))

	return mux
}
