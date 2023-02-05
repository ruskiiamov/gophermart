package httpserver

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const loginURL = "/api/user/login"

func TestLogin(t *testing.T) {
	ua := new(mockedUserAuthorizer)
	bm := new(mockedBonusManager)
	handler := createHandler(ua, bm)

	t.Run("405", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, loginURL, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		response := w.Result()

		assert.Equal(t, http.StatusMethodNotAllowed, response.StatusCode)
		assert.Equal(t, http.MethodPost, response.Header.Get("Allow"))
	})

	t.Run("200", func(t *testing.T) {
		content := `{"login": "test_login","password": "test_pass"}`
		body := strings.NewReader(content)
		r := httptest.NewRequest(http.MethodPost, loginURL, body)
		r.Header.Add(contTypeHeader, appJSON)

		accessToken := "Bearer XaSCghyfce324G5ef53dbhU643"
		ua.On("Login", mock.Anything, "test_login", "test_pass").Return(accessToken, nil).Once()

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		response := w.Result()

		ua.AssertExpectations(t)

		resContent, err := io.ReadAll(response.Body)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.NotEmpty(t, response.Header.Get(authHeader))
		assert.Regexp(t, "^Bearer\\s[\\w-.]+\\w$", response.Header.Get(authHeader))
		assert.Empty(t, resContent)
	})

	t.Run("400 empty header", func(t *testing.T) {
		content := `{"login": "test_login","password": "test_pass"}`
		body := strings.NewReader(content)
		r := httptest.NewRequest(http.MethodPost, loginURL, body)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		response := w.Result()

		resContent, err := io.ReadAll(response.Body)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		assert.Empty(t, response.Header.Get(authHeader))
		assert.Empty(t, resContent)
	})

	t.Run("400 invalid json", func(t *testing.T) {
		content := `{"login": "test_login","password": "test_pass",}`
		body := strings.NewReader(content)
		r := httptest.NewRequest(http.MethodPost, loginURL, body)
		r.Header.Add(contTypeHeader, appJSON)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		response := w.Result()

		resContent, err := io.ReadAll(response.Body)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		assert.Empty(t, response.Header.Get(authHeader))
		assert.Empty(t, string(resContent))
	})

	t.Run("400 wrong json structure", func(t *testing.T) {
		content := `{"login": "test_login","pass": "test_pass"}`
		body := strings.NewReader(content)
		r := httptest.NewRequest(http.MethodPost, loginURL, body)
		r.Header.Add(contTypeHeader, appJSON)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		response := w.Result()

		resContent, err := io.ReadAll(response.Body)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		assert.Empty(t, response.Header.Get(authHeader))
		assert.Empty(t, resContent)
	})

	t.Run("401 wrong login-password pair", func(t *testing.T) {
		content := `{"login": "test_login","password": "test_pass"}`
		body := strings.NewReader(content)
		r := httptest.NewRequest(http.MethodPost, loginURL, body)
		r.Header.Add(contTypeHeader, appJSON)

		ua.On("Login", mock.Anything, "test_login", "test_pass").Return("", ErrLoginPassword).Once()

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		response := w.Result()

		ua.AssertExpectations(t)

		resContent, err := io.ReadAll(response.Body)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
		assert.Empty(t, response.Header.Get(authHeader))
		assert.Empty(t, resContent)
	})
}
