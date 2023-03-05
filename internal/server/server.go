package server

import (
	"context"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ruskiiamov/gophermart/internal/access"
	"github.com/ruskiiamov/gophermart/internal/bonus"
	"github.com/ruskiiamov/gophermart/internal/queue"
)

const (
	contTypeHeader        = "Content-Type"
	appJSON               = "application/json"
	authHeader            = "Authorization"
	userIDKey      ctxKey = "auth_user_id"
)

type ctxKey string

func NewServer(
	ctx context.Context,
	address string,
	accessManager *access.Manager,
	bonusManager *bonus.Manager,
	taskDispatcher *queue.Dispatcher,
) *http.Server {
	r := chi.NewRouter()

	r.Use(middleware.Compress(5))

	r.Route("/api/user", func(r chi.Router) {
		r.With(middleware.AllowContentType(appJSON)).Post("/register", registerHnadler(accessManager))
		r.With(middleware.AllowContentType(appJSON)).Post("/login", loginHandler(accessManager))

		r.With(authMiddleware(accessManager)).Route("/", func(r chi.Router) {
			r.Route("/orders", func(r chi.Router) {
				r.Post("/", postOrderHandler(bonusManager, taskDispatcher))
				r.Get("/", getOrdersHandler(bonusManager))
			})

			r.Route("/balance", func(r chi.Router) {
				r.Get("/", balanceHandler(bonusManager))
				r.With(middleware.AllowContentType(appJSON)).Post("/withdraw", withdrawHandler(bonusManager))
			})
			r.Get("/withdrawals", withdrawalsHandler(bonusManager))
		})
	})

	return &http.Server{
		Addr:    address,
		Handler: r,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}
}
