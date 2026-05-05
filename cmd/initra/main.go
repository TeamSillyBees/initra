package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/teamsillybees/initra/templates"
)

const frameworkModule = "github.com/teamsillybees/initra"

var (
	goPackageNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	safeNamePattern      = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
)

var version = "dev"

type templateData struct {
	ModulePath       string
	AppName          string
	FrameworkModule  string
	FrameworkVersion string
	LocalReplacePath string
}

func main() {
	if err := run(os.Args[1:], os.Stdout, version); err != nil {
		log.Fatalf("initra: %v", err)
	}
}

func run(args []string, stdout io.Writer, cliVersion string) error {
	if len(args) == 0 {
		return fmt.Errorf("用法: initra <command>")
	}
	switch args[0] {
	case "new":
		return runNew(args[1:], stdout, cliVersion)
	case "module":
		return runModule(args[1:], stdout)
	case "crud":
		return runCRUD(args[1:], stdout)
	case "config":
		return runConfig(args[1:], stdout)
	case "migrate":
		return runMigrate(args[1:], stdout)
	case "doctor":
		return runDoctor(args[1:], stdout)
	default:
		return fmt.Errorf("未知命令 %q", args[0])
	}
}

func runNew(args []string, stdout io.Writer, cliVersion string) error {
	if len(args) == 0 {
		return fmt.Errorf("用法: initra new <dir>")
	}

	targetDir := args[0]
	flags := flag.NewFlagSet("new", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	modulePath := flags.String("module", "", "生成项目的 Go module path")
	appName := flags.String("app-name", "", "应用名称")
	projectType := flags.String("type", "api", "项目类型：api 或 worker")
	templateName := flags.String("template", "", "兼容旧参数，等同于 --type")
	frameworkVersion := flags.String("framework-version", "", "initra 框架版本")
	localReplacePath := flags.String("replace", "", "本地 initra 仓库路径，用于 go.mod replace")

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		return fmt.Errorf("new 命令只接受一个目标目录")
	}
	resolvedType := strings.TrimSpace(*projectType)
	if resolvedType == "" {
		resolvedType = "api"
	}
	if template := strings.TrimSpace(*templateName); template != "" {
		if resolvedType != "api" && resolvedType != template {
			return fmt.Errorf("--type 与 --template 不能指定不同项目类型")
		}
		resolvedType = template
	}
	if resolvedType == "basic" {
		resolvedType = "api"
	}
	if resolvedType != "api" && resolvedType != "worker" {
		return fmt.Errorf("暂不支持项目类型 %q", resolvedType)
	}

	normalizedAppName := strings.TrimSpace(*appName)
	if normalizedAppName == "" {
		normalizedAppName = filepath.Base(filepath.Clean(targetDir))
	}
	normalizedModulePath := strings.TrimSpace(*modulePath)
	if normalizedModulePath == "" {
		normalizedModulePath = normalizedAppName
	}

	resolvedReplace, err := normalizeReplacePath(*localReplacePath)
	if err != nil {
		return err
	}
	resolvedVersion, err := resolveFrameworkVersion(*frameworkVersion, cliVersion, resolvedReplace)
	if err != nil {
		return err
	}

	if err := ensureWritableTarget(targetDir); err != nil {
		return err
	}

	data := templateData{
		ModulePath:       normalizedModulePath,
		AppName:          normalizedAppName,
		FrameworkModule:  frameworkModule,
		FrameworkVersion: resolvedVersion,
		LocalReplacePath: resolvedReplace,
	}
	if err := renderTemplate(resolvedType, targetDir, data); err != nil {
		return err
	}

	if stdout != nil {
		_, _ = fmt.Fprintf(stdout, "created %s\n", targetDir)
	}
	return nil
}

func runModule(args []string, stdout io.Writer) error {
	if len(args) < 2 || args[0] != "add" {
		return fmt.Errorf("用法: initra module add <name>")
	}
	if len(args) > 2 {
		return fmt.Errorf("module add 只接受一个模块名")
	}
	name, err := normalizeGoPackageName(args[1])
	if err != nil {
		return err
	}

	moduleDir := filepath.Join("internal", "module", name)
	if err := os.MkdirAll(moduleDir, 0o755); err != nil {
		return err
	}

	files := map[string]string{
		name + ".model.go":   moduleModelTemplate(name),
		name + ".service.go": moduleServiceTemplate(name),
		name + ".repo.go":    moduleRepoTemplate(name),
		name + ".handler.go": moduleHandlerTemplate(name),
		name + ".routes.go":  moduleRoutesTemplate(name),
		"providers.go":       moduleProvidersTemplate(name),
		name + "_test.go":    moduleTestTemplate(name),
	}
	for filename, content := range files {
		if err := writeNewFile(filepath.Join(moduleDir, filename), content); err != nil {
			return err
		}
	}

	if stdout != nil {
		_, _ = fmt.Fprintf(stdout, "created module %s\n", name)
	}
	return nil
}

func runCRUD(args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "add" {
		return fmt.Errorf("用法: initra crud add <module> --table <table>")
	}
	if len(args) < 2 {
		return fmt.Errorf("crud add 缺少模块名")
	}
	moduleName, err := normalizeGoPackageName(args[1])
	if err != nil {
		return err
	}

	flags := flag.NewFlagSet("crud add", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	tableName := flags.String("table", "", "数据表名")
	if err := flags.Parse(args[2:]); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		return fmt.Errorf("crud add 只接受一个模块名")
	}
	table := strings.TrimSpace(*tableName)
	if table == "" {
		return fmt.Errorf("crud add 必须提供 --table")
	}

	moduleDir := filepath.Join("internal", "module", moduleName)
	if _, err := os.Stat(moduleDir); err != nil {
		return fmt.Errorf("模块 %s 不存在，请先执行 initra module add %s", moduleName, moduleName)
	}
	if err := writeNewFile(filepath.Join(moduleDir, moduleName+".crud.go"), crudTemplate(moduleName, table)); err != nil {
		return err
	}
	if stdout != nil {
		_, _ = fmt.Fprintf(stdout, "created crud sample for %s\n", moduleName)
	}
	return nil
}

func runConfig(args []string, stdout io.Writer) error {
	if len(args) != 2 || args[0] != "add" {
		return fmt.Errorf("用法: initra config add <capability>")
	}
	capability, err := normalizeGoPackageName(args[1])
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join("internal", "boot"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll("configs", 0o755); err != nil {
		return err
	}
	if err := writeNewFile(filepath.Join("internal", "boot", capability+".config.go"), configGoTemplate(capability)); err != nil {
		return err
	}
	if err := writeNewFile(filepath.Join("configs", capability+".yaml"), configYAMLTemplate(capability)); err != nil {
		return err
	}
	if stdout != nil {
		_, _ = fmt.Fprintf(stdout, "created config %s\n", capability)
	}
	return nil
}

func runMigrate(args []string, stdout io.Writer) error {
	if len(args) != 2 {
		return fmt.Errorf("用法: initra migrate <new|diff> <name>")
	}
	name, err := normalizeSafeName(args[1])
	if err != nil {
		return err
	}

	switch args[0] {
	case "new":
		if err := os.MkdirAll(filepath.Join("db", "migrations"), 0o755); err != nil {
			return err
		}
		filename := time.Now().UTC().Format("20060102150405") + "_" + name + ".sql"
		path := filepath.Join("db", "migrations", filename)
		if err := writeNewFile(path, "-- "+name+"\n\n"); err != nil {
			return err
		}
		if stdout != nil {
			_, _ = fmt.Fprintf(stdout, "created migration %s\n", path)
		}
		return nil
	case "diff":
		if err := os.MkdirAll("scripts", 0o755); err != nil {
			return err
		}
		path := filepath.Join("scripts", "migrate-diff-"+name+".ps1")
		content := "param(\n    [string]$Env = \"local\"\n)\n\natlas migrate diff " + name + " --env $Env -c file://db/atlas.hcl\n"
		if err := writeNewFile(path, content); err != nil {
			return err
		}
		if stdout != nil {
			_, _ = fmt.Fprintf(stdout, "created migrate diff script %s\n", path)
		}
		return nil
	default:
		return fmt.Errorf("未知 migrate 子命令 %q", args[0])
	}
}

func runDoctor(args []string, stdout io.Writer) error {
	if len(args) != 0 {
		return fmt.Errorf("用法: initra doctor")
	}
	if stdout == nil {
		stdout = io.Discard
	}

	reportTool(stdout, "Go", "go", "version")
	reportTool(stdout, "Atlas", "atlas", "version")
	reportTool(stdout, "go-jet", "jet", "-version")
	reportTool(stdout, "golangci-lint", "golangci-lint", "version")
	reportFile(stdout, "config.local.yaml", filepath.Join("configs", "config.local.yaml"))
	reportFile(stdout, "Atlas config", filepath.Join("db", "atlas.hcl"))
	return nil
}

func resolveFrameworkVersion(flagVersion string, cliVersion string, replacePath string) (string, error) {
	if version := strings.TrimSpace(flagVersion); version != "" {
		return version, nil
	}
	if replacePath != "" {
		return "v0.0.0", nil
	}
	if version := strings.TrimSpace(cliVersion); version != "" && version != "dev" {
		return version, nil
	}
	return "", fmt.Errorf("开发版 CLI 必须提供 --framework-version 或 --replace")
}

func normalizeReplacePath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", nil
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("解析 replace 路径失败: %w", err)
	}
	return filepath.ToSlash(absolute), nil
}

func ensureWritableTarget(targetDir string) error {
	entries, err := os.ReadDir(targetDir)
	switch {
	case err == nil && len(entries) > 0:
		return fmt.Errorf("目标目录 %s 已存在且非空", targetDir)
	case err == nil:
		return nil
	case errors.Is(err, os.ErrNotExist):
		return os.MkdirAll(targetDir, 0o755)
	default:
		return err
	}
}

func renderTemplate(templateName string, targetDir string, data templateData) error {
	root := templateName
	return fs.WalkDir(templates.FS, root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}

		relativePath, err := filepath.Rel(root, filepath.ToSlash(path))
		if err != nil {
			return err
		}
		outputPath := filepath.Join(targetDir, filepath.FromSlash(strings.TrimSuffix(relativePath, ".tmpl")))

		if entry.IsDir() {
			return os.MkdirAll(outputPath, 0o755)
		}
		if !strings.HasSuffix(relativePath, ".tmpl") {
			return nil
		}

		content, err := templates.FS.ReadFile(path)
		if err != nil {
			return err
		}
		parsed, err := template.New(relativePath).Option("missingkey=error").Parse(string(content))
		if err != nil {
			return fmt.Errorf("解析模板 %s 失败: %w", relativePath, err)
		}

		var rendered bytes.Buffer
		if err := parsed.Execute(&rendered, data); err != nil {
			return fmt.Errorf("渲染模板 %s 失败: %w", relativePath, err)
		}

		if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(outputPath, rendered.Bytes(), 0o644)
	})
}

func normalizeGoPackageName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if !goPackageNamePattern.MatchString(name) {
		return "", fmt.Errorf("名称 %q 必须匹配 %s", name, goPackageNamePattern.String())
	}
	return name, nil
}

func normalizeSafeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if !safeNamePattern.MatchString(name) {
		return "", fmt.Errorf("名称 %q 只能包含字母、数字、下划线和中划线", name)
	}
	return name, nil
}

func writeNewFile(path string, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("文件 %s 已存在", path)
		}
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	return err
}

func reportTool(stdout io.Writer, label string, command string, args ...string) {
	path, err := exec.LookPath(command)
	if err != nil {
		_, _ = fmt.Fprintf(stdout, "%s: MISSING (%s)\n", label, command)
		return
	}
	output, err := exec.Command(command, args...).CombinedOutput()
	if err != nil {
		_, _ = fmt.Fprintf(stdout, "%s: FOUND %s\n", label, path)
		return
	}
	_, _ = fmt.Fprintf(stdout, "%s: OK %s\n", label, firstLine(string(output)))
}

func reportFile(stdout io.Writer, label string, path string) {
	if _, err := os.Stat(path); err != nil {
		_, _ = fmt.Fprintf(stdout, "%s: MISSING %s\n", label, path)
		return
	}
	_, _ = fmt.Fprintf(stdout, "%s: OK %s\n", label, path)
}

func firstLine(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if index := strings.IndexAny(value, "\r\n"); index >= 0 {
		return value[:index]
	}
	return value
}

func exportedName(name string) string {
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '_' || r == '-'
	})
	var builder strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		builder.WriteString(strings.ToUpper(part[:1]))
		if len(part) > 1 {
			builder.WriteString(part[1:])
		}
	}
	return builder.String()
}

func pluralName(name string) string {
	if strings.HasSuffix(name, "s") {
		return name
	}
	return name + "s"
}

func moduleModelTemplate(name string) string {
	typeName := exportedName(name)
	return fmt.Sprintf(`package %s

// %s 是 %s 模块的领域占位模型。
type %s struct {
	ID int64 `+"`json:\"id\"`"+`
}
`, name, typeName, name, typeName)
}

func moduleServiceTemplate(name string) string {
	typeName := exportedName(name)
	return fmt.Sprintf(`package %s

import "context"

// Service 是 %s 模块的应用服务。
type Service struct{}

// NewService 创建 %s 模块应用服务。
func NewService() *Service {
	return &Service{}
}

// Get 返回 %s 详情占位数据。
func (s *Service) Get(ctx context.Context, id int64) (*%s, error) {
	_ = s
	_ = ctx
	return &%s{ID: id}, nil
}
`, name, name, name, name, typeName, typeName)
}

func moduleRepoTemplate(name string) string {
	return fmt.Sprintf(`package %s

// Repository 是 %s 模块的数据访问占位实现。
type Repository struct{}

// NewRepository 创建 %s 模块仓储。
func NewRepository() *Repository {
	return &Repository{}
}
`, name, name, name)
}

func moduleHandlerTemplate(name string) string {
	typeName := exportedName(name)
	return fmt.Sprintf(`package %s

import (
	"context"

	"github.com/teamsillybees/initra/pkg/requestctx"
	"github.com/teamsillybees/initra/pkg/response"
)

// Get%sInput 描述 %s 详情查询路径参数。
type Get%sInput struct {
	ID int64 `+"`path:\"id\" doc:\"ID\"`"+`
}

// %sResponse 描述 %s 对外响应。
type %sResponse struct {
	ID int64 `+"`json:\"id\"`"+`
}

type get%sOutput struct {
	Body response.SuccessResponse[%sResponse]
}

// Handler 封装 %s 模块 HTTP 适配逻辑。
type Handler struct {
	service *Service
}

// NewHandler 创建 %s 模块 Handler。
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Get 返回 %s 详情。
func (h *Handler) Get(ctx context.Context, input *Get%sInput) (*get%sOutput, error) {
	item, err := h.service.Get(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	return &get%sOutput{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), %sResponse{ID: item.ID}),
	}, nil
}
`, name, typeName, name, typeName, typeName, name, typeName, typeName, typeName, name, name, name, typeName, typeName, typeName, typeName)
}

func moduleRoutesTemplate(name string) string {
	typeName := exportedName(name)
	path := "/api/v1/" + pluralName(name) + "/{id}"
	return fmt.Sprintf(`package %s

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	platformauth "github.com/teamsillybees/initra/pkg/auth"
	"github.com/teamsillybees/initra/pkg/server"
)

// Module 负责 %s 模块路由注册。
type Module struct {
	handler *Handler
}

// NewModule 创建 %s 模块实例。
func NewModule(handler *Handler) *Module {
	return &Module{handler: handler}
}

// Register 将 %s 模块注册到应用。
func (m *Module) Register(api huma.API, registry *server.RouteRegistry) {
	huma.Register(api, huma.Operation{
		OperationID: "get-%s",
		Method:      http.MethodGet,
		Path:        "%s",
		Summary:     "查询%s详情",
		Tags:        []string{"%s"},
	}, m.handler.Get)
	registry.Register(http.MethodGet, "%s", platformauth.RouteSecurity{Resource: "%s", Action: "read"})
}
`, name, name, name, name, name, path, typeName, typeName, path, name)
}

func moduleProvidersTemplate(name string) string {
	return fmt.Sprintf(`package %s

import "github.com/samber/do"

const (
	%sServiceName = "%s.service"
	%sHandlerName = "%s.handler"
)

// Provide 使用 do 注册 %s 模块依赖。
func Provide(injector *do.Injector) {
	do.ProvideNamed(injector, %sServiceName, func(i *do.Injector) (*Service, error) {
		return NewService(), nil
	})
	do.ProvideNamed(injector, %sHandlerName, func(i *do.Injector) (*Handler, error) {
		service := do.MustInvokeNamed[*Service](i, %sServiceName)
		return NewHandler(service), nil
	})
	do.Provide(injector, func(i *do.Injector) (*Module, error) {
		handler := do.MustInvokeNamed[*Handler](i, %sHandlerName)
		return NewModule(handler), nil
	})
}
`, name, name, name, name, name, name, name, name, name, name)
}

func moduleTestTemplate(name string) string {
	typeName := exportedName(name)
	return fmt.Sprintf(`package %s

import (
	"context"
	"testing"
)

// TestServiceGet 验证 %s 模块服务占位逻辑可调用。
func TestServiceGet(t *testing.T) {
	item, err := NewService().Get(context.Background(), 1001)
	if err != nil {
		t.Fatalf("Get() error = %%v", err)
	}
	if item.ID != 1001 {
		t.Fatalf("Get() ID = %%d, want 1001", item.ID)
	}
}

var _ = (*%s)(nil)
`, name, name, typeName)
}

func crudTemplate(moduleName string, tableName string) string {
	typeName := exportedName(moduleName)
	return fmt.Sprintf(`package %s

const %sCRUDTable = %q

// %sCRUD 是基于数据表 %s 的 CRUD 样例占位。
type %sCRUD struct{}

// New%sCRUD 创建 CRUD 样例。
func New%sCRUD() *%sCRUD {
	return &%sCRUD{}
}
`, moduleName, moduleName, tableName, typeName, tableName, typeName, typeName, typeName, typeName, typeName)
}

func configGoTemplate(capability string) string {
	typeName := exportedName(capability) + "Config"
	return fmt.Sprintf(`package boot

// %s 描述 %s 能力的配置占位。
type %s struct {
	Enabled bool `+"`mapstructure:\"enabled\"`"+`
}
`, typeName, capability, typeName)
}

func configYAMLTemplate(capability string) string {
	return fmt.Sprintf(`%s:
  enabled: false
`, capability)
}
