package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-jet/jet/v2/generator/metadata"
	"github.com/go-jet/jet/v2/generator/postgres"
	gentemplate "github.com/go-jet/jet/v2/generator/template"
	jetpostgres "github.com/go-jet/jet/v2/postgres"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/teamsillybees/initra/examples/api/internal/boot"
)

type generateRequest struct {
	// DSN 是从应用配置读取的数据库连接串，避免生成入口和运行时配置漂移。
	DSN string
	// Schema 指定 go-jet introspect 的 PostgreSQL schema。
	Schema string
	// DestDir 指定生成代码输出目录，目录内容会由 go-jet 生成器维护。
	DestDir string
}

type generatorFunc func(generateRequest) error

type dbOpenerFunc func(context.Context, string) (*sql.DB, func() error, error)
type generateDBFunc func(*sql.DB, string, string, ...gentemplate.Template) error

// main 从配置中读取数据库连接，并委托 go-jet 生成 PostgreSQL 查询代码。
func main() {
	if err := run(os.Args[1:], os.Getenv, generateJet); err != nil {
		log.Fatalf("jetgen failed: %v", err)
	}
}

func run(args []string, getenv func(string) string, generate generatorFunc) error {
	if getenv == nil {
		getenv = os.Getenv
	}
	if generate == nil {
		return fmt.Errorf("generator function 不能为空")
	}

	defaultEnv := strings.TrimSpace(getenv("APP_ENV"))
	if defaultEnv == "" {
		defaultEnv = "dev"
	}

	flags := flag.NewFlagSet("jetgen", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	env := flags.String("env", defaultEnv, "配置环境名称，默认读取 APP_ENV")
	configDir := flags.String("config-dir", filepath.Join(".", "configs"), "配置文件目录")
	schema := flags.String("schema", "public", "PostgreSQL schema 名称")
	destDir := flags.String("dest", filepath.Join("internal", "gen", "jet"), "Jet 代码输出目录")

	if err := flags.Parse(args); err != nil {
		return err
	}

	normalizedEnv := strings.TrimSpace(*env)
	if normalizedEnv == "" {
		return fmt.Errorf("env 不能为空")
	}

	normalizedConfigDir := strings.TrimSpace(*configDir)
	if normalizedConfigDir == "" {
		return fmt.Errorf("config-dir 不能为空")
	}

	normalizedSchema := strings.TrimSpace(*schema)
	if normalizedSchema == "" {
		return fmt.Errorf("schema 不能为空")
	}

	normalizedDestDir := strings.TrimSpace(*destDir)
	if normalizedDestDir == "" {
		return fmt.Errorf("dest 不能为空")
	}

	cfg, err := boot.LoadConfig(normalizedEnv, normalizedConfigDir)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	if driver := strings.ToLower(strings.TrimSpace(cfg.Database.Driver)); driver != "postgres" {
		return fmt.Errorf("db.driver 仅支持 postgres，当前为 %q", cfg.Database.Driver)
	}

	req := generateRequest{
		DSN:     cfg.Database.DSN,
		Schema:  normalizedSchema,
		DestDir: normalizedDestDir,
	}
	if err := generate(req); err != nil {
		return fmt.Errorf("生成 Jet 代码失败: %w", err)
	}

	return nil
}

func generateJet(req generateRequest) error {
	return generateJetWith(context.Background(), req, openPostgresDB, postgres.GenerateDB)
}

func generateJetWith(ctx context.Context, req generateRequest, openDB dbOpenerFunc, generateDB generateDBFunc) error {
	db, closeDB, err := openDB(ctx, req.DSN)
	if err != nil {
		return err
	}
	defer func() {
		_ = closeDB()
	}()

	return generateDB(db, req.Schema, req.DestDir, generatorTemplate())
}

func openPostgresDB(ctx context.Context, dsn string) (*sql.DB, func() error, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("打开 PostgreSQL 连接失败: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("连接 PostgreSQL 失败: %w", err)
	}

	return db, db.Close, nil
}

func generatorTemplate() gentemplate.Template {
	return gentemplate.Default(jetpostgres.Dialect).
		UseSchema(func(schemaMetaData metadata.Schema) gentemplate.Schema {
			return gentemplate.DefaultSchema(schemaMetaData).UsePath(".")
		})
}
