package database

import (
	"context"
	"database/sql"
)

// txContextKey 是事务执行器在 context 中的私有 key。
type txContextKey struct{}

// Executor 统一了 *sql.DB 与 *sql.Tx 的最小执行接口，便于仓储复用。
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// WithinTx 在同一个事务中执行回调函数，并把事务句柄写入上下文。
func WithinTx(ctx context.Context, db *sql.DB, fn func(context.Context) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	txCtx := context.WithValue(ctx, txContextKey{}, tx)
	if err := fn(txCtx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

// ExecutorFromContext 优先返回当前上下文中的事务执行器，否则回退到 DB。
func ExecutorFromContext(ctx context.Context, db *sql.DB) Executor {
	if tx, ok := ctx.Value(txContextKey{}).(*sql.Tx); ok {
		return tx
	}
	return db
}
