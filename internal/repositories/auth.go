package repositories

import (
	"context"
	"diploma-2/pkg/db"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type ctxKey string

// UserLoginKey используется как ключ для сохранения логина в контексте запроса.
const (
	UserLoginKey ctxKey = "user_login"
)

// ErrLoginExists возвращается, если логин уже занят другим пользователем.
// ErrInvalidCredentials возвращается при неверной паре логин/пароль.
var (
	ErrLoginExists        = errors.New("login already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// CreateUser создаёт нового пользователя с указанным логином и паролем в БД.
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

// AuthenticateUser проверяет логин и пароль пользователя по данным в БД.
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

// GetLoginFromCtx извлекает логин пользователя из gin.Context, если он там есть.
func GetLoginFromCtx(c *gin.Context) (string, bool) {
	v, ok := c.Get(string(UserLoginKey))
	if !ok {
		return "", false
	}
	login, _ := v.(string)
	return login, login != ""
}
