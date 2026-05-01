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
	"path/filepath"
	"strings"
	"text/template"

	"github.com/teamsillybees/initra/templates"
)

const frameworkModule = "github.com/teamsillybees/initra"

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
		return fmt.Errorf("用法: initra new <dir>")
	}
	switch args[0] {
	case "new":
		return runNew(args[1:], stdout, cliVersion)
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
	templateName := flags.String("template", "basic", "模板名称")
	frameworkVersion := flags.String("framework-version", "", "initra 框架版本")
	localReplacePath := flags.String("replace", "", "本地 initra 仓库路径，用于 go.mod replace")

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		return fmt.Errorf("new 命令只接受一个目标目录")
	}
	if *templateName != "basic" {
		return fmt.Errorf("暂不支持模板 %q", *templateName)
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
	if err := renderTemplate("basic", targetDir, data); err != nil {
		return err
	}

	if stdout != nil {
		_, _ = fmt.Fprintf(stdout, "created %s\n", targetDir)
	}
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
