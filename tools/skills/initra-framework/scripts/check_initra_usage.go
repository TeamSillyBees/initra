package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// rule 描述一条静态扫描规则。
type rule struct {
	id               string
	severity         string
	message          string
	patterns         []string
	regexps          []*regexp.Regexp
	filenameSuffixes []string
	allowed          func(path string) bool
}

// finding 描述一次规则命中。
type finding struct {
	path     string
	line     int
	severity string
	ruleID   string
	message  string
	text     string
}

// main 解析命令行参数并执行 initra 用法扫描。
func main() {
	root := flag.String("root", ".", "要扫描的项目根目录")
	includeTests := flag.Bool("include-tests", false, "是否扫描 *_test.go 文件")
	flag.Parse()

	absRoot, err := filepath.Abs(*root)
	if err != nil {
		exitWithError(err)
	}
	findings, err := scan(absRoot, *includeTests)
	if err != nil {
		exitWithError(err)
	}
	if len(findings) == 0 {
		fmt.Println("initra 用法检查通过")
		return
	}
	for _, f := range findings {
		rel, _ := filepath.Rel(absRoot, f.path)
		fmt.Printf("%s:%d: [%s] %s: %s\n", rel, f.line, f.severity, f.ruleID, f.message)
		if strings.TrimSpace(f.text) != "" {
			fmt.Printf("  %s\n", strings.TrimSpace(f.text))
		}
	}
	os.Exit(1)
}

// scan 扫描项目中的 Go 文件并返回所有命中项。
func scan(root string, includeTests bool) ([]finding, error) {
	rules := defaultRules()
	var findings []finding
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if shouldSkipDir(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if !includeTests && strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fileFindings, err := scanFile(path, rules)
		if err != nil {
			return err
		}
		findings = append(findings, fileFindings...)
		return nil
	})
	return findings, err
}

// scanFile 扫描单个 Go 文件。
func scanFile(path string, rules []rule) ([]finding, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var findings []finding
	findings = append(findings, filenameFindings(path, rules)...)

	scanner := bufio.NewScanner(file)
	line := 0
	for scanner.Scan() {
		line++
		text := scanner.Text()
		for _, r := range rules {
			if r.allowed != nil && r.allowed(path) {
				continue
			}
			if matchesRule(text, r) {
				findings = append(findings, finding{
					path:     path,
					line:     line,
					severity: r.severity,
					ruleID:   r.id,
					message:  r.message,
					text:     text,
				})
			}
		}
	}
	return findings, scanner.Err()
}

// filenameFindings 根据文件名规则生成命中项。
func filenameFindings(path string, rules []rule) []finding {
	var findings []finding
	for _, r := range rules {
		if len(r.filenameSuffixes) == 0 {
			continue
		}
		if r.allowed != nil && r.allowed(path) {
			continue
		}
		for _, suffix := range r.filenameSuffixes {
			if strings.HasSuffix(filepath.ToSlash(path), suffix) {
				findings = append(findings, finding{
					path:     path,
					line:     1,
					severity: r.severity,
					ruleID:   r.id,
					message:  r.message,
				})
			}
		}
	}
	return findings
}

// defaultRules 返回内置 initra 用法扫描规则。
func defaultRules() []rule {
	return []rule{
		{
			id:       "redis-client-in-business",
			severity: "error",
			message:  "使用 boot 层的 redisx.Register，并注入 Redis 接口；不要在业务代码中创建 client",
			patterns: []string{"redis.NewClient(", "redis.NewUniversalClient(", "redis.NewFailoverClient(", "redis.NewClusterClient("},
			allowed:  allowFrameworkOrBoot,
		},
		{
			id:       "redis-keys",
			severity: "error",
			message:  "生产 Redis 代码禁止使用 KEYS；优先使用 redisx ScanPrefix/UnlinkByPrefix",
			regexps:  mustRegexps(`\.Keys\s*\(`, `\bKEYS\b`),
			allowed:  allowFrameworkOrBoot,
		},
		{
			id:       "adhoc-http-client",
			severity: "warning",
			message:  "远程服务调用使用 httpclient.Register 和命名 httpclient.Client 依赖",
			patterns: []string{"http.DefaultClient", "http.Client{", "&http.Client{"},
			allowed:  allowFrameworkOrBoot,
		},
		{
			id:       "direct-storage-provider",
			severity: "error",
			message:  "业务代码应依赖 storage.Service，不应依赖 provider-specific package 或云厂商 SDK",
			patterns: []string{
				"github.com/teamsillybees/initra/pkg/storage/aliyunoss",
				"github.com/teamsillybees/initra/pkg/storage/awss3",
				"github.com/teamsillybees/initra/pkg/storage/tencentcos",
				"github.com/aliyun/aliyun-oss-go-sdk",
				"github.com/aws/aws-sdk-go-v2/service/s3",
				"github.com/tencentyun/cos-go-sdk-v5",
			},
			allowed: allowFrameworkOrBoot,
		},
		{
			id:       "direct-asynq-in-business",
			severity: "error",
			message:  "业务代码应依赖 pkg/task，不要直接 import Asynq；Asynq 类型只允许在 pkg/task/asynqadapter 内出现",
			patterns: []string{"github.com/hibiken/asynq"},
			allowed:  allowFrameworkOrBoot,
		},
		{
			id:       "custom-config-loader",
			severity: "warning",
			message:  "使用 pkg/config LoadInto 和 Config.Validate，不要重复实现 Viper loader",
			patterns: []string{"viper.New(", "viper.SetConfig", "viper.ReadInConfig"},
			allowed:  allowFrameworkOrBoot,
		},
		{
			id:       "custom-logger",
			severity: "warning",
			message:  "使用 logx.Register 和注入的 *logx.Logger，不要初始化第二套全局 logger",
			patterns: []string{"zap.NewProduction(", "zap.NewDevelopment(", "zap.Config{"},
			allowed:  allowFrameworkOrBoot,
		},
		{
			id:       "root-internal-import",
			severity: "error",
			message:  "业务项目和模板不得 import initra 根仓库 internal package",
			patterns: []string{"github.com/teamsillybees/initra/internal/"},
			allowed:  allowFrameworkOrBoot,
		},
		{
			id:       "injector-in-business-logic",
			severity: "warning",
			message:  "在 boot/provider 文件中解析依赖；service 和 handler 应通过构造函数接收依赖",
			patterns: []string{"do.MustInvoke", "do.Invoke"},
			allowed:  allowProviderFile,
		},
		{
			id:       "two-stage-config-registration",
			severity: "warning",
			message:  "配置应直接传给 Register(injector, cfg)，不要用 do.ProvideValue 做两阶段配置注册",
			patterns: []string{"do.ProvideValue"},
			allowed:  allowFrameworkOrTests,
		},
		{
			id:               "deprecated-module-layer",
			severity:         "warning",
			message:          "标准 API 模块默认不创建 repo/model 层；需要数据库的 service 直接使用 Ent Client",
			filenameSuffixes: []string{".repo.go", ".model.go"},
			allowed:          allowFrameworkOrGenerated,
		},
	}
}

// matchesRule 判断一行文本是否命中规则。
func matchesRule(text string, r rule) bool {
	for _, pattern := range r.patterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	for _, re := range r.regexps {
		if re.MatchString(text) {
			return true
		}
	}
	return false
}

// shouldSkipDir 判断扫描时应跳过的目录。
func shouldSkipDir(name string) bool {
	switch name {
	case ".agents", ".codex", ".git", ".idea", ".vscode", "node_modules", "vendor", "tmp", "var":
		return true
	default:
		return false
	}
}

// allowFrameworkOrBoot 放行框架源码、boot 装配层和 skill 自身。
func allowFrameworkOrBoot(path string) bool {
	normalized := filepath.ToSlash(path)
	return strings.Contains(normalized, "/pkg/") ||
		strings.Contains(normalized, "/cmd/initra/") ||
		strings.Contains(normalized, "/internal/boot/") ||
		strings.Contains(normalized, "/templates/") ||
		strings.Contains(normalized, "/tools/skills/")
}

// allowProviderFile 放行 provider 文件中的依赖解析。
func allowProviderFile(path string) bool {
	normalized := filepath.ToSlash(path)
	base := filepath.Base(path)
	return allowFrameworkOrBoot(path) ||
		base == "providers.go" ||
		strings.HasSuffix(normalized, ".go.tmpl")
}

// allowFrameworkOrTests 放行框架源码和测试文件。
func allowFrameworkOrTests(path string) bool {
	return allowFrameworkOrBoot(path) || strings.HasSuffix(path, "_test.go")
}

// allowFrameworkOrGenerated 放行模板、生成代码和框架自身。
func allowFrameworkOrGenerated(path string) bool {
	normalized := filepath.ToSlash(path)
	return allowFrameworkOrBoot(path) ||
		strings.Contains(normalized, "/templates/") ||
		strings.Contains(normalized, "/internal/data/ent/")
}

// mustRegexps 编译规则正则表达式。
func mustRegexps(expressions ...string) []*regexp.Regexp {
	result := make([]*regexp.Regexp, 0, len(expressions))
	for _, expression := range expressions {
		result = append(result, regexp.MustCompile(expression))
	}
	return result
}

// exitWithError 输出错误并使用检查失败状态码退出。
func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "initra 用法检查失败: %v\n", err)
	os.Exit(2)
}
