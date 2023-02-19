package server

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ruskiiamov/gophermart/internal/bonus"
	"github.com/ruskiiamov/gophermart/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const withdrawURL = "/api/user/balance/withdraw"

func TestWithdraw(t *testing.T) {
	ua := new(mockedUserAuthorizer)
	bm := new(mockedBonusManager)
	qc := new(mockedQueueController)
	handler := createHandler(ua, bm, qc)

	t.Run("405", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, withdrawURL, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		response := w.Result()
		defer response.Body.Close()

		assert.Equal(t, http.StatusMethodNotAllowed, response.StatusCode)
		assert.Equal(t, http.MethodPost, response.Header.Get("Allow"))
	})

	accessTokens := []string{"", "451hghjgi7r", "Bearer 1234abcd5678efgh"}
	for _, accessToken := range accessTokens {
		t.Run("401", func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, withdrawURL, nil)
			if accessToken != "" {
				r.Header.Add(authHeader, accessToken)
				ua.On("AuthByToken", mock.Anything, accessToken).Return("", user.ErrTokenNotValid).Once()
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			response := w.Result()
			defer response.Body.Close()

			resContent, err := io.ReadAll(response.Body)
			assert.NoError(t, err)

			assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
			assert.Empty(t, resContent)
		})
	}

	t.Run("400 empty Content-Type", func(t *testing.T) {
		body := strings.NewReader(`{"order": "2377225624", "sum": 751}`)
		r := httptest.NewRequest(http.MethodPost, withdrawURL, body)
		accessToken := "Bearer 1234abcd5678efgh"
		r.Header.Add(authHeader, accessToken)

		ua.On("AuthByToken", mock.Anything, accessToken).Return("aaa-bbb", nil).Once()

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		response := w.Result()
		defer response.Body.Close()

		resContent, err := io.ReadAll(response.Body)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		assert.Empty(t, response.Header.Get(authHeader))
		assert.Empty(t, resContent)
	})

	contents := []string{
		`{"order": "2377225624", "sum": 751,}`,
		`{"order_id": "2377225624", "sum": 751}`,
		`{"order": "23s7f7g2h25624", "sum": 751}`,
	}
	for _, content := range contents {
		t.Run("400 invalid json", func(t *testing.T) {
			body := strings.NewReader(content)
			r := httptest.NewRequest(http.MethodPost, withdrawURL, body)
			accessToken := "Bearer 1234abcd5678efgh"
			r.Header.Add(authHeader, accessToken)
			r.Header.Add(contTypeHeader, appJSON)

			ua.On("AuthByToken", mock.Anything, accessToken).Return("aaa-bbb", nil).Once()

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			response := w.Result()
			defer response.Body.Close()

			resContent, err := io.ReadAll(response.Body)
			assert.NoError(t, err)

			assert.Equal(t, http.StatusBadRequest, response.StatusCode)
			assert.Empty(t, response.Header.Get(authHeader))
			assert.Empty(t, string(resContent))
		})
	}

	errTests := []struct {
		err    error
		status int
	}{
		{nil, http.StatusOK},
		{bonus.ErrNotEnough, http.StatusPaymentRequired},
		{bonus.ErrLuhnAlgo, http.StatusUnprocessableEntity},
		{bonus.ErrOrderExists, http.StatusUnprocessableEntity},
		{bonus.ErrWrongSum, http.StatusBadRequest},
		{errors.New("test"), http.StatusInternalServerError},
	}
	for _, tt := range errTests {
		t.Run("200, 402, 422, 500", func(t *testing.T) {
			body := strings.NewReader(`{"order": "2377225624", "sum": 751}`)
			r := httptest.NewRequest(http.MethodPost, withdrawURL, body)
			accessToken := "Bearer 1234abcd5678efgh"
			r.Header.Add(authHeader, accessToken)
			r.Header.Add(contTypeHeader, appJSON)

			userID := "aaaa-bbbb-cccc-dddd"
			order := 2377225624
			sum := 75100
			ua.On("AuthByToken", mock.Anything, accessToken).Return(userID, nil).Once()
			bm.On("Withdraw", mock.Anything, userID, order, sum).Return(tt.err).Once()

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			response := w.Result()
			defer response.Body.Close()

			resContent, err := io.ReadAll(response.Body)
			assert.NoError(t, err)

			assert.Equal(t, tt.status, response.StatusCode)
			assert.Empty(t, resContent)
		})
	}
}
