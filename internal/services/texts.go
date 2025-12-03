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

type TextsService struct {
	cfg   *config.Config
	db    *db.SqlConnection
	users *UsersService
}

type TextRequest struct {
	ID   string  `json:"id" binding:"required"`   // заголовок/ID записи
	Text string  `json:"text" binding:"required"` // сам текст
	Meta *string `json:"meta"`                    // произвольная мета
}

type TextResponse struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Meta      *string   `json:"meta,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsDeleted bool      `json:"is_deleted"`
}

func NewTextsService(cfg *config.Config, dbConn *db.SqlConnection, users *UsersService) *TextsService {
	return &TextsService{
		cfg:   cfg,
		db:    dbConn,
		users: users,
	}
}

func (s *TextsService) CreateOrUpdateText(
	ctx context.Context,
	login string,
	req *TextRequest,
) (*TextResponse, error) {
	userID, err := s.users.GetUserIDByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	secret := repositories.TextSecret{Text: req.Text}
	encText, err := cryptoenc.EncryptJSON(secret, s.cfg.DataEncKey)
	if err != nil {
		return nil, err
	}

	rec := &repositories.TextRecord{
		UserID:    userID,
		ID:        req.ID,
		Text:      encText,
		Meta:      req.Meta,
		CreatedAt: time.Time{},
		IsDeleted: false,
	}

	if err := repositories.UpsertText(ctx, s.db, rec); err != nil {
		return nil, err
	}

	saved, err := repositories.GetText(ctx, s.db, userID, req.ID)
	if err != nil {
		return nil, err
	}
	if saved == nil {
		return nil, errors.New("failed to fetch saved text record")
	}

	// Зашифрованное значение
	return &TextResponse{
		ID:        saved.ID,
		Text:      saved.Text,
		Meta:      saved.Meta,
		CreatedAt: saved.CreatedAt,
		UpdatedAt: saved.UpdatedAt,
		IsDeleted: saved.IsDeleted,
	}, nil
}

// GetText — логика GET /api/v1/texts/:id
func (s *TextsService) GetText(
	ctx context.Context,
	login, id string,
) (*TextResponse, error) {
	userID, err := s.users.GetUserIDByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	rec, err := repositories.GetText(ctx, s.db, userID, id)
	if err != nil {
		return nil, err
	}
	if rec == nil || rec.IsDeleted {
		return nil, ErrTextNotFound
	}

	var secret repositories.TextSecret
	if err := cryptoenc.DecryptJSON(rec.Text, s.cfg.DataEncKey, &secret); err != nil {
		return nil, ErrDecryptionFailed
	}

	return &TextResponse{
		ID:        rec.ID,
		Text:      secret.Text,
		Meta:      rec.Meta,
		CreatedAt: rec.CreatedAt,
		UpdatedAt: rec.UpdatedAt,
		IsDeleted: rec.IsDeleted,
	}, nil
}

// ListTexts — логика GET /api/v1/texts
func (s *TextsService) ListTexts(
	ctx context.Context,
	login string,
) ([]TextResponse, error) {
	userID, err := s.users.GetUserIDByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	recs, err := repositories.ListTexts(ctx, s.db, userID)
	if err != nil {
		return nil, err
	}

	resp := make([]TextResponse, 0, len(recs))
	for _, r := range recs {
		var secret repositories.TextSecret
		if err := cryptoenc.DecryptJSON(r.Text, s.cfg.DataEncKey, &secret); err != nil {
			return nil, ErrDecryptionFailed
		}
		resp = append(resp, TextResponse{
			ID:        r.ID,
			Text:      secret.Text,
			Meta:      r.Meta,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
			IsDeleted: r.IsDeleted,
		})
	}

	return resp, nil
}
