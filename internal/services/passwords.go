package services

import (
	"context"
	"diploma-2/internal/repositories"
	"diploma-2/pkg/config"
	"diploma-2/pkg/cryptoenc"
	"diploma-2/pkg/db"
	"errors"
	"time"
)

// PasswordsService управляет созданием, хранением и выдачей паролей пользователя.
type PasswordsService struct {
	cfg   *config.Config
	db    *db.SqlConnection
	users *UsersService
}

// PasswordRequest конструкция входящего запроса
type PasswordRequest struct {
	ID       string  `json:"id" binding:"required"`
	Login    string  `json:"login" binding:"required"`
	Password string  `json:"password" binding:"required"`
	Meta     *string `json:"meta"`
}

// PasswordResponse конструкция исходящего ответа
type PasswordResponse struct {
	ID        string    `json:"id"`
	Login     string    `json:"login"`
	Password  string    `json:"password"`
	Meta      *string   `json:"meta,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsDeleted bool      `json:"is_deleted"`
}

// NewPasswordsService создаёт новый сервис паролей.
func NewPasswordsService(cfg *config.Config, dbConn *db.SqlConnection, users *UsersService) *PasswordsService {
	return &PasswordsService{
		cfg:   cfg,
		db:    dbConn,
		users: users,
	}
}

// CreateOrUpdatePassword создаёт или обновляет запись пароля для указанного логина.
func (s *PasswordsService) CreateOrUpdatePassword(
	ctx context.Context,
	login string,
	req *PasswordRequest,
) (*PasswordResponse, error) {
	userID, err := s.users.GetUserIDByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	secret := repositories.PasswordSecret{Password: req.Password}
	encPassword, err := cryptoenc.EncryptJSON(secret, s.cfg.DataEncKey)
	if err != nil {
		return nil, err
	}

	rec := &repositories.PasswordRecord{
		UserID:    userID,
		ID:        req.ID,
		Login:     req.Login,
		Password:  encPassword,
		Meta:      req.Meta,
		CreatedAt: time.Time{},
		IsDeleted: false,
	}

	if err := repositories.UpsertPassword(ctx, s.db, rec); err != nil {
		return nil, err
	}

	saved, err := repositories.GetPassword(ctx, s.db, userID, req.ID)
	if err != nil {
		return nil, err
	}
	if saved == nil {
		return nil, errors.New("failed to fetch saved record")
	}

	return &PasswordResponse{
		ID:        saved.ID,
		Login:     saved.Login,
		Password:  saved.Password,
		Meta:      saved.Meta,
		CreatedAt: saved.CreatedAt,
		UpdatedAt: saved.UpdatedAt,
		IsDeleted: saved.IsDeleted,
	}, nil
}

// GetPassword возвращает одну запись пароля по логину пользователя и идентификатору записи.
func (s *PasswordsService) GetPassword(
	ctx context.Context,
	login, id string,
) (*PasswordResponse, error) {
	userID, err := s.users.GetUserIDByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	rec, err := repositories.GetPassword(ctx, s.db, userID, id)
	if err != nil {
		return nil, err
	}
	if rec == nil || rec.IsDeleted {
		return nil, ErrPasswordNotFound
	}

	var secret repositories.PasswordSecret
	if err := cryptoenc.DecryptJSON(rec.Password, s.cfg.DataEncKey, &secret); err != nil {
		return nil, ErrDecryptionFailed
	}

	return &PasswordResponse{
		ID:        rec.ID,
		Login:     rec.Login,
		Password:  secret.Password,
		Meta:      rec.Meta,
		CreatedAt: rec.CreatedAt,
		UpdatedAt: rec.UpdatedAt,
		IsDeleted: rec.IsDeleted,
	}, nil
}

// ListPasswords возвращает все пароли пользователя.
func (s *PasswordsService) ListPasswords(
	ctx context.Context,
	login string,
) ([]PasswordResponse, error) {
	userID, err := s.users.GetUserIDByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	recs, err := repositories.ListPasswords(ctx, s.db, userID)
	if err != nil {
		return nil, err
	}

	resp := make([]PasswordResponse, 0, len(recs))
	for _, r := range recs {
		var secret repositories.PasswordSecret
		if err := cryptoenc.DecryptJSON(r.Password, s.cfg.DataEncKey, &secret); err != nil {
			return nil, ErrDecryptionFailed
		}
		resp = append(resp, PasswordResponse{
			ID:        r.ID,
			Login:     r.Login,
			Password:  secret.Password,
			Meta:      r.Meta,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
			IsDeleted: r.IsDeleted,
		})
	}

	return resp, nil
}
