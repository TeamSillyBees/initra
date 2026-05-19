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

type rule struct {
	id       string
	severity string
	message  string
	patterns []string
	regexps  []*regexp.Regexp
	allowed  func(path string) bool
}

type finding struct {
	path     string
	line     int
	severity string
	ruleID   string
	message  string
	text     string
}

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
		fmt.Printf("  %s\n", strings.TrimSpace(f.text))
	}
	os.Exit(1)
}

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

func scanFile(path string, rules []rule) ([]finding, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var findings []finding
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
			message:  "生产 Redis 代码禁止使用 KEYS；优先使用 redisx scanner helper 或 SCAN 风格逻辑",
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
			message:  "使用 logging.Register 和注入的 *zap.Logger，不要初始化第二套全局 logger",
			patterns: []string{"zap.NewProduction(", "zap.NewDevelopment(", "zap.Config{"},
			allowed:  allowFrameworkOrBoot,
		},
		{
			id:       "root-internal-import",
			severity: "error",
			message:  "业务项目和模板不得 import initra 根仓库 internal package",
			patterns: []string{"github.com/teamsillybees/initra/internal/"},
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
	}
}

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

func shouldSkipDir(name string) bool {
	switch name {
	case ".agents", ".claude", ".codex", ".git", ".idea", ".vscode", "node_modules", "vendor", "tmp", "var":
		return true
	default:
		return false
	}
}

func allowFrameworkOrBoot(path string) bool {
	normalized := filepath.ToSlash(path)
	return strings.Contains(normalized, "/pkg/") ||
		strings.Contains(normalized, "/internal/boot/") ||
		strings.Contains(normalized, "/tools/skills/")
}

func allowProviderFile(path string) bool {
	normalized := filepath.ToSlash(path)
	base := filepath.Base(path)
	return allowFrameworkOrBoot(path) ||
		base == "providers.go" ||
		strings.HasSuffix(normalized, ".go.tmpl")
}

func allowFrameworkOrTests(path string) bool {
	return allowFrameworkOrBoot(path) || strings.HasSuffix(path, "_test.go")
}

func mustRegexps(expressions ...string) []*regexp.Regexp {
	result := make([]*regexp.Regexp, 0, len(expressions))
	for _, expression := range expressions {
		result = append(result, regexp.MustCompile(expression))
	}
	return result
}

func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "initra 用法检查失败: %v\n", err)
	os.Exit(2)
}
