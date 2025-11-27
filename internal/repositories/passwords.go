package repositories

import (
	"context"
	"diploma-2/pkg/db"
	"fmt"
	"time"
)

type PasswordSecret struct {
	Password string `json:"password"`
}

// PasswordRecord — доменная модель для таблицы passwords.
type PasswordRecord struct {
	UserID    int64     // FK на logins.id
	ID        string    // ID записи (то, что задаёт клиент)
	Login     string    // логин для этого ресурса
	Password  string    // сам пароль (шифрованный)
	Meta      *string   // произвольное описание / мета
	CreatedAt time.Time // когда запись создана
	UpdatedAt time.Time // когда запись обновлена
	IsDeleted bool      // удаление
}

// PasswordRequest — тело запроса на создание/обновление записи.
type PasswordRequest struct {
	ID       string  `json:"id" binding:"required"`       // ID записи, задаётся клиентом
	Login    string  `json:"login" binding:"required"`    // логин для внешнего сервиса
	Password string  `json:"password" binding:"required"` // сам пароль (пока в открытом виде)
	Meta     *string `json:"meta"`                        // произвольное описание/мета
}

// PasswordResponse — то, что отдаём клиенту.
type PasswordResponse struct {
	ID        string    `json:"id"`
	Login     string    `json:"login"`
	Password  string    `json:"password"`
	Meta      *string   `json:"meta,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsDeleted bool      `json:"is_deleted"`
}

// UpsertPassword создаёт или обновляет запись паролей.
func UpsertPassword(ctx context.Context, conn *db.SqlConnection, rec *PasswordRecord) error {
	if rec == nil {
		return fmt.Errorf("nil PasswordRecord")
	}

	// ON CONFLICT по составному ключу (user_id, id)
	const q = `
		INSERT INTO passwords (
			user_id, id, login, password, meta, created_at, updated_at, is_deleted
		) VALUES (
			$1, $2, $3, $4, $5, COALESCE($6, now()), now(), $7
		)
		ON CONFLICT (user_id, id) DO UPDATE SET
			login      = EXCLUDED.login,
			password   = EXCLUDED.password,
			meta       = EXCLUDED.meta,
			updated_at = now(),
			is_deleted = EXCLUDED.is_deleted
	`

	created := rec.CreatedAt
	if created.IsZero() {
		// Если не передано rec.CreatedAt - по умолчанию created_date - now()
		returnWithCreatedNil := (func() error {
			_, err := db.ExecWithTimeout(ctx, conn.PgSql, conn.Timeout, q,
				rec.UserID,
				rec.ID,
				rec.Login,
				rec.Password,
				rec.Meta,
				nil,
				rec.IsDeleted,
			)
			return err
		})()
		return returnWithCreatedNil
	}

	_, err := db.ExecWithTimeout(ctx, conn.PgSql, conn.Timeout, q,
		rec.UserID,
		rec.ID,
		rec.Login,
		rec.Password,
		rec.Meta,
		rec.CreatedAt,
		rec.IsDeleted,
	)
	return err
}

// GetPassword возвращает одну запись по user_id + id.
// Если записей нет — (*PasswordRecord, nil, nil) => (nil, nil).
func GetPassword(ctx context.Context, conn *db.SqlConnection, userID int64, id string) (*PasswordRecord, error) {
	const q = `
		SELECT user_id, id, login, password, meta, created_at, updated_at, is_deleted
		FROM passwords
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

	rec, err := scanPasswordRow(row)
	if err != nil {
		return nil, err
	}
	return rec, nil
}

// ListPasswords возвращает все записи пользователя, отсортированные по updated_at.
func ListPasswords(ctx context.Context, conn *db.SqlConnection, userID int64) ([]*PasswordRecord, error) {
	const q = `
		SELECT user_id, id, login, password, meta, created_at, updated_at, is_deleted
		FROM passwords
		WHERE user_id = $1
		ORDER BY updated_at ASC
	`

	rows, err := db.ExecuteDBQueryAll(ctx, conn, q, userID)
	if err != nil {
		return nil, err
	}

	result := make([]*PasswordRecord, 0, len(rows))
	for _, r := range rows {
		rec, err := scanPasswordRow(r)
		if err != nil {
			return nil, err
		}
		result = append(result, rec)
	}
	return result, nil
}

// scanPasswordRow — хелпер для преобразования []any в PasswordRecord.
// Для того чтобы преобразовать то, как pgx возвращает типы для bigint/text/timestamptz/bool.
func scanPasswordRow(row []any) (*PasswordRecord, error) {
	if len(row) != 8 {
		return nil, fmt.Errorf("unexpected column count: %d", len(row))
	}

	var (
		userIDRaw    = row[0]
		idRaw        = row[1]
		loginRaw     = row[2]
		passwordRaw  = row[3]
		metaRaw      = row[4]
		createdRaw   = row[5]
		updatedRaw   = row[6]
		isDeletedRaw = row[7]
	)

	rec := &PasswordRecord{}

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

	// login (text)
	if s, ok := loginRaw.(string); ok {
		rec.Login = s
	} else {
		return nil, fmt.Errorf("bad type for login: %T", loginRaw)
	}

	// password (text)
	if s, ok := passwordRaw.(string); ok {
		rec.Password = s
	} else {
		return nil, fmt.Errorf("bad type for password: %T", passwordRaw)
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
