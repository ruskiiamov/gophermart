package httpserver

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ruskiiamov/gophermart/internal/bonus"
	"github.com/ruskiiamov/gophermart/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const ordersURL = "/api/user/orders"

func TestOrders(t *testing.T) {
	ua := new(mockedUserAuthorizer)
	bm := new(mockedBonusManager)
	handler := createHandler(ua, bm)

	t.Run("405", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPut, ordersURL, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		response := w.Result()

		assert.Equal(t, http.StatusMethodNotAllowed, response.StatusCode)
		assert.ElementsMatch(
			t,
			[]string{http.MethodPost, http.MethodGet},
			strings.Split(response.Header.Get("Allow"), ", "),
		)
	})
}

func TestPostOrders(t *testing.T) {
	ua := new(mockedUserAuthorizer)
	bm := new(mockedBonusManager)
	handler := createHandler(ua, bm)

	accessTokens := []string{"", "451hghjgi7r", "Bearer 1234abcd5678efgh"}
	for _, accessToken := range accessTokens {
		t.Run("401", func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, ordersURL, nil)
			if accessToken != "" {
				r.Header.Add(authHeader, accessToken)
				ua.On("AuthByToken", mock.Anything, accessToken).Return("", user.ErrTokenNotValid).Once()
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			response := w.Result()

			resContent, err := io.ReadAll(response.Body)
			assert.NoError(t, err)

			assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
			assert.Empty(t, resContent)
		})
	}

	orderIDtests := []struct {
		orderID string
		status  int
	}{
		{"", http.StatusBadRequest},
		{"1a2s3d4f5g6h7j8k", http.StatusBadRequest},
	}
	for _, tt := range orderIDtests {
		t.Run("400", func(t *testing.T) {
			body := strings.NewReader(tt.orderID)
			r := httptest.NewRequest(http.MethodPost, ordersURL, body)
			accessToken := "Bearer 1234abcd5678efgh"
			r.Header.Add(authHeader, accessToken)

			ua.On("AuthByToken", mock.Anything, accessToken).Return("aaaa-bbbb-cccc-dddd", nil).Once()

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			response := w.Result()

			resContent, err := io.ReadAll(response.Body)
			assert.NoError(t, err)

			assert.Equal(t, tt.status, response.StatusCode)
			assert.Empty(t, resContent)
		})
	}

	t.Run("422", func(t *testing.T) {
		body := strings.NewReader("79927398714")
		r := httptest.NewRequest(http.MethodPost, ordersURL, body)
		accessToken := "Bearer 1234abcd5678efgh"
		r.Header.Add(authHeader, accessToken)

		userID := "aaaa-bbbb-cccc-dddd"
		ua.On("AuthByToken", mock.Anything, accessToken).Return(userID, nil).Once()
		bm.On("AddOrder", mock.Anything, userID, 79927398714).Return(bonus.ErrLuhnAlgo).Once()

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		response := w.Result()

		resContent, err := io.ReadAll(response.Body)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusUnprocessableEntity, response.StatusCode)
		assert.Empty(t, resContent)
	})

	errorTests := []struct {
		err    error
		status int
	}{
		{bonus.ErrUserHasOrder, http.StatusOK},
		{bonus.ErrOrderExists, http.StatusConflict},
		{nil, http.StatusAccepted},
		{errors.New("test"), http.StatusInternalServerError},
	}
	for _, tt := range errorTests {
		t.Run("200, 202, 409, 500", func(t *testing.T) {
			orderID := "79927398713"
			intOrderID := 79927398713
			body := strings.NewReader(orderID)
			r := httptest.NewRequest(http.MethodPost, ordersURL, body)
			accessToken := "Bearer 1234abcd5678efgh"
			r.Header.Add(authHeader, accessToken)

			userID := "aaaa-bbbb-cccc-dddd"
			ua.On("AuthByToken", mock.Anything, accessToken).Return(userID, nil).Once()
			bm.On("AddOrder", mock.Anything, userID, intOrderID).Return(tt.err).Once()

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			response := w.Result()

			resContent, err := io.ReadAll(response.Body)
			assert.NoError(t, err)

			assert.Equal(t, tt.status, response.StatusCode)
			assert.Empty(t, resContent)
		})
	}
}

func TestGetOrders(t *testing.T) {
	ua := new(mockedUserAuthorizer)
	bm := new(mockedBonusManager)
	handler := createHandler(ua, bm)

	accessTokens := []string{"", "451hghjgi7r", "Bearer 1234abcd5678efgh"}
	for _, accessToken := range accessTokens {
		t.Run("401", func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, ordersURL, nil)
			if accessToken != "" {
				r.Header.Add(authHeader, accessToken)
				ua.On("AuthByToken", mock.Anything, accessToken).Return("", user.ErrTokenNotValid).Once()
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			response := w.Result()

			resContent, err := io.ReadAll(response.Body)
			assert.NoError(t, err)

			assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
			assert.Empty(t, resContent)
		})
	}

	getTests := []struct {
		orders  []bonus.Order
		status  int
		content string
	}{
		{
			orders:  []bonus.Order{},
			status:  http.StatusNoContent,
			content: "",
		},
		{
			orders: []bonus.Order{
				{
					ID:      9278923470,
					UserID:  "aaaa-bbbb-cccc-dddd",
					Status:  "PROCESSED",
					Accrual: 500,
					CreatedAt: func() time.Time {
						t, _ := time.Parse(time.RFC3339, "2020-12-10T15:15:45+03:00")
						return t
					}(),
				},
				{
					ID:      12345678903,
					UserID:  "aaaa-bbbb-cccc-dddd",
					Status:  "PROCESSING",
					Accrual: 0,
					CreatedAt: func() time.Time {
						t, _ := time.Parse(time.RFC3339, "2020-12-10T15:12:01+03:00")
						return t
					}(),
				},
				{
					ID:      346436439,
					UserID:  "aaaa-bbbb-cccc-dddd",
					Status:  "INVALID",
					Accrual: 0,
					CreatedAt: func() time.Time {
						t, _ := time.Parse(time.RFC3339, "2020-12-09T16:09:53+03:00")
						return t
					}(),
				},
			},
			status: http.StatusOK,
			content: `[
				{
					"number": "9278923470",
					"status": "PROCESSED",
					"accrual": 500,
					"uploaded_at": "2020-12-10T15:15:45+03:00"
				},
				{
					"number": "12345678903",
					"status": "PROCESSING",
					"uploaded_at": "2020-12-10T15:12:01+03:00"
				},
				{
					"number": "346436439",
					"status": "INVALID",
					"uploaded_at": "2020-12-09T16:09:53+03:00"
				}
			]`,
		},
	}
	for _, tt := range getTests {
		t.Run("204, 200", func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, ordersURL, nil)
			accessToken := "Bearer 1234abcd5678efgh"
			r.Header.Add(authHeader, accessToken)

			userID := "aaaa-bbbb-cccc-dddd"
			ua.On("AuthByToken", mock.Anything, accessToken).Return(userID, nil).Once()
			bm.On("GetOrders", mock.Anything, userID).Return(tt.orders, nil).Once()

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			response := w.Result()

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
