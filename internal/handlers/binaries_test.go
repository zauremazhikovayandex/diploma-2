package handlers

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTestBinaryAPI() *API {
	return &API{}
}

func TestPostBinary_UnauthorizedWhenNoLoginInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := newTestBinaryAPI()

	c, w := newTestContext("POST", "/api/v1/binaries", []byte(`{}`))

	api.PostBinary(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestPostBinary_BadRequestOnInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := newTestBinaryAPI()

	c, w := newTestContext("POST", "/api/v1/binaries", []byte(`{invalid json`))

	c.Set("user_login", "testuser")

	api.PostBinary(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestGetBinaryHandler_UnauthorizedWhenNoLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := newTestBinaryAPI()

	c, w := newTestContext("GET", "/api/v1/binaries/some-id", nil)

	api.GetBinaryHandler(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
