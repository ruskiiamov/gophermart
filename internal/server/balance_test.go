package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ruskiiamov/gophermart/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const balanceURL = "/api/user/balance"

func TestBalance(t *testing.T) {
	ua := new(mockedUserAuthorizer)
	bm := new(mockedBonusManager)
	qc := new(mockedQueueController)
	handler := createHandler(ua, bm, qc)

	t.Run("405", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, balanceURL, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		response := w.Result()
		defer response.Body.Close()

		assert.Equal(t, http.StatusMethodNotAllowed, response.StatusCode)
		assert.Equal(t, http.MethodGet, response.Header.Get("Allow"))
	})

	accessTokens := []string{"", "451hghjgi7r", "Bearer 1234abcd5678efgh"}
	for _, accessToken := range accessTokens {
		t.Run("401", func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, balanceURL, nil)
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

	t.Run("200", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, balanceURL, nil)
		accessToken := "Bearer 1234abcd5678efgh"
		r.Header.Add(authHeader, accessToken)

		userID := "aaaa-bbbb-cccc-dddd"
		current := 50050
		withdrawn := 4200
		ua.On("AuthByToken", mock.Anything, accessToken).Return(userID, nil).Once()
		bm.On("GetBalance", mock.Anything, userID).Return(current, withdrawn, nil).Once()

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		response := w.Result()
		defer response.Body.Close()

		resContent, err := io.ReadAll(response.Body)
		assert.NoError(t, err)

		content := `{"current": 500.5,"withdrawn": 42}`

		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, appJSON, response.Header.Get(contTypeHeader))
		assert.JSONEq(t, content, string(resContent))
	})
}
