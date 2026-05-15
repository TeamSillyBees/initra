package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/teamsillybees/initra/examples/api/internal/ent"
	"github.com/teamsillybees/initra/pkg/auth"
	"github.com/teamsillybees/initra/pkg/entx"
	"github.com/teamsillybees/initra/pkg/idgen"

	_ "github.com/lib/pq"
)

// DatabaseConfig 描述 PostgreSQL 连接配置。
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// NewEntClient 创建 Ent Client，配置连接池并注册统一运行时 Hook。
func NewEntClient(cfg DatabaseConfig, generator *idgen.Generator) (*ent.Client, error) {
	sqlDB, err := NewSQLDB(cfg)
	if err != nil {
		return nil, err
	}
	return NewEntClientFromDB(sqlDB, generator), nil
}

// NewSQLDB 创建并配置 PostgreSQL 连接池。
func NewSQLDB(cfg DatabaseConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.DBName, cfg.Password,
	)

	sqlDB, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return sqlDB, nil
}

// NewEntClientFromDB 基于已有 *sql.DB 创建 Ent Client 并注册 Hook，用于 sqlmock 等测试场景。
func NewEntClientFromDB(db *sql.DB, generator *idgen.Generator) *ent.Client {
	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))

	client.Use(entx.AuditHook(entx.AuditHookOptions{
		IDGen: generator,
		Now:   time.Now,
		Operator: func(ctx context.Context) (int64, bool) {
			if id, ok := entx.OperatorIDFromContext(ctx); ok {
				return id, true
			}
			principal, ok := auth.PrincipalFromContext(ctx)
			if !ok {
				return 0, false
			}
			return principal.UserID, true
		},
	}))
	client.Use(entx.RejectDeleteHook())

	return client
}
