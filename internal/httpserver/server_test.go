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
	qc := new(mockedQueueController)
	h := createHandler(ua, bm, qc)

	t.Run("not found", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		response := w.Result()

		assert.Equal(t, http.StatusNotFound, response.StatusCode)
	})
}
