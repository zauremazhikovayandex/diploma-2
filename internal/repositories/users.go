package repositories

import (
	"context"
	"diploma-2/pkg/db"
	"fmt"
)

// ErrUserNotFound Если пользователь не найден — вернёт (0, ErrUserNotFound).
var ErrUserNotFound = fmt.Errorf("user not found")

// GetUserIDByLogin возвращает id пользователя по его логину.
func GetUserIDByLogin(ctx context.Context, conn *db.SqlConnection, login string) (int64, error) {
	const q = `
		SELECT id
		FROM logins
		WHERE login = $1
		LIMIT 1
	`

	row, err := db.ExecuteDBQuery(ctx, conn, q, login)
	if err != nil {
		return 0, err
	}
	if row == nil {
		return 0, ErrUserNotFound
	}

	v := row[0]
	id, ok := v.(int64)
	if !ok {
		return 0, fmt.Errorf("bad type for id: %T", v)
	}
	return id, nil
}
