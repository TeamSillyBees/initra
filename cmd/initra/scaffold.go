package main

import (
	"fmt"
	"go/format"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var (
	goPackageNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	safeNamePattern      = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
)

type snippetAddOptions struct {
	tableName string
}

func newModuleCommand(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "module",
		Short:         "管理业务模块骨架",
		Long:          "管理标准项目的 internal/modules/<name> 业务模块骨架。模块按 flat package 组织，包含 handler、service、dto、routes、providers 和测试文件。",
		Example:       "  initra module add order",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showCommandHelp(cmd)
		},
	}
	configureCommand(cmd, stdout)
	cmd.AddCommand(newModuleAddCommand(stdout))
	return cmd
}

func newModuleAddCommand(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "add <name>",
		Short:         "生成 flat package 业务模块骨架",
		Long:          "在当前项目的 internal/modules 下生成一个业务模块骨架，模块名必须是合法 Go package 名称。",
		Example:       "  initra module add order",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          requireExactArgs(1, "模块名"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return addModule(args[0], cmd.OutOrStdout())
		},
	}
	configureCommand(cmd, stdout)
	return cmd
}

func newSnippetCommand(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "snippet",
		Short:         "管理显式代码片段",
		Long:          "为已存在的业务模块追加不带持久化或路由接线承诺的代码片段。",
		Example:       "  initra snippet add order --table sys_order",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showCommandHelp(cmd)
		},
	}
	configureCommand(cmd, stdout)
	cmd.AddCommand(newSnippetAddCommand(stdout))
	return cmd
}

func newSnippetAddCommand(stdout io.Writer) *cobra.Command {
	opts := snippetAddOptions{}
	cmd := &cobra.Command{
		Use:           "add <module>",
		Short:         "为现有模块生成数据表名片段",
		Long:          "在现有业务模块目录中生成 <module>.table.go，只声明规范化后的数据表名，不承诺 CRUD、持久化或路由接线。",
		Example:       "  initra snippet add order --table sys_order",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          requireExactArgs(1, "模块名"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return addTableSnippet(args[0], opts, cmd.OutOrStdout())
		},
	}
	configureCommand(cmd, stdout)
	cmd.Flags().StringVar(&opts.tableName, "table", "", "数据表名")
	return cmd
}

func newConfigCommand(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "config",
		Short:         "管理应用聚合配置",
		Long:          "为标准项目追加可被 LoadConfig 实际加载的能力配置，并更新 internal/boot.Config 与 configs/config.yaml。",
		Example:       "  initra config add feature_flag",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showCommandHelp(cmd)
		},
	}
	configureCommand(cmd, stdout)
	cmd.AddCommand(newConfigAddCommand(stdout))
	return cmd
}

func newConfigAddCommand(stdout io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "add <capability>",
		Short:         "接入新的能力配置",
		Long:          "事务化生成 internal/boot/<capability>.config.go，并把能力字段接入 boot.Config 与 configs/config.yaml。",
		Example:       "  initra config add feature_flag",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          requireExactArgs(1, "能力名"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return addConfigSnippet(args[0], cmd.OutOrStdout())
		},
	}
	configureCommand(cmd, stdout)
	return cmd
}

func addModule(name string, stdout io.Writer) error {
	name, err := normalizeGoPackageName(name)
	if err != nil {
		return err
	}

	files := map[string]string{
		name + ".service.go": moduleServiceTemplate(name),
		name + ".dto.go":     moduleDTOTemplate(name),
		name + ".handler.go": moduleHandlerTemplate(name),
		name + ".routes.go":  moduleRoutesTemplate(name),
		"providers.go":       moduleProvidersTemplate(name),
		name + "_test.go":    moduleTestTemplate(name),
	}
	for filename, content := range files {
		formatted, err := format.Source([]byte(content))
		if err != nil {
			return fmt.Errorf("格式化模块文件 %s 失败: %w", filename, err)
		}
		files[filename] = string(formatted)
	}
	moduleDir := filepath.Join("internal", "modules", name)
	if err := writeModuleDirectoryTransaction(moduleDir, files); err != nil {
		return err
	}

	if stdout != nil {
		_, _ = fmt.Fprintf(stdout, "created module %s\n", name)
	}
	return nil
}

func addTableSnippet(moduleName string, opts snippetAddOptions, stdout io.Writer) error {
	moduleName, err := normalizeGoPackageName(moduleName)
	if err != nil {
		return err
	}
	table := strings.TrimSpace(opts.tableName)
	if table == "" {
		return fmt.Errorf("snippet add 必须提供 --table")
	}
	if _, err := normalizeSafeName(table); err != nil {
		return fmt.Errorf("数据表名无效: %w", err)
	}

	moduleDir := filepath.Join("internal", "modules", moduleName)
	if _, err := os.Stat(moduleDir); err != nil {
		return fmt.Errorf("模块 %s 不存在，请先执行 initra module add %s", moduleName, moduleName)
	}
	content, err := format.Source([]byte(tableSnippetTemplate(moduleName, table)))
	if err != nil {
		return fmt.Errorf("格式化表名片段失败: %w", err)
	}
	if err := applyFileChangesTransaction([]fileChange{{
		path:       filepath.Join(moduleDir, moduleName+".table.go"),
		content:    content,
		createOnly: true,
	}}); err != nil {
		return err
	}
	if stdout != nil {
		_, _ = fmt.Fprintf(stdout, "created table snippet for %s\n", moduleName)
	}
	return nil
}

func addConfigSnippet(capability string, stdout io.Writer) error {
	capability, err := normalizeGoPackageName(capability)
	if err != nil {
		return err
	}
	configPath := filepath.Join("internal", "boot", "config.go")
	yamlPath := filepath.Join("configs", "config.yaml")
	configSource, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取 %s 失败，请在标准 initra 项目根目录执行: %w", configPath, err)
	}
	yamlSource, err := os.ReadFile(yamlPath)
	if err != nil {
		return fmt.Errorf("读取 %s 失败，请在标准 initra 项目根目录执行: %w", yamlPath, err)
	}
	updatedConfig, err := addConfigFieldToAggregate(configSource, capability)
	if err != nil {
		return err
	}
	updatedYAML, err := appendConfigYAML(yamlSource, capability)
	if err != nil {
		return err
	}
	configTypeSource, err := format.Source([]byte(configGoTemplate(capability)))
	if err != nil {
		return fmt.Errorf("格式化能力配置失败: %w", err)
	}
	if err := applyFileChangesTransaction([]fileChange{
		{path: configPath, content: updatedConfig, mustExist: true},
		{path: yamlPath, content: updatedYAML, mustExist: true},
		{path: filepath.Join("internal", "boot", capability+".config.go"), content: configTypeSource, createOnly: true},
	}); err != nil {
		return err
	}
	if stdout != nil {
		_, _ = fmt.Fprintf(stdout, "created config %s\n", capability)
	}
	return nil
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

func moduleServiceTemplate(name string) string {
	typeName := exportedName(name)
	return fmt.Sprintf(`package %s

import (
	"context"

	"github.com/teamsillybees/initra/pkg/idgen"
)

// Service 是 %s 模块的应用服务。
type Service struct{}

// NewService 创建 %s 模块应用服务。
func NewService() *Service {
	return &Service{}
}

// Get 返回 %s 详情占位数据。
func (s *Service) Get(ctx context.Context, id idgen.ID) (%sVO, error) {
	_ = s
	_ = ctx
	return %sVO{ID: id}, nil
}
`, name, name, name, name, typeName, typeName)
}

func moduleHandlerTemplate(name string) string {
	typeName := exportedName(name)
	return fmt.Sprintf(`package %s

import (
	"context"

	"github.com/teamsillybees/initra/pkg/response"
)

// Handler 封装 %s 模块 HTTP 适配逻辑。
type Handler struct {
	service *Service
}

// NewHandler 创建 %s 模块 Handler。
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) get(ctx context.Context, input *get%sRequest) (*get%sResponse, error) {
	item, err := h.service.Get(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	return &get%sResponse{
		Body: response.OK(ctx, item),
	}, nil
}
`, name, name, name, typeName, typeName, typeName)
}

func moduleDTOTemplate(name string) string {
	typeName := exportedName(name)
	return fmt.Sprintf(`package %s

import (
	"github.com/teamsillybees/initra/pkg/idgen"
	"github.com/teamsillybees/initra/pkg/response"
)

type get%sRequest struct {
	ID idgen.ID `+"`path:\"id\" example:\"1771234567890123456\" doc:\"ID\"`"+`
}

// %sVO 描述 %s 对外 JSON DTO。
type %sVO struct {
	ID idgen.ID `+"`json:\"id\"`"+`
}

type get%sResponse struct {
	Body response.SuccessVO[%sVO]
}
`, name, typeName, typeName, name, typeName, typeName, typeName)
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
	}, m.handler.get)
	registry.Register(http.MethodGet, "%s", platformauth.RouteSecurity{AccessMode: platformauth.AccessModePermission, Resource: "%s", Action: "read"})
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

`, name, name)
}

func tableSnippetTemplate(moduleName string, tableName string) string {
	typeName := exportedName(moduleName)
	return fmt.Sprintf(`package %s

// %sTableName 记录模块约定的数据表名；该片段不提供 CRUD 或持久化接线。
const %sTableName = %q
`, moduleName, typeName, typeName, tableName)
}

func configGoTemplate(capability string) string {
	typeName := exportedName(capability) + "Config"
	return fmt.Sprintf(`package boot

// %s 描述 %s 能力配置；零值表示默认关闭。
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
