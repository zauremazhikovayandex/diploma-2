package handlers

import (
	"diploma-2/internal/auth"
	"diploma-2/internal/services"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

// PostPassword POST /api/v1/passwords
func (a *API) PostPassword(c *gin.Context) {
	// 1. логин из контекста
	login, ok := auth.GetLoginFromCtx(c)
	if !ok || login == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 2. парсим JSON
	var req services.PasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 3. вызываем сервисный слой
	resp, err := a.passwordsService.CreateOrUpdatePassword(c.Request.Context(), login, &req)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// 4. HTTP-ответ
	c.JSON(http.StatusOK, resp)
}

// GetPasswordHandler GET /api/v1/passwords/:id
func (a *API) GetPasswordHandler(c *gin.Context) {
	login, ok := auth.GetLoginFromCtx(c)
	if !ok || login == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	resp, err := a.passwordsService.GetPassword(c.Request.Context(), login, id)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		if errors.Is(err, services.ErrPasswordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if errors.Is(err, services.ErrDecryptionFailed) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "decryption failed"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetPasswordsListHandler GET /api/v1/passwords
func (a *API) GetPasswordsListHandler(c *gin.Context) {
	login, ok := auth.GetLoginFromCtx(c)
	if !ok || login == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	resp, err := a.passwordsService.ListPasswords(c.Request.Context(), login)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		if errors.Is(err, services.ErrDecryptionFailed) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "decryption failed"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
