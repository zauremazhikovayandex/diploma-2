package auth

import (
	"diploma-2/internal/repositories"
	"diploma-2/pkg/config"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func testMiddlewareConfig() *config.Config {
	return &config.Config{
		JWTCookieName: "auth_token",
		JWTSecretKey:  "test-secret-key",
		JWTTokenExp:   time.Hour,
		Env:           "test",
	}
}

func TestMiddlewareAuth_PublicPathAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := testMiddlewareConfig()

	r := gin.New()
	r.Use(MiddlewareAuth(cfg))

	// /health считается публичным
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestMiddlewareAuth_ProtectedPathUnauthorizedWithoutToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := testMiddlewareConfig()

	r := gin.New()
	r.Use(MiddlewareAuth(cfg))

	// защищённый эндпоинт
	r.GET("/api/v1/passwords", func(c *gin.Context) {
		c.String(http.StatusOK, "SHOULD NOT BE REACHED")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/passwords", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMiddlewareAuth_ProtectedPathWithValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := testMiddlewareConfig()

	r := gin.New()
	r.Use(MiddlewareAuth(cfg))

	// эндпоинт, который читает логин из контекста
	r.GET("/api/v1/passwords", func(c *gin.Context) {
		loginIfc, exists := c.Get(string(repositories.UserLoginKey))
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no login in context"})
			return
		}
		login, _ := loginIfc.(string)
		c.JSON(http.StatusOK, gin.H{"login": login})
	})

	// создаём валидный токен
	token, err := GenerateJWTToken("testuser", cfg)
	if err != nil {
		t.Fatalf("GenerateJWTToken error: %v", err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/passwords", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}
