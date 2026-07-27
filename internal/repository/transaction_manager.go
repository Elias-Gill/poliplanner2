package repository

import (
	"context"
	"database/sql"
)

// Executor is the common interface between sql.Db and sql.Tx
type Executor interface {
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)

	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)

	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row

	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// ===================
// Main interface
// ===================

// TxManager define cómo envolvemos una operación en una transacción
type TxManager interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
