package auth

import (
	"diploma-2/pkg/config"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"strings"
	"time"
)

var ErrNoCookie = errors.New("no cookie")
var ErrNoAuthHeader = errors.New("no auth header")
var ErrBadAuthHeader = errors.New("bad auth header")
var ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
var ErrInvalidToken = errors.New("invalid token")

func GenerateJWTToken(login string, cfg *config.Config) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.JWTTokenExp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Login: login,
	})
	return token.SignedString([]byte(cfg.JWTSecretKey))
}

func parseTokenFromCookie(r *http.Request, cfg *config.Config) (string, error) {
	c, err := r.Cookie(cfg.JWTCookieName)
	if err != nil || c == nil || c.Value == "" {
		return "", ErrNoCookie
	}
	return c.Value, nil
}

func parseTokenFromAuthHeader(r *http.Request) (string, error) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", ErrNoAuthHeader
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", ErrBadAuthHeader
	}
	return parts[1], nil
}

func parseLoginFromToken(raw string, cfg *config.Config) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok || t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, ErrUnexpectedSigningMethod
		}
		return []byte(cfg.JWTSecretKey), nil
	})
	if err != nil || !token.Valid {
		return "", ErrInvalidToken
	}
	return claims.Login, nil
}
