package db

import (
	"fmt"

	"diploma-2/pkg/logger"
	"diploma-2/pkg/logger/message"

	"github.com/golang-migrate/migrate/v4"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file" // file://
)

// RunMigrations применяет все доступные up-миграции к указанной базе данных.
func RunMigrations(conn *SqlConnection, migrationsDir, dbName string) error {
	driver, err := migratepgx.WithInstance(conn.SqlDB, &migratepgx.Config{})
	if err != nil {
		return fmt.Errorf("create pgx v5 driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsDir, // путь к папке с .sql миграциями
		dbName,                  // просто имя БД, например "diploma"
		driver,
	)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

// PrepareDB запускает миграции для сервисной базы данных и логирует результат.
func PrepareDB(conn *SqlConnection) {
	if err := RunMigrations(conn, "migrations", "diploma"); err != nil {
		logger.Log.Error(&message.LogMessage{
			Message: fmt.Sprintf("DB migrations ERROR: %s", err),
		})
		return
	}

	logger.Log.Info(&message.LogMessage{
		Message: "DB migrations applied successfully",
	})
}
