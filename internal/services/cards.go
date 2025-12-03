package services

import (
	"context"
	"diploma-2/internal/repositories"
	"diploma-2/pkg/config"
	"diploma-2/pkg/cryptoenc"
	"diploma-2/pkg/db"
	"errors"
	"fmt"
	"strings"
	"time"
)

type CardsService struct {
	cfg   *config.Config
	db    *db.SqlConnection
	users *UsersService
}

type CardRequest struct {
	Number string  `json:"number" binding:"required"` // полный номер карты
	Holder string  `json:"holder" binding:"required"` // имя владельца
	Expire string  `json:"expire" binding:"required"` // срок действия (например, 12/30)
	CVV    string  `json:"cvv" binding:"required"`    // CVV
	Meta   *string `json:"meta"`                      // произвольная мета
}

type CardResponse struct {
	CardPAN string  `json:"card_pan"` // маскированный PAN, например 4600********5363
	Number  string  `json:"number"`   // полный номер (для владельца, уже расшифрованный)
	Holder  string  `json:"holder"`
	Expire  string  `json:"expire"`
	CVV     string  `json:"cvv"`
	Meta    *string `json:"meta,omitempty"`
}

// NewCardsService создаёт сервис карт.
func NewCardsService(cfg *config.Config, dbConn *db.SqlConnection, users *UsersService) *CardsService {
	return &CardsService{
		cfg:   cfg,
		db:    dbConn,
		users: users,
	}
}

// maskCardPAN уникальный id по номеру карты: первые 4 цифры + 8 * + последние 4 цифры.
func maskCardPAN(number string) (string, error) {
	num := strings.ReplaceAll(number, " ", "")
	num = strings.ReplaceAll(num, "-", "")

	if len(num) < 8 {
		return "", fmt.Errorf("card number too short")
	}
	return num[:4] + "********" + num[len(num)-4:], nil
}

// CreateOrUpdateCard создаёт или обновляет запись карты пользователя.
func (s *CardsService) CreateOrUpdateCard(
	ctx context.Context,
	login string,
	req *CardRequest,
) (*CardResponse, error) {
	userID, err := s.users.GetUserIDByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	cardPAN, err := maskCardPAN(req.Number)
	if err != nil {
		return nil, err
	}

	secret := repositories.CardSecret{
		Number: req.Number,
		Holder: req.Holder,
		Expire: req.Expire,
		CVV:    req.CVV,
	}
	encData, err := cryptoenc.EncryptJSON(secret, s.cfg.DataEncKey)
	if err != nil {
		return nil, err
	}

	rec := &repositories.CardRecord{
		UserID:    userID,
		CardPAN:   cardPAN,
		Data:      encData,
		Meta:      req.Meta,
		CreatedAt: time.Time{},
		IsDeleted: false,
	}

	if err := repositories.UpsertCard(ctx, s.db, rec); err != nil {
		return nil, err
	}

	saved, err := repositories.GetCard(ctx, s.db, userID, cardPAN)
	if err != nil {
		return nil, err
	}
	if saved == nil {
		return nil, errors.New("failed to fetch saved card record")
	}

	var secretOut repositories.CardSecret
	if err := cryptoenc.DecryptJSON(saved.Data, s.cfg.DataEncKey, &secretOut); err != nil {
		return nil, ErrDecryptionFailed
	}

	return &CardResponse{
		CardPAN: saved.CardPAN,
		Number:  secretOut.Number,
		Holder:  secretOut.Holder,
		Expire:  secretOut.Expire,
		CVV:     secretOut.CVV,
		Meta:    saved.Meta,
	}, nil
}

// GetCard — логика GET /api/v1/cards/:id (id = card_pan).
func (s *CardsService) GetCard(
	ctx context.Context,
	login, cardPAN string,
) (*CardResponse, error) {
	userID, err := s.users.GetUserIDByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	rec, err := repositories.GetCard(ctx, s.db, userID, cardPAN)
	if err != nil {
		return nil, err
	}
	if rec == nil || rec.IsDeleted {
		return nil, ErrCardNotFound
	}

	var secret repositories.CardSecret
	if err := cryptoenc.DecryptJSON(rec.Data, s.cfg.DataEncKey, &secret); err != nil {
		return nil, ErrDecryptionFailed
	}

	return &CardResponse{
		CardPAN: rec.CardPAN,
		Number:  secret.Number,
		Holder:  secret.Holder,
		Expire:  secret.Expire,
		CVV:     secret.CVV,
		Meta:    rec.Meta,
	}, nil
}

// ListCards — логика GET /api/v1/cards.
func (s *CardsService) ListCards(
	ctx context.Context,
	login string,
) ([]CardResponse, error) {
	userID, err := s.users.GetUserIDByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	recs, err := repositories.ListCards(ctx, s.db, userID)
	if err != nil {
		return nil, err
	}

	resp := make([]CardResponse, 0, len(recs))
	for _, r := range recs {
		if r.IsDeleted {
			continue
		}
		var secret repositories.CardSecret
		if err := cryptoenc.DecryptJSON(r.Data, s.cfg.DataEncKey, &secret); err != nil {
			return nil, ErrDecryptionFailed
		}
		resp = append(resp, CardResponse{
			CardPAN: r.CardPAN,
			Number:  secret.Number,
			Holder:  secret.Holder,
			Expire:  secret.Expire,
			CVV:     secret.CVV,
			Meta:    r.Meta,
		})
	}

	return resp, nil
}
