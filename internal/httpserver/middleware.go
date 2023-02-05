package httpserver

import (
	"net/http"
	"strings"
)

const allowHeader = "Allow"

type middleware func(http.Handler) http.Handler

func withMiddleware(h http.Handler, ms ...middleware) http.Handler {
	for _, m := range ms {
		h = m(h)
	}
	return h
}

func post(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Add(allowHeader, http.MethodPost)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func get(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Add(allowHeader, http.MethodGet)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getOrPost(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			w.Header().Add(allowHeader, strings.Join([]string{http.MethodGet, http.MethodPost}, ", "))
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func contType(value string) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contTypeValue := r.Header.Get(contTypeHeader)
			if contTypeValue == "" || contTypeValue != value {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
