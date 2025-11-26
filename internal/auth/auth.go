package auth

import (
	"context"
	"diploma-2/pkg/config"
	"diploma-2/pkg/db"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jackc/pgconn"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

type Claims struct {
	jwt.RegisteredClaims
	Login string `json:"login"`
}

type ctxKey string

const (
	UserLoginKey ctxKey = "user_login"
)

type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

var ErrLoginExists = errors.New("login already exists")
var ErrLoginNotExists = errors.New("login does not exist")
var ErrIncorrectPassword = errors.New("incorrect password")
var ErrInvalidCredentials = errors.New("invalid credentials")

func SetTokenCookie(c *gin.Context, cfg *config.Config, w http.ResponseWriter, login string) (string, error) {
	token, err := GenerateJWTToken(login, cfg)
	if err != nil {
		return "", err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     cfg.JWTCookieName,
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(cfg.JWTTokenExp),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   cfg.Env == "prod",
	})
	c.Header("Authorization", "Bearer "+token)
	return token, nil
}

func CreateUser(ctx context.Context, conn *db.SqlConnection, login string, password string) error {

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.ExecuteDBExec(ctx, conn,
		"INSERT INTO logins (login, password_hash) VALUES ($1, $2)", login, string(hashedPass),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrLoginExists
		}
		return err
	}
	return nil
}

func AuthenticateUser(ctx context.Context, conn *db.SqlConnection, login, password string) error {
	// Берём только хэш
	row, err := db.ExecuteDBQuery(ctx, conn, "SELECT password_hash FROM logins WHERE login = $1 LIMIT 1", login)
	if err != nil {
		return err
	}
	if row == nil {
		// Нет такого логина — не раскрываем детали
		return ErrInvalidCredentials
	}

	var hashStr string
	switch v := row[0].(type) {
	case string:
		hashStr = v
	case []byte:
		hashStr = string(v)
	default:
		// Непредвиденный тип — считаем это внутр. ошибкой
		return errors.New("bad password_hash type")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashStr), []byte(password)); err != nil {
		return ErrInvalidCredentials
	}
	return nil
}

func GetLoginFromCtx(c *gin.Context) (string, bool) {
	v, ok := c.Get(string(UserLoginKey))
	if !ok {
		return "", false
	}
	login, _ := v.(string)
	return login, login != ""
}
