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

type BinariesService struct {
	cfg   *config.Config
	db    *db.SqlConnection
	users *UsersService
}

type BinaryRequest struct {
	ID   string  `json:"id" binding:"required"`   // ID записи (задаётся клиентом)
	Data string  `json:"data" binding:"required"` // произвольные бинарные данные (например, base64)
	Meta *string `json:"meta"`                    // произвольная мета
}

type BinaryResponse struct {
	ID        string    `json:"id"`
	Data      string    `json:"data"`
	Meta      *string   `json:"meta,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsDeleted bool      `json:"is_deleted"`
}

func NewBinariesService(cfg *config.Config, dbConn *db.SqlConnection, users *UsersService) *BinariesService {
	return &BinariesService{
		cfg:   cfg,
		db:    dbConn,
		users: users,
	}
}

// CreateOrUpdateBinary создаёт или обновляет запись с бинарными данными.
func (s *BinariesService) CreateOrUpdateBinary(
	ctx context.Context,
	login string,
	req *BinaryRequest,
) (*BinaryResponse, error) {
	userID, err := s.users.GetUserIDByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	secret := repositories.BinarySecret{Data: req.Data}
	encData, err := cryptoenc.EncryptJSON(secret, s.cfg.DataEncKey)
	if err != nil {
		return nil, err
	}

	rec := &repositories.BinaryRecord{
		UserID:    userID,
		ID:        req.ID,
		Data:      encData,
		Meta:      req.Meta,
		CreatedAt: time.Time{},
		IsDeleted: false,
	}

	if err := repositories.UpsertBinary(ctx, s.db, rec); err != nil {
		return nil, err
	}

	saved, err := repositories.GetBinary(ctx, s.db, userID, req.ID)
	if err != nil {
		return nil, err
	}
	if saved == nil {
		return nil, errors.New("failed to fetch saved binary record")
	}

	// Как и в Passwords/Text CreateOrUpdate — отдаём зашифрованное значение.
	return &BinaryResponse{
		ID:        saved.ID,
		Data:      saved.Data,
		Meta:      saved.Meta,
		CreatedAt: saved.CreatedAt,
		UpdatedAt: saved.UpdatedAt,
		IsDeleted: saved.IsDeleted,
	}, nil
}

// GetBinary — логика GET /api/v1/binaries/:id
func (s *BinariesService) GetBinary(
	ctx context.Context,
	login, id string,
) (*BinaryResponse, error) {
	userID, err := s.users.GetUserIDByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	rec, err := repositories.GetBinary(ctx, s.db, userID, id)
	if err != nil {
		return nil, err
	}
	if rec == nil || rec.IsDeleted {
		return nil, ErrBinaryNotFound
	}

	var secret repositories.BinarySecret
	if err := cryptoenc.DecryptJSON(rec.Data, s.cfg.DataEncKey, &secret); err != nil {
		return nil, ErrDecryptionFailed
	}

	return &BinaryResponse{
		ID:        rec.ID,
		Data:      secret.Data, // уже расшифрованное содержимое
		Meta:      rec.Meta,
		CreatedAt: rec.CreatedAt,
		UpdatedAt: rec.UpdatedAt,
		IsDeleted: rec.IsDeleted,
	}, nil
}

// ListBinaries — логика GET /api/v1/binaries
func (s *BinariesService) ListBinaries(
	ctx context.Context,
	login string,
) ([]BinaryResponse, error) {
	userID, err := s.users.GetUserIDByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	recs, err := repositories.ListBinaries(ctx, s.db, userID)
	if err != nil {
		return nil, err
	}

	resp := make([]BinaryResponse, 0, len(recs))
	for _, r := range recs {
		if r.IsDeleted {
			continue
		}
		var secret repositories.BinarySecret
		if err := cryptoenc.DecryptJSON(r.Data, s.cfg.DataEncKey, &secret); err != nil {
			return nil, ErrDecryptionFailed
		}
		resp = append(resp, BinaryResponse{
			ID:        r.ID,
			Data:      secret.Data,
			Meta:      r.Meta,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
			IsDeleted: r.IsDeleted,
		})
	}

	return resp, nil
}
