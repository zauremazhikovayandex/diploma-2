package handlers

import (
	"diploma-2/internal/auth"
	"diploma-2/internal/services"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

// PostCard создаёт или обновляет данные банковской карты пользователя.
func (a *API) PostCard(c *gin.Context) {
	login, ok := auth.GetLoginFromCtx(c)
	if !ok || login == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req services.CardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	resp, err := a.cardsService.CreateOrUpdateCard(c.Request.Context(), login, &req)
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

// GetCardHandler возвращает данные карты по её маскированному PAN.
func (a *API) GetCardHandler(c *gin.Context) {
	login, ok := auth.GetLoginFromCtx(c)
	if !ok || login == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	cardPAN := c.Param("id")
	if cardPAN == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	resp, err := a.cardsService.GetCard(c.Request.Context(), login, cardPAN)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		if errors.Is(err, services.ErrCardNotFound) {
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

// GetCardsListHandler возвращает список всех карт пользователя.
func (a *API) GetCardsListHandler(c *gin.Context) {
	login, ok := auth.GetLoginFromCtx(c)
	if !ok || login == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	resp, err := a.cardsService.ListCards(c.Request.Context(), login)
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
