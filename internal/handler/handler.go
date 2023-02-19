package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ruskiiamov/gophermart/internal/bonus"
	"github.com/ruskiiamov/gophermart/internal/task"
	"github.com/ruskiiamov/gophermart/internal/user"
)

const (
	contTypeHeader = "Content-Type"
	appJSON        = "application/json"
)

func NewHandler(ua *user.Authorizer, bm *bonus.Manager, td *task.Dispatcher) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Compress(5))

	r.Route("/api/user", func(r chi.Router) {
		r.With(middleware.AllowContentType(appJSON)).Post("/register", register(ua))
		r.With(middleware.AllowContentType(appJSON)).Post("/login", login(ua))

		r.With(auth(ua)).Route("/", func(r chi.Router) {
			r.Route("/orders", func(r chi.Router) {
				r.Post("/", postOrder(bm, td))
				r.Get("/", getOrders(bm))
			})

			r.Route("/balance", func(r chi.Router) {
				r.Get("/", balance(bm))
				r.With(middleware.AllowContentType(appJSON)).Post("/withdraw", withdraw(bm))
			})
			r.Get("/withdrawals", withdrawals(bm))
		})
	})

	return r
}
