package httpserver

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/ruskiiamov/gophermart/internal/bonus"
	"github.com/ruskiiamov/gophermart/internal/user"
)

const (
	contTypeHeader = "Content-Type"
	authHeader     = "Authorization"
	appJSON        = "application/json"
)

func NewServer(ctx context.Context, address string, ua user.Authorizer, bm bonus.Manager) *http.Server {
	return &http.Server{
		Addr:    address,
		Handler: createHandler(ua, bm),
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}
}

func createHandler(ua user.Authorizer, bm bonus.Manager) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", notfound)

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
				log.Println("POST")
				withMiddleware(postOrders(bm), auth(ua)).ServeHTTP(w, r)
			case http.MethodGet:
				log.Println("GET")
				withMiddleware(getOrders(bm), auth(ua)).ServeHTTP(w, r)
			}
		}),
		getOrPost,
	))

	mux.Handle("/api/user/balance", withMiddleware(
		balance(bm),
		get,
	))

	mux.Handle("/api/user/balance/withdraw", withMiddleware(
		withdraw(bm),
		post,
	))

	mux.Handle("/api/user/withdrawals", withMiddleware(
		withdrawals(bm),
		get,
	))

	return mux
}

func notfound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
