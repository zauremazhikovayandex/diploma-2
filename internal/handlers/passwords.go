package handlers

import (
	"diploma-2/internal/auth"
	"diploma-2/internal/repositories"
	"diploma-2/pkg/cryptoenc"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// PostPassword создаёт или обновляет запись логин/пароль для текущего пользователя
func (a *API) PostPassword(c *gin.Context) {
	// 1. Достаём логин пользователя из контекста (middleware должен был его положить)
	login, ok := auth.GetLoginFromCtx(c)
	if !ok || login == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 2. Парсим тело запроса — если JSON сломан, даже не трогаем БД
	var req repositories.PasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 3. шифруем пароль
	secret := repositories.PasswordSecret{
		Password: req.Password,
	}

	encPassword, err := cryptoenc.EncryptJSON(secret, a.cfg.DataEncKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "encryption failed"})
		return
	}

	// 4. Получаем user_id по логину.
	userID, err := repositories.GetUserIDByLogin(c.Request.Context(), a.db, login)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// 5. Собираем доменную модель
	rec := &repositories.PasswordRecord{
		UserID:    userID,
		ID:        req.ID,
		Login:     req.Login,
		Password:  encPassword,
		Meta:      req.Meta,
		CreatedAt: time.Time{},
		IsDeleted: false,
	}

	// 6. Upsert
	if err := repositories.UpsertPassword(c.Request.Context(), a.db, rec); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// 7. Читаем обратно (для created_at/updated_at)
	saved, err := repositories.GetPassword(c.Request.Context(), a.db, userID, req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if saved == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch saved record"})
		return
	}

	resp := repositories.PasswordResponse{
		ID:        saved.ID,
		Login:     saved.Login,
		Password:  saved.Password,
		Meta:      saved.Meta,
		CreatedAt: saved.CreatedAt,
		UpdatedAt: saved.UpdatedAt,
		IsDeleted: saved.IsDeleted,
	}

	c.JSON(http.StatusOK, resp)
}

// GetPasswordHandler возвращает одну запись логин/пароль по её id для текущего пользователя
func (a *API) GetPasswordHandler(c *gin.Context) {
	// 1. Авторизованный пользователь
	login, ok := auth.GetLoginFromCtx(c)
	if !ok || login == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, err := repositories.GetUserIDByLogin(c.Request.Context(), a.db, login)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// 2. Берём id из path-параметра
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	// 3. Достаём запись из репозитория
	rec, err := repositories.GetPassword(c.Request.Context(), a.db, userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if rec == nil || rec.IsDeleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	// 4. дешифруем
	var secret repositories.PasswordSecret
	if err := cryptoenc.DecryptJSON(rec.Password, a.cfg.DataEncKey, &secret); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "decryption failed"})
		return
	}

	resp := repositories.PasswordResponse{
		ID:        rec.ID,
		Login:     rec.Login,
		Password:  secret.Password,
		Meta:      rec.Meta,
		CreatedAt: rec.CreatedAt,
		UpdatedAt: rec.UpdatedAt,
		IsDeleted: rec.IsDeleted,
	}

	c.JSON(http.StatusOK, resp)
}

// GetPasswordsListHandler возвращает все записи логин/пароль текущего пользователя
func (a *API) GetPasswordsListHandler(c *gin.Context) {
	// 1. Авторизованный пользователь
	login, ok := auth.GetLoginFromCtx(c)
	if !ok || login == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, err := repositories.GetUserIDByLogin(c.Request.Context(), a.db, login)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// 2. Достаём все записи пользователя
	recs, err := repositories.ListPasswords(c.Request.Context(), a.db, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// 3. Мапим в DTO
	resp := make([]repositories.PasswordResponse, 0, len(recs))
	for _, r := range recs {
		// 4. дешифруем
		var secret repositories.PasswordSecret
		if err := cryptoenc.DecryptJSON(r.Password, a.cfg.DataEncKey, &secret); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "decryption failed"})
			return
		}
		resp = append(resp, repositories.PasswordResponse{
			ID:        r.ID,
			Login:     r.Login,
			Password:  secret.Password,
			Meta:      r.Meta,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
			IsDeleted: r.IsDeleted,
		})
	}

	c.JSON(http.StatusOK, resp)
}
