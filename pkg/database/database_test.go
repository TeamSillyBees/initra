package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/samber/do"
	"github.com/stretchr/testify/require"
)

var testDriverID atomic.Int64

func TestOpenPingsDatabase(t *testing.T) {
	name := registerTestDriver(nil)

	db, err := Open(context.Background(), Config{
		DriverName:     name,
		DataSourceName: "test",
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})
}

func TestOpenReturnsPingError(t *testing.T) {
	name := registerTestDriver(errors.New("database unavailable"))

	db, err := Open(context.Background(), Config{
		DriverName:     name,
		DataSourceName: "test",
	})

	require.Nil(t, db)
	require.ErrorContains(t, err, "database ping failed")
}

func TestRegisterProvidesPingingDatabase(t *testing.T) {
	name := registerTestDriver(nil)
	injector := do.New()

	Register(injector, Config{
		DriverName:     name,
		DataSourceName: "test",
	})

	db := do.MustInvoke[*sql.DB](injector)
	require.NotNil(t, db)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})
}

func registerTestDriver(pingErr error) string {
	name := fmt.Sprintf("initra_database_test_%d", testDriverID.Add(1))
	sql.Register(name, &pingDriver{pingErr: pingErr})
	return name
}

type pingDriver struct {
	pingErr error
}

func (d *pingDriver) Open(_ string) (driver.Conn, error) {
	return &pingConn{pingErr: d.pingErr}, nil
}

type pingConn struct {
	pingErr error
}

func (c *pingConn) Ping(_ context.Context) error {
	return c.pingErr
}

func (c *pingConn) Prepare(_ string) (driver.Stmt, error) {
	return nil, errors.New("prepare is unused")
}

func (c *pingConn) Close() error {
	return nil
}

func (c *pingConn) Begin() (driver.Tx, error) {
	return nil, errors.New("begin is unused")
}
