package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotFound(t *testing.T) {
	ua := new(mockedUserAuthorizer)
	bm := new(mockedBonusManager)
	h := createHandler(ua, bm)

	t.Run("not found", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		response := w.Result()

		assert.Equal(t, http.StatusNotFound, response.StatusCode)
	})
}

func TestBalance(t *testing.T) {
	target := "/api/user/balance"
	ua := new(mockedUserAuthorizer)
	bm := new(mockedBonusManager)
	h := createHandler(ua, bm)

	t.Run("method error", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, target, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		response := w.Result()

		assert.Equal(t, http.StatusMethodNotAllowed, response.StatusCode)
		assert.Equal(t, http.MethodGet, response.Header.Get("Allow"))
	})

	t.Run("ok", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, target, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		response := w.Result()

		assert.Equal(t, http.StatusOK, response.StatusCode)
	})
}

func TestBalanceWithdraw(t *testing.T) {
	target := "/api/user/balance/withdraw"
	ua := new(mockedUserAuthorizer)
	bm := new(mockedBonusManager)
	h := createHandler(ua, bm)

	t.Run("method error", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, target, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		response := w.Result()

		assert.Equal(t, http.StatusMethodNotAllowed, response.StatusCode)
		assert.Equal(t, http.MethodPost, response.Header.Get("Allow"))
	})

	t.Run("ok", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, target, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		response := w.Result()

		assert.Equal(t, http.StatusOK, response.StatusCode)
	})
}

func TestWithdrawals(t *testing.T) {
	target := "/api/user/withdrawals"
	ua := new(mockedUserAuthorizer)
	bm := new(mockedBonusManager)
	h := createHandler(ua, bm)

	t.Run("method error", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, target, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		response := w.Result()

		assert.Equal(t, http.StatusMethodNotAllowed, response.StatusCode)
		assert.Equal(t, http.MethodGet, response.Header.Get("Allow"))
	})

	t.Run("ok", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, target, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		response := w.Result()

		assert.Equal(t, http.StatusOK, response.StatusCode)
	})
}
