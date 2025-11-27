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

		`CREATE TABLE IF NOT EXISTS orders (
			number       text PRIMARY KEY,
			user_id      bigint NOT NULL REFERENCES logins(id),
			status       text NOT NULL DEFAULT 'NEW',
			accrual      numeric(12,2),
			uploaded_at  timestamptz NOT NULL DEFAULT now()
		);
		CREATE INDEX IF NOT EXISTS ix_orders_user_id ON orders(user_id);`,

		`CREATE TABLE IF NOT EXISTS balances (
			user_id    bigint PRIMARY KEY REFERENCES logins(id),
			current    numeric(14,2) NOT NULL DEFAULT 0,
			withdrawn  numeric(14,2) NOT NULL DEFAULT 0,
			updated_at timestamptz   NOT NULL DEFAULT now()
		);`,

		`CREATE TABLE IF NOT EXISTS withdrawals (
			id           bigserial PRIMARY KEY,
			user_id      bigint NOT NULL REFERENCES logins(id),
			order_number text   NOT NULL UNIQUE,
			amount       numeric(14,2) NOT NULL,
			processed_at timestamptz   NOT NULL DEFAULT now()
		);
		CREATE INDEX IF NOT EXISTS ix_withdrawals_user_id ON withdrawals(user_id);`,
	}

	for i, sqlQuery := range sqlQueries {
		_, err := db.PgSql.Exec(ctx, sqlQuery)
		if err != nil {
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
