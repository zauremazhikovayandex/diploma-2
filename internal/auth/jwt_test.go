package auth

import (
	"diploma-2/pkg/config"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testConfig() *config.Config {
	return &config.Config{
		JWTCookieName: "auth_token",
		JWTSecretKey:  "test-secret-key",
		JWTTokenExp:   time.Hour,
		Env:           "test",
	}
}

func TestGenerateAndParseJWTToken(t *testing.T) {
	cfg := testConfig()
	login := "testuser"

	tokenStr, err := GenerateJWTToken(login, cfg)
	if err != nil {
		t.Fatalf("GenerateJWTToken error: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("expected non-empty token")
	}

	gotLogin, err := parseLoginFromToken(tokenStr, cfg)
	if err != nil {
		t.Fatalf("parseLoginFromToken error: %v", err)
	}
	if gotLogin != login {
		t.Fatalf("got login %q, want %q", gotLogin, login)
	}
}

func TestParseTokenFromCookie(t *testing.T) {
	cfg := testConfig()

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// без куки
	_, err := parseTokenFromCookie(req, cfg)
	if err == nil {
		t.Fatal("expected error when no cookie")
	}
	if !errors.Is(err, ErrNoCookie) {
		t.Fatalf("expected ErrNoCookie, got %v", err)
	}

	// с корректной кукой
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.AddCookie(&http.Cookie{
		Name:  cfg.JWTCookieName,
		Value: "token-value",
	})

	token, err := parseTokenFromCookie(req2, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "token-value" {
		t.Fatalf("got %q, want %q", token, "token-value")
	}
}

func TestParseTokenFromAuthHeader(t *testing.T) {
	// нет заголовка
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if _, err := parseTokenFromAuthHeader(req); err != ErrNoAuthHeader {
		t.Fatalf("expected ErrNoAuthHeader, got %v", err)
	}

	// неправильный формат
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("Authorization", "Bearer") // без токена
	if _, err := parseTokenFromAuthHeader(req2); err != ErrBadAuthHeader {
		t.Fatalf("expected ErrBadAuthHeader, got %v", err)
	}

	// правильный формат
	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	req3.Header.Set("Authorization", "Bearer abc.def.ghi")
	token, err := parseTokenFromAuthHeader(req3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "abc.def.ghi" {
		t.Fatalf("got %q, want %q", token, "abc.def.ghi")
	}
}

func TestParseLoginFromToken_InvalidAlgorithm(t *testing.T) {
	cfg := testConfig()

	// токен с другим алгоритмом (HS512)
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		Login: "user",
	})
	tokenStr, err := token.SignedString([]byte(cfg.JWTSecretKey))
	if err != nil {
		t.Fatalf("SignedString error: %v", err)
	}

	_, err = parseLoginFromToken(tokenStr, cfg)
	if err == nil {
		t.Fatal("expected error for unexpected signing method, got nil")
	}
	if !errors.Is(err, ErrInvalidToken) && !strings.Contains(err.Error(), "invalid token") {
		// parseLoginFromToken возвращает ErrInvalidToken при любой проблеме с токеном
		t.Fatalf("expected ErrInvalidToken-like error, got %v", err)
	}
}
