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
