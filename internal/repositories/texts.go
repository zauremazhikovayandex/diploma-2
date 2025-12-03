package repositories

import (
	"context"
	"diploma-2/pkg/db"
	"fmt"
	"time"
)

type TextSecret struct {
	Text string `json:"text"`
}

// TextRecord — доменная модель для таблицы texts.
type TextRecord struct {
	UserID    int64
	ID        string
	Text      string
	Meta      *string
	CreatedAt time.Time
	UpdatedAt time.Time
	IsDeleted bool
}

// UpsertText создаёт или обновляет текстовую запись.
func UpsertText(ctx context.Context, conn *db.SqlConnection, rec *TextRecord) error {
	if rec == nil {
		return fmt.Errorf("nil TextRecord")
	}

	const q = `
		INSERT INTO texts (
			user_id, id, text, meta, created_at, updated_at, is_deleted
		) VALUES (
			$1, $2, $3, $4, COALESCE($5, now()), now(), $6
		)
		ON CONFLICT (user_id, id) DO UPDATE SET
			text       = EXCLUDED.text,
			meta       = EXCLUDED.meta,
			updated_at = now(),
			is_deleted = EXCLUDED.is_deleted
	`

	created := rec.CreatedAt
	if created.IsZero() {
		_, err := db.ExecWithTimeout(ctx, conn.PgSql, conn.Timeout, q,
			rec.UserID,
			rec.ID,
			rec.Text,
			rec.Meta,
			nil,
			rec.IsDeleted,
		)
		return err
	}

	_, err := db.ExecWithTimeout(ctx, conn.PgSql, conn.Timeout, q,
		rec.UserID,
		rec.ID,
		rec.Text,
		rec.Meta,
		rec.CreatedAt,
		rec.IsDeleted,
	)
	return err
}

// GetText возвращает одну текстовую запись по user_id + id.
func GetText(ctx context.Context, conn *db.SqlConnection, userID int64, id string) (*TextRecord, error) {
	const q = `
		SELECT user_id, id, text, meta, created_at, updated_at, is_deleted
		FROM texts
		WHERE user_id = $1 AND id = $2
		LIMIT 1
	`

	row, err := db.ExecuteDBQuery(ctx, conn, q, userID, id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, nil
	}

	rec, err := scanTextRow(row)
	if err != nil {
		return nil, err
	}
	return rec, nil
}

// ListTexts возвращает все записи пользователя, отсортированные по updated_at.
func ListTexts(ctx context.Context, conn *db.SqlConnection, userID int64) ([]*TextRecord, error) {
	const q = `
		SELECT user_id, id, text, meta, created_at, updated_at, is_deleted
		FROM texts
		WHERE user_id = $1
		ORDER BY updated_at ASC
	`

	rows, err := db.ExecuteDBQueryAll(ctx, conn, q, userID)
	if err != nil {
		return nil, err
	}

	result := make([]*TextRecord, 0, len(rows))
	for _, r := range rows {
		rec, err := scanTextRow(r)
		if err != nil {
			return nil, err
		}
		result = append(result, rec)
	}
	return result, nil
}

// scanTextRow — преобразует []any в TextRecord.
func scanTextRow(row []any) (*TextRecord, error) {
	if len(row) != 7 {
		return nil, fmt.Errorf("unexpected column count: %d", len(row))
	}

	var (
		userIDRaw    = row[0]
		idRaw        = row[1]
		textRaw      = row[2]
		metaRaw      = row[3]
		createdRaw   = row[4]
		updatedRaw   = row[5]
		isDeletedRaw = row[6]
	)

	rec := &TextRecord{}

	// user_id (bigint)
	switch v := userIDRaw.(type) {
	case int64:
		rec.UserID = v
	default:
		return nil, fmt.Errorf("bad type for user_id: %T", userIDRaw)
	}

	// id (text)
	if s, ok := idRaw.(string); ok {
		rec.ID = s
	} else {
		return nil, fmt.Errorf("bad type for id: %T", idRaw)
	}

	// text (text)
	if s, ok := textRaw.(string); ok {
		rec.Text = s
	} else {
		return nil, fmt.Errorf("bad type for text: %T", textRaw)
	}

	// meta (nullable text)
	if metaRaw == nil {
		rec.Meta = nil
	} else if s, ok := metaRaw.(string); ok {
		rec.Meta = &s
	} else {
		return nil, fmt.Errorf("bad type for meta: %T", metaRaw)
	}

	// created_at (timestamptz)
	if t, ok := createdRaw.(time.Time); ok {
		rec.CreatedAt = t
	} else {
		return nil, fmt.Errorf("bad type for created_at: %T", createdRaw)
	}

	// updated_at (timestamptz)
	if t, ok := updatedRaw.(time.Time); ok {
		rec.UpdatedAt = t
	} else {
		return nil, fmt.Errorf("bad type for updated_at: %T", updatedRaw)
	}

	// is_deleted (bool)
	if b, ok := isDeletedRaw.(bool); ok {
		rec.IsDeleted = b
	} else {
		return nil, fmt.Errorf("bad type for is_deleted: %T", isDeletedRaw)
	}

	return rec, nil
}
