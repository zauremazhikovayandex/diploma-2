package db

import (
	"context"
	"diploma-2/pkg/logger"
	"diploma-2/pkg/logger/message"
	"fmt"
)

func CreateTables(db *SqlConnection) error {
	ctx := context.Background()

	sqlQueries := []string{
		`CREATE TABLE IF NOT EXISTS logins (
			id            bigserial PRIMARY KEY,
			login         text NOT NULL,
			password_hash text NOT NULL,
			created_at    timestamptz NOT NULL DEFAULT now()
		);
		CREATE UNIQUE INDEX IF NOT EXISTS ux_logins_login ON logins (login);`,

		// ====== ПАРЫ ЛОГИН/ПАРОЛЬ ======
		`CREATE TABLE IF NOT EXISTS passwords (
			user_id     bigint NOT NULL REFERENCES logins(id),
			id          text   NOT NULL,                         -- ID записи, задаётся клиентом
			login       text   NOT NULL,                         -- логин для этого ресурса
			password    text   NOT NULL,                         -- позже можно шифровать
			meta        text,                                    -- произвольное описание/мета
			created_at  timestamptz NOT NULL DEFAULT now(),
			updated_at  timestamptz NOT NULL DEFAULT now(),
			is_deleted  bool        NOT NULL DEFAULT false,
			PRIMARY KEY (user_id, id)
		);
		CREATE INDEX IF NOT EXISTS ix_passwords_user_id ON passwords(user_id);
		CREATE INDEX IF NOT EXISTS ix_passwords_user_id_updated_at
			ON passwords(user_id, updated_at);`,
	}

	for i, sqlQuery := range sqlQueries {
		if _, err := db.PgSql.Exec(ctx, sqlQuery); err != nil {
			return fmt.Errorf("error executing query %d: %w", i, err)
		}
	}

	return nil
}

func PrepareDB(db *SqlConnection) {
	err := CreateTables(db)
	if err != nil {
		logger.Log.Error(&message.LogMessage{Message: fmt.Sprintf("DB CreateTables ERROR: %s", err)})
	}
}
