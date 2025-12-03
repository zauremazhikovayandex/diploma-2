package db

import (
	"context"
	"database/sql"
	"diploma-2/pkg/config"
	"diploma-2/pkg/logger"
	"diploma-2/pkg/logger/message"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"sync"
	"time"
)

// SqlConnection инкапсулирует пул соединений pgx и адаптер database/sql,
// а также общий таймаут для запросов к БД.
type SqlConnection struct {
	PgSql   *pgxpool.Pool // основной пул для работы приложения (pgx)
	SqlDB   *sql.DB       // адаптер для database/sql — нужен migrate (database/pgx/v5)
	Timeout time.Duration
}

var (
	inst *SqlConnection
	once sync.Once
)

// SqlInstance создаёт и инициализирует глобальное подключение к БД Postgres.
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

		// В pgx/v5 вместо ConnectConfig используем NewWithConfig
		dbPool, err := pgxpool.NewWithConfig(ctx, pgCfg)
		if err != nil {
			initErr = fmt.Errorf("pgxpool.NewWithConfig: %w", err)
			return
		}

		if err := dbPool.Ping(ctx); err != nil {
			dbPool.Close()
			initErr = fmt.Errorf("failed ping: %w", err)
			return
		}

		// Адаптер database/sql поверх pgxpool.Pool — нужен для migrate/database/pgx/v5
		sqlDB := stdlib.OpenDBFromPool(dbPool)

		inst = &SqlConnection{
			PgSql:   dbPool,
			SqlDB:   sqlDB,
			Timeout: timeout,
		}
	})

	if initErr != nil || inst == nil {
		return nil, initErr
	}
	return inst, nil
}

// CloseSqlInstance корректно закрывает пул соединений и связанные ресурсы БД.
func (s *SqlConnection) CloseSqlInstance() {
	if s == nil {
		return
	}

	// Закрываем адаптер *sql.DB (он НЕ закрывает PgSql, только свои ресурсы)
	if s.SqlDB != nil {
		_ = s.SqlDB.Close()
	}

	if s.PgSql != nil {
		logger.Log.Info(&message.LogMessage{Message: "Closing database connection pool..."})
		s.PgSql.Close()
		logger.Log.Info(&message.LogMessage{Message: "Database connection pool closed."})
	}
}
