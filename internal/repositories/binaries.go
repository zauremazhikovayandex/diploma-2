package repositories

import (
	"context"
	"diploma-2/pkg/db"
	"fmt"
	"time"
)

type BinarySecret struct {
	// хранится в виде base64-строки, уже внутри зашифрованного JSON.
	Data string `json:"data"`
}

// BinaryRecord — доменная модель для таблицы binaries.
type BinaryRecord struct {
	UserID    int64
	ID        string
	Data      string
	Meta      *string
	CreatedAt time.Time
	UpdatedAt time.Time
	IsDeleted bool
}

// UpsertBinary создаёт или обновляет бинарную запись.
func UpsertBinary(ctx context.Context, conn *db.SqlConnection, rec *BinaryRecord) error {
	if rec == nil {
		return fmt.Errorf("nil BinaryRecord")
	}

	const q = `
		INSERT INTO binaries (
			user_id, id, data, meta, created_at, updated_at, is_deleted
		) VALUES (
			$1, $2, $3, $4, COALESCE($5, now()), now(), $6
		)
		ON CONFLICT (user_id, id) DO UPDATE SET
			data       = EXCLUDED.data,
			meta       = EXCLUDED.meta,
			updated_at = now(),
			is_deleted = EXCLUDED.is_deleted
	`

	created := rec.CreatedAt
	if created.IsZero() {
		_, err := db.ExecWithTimeout(ctx, conn.PgSql, conn.Timeout, q,
			rec.UserID,
			rec.ID,
			rec.Data,
			rec.Meta,
			nil,
			rec.IsDeleted,
		)
		return err
	}

	_, err := db.ExecWithTimeout(ctx, conn.PgSql, conn.Timeout, q,
		rec.UserID,
		rec.ID,
		rec.Data,
		rec.Meta,
		rec.CreatedAt,
		rec.IsDeleted,
	)
	return err
}

// GetBinary возвращает одну запись по user_id + id.
func GetBinary(ctx context.Context, conn *db.SqlConnection, userID int64, id string) (*BinaryRecord, error) {
	const q = `
		SELECT user_id, id, data, meta, created_at, updated_at, is_deleted
		FROM binaries
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

	rec, err := scanBinaryRow(row)
	if err != nil {
		return nil, err
	}
	return rec, nil
}

// ListBinaries возвращает все записи пользователя.
func ListBinaries(ctx context.Context, conn *db.SqlConnection, userID int64) ([]*BinaryRecord, error) {
	const q = `
		SELECT user_id, id, data, meta, created_at, updated_at, is_deleted
		FROM binaries
		WHERE user_id = $1
		ORDER BY updated_at ASC
	`

	rows, err := db.ExecuteDBQueryAll(ctx, conn, q, userID)
	if err != nil {
		return nil, err
	}

	result := make([]*BinaryRecord, 0, len(rows))
	for _, r := range rows {
		rec, err := scanBinaryRow(r)
		if err != nil {
			return nil, err
		}
		result = append(result, rec)
	}
	return result, nil
}

// scanBinaryRow — преобразует []any в BinaryRecord.
func scanBinaryRow(row []any) (*BinaryRecord, error) {
	if len(row) != 7 {
		return nil, fmt.Errorf("unexpected column count: %d", len(row))
	}

	var (
		userIDRaw    = row[0]
		idRaw        = row[1]
		dataRaw      = row[2]
		metaRaw      = row[3]
		createdRaw   = row[4]
		updatedRaw   = row[5]
		isDeletedRaw = row[6]
	)

	rec := &BinaryRecord{}

	// user_id
	switch v := userIDRaw.(type) {
	case int64:
		rec.UserID = v
	default:
		return nil, fmt.Errorf("bad type for user_id: %T", userIDRaw)
	}

	// id
	if s, ok := idRaw.(string); ok {
		rec.ID = s
	} else {
		return nil, fmt.Errorf("bad type for id: %T", idRaw)
	}

	// data
	if s, ok := dataRaw.(string); ok {
		rec.Data = s
	} else {
		return nil, fmt.Errorf("bad type for data: %T", dataRaw)
	}

	// meta (nullable)
	if metaRaw == nil {
		rec.Meta = nil
	} else if s, ok := metaRaw.(string); ok {
		rec.Meta = &s
	} else {
		return nil, fmt.Errorf("bad type for meta: %T", metaRaw)
	}

	// created_at
	if t, ok := createdRaw.(time.Time); ok {
		rec.CreatedAt = t
	} else {
		return nil, fmt.Errorf("bad type for created_at: %T", createdRaw)
	}

	// updated_at
	if t, ok := updatedRaw.(time.Time); ok {
		rec.UpdatedAt = t
	} else {
		return nil, fmt.Errorf("bad type for updated_at: %T", updatedRaw)
	}

	// is_deleted
	if b, ok := isDeletedRaw.(bool); ok {
		rec.IsDeleted = b
	} else {
		return nil, fmt.Errorf("bad type for is_deleted: %T", isDeletedRaw)
	}

	return rec, nil
}
