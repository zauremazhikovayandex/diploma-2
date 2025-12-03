package handlers

import (
	"diploma-2/internal/auth"
	"diploma-2/internal/services"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

// PostText создаёт или обновляет текстовую запись пользователя.
func (a *API) PostText(c *gin.Context) {
	login, ok := auth.GetLoginFromCtx(c)
	if !ok || login == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req services.TextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	resp, err := a.textsService.CreateOrUpdateText(c.Request.Context(), login, &req)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetTextHandler возвращает одну текстовую запись по её идентификатору.
func (a *API) GetTextHandler(c *gin.Context) {
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

	resp, err := a.textsService.GetText(c.Request.Context(), login, id)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		if errors.Is(err, services.ErrTextNotFound) {
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

// GetTextsListHandler возвращает список всех текстовых записей пользователя.
func (a *API) GetTextsListHandler(c *gin.Context) {
	login, ok := auth.GetLoginFromCtx(c)
	if !ok || login == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	resp, err := a.textsService.ListTexts(c.Request.Context(), login)
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
