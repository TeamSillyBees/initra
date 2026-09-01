package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileContentChangedIgnoresLineEndings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample.tmpl")
	if err := os.WriteFile(path, []byte("first\r\nsecond\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	changed, err := fileContentChanged(path, []byte("first\nsecond\n"))
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("line-ending-only difference must not be reported as drift")
	}
}

func TestSyncTemplatesCheckReportsDriftWithoutWriting(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	target := filepath.Join(root, "target")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "sample.txt"), []byte("new\r\nvalue\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	targetPath := filepath.Join(target, "sample.txt.tmpl")
	if err := os.WriteFile(targetPath, []byte("old\nvalue\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	opts := syncOptions{source: source, target: target, check: true, delete: true}

	actions, err := syncTemplates(opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0].kind != "update" {
		t.Fatalf("actions = %#v", actions)
	}
	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "old\nvalue\n" {
		t.Fatalf("check mode modified target: %q", content)
	}
	var output bytes.Buffer
	if err := reportSyncActions(&output, opts, actions); err == nil {
		t.Fatal("check mode must return a non-zero error when drift exists")
	}
	if !strings.Contains(output.String(), "check: 1 pending changes") {
		t.Fatalf("output = %q", output.String())
	}
}

func TestSyncTemplatesWritesLF(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	target := filepath.Join(root, "target")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "sample.txt"), []byte("first\r\nsecond\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	actions, err := syncTemplates(syncOptions{source: source, target: target, delete: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0].kind != "add" {
		t.Fatalf("actions = %#v", actions)
	}
	content, err := os.ReadFile(filepath.Join(target, "sample.txt.tmpl"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(content, []byte{'\r'}) {
		t.Fatalf("template contains CR line endings: %q", content)
	}
}

func TestTransformTemplateContentPreservesGeneratedProjectValues(t *testing.T) {
	config := `app:
  name: initra
  slug: initra
database:
  user: "initra"
  dbname: "initra"
  application_name: initra
auth:
  jwt:
    issuer: initra
    secret: "local-only-change-me-0123456789abcdef"
httpclient:
  clients:
    demo:
      headers:
        app_id: initra-httpdemo
        X-App-Id: initra
`
	rendered, err := transformTemplateContent("configs/config.yaml", config)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		`name: {{ printf "%q" .AppName }}`,
		"slug: {{ .AppSlug }}",
		`user: "{{ .AppSlug }}"`,
		`dbname: "{{ .AppSlug }}"`,
		"application_name: {{ .AppSlug }}",
		"issuer: {{ .AppSlug }}",
		`secret: "{{ .LocalJWTSecret }}"`,
	} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("rendered config does not contain %q:\n%s", expected, rendered)
		}
	}

	seed := "-- 默认管理员账号为 admin；示例仓库不提供初始密码明文。\nVALUES (\n    1000000000001,\n    'admin',\n    '$2a$10$2JNGca7Fqsq/IAvSmG2QC.XDzt5VRO9ofixT0jmZsnAic6pzBnV7C',\n    true\n)\n"
	renderedSeed, err := transformTemplateContent("db/seeds/001_seed_admin.sql", seed)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(renderedSeed, "'{{ .AdminPasswordHash }}',") {
		t.Fatalf("rendered seed = %s", renderedSeed)
	}
}

func TestTransformTemplateContentRejectsMissingCriticalAnchor(t *testing.T) {
	_, err := transformTemplateContent("configs/config.yaml", "app:\n  name: initra\n")
	if err == nil || !strings.Contains(err.Error(), "expected exactly once") {
		t.Fatalf("err = %v", err)
	}
}

func TestTransformHistoricalForeignKeyMigrationToNoOp(t *testing.T) {
	source := strings.Repeat("ALTER TABLE x ADD FOREIGN KEY (id) REFERENCES y (id);\n", 5)
	rendered, err := transformTemplateContent("db/migrations/20260715000000_add_relationship_foreign_keys.sql", source)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.ToUpper(rendered), "FOREIGN KEY (") {
		t.Fatalf("rendered migration still creates a physical foreign key: %s", rendered)
	}
	if !strings.Contains(rendered, "新生成项目不建立物理外键") {
		t.Fatalf("rendered migration does not explain the compatibility no-op: %s", rendered)
	}
}
