package handlers

import (
	"diploma-2/internal/auth"
	"diploma-2/pkg/config"
	"diploma-2/pkg/db"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Deps struct {
	Cfg *config.Config
	DB  *db.SqlConnection
}

type API struct {
	cfg *config.Config
	db  *db.SqlConnection
}

func New(d Deps) *API {
	return &API{
		cfg: d.Cfg,
		db:  d.DB,
	}
}

func (a *API) GetHealthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "OK")
}

func (a *API) PostRegister(ctx *gin.Context) {
	var req auth.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := auth.CreateUser(ctx.Request.Context(), a.db, req.Login, req.Password); err != nil {
		if errors.Is(err, auth.ErrLoginExists) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "login already exists"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if _, err := auth.SetTokenCookie(ctx, a.cfg, ctx.Writer, req.Login); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	//  200 — «пользователь успешно зарегистрирован и аутентифицирован»
	ctx.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func (a *API) PostLogin(ctx *gin.Context) {
	var req auth.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Проверяем логин/пароль
	if err := auth.AuthenticateUser(ctx.Request.Context(), a.db, req.Login, req.Password); err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid login or password"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Устанавливаем cookie и дублируем токен в Authorization ответа
	if _, err := auth.SetTokenCookie(ctx, a.cfg, ctx.Writer, req.Login); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "OK"})
}
