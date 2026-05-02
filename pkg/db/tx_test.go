package db

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

// TestWithinTxCommitsOnSuccess 验证事务回调成功时会提交事务。
func TestWithinTxCommitsOnSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	err = WithinTx(context.Background(), db, func(context.Context) error {
		return nil
	})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestWithinTxRollsBackOnError 验证事务回调返回错误时会回滚事务。
func TestWithinTxRollsBackOnError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	expectedErr := errors.New("boom")
	err = WithinTx(context.Background(), db, func(context.Context) error {
		return expectedErr
	})
	require.ErrorIs(t, err, expectedErr)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestWithinTxRollsBackOnPanic 验证事务回调 panic 时仍会先回滚再继续抛出 panic。
func TestWithinTxRollsBackOnPanic(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	require.Panics(t, func() {
		_ = WithinTx(context.Background(), db, func(context.Context) error {
			panic("boom")
		})
	})
	require.NoError(t, mock.ExpectationsWereMet())
}
