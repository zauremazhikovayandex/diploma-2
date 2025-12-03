package handlers

import (
	"diploma-2/internal/auth"
	"diploma-2/internal/services"
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
	cfg              *config.Config
	passwordsService *services.PasswordsService
	textsService     *services.TextsService
	cardsService     *services.CardsService
	binariesService  *services.BinariesService
	usersService     *services.UsersService
}

func New(d Deps) *API {
	usersSvc := services.NewUsersService(d.DB)
	passwordsSvc := services.NewPasswordsService(d.Cfg, d.DB, usersSvc)
	textsSvc := services.NewTextsService(d.Cfg, d.DB, usersSvc)
	cardsSvc := services.NewCardsService(d.Cfg, d.DB, usersSvc)
	binariesSvc := services.NewBinariesService(d.Cfg, d.DB, usersSvc)

	return &API{
		cfg:              d.Cfg,
		passwordsService: passwordsSvc,
		textsService:     textsSvc,
		cardsService:     cardsSvc,
		binariesService:  binariesSvc,
		usersService:     usersSvc,
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

	// Вызываем сервис вместо repositories.CreateUser
	if err := a.usersService.Register(ctx.Request.Context(), req.Login, req.Password); err != nil {
		if errors.Is(err, services.ErrLoginExists) {
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

	ctx.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func (a *API) PostLogin(ctx *gin.Context) {
	var req auth.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Проверяем логин/пароль через сервис
	if err := a.usersService.Authenticate(ctx.Request.Context(), req.Login, req.Password); err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
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
