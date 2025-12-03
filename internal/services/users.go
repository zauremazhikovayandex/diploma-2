package services

import (
	"context"
	"diploma-2/internal/repositories"
	"diploma-2/pkg/db"
	"errors"
)

// Общие ошибки уровня сервиса, которые возвращаются хендлерам и другим сервисам.
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrLoginExists        = errors.New("login already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTextNotFound       = errors.New("text not found")
	ErrPasswordNotFound   = errors.New("password not found")
	ErrCardNotFound       = errors.New("card not found")
	ErrBinaryNotFound     = errors.New("binary not found")
	ErrDecryptionFailed   = errors.New("decryption failed")
)

// UsersService инкапсулирует операции над пользователями: регистрацию,
// аутентификацию и получение идентификатора по логину.
type UsersService struct {
	db *db.SqlConnection
}

// NewUsersService создаёт новый сервис пользователей поверх соединения с БД.
func NewUsersService(dbConn *db.SqlConnection) *UsersService {
	return &UsersService{db: dbConn}
}

// Register регистрирует нового пользователя с указанными логином и паролем.
func (s *UsersService) Register(ctx context.Context, login, password string) error {
	err := repositories.CreateUser(ctx, s.db, login, password)
	if err != nil {
		if errors.Is(err, repositories.ErrLoginExists) {
			return ErrLoginExists
		}
		return err
	}
	return nil
}

// Authenticate проверяет логин и пароль пользователя.
func (s *UsersService) Authenticate(ctx context.Context, login, password string) error {
	err := repositories.AuthenticateUser(ctx, s.db, login, password)
	if err != nil {
		if errors.Is(err, repositories.ErrInvalidCredentials) {
			return ErrInvalidCredentials
		}
		return err
	}
	return nil
}

// GetUserIDByLogin возвращает идентификатор пользователя по его логину.
func (s *UsersService) GetUserIDByLogin(ctx context.Context, login string) (int64, error) {
	id, err := repositories.GetUserIDByLogin(ctx, s.db, login)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return 0, ErrUserNotFound
		}
		return 0, err
	}
	return id, nil
}
