package handlers

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTestTextAPI() *API {
	return &API{}
}

func TestPostText_UnauthorizedWhenNoLoginInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := newTestTextAPI()

	c, w := newTestContext("POST", "/api/v1/texts", []byte(`{}`))

	api.PostText(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestPostText_BadRequestOnInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := newTestTextAPI()

	c, w := newTestContext("POST", "/api/v1/texts", []byte(`{invalid json`))

	c.Set("user_login", "testuser")

	api.PostText(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestGetTextHandler_UnauthorizedWhenNoLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := newTestTextAPI()

	c, w := newTestContext("GET", "/api/v1/texts/some-id", nil)

	api.GetTextHandler(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
