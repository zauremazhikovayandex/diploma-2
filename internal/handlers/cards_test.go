package handlers

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTestCardAPI() *API {
	return &API{}
}

func TestPostCard_UnauthorizedWhenNoLoginInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := newTestCardAPI()

	c, w := newTestContext("POST", "/api/v1/cards", []byte(`{}`))

	api.PostCard(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestPostCard_BadRequestOnInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := newTestCardAPI()

	c, w := newTestContext("POST", "/api/v1/cards", []byte(`{invalid json`))

	c.Set("user_login", "testuser")

	api.PostCard(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestGetCardHandler_UnauthorizedWhenNoLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := newTestCardAPI()

	c, w := newTestContext("GET", "/api/v1/cards/some-pan", nil)

	api.GetCardHandler(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
