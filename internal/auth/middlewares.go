package auth

import (
	"context"
	"diploma-2/pkg/config"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func isPublicPath(path string) bool {
	public := []string{
		"/health",
		"/api/user/register",
		"/api/user/login",
	}
	// публичный accrual-эндпоинт
	if strings.HasPrefix(path, "/api/secret/") {
		return true
	}
	for _, p := range public {
		if path == p {
			return true
		}
	}
	return path == "health"
}

func MiddlewareAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if isPublicPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		var rawToken string
		var login string
		var err error

		if t, e := parseTokenFromCookie(c.Request, cfg); e == nil {
			rawToken = t
		} else {
			if t, e2 := parseTokenFromAuthHeader(c.Request); e2 == nil {
				rawToken = t
			}
		}

		if rawToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		login, err = parseLoginFromToken(rawToken, cfg)
		if err != nil || strings.TrimSpace(login) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		// Скользящее продление — при необходимости можно выключить.
		_, _ = SetTokenCookie(c, cfg, c.Writer, login)

		// Кладём login в контекст
		ctx := context.WithValue(c.Request.Context(), UserLoginKey, login)
		c.Request = c.Request.WithContext(ctx)
		c.Set(string(UserLoginKey), login)

		c.Next()
	}
}
