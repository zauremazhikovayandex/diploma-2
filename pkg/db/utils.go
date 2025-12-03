package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SelectWithTimeout выполняет SELECT-запрос и возвращает все строки в виде [][]any
// с использованием контекста с таймаутом.
func SelectWithTimeout(ctx context.Context, pool *pgxpool.Pool, queryTimeout time.Duration, query string, args ...any) ([][]any, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	rows, err := pool.Query(timeoutCtx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results [][]any
	for rows.Next() {
		vals, err := rows.Values() // []any для текущей строки
		if err != nil {
			return nil, err
		}
		results = append(results, vals)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// SelectOneWithTimeout выполняет SELECT-запрос и возвращает только первую строку.
func SelectOneWithTimeout(ctx context.Context, pool *pgxpool.Pool, queryTimeout time.Duration, query string, args ...any) ([]any, error) {
	rows, err := SelectWithTimeout(ctx, pool, queryTimeout, query, args...)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
}

// ExecuteDBQueryAll — универсальный helper для SELECT-запросов, возвращающих несколько строк.
func ExecuteDBQueryAll(ctx context.Context, conn *SqlConnection, query string, params ...any) ([][]any, error) {
	return SelectWithTimeout(ctx, conn.PgSql, conn.Timeout, query, params...)
}

// ExecuteDBQuery — helper для SELECT-запросов, возвращающих одну строку.
func ExecuteDBQuery(ctx context.Context, conn *SqlConnection, query string, params ...any) ([]any, error) {
	return SelectOneWithTimeout(ctx, conn.PgSql, conn.Timeout, query, params...)
}

// ExecWithTimeout выполняет командный SQL-запрос (INSERT/UPDATE/DELETE) без возврата строк.
func ExecWithTimeout(ctx context.Context, pool *pgxpool.Pool, queryTimeout time.Duration, query string, args ...any) (int64, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	ct, err := pool.Exec(timeoutCtx, query, args...)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}

// ExecuteDBExec выполняет командный запрос к БД через SqlConnection.
func ExecuteDBExec(ctx context.Context, conn *SqlConnection, query string, params ...any) (int64, error) {
	return ExecWithTimeout(ctx, conn.PgSql, conn.Timeout, query, params...)
}
