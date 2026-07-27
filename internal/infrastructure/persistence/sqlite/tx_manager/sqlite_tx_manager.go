package txManager

import (
	"context"
	"database/sql"

	"github.com/elias-gill/poliplanner2/internal/repository"
)

type SQLTxManager struct {
	db *sql.DB
}

func NewSQLTxManager(db *sql.DB) *SQLTxManager {
	return &SQLTxManager{db: db}
}

// !IMPORTANT. To have different keys we can implement a function that takes a custom key to extract
// from the context. This is usefull for NESTED transactions.
type txKey struct{}

// WithTransaction executes a function inside a SQL transaction
func (m *SQLTxManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Inject the transaction object inside the context so the repositories can access it
	ctxWithTx := context.WithValue(ctx, txKey{}, tx)

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // Re-panic after rollback error
		}
	}()

	if err := fn(ctxWithTx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetExecutor extracts the transcation object from the context. Returns the provided
// db connection if not exists.
//
// Recibes the context to extract the transaction and the database connection to return as a
// default value.
func GetExecutor(ctx context.Context, db *sql.DB) repository.Executor {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}
	return db
}
