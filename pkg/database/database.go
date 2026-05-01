package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Config 描述 PostgreSQL 连接池配置。
type Config struct {
	Driver       string `mapstructure:"driver"`
	DSN          string `mapstructure:"dsn"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

// Open 初始化 PostgreSQL 连接池并执行一次启动探活。
func Open(ctx context.Context, cfg Config) (*sql.DB, error) {
	driver := strings.TrimSpace(cfg.Driver)
	if driver == "" {
		driver = "postgres"
	}
	if !strings.EqualFold(driver, "postgres") && !strings.EqualFold(driver, "pgx") {
		return nil, fmt.Errorf("database.driver 仅支持 postgres，当前为 %q", cfg.Driver)
	}
	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}
