package httpserver

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ruskiiamov/gophermart/internal/bonus"
	"github.com/ruskiiamov/gophermart/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const withdrawalsURL = "/api/user/withdrawals"

func TestWithdrawals(t *testing.T) {
	ua := new(mockedUserAuthorizer)
	bm := new(mockedBonusManager)
	qc := new(mockedQueueController)
	handler := createHandler(ua, bm, qc)

	t.Run("405", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, withdrawalsURL, nil)
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
			r := httptest.NewRequest(http.MethodGet, withdrawalsURL, nil)
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

	tests := []struct {
		withdrawals []*bonus.Withdrawal
		status      int
		content     string
	}{
		{
			withdrawals: []*bonus.Withdrawal{},
			status:      http.StatusNoContent,
			content:     "",
		},
		{
			withdrawals: []*bonus.Withdrawal{
				{
					ID:     2377225624,
					UserID: "aaaa-bbbb-cccc-dddd",
					Sum:    50000,
					CreatedAt: func() time.Time {
						t, _ := time.Parse(time.RFC3339, "2020-12-09T16:09:57+03:00")
						return t
					}(),
				},
			},
			status: http.StatusOK,
			content: `[
				{
					"order": "2377225624",
					"sum": 500,
					"processed_at": "2020-12-09T16:09:57+03:00"
				}
			]`,
		},
	}
	for _, tt := range tests {
		t.Run("200", func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, withdrawalsURL, nil)
			accessToken := "Bearer 1234abcd5678efgh"
			r.Header.Add(authHeader, accessToken)

			userID := "aaaa-bbbb-cccc-dddd"
			ua.On("AuthByToken", mock.Anything, accessToken).Return(userID, nil).Once()
			bm.On("GetWithdrawals", mock.Anything, userID).Return(tt.withdrawals, nil).Once()

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			response := w.Result()
			defer response.Body.Close()

			resContent, err := io.ReadAll(response.Body)
			assert.NoError(t, err)

			assert.Equal(t, tt.status, response.StatusCode)
			if tt.content != "" {
				assert.Equal(t, appJSON, response.Header.Get(contTypeHeader))
				assert.JSONEq(t, tt.content, string(resContent))
			} else {
				assert.Empty(t, resContent)
			}
		})
	}
}
