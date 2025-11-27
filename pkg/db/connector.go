package db

import (
	"context"
	"diploma-2/pkg/config"
	"diploma-2/pkg/logger"
	"diploma-2/pkg/logger/message"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"sync"
	"time"
)

type SqlConnection struct {
	PgSql   *pgxpool.Pool
	Timeout time.Duration
}

var (
	inst *SqlConnection
	once sync.Once
)

func SqlInstance(cfg *config.DBConfig) (*SqlConnection, error) {
	var initErr error
	timeout := time.Duration(cfg.DBTimeout) * time.Millisecond // ← умножаем ОДИН раз

	once.Do(func() {
		pgCfg, err := pgxpool.ParseConfig(cfg.DatabaseUri)
		if err != nil {
			initErr = err
			return
		}
		pgCfg.MaxConns = int32(cfg.PoolSize)
		pgCfg.MaxConnLifetime = 30 * time.Minute
		pgCfg.MaxConnIdleTime = 5 * time.Minute
		pgCfg.HealthCheckPeriod = time.Minute

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		dbPool, err := pgxpool.ConnectConfig(ctx, pgCfg)
		if err != nil {
			initErr = err
			return
		}

		if err := dbPool.Ping(ctx); err != nil {
			dbPool.Close()
			initErr = fmt.Errorf("failed ping: %w", err)
			return
		}
		inst = &SqlConnection{
			PgSql:   dbPool,
			Timeout: timeout,
		}
	})

	if initErr != nil || inst == nil {
		return nil, initErr
	}
	return inst, nil
}

func (s *SqlConnection) CloseSqlInstance() {
	if s != nil && s.PgSql != nil {
		logger.Log.Info(&message.LogMessage{Message: "Closing database connection pool..."})
		s.PgSql.Close()
		logger.Log.Info(&message.LogMessage{Message: "Database connection pool closed."})
	}
}

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
