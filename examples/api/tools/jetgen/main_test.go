package main

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-jet/jet/v2/generator/metadata"
	gentemplate "github.com/go-jet/jet/v2/generator/template"
	"github.com/stretchr/testify/require"
)

// TestRunLoadsConfigAndInvokesGenerator 验证生成器从项目配置读取数据库连接并转交给 go-jet。
func TestRunLoadsConfigAndInvokesGenerator(t *testing.T) {
	configDir := t.TempDir()
	writeTestConfig(t, configDir, "postgres", "postgres://postgres:postgres@127.0.0.1:5432/initra?sslmode=disable")

	destDir := filepath.Join(t.TempDir(), "jet")
	var got generateRequest
	calls := 0

	err := run([]string{
		"-env", "local",
		"-config-dir", configDir,
		"-schema", "custom_schema",
		"-dest", destDir,
	}, func(key string) string {
		return ""
	}, func(req generateRequest) error {
		calls++
		got = req
		return nil
	})

	require.NoError(t, err)
	require.Equal(t, 1, calls)
	require.Equal(t, "postgres://postgres:postgres@127.0.0.1:5432/initra?sslmode=disable", got.DSN)
	require.Equal(t, "custom_schema", got.Schema)
	require.Equal(t, destDir, got.DestDir)
}

// TestRunUsesAPPEnvWhenEnvFlagOmitted 验证命令默认沿用服务启动时的 APP_ENV 选择配置文件。
func TestRunUsesAPPEnvWhenEnvFlagOmitted(t *testing.T) {
	configDir := t.TempDir()
	writeConfigFile(t, configDir, "dev", "postgres", "postgres://postgres:postgres@127.0.0.1:5432/initra_dev?sslmode=disable")

	var got generateRequest
	err := run([]string{
		"-config-dir", configDir,
	}, func(key string) string {
		if key == "APP_ENV" {
			return "dev"
		}
		return ""
	}, func(req generateRequest) error {
		got = req
		return nil
	})

	require.NoError(t, err)
	require.Equal(t, "postgres://postgres:postgres@127.0.0.1:5432/initra_dev?sslmode=disable", got.DSN)
	require.Equal(t, "public", got.Schema)
	require.Equal(t, filepath.Join("internal", "gen", "jet"), got.DestDir)
}

// TestRunRejectsNonPostgresDriver 验证当前工具只调用 postgres 生成器，避免配置与生成方言不一致。
func TestRunRejectsNonPostgresDriver(t *testing.T) {
	configDir := t.TempDir()
	writeTestConfig(t, configDir, "mysql", "root:root@tcp(127.0.0.1:3306)/initra")

	err := run([]string{
		"-env", "local",
		"-config-dir", configDir,
	}, func(key string) string {
		return ""
	}, func(req generateRequest) error {
		t.Fatal("非 postgres driver 不应调用生成器")
		return nil
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "database.driver")
}

// TestGenerateJetKeepsConfiguredDestAsGeneratedRoot 验证 dest 目录就是生成根目录，不再追加数据库名和 schema 名。
func TestGenerateJetKeepsConfiguredDestAsGeneratedRoot(t *testing.T) {
	req := generateRequest{
		DSN:     "postgres://postgres:postgres@127.0.0.1:5432/initra_dev?sslmode=disable",
		Schema:  "public",
		DestDir: filepath.Join("internal", "gen", "jet"),
	}

	var gotDSN string
	var gotSchema string
	var gotDestDir string
	var gotSchemaPath string
	closeCalled := false

	err := generateJetWith(context.Background(), req, func(ctx context.Context, dsn string) (*sql.DB, func() error, error) {
		gotDSN = dsn
		return nil, func() error {
			closeCalled = true
			return nil
		}, nil
	}, func(db *sql.DB, schema, destDir string, templates ...gentemplate.Template) error {
		gotSchema = schema
		gotDestDir = destDir
		require.Len(t, templates, 1)
		gotSchemaPath = templates[0].Schema(metadata.Schema{Name: schema}).Path
		return nil
	})

	require.NoError(t, err)
	require.True(t, closeCalled)
	require.Equal(t, req.DSN, gotDSN)
	require.Equal(t, req.Schema, gotSchema)
	require.Equal(t, req.DestDir, gotDestDir)
	require.Equal(t, ".", gotSchemaPath)
}

func writeTestConfig(t *testing.T, configDir, driver, dsn string) {
	t.Helper()
	writeConfigFile(t, configDir, "local", driver, dsn)
}

func writeConfigFile(t *testing.T, configDir, env, driver, dsn string) {
	t.Helper()

	content := []byte(`
app:
  name: initra
  env: ` + env + `
server:
  addr: ":8080"
database:
  driver: ` + driver + `
  dsn: ` + dsn + `
casbin:
  model_path: ./configs/rbac_model.conf
  policy_path: ./configs/rbac_policy.csv
auth:
  enabled: true
  jwt:
    issuer: initra
    secret: test-secret
`)
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config."+env+".yaml"), content, 0o600))
}
