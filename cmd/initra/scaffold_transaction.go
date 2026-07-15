package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type fileChange struct {
	path       string
	content    []byte
	createOnly bool
	mustExist  bool
}

type preparedFileChange struct {
	target    string
	staged    string
	backup    string
	existed   bool
	backedUp  bool
	installed bool
}

// writeModuleDirectoryTransaction 先在临时目录生成完整模块，再一次性提交模块目录。
func writeModuleDirectoryTransaction(moduleDir string, files map[string]string) error {
	if _, err := os.Lstat(moduleDir); err == nil {
		return fmt.Errorf("模块目录 %s 已存在", moduleDir)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	stagedDir, err := os.MkdirTemp(".", ".initra-module-*")
	if err != nil {
		return fmt.Errorf("创建模块临时目录失败: %w", err)
	}
	defer os.RemoveAll(stagedDir)
	if err := os.Chmod(stagedDir, 0o755); err != nil {
		return fmt.Errorf("设置模块临时目录权限失败: %w", err)
	}

	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(stagedDir, name), []byte(files[name]), 0o644); err != nil {
			return fmt.Errorf("写入模块临时文件 %s 失败: %w", name, err)
		}
	}

	if _, err := os.Lstat(moduleDir); err == nil {
		return fmt.Errorf("模块目录 %s 在生成期间已被创建", moduleDir)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	parent := filepath.Dir(moduleDir)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return err
	}
	if err := os.Rename(stagedDir, moduleDir); err != nil {
		return fmt.Errorf("提交模块目录失败: %w", err)
	}
	return nil
}

// addConfigFieldToAggregate 把新能力配置加入 boot.Config，使标准 LoadConfig 能实际加载它。
func addConfigFieldToAggregate(source []byte, capability string) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "config.go", source, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("解析 internal/boot/config.go 失败: %w", err)
	}
	var configStruct *ast.StructType
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.TYPE {
			continue
		}
		for _, spec := range general.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != "Config" {
				continue
			}
			configStruct, ok = typeSpec.Type.(*ast.StructType)
			if !ok {
				return nil, fmt.Errorf("internal/boot.Config 必须是 struct")
			}
		}
	}
	if configStruct == nil {
		return nil, fmt.Errorf("internal/boot/config.go 未找到 Config 聚合结构")
	}

	tagFragment := fmt.Sprintf(`mapstructure:"%s"`, capability)
	fieldName := exportedName(capability)
	for _, field := range configStruct.Fields.List {
		for _, name := range field.Names {
			if name.Name == fieldName {
				return nil, fmt.Errorf("boot.Config 已包含字段 %s", fieldName)
			}
		}
		if field.Tag == nil {
			continue
		}
		tag, unquoteErr := strconv.Unquote(field.Tag.Value)
		mapstructureName := strings.Split(reflect.StructTag(tag).Get("mapstructure"), ",")[0]
		if unquoteErr == nil && mapstructureName == capability {
			return nil, fmt.Errorf("配置能力 %s 已接入 boot.Config", capability)
		}
	}
	configStruct.Fields.List = append(configStruct.Fields.List, &ast.Field{
		Names: []*ast.Ident{ast.NewIdent(fieldName)},
		Type:  ast.NewIdent(fieldName + "Config"),
		Tag: &ast.BasicLit{
			Kind:  token.STRING,
			Value: "`" + tagFragment + "`",
		},
	})

	var output bytes.Buffer
	if err := format.Node(&output, fset, file); err != nil {
		return nil, fmt.Errorf("格式化 internal/boot/config.go 失败: %w", err)
	}
	return output.Bytes(), nil
}

func appendConfigYAML(source []byte, capability string) ([]byte, error) {
	keyPattern := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(capability) + `\s*:`)
	if keyPattern.Match(source) {
		return nil, fmt.Errorf("configs/config.yaml 已包含配置能力 %s", capability)
	}
	trimmed := strings.TrimRight(string(source), "\r\n")
	return []byte(trimmed + "\n\n" + configYAMLTemplate(capability)), nil
}

// applyFileChangesTransaction 对全部目标做预检和暂存，提交失败时恢复原文件。
func applyFileChangesTransaction(changes []fileChange) (returnErr error) {
	if len(changes) == 0 {
		return nil
	}
	stageRoot, err := os.MkdirTemp(".", ".initra-files-*")
	if err != nil {
		return fmt.Errorf("创建文件事务临时目录失败: %w", err)
	}
	preserveStage := false
	defer func() {
		if !preserveStage {
			_ = os.RemoveAll(stageRoot)
		}
	}()

	prepared := make([]preparedFileChange, 0, len(changes))
	seen := make(map[string]struct{}, len(changes))
	for index, change := range changes {
		target, err := filepath.Abs(change.path)
		if err != nil {
			return err
		}
		if _, duplicate := seen[target]; duplicate {
			return fmt.Errorf("文件事务包含重复目标 %s", change.path)
		}
		seen[target] = struct{}{}

		info, statErr := os.Stat(target)
		existed := statErr == nil
		switch {
		case statErr != nil && !errors.Is(statErr, os.ErrNotExist):
			return statErr
		case change.createOnly && existed:
			return fmt.Errorf("文件 %s 已存在", change.path)
		case change.mustExist && !existed:
			return fmt.Errorf("文件 %s 不存在，请在标准 initra 项目根目录执行", change.path)
		}
		mode := os.FileMode(0o644)
		if existed {
			mode = info.Mode().Perm()
		}
		staged := filepath.Join(stageRoot, "new", fmt.Sprintf("%03d", index))
		if err := os.MkdirAll(filepath.Dir(staged), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(staged, change.content, mode); err != nil {
			return err
		}
		prepared = append(prepared, preparedFileChange{
			target:  target,
			staged:  staged,
			backup:  filepath.Join(stageRoot, "backup", fmt.Sprintf("%03d", index)),
			existed: existed,
		})
	}

	defer func() {
		if returnErr == nil {
			return
		}
		if rollbackErr := rollbackFileChanges(prepared); rollbackErr != nil {
			preserveStage = true
			returnErr = errors.Join(
				returnErr,
				fmt.Errorf("回滚文件事务失败，备份保留在 %s: %w", stageRoot, rollbackErr),
			)
		}
	}()
	for index := range prepared {
		change := &prepared[index]
		if err := os.MkdirAll(filepath.Dir(change.target), 0o755); err != nil {
			return err
		}
		if change.existed {
			if err := os.MkdirAll(filepath.Dir(change.backup), 0o755); err != nil {
				return err
			}
			if err := os.Rename(change.target, change.backup); err != nil {
				return fmt.Errorf("备份 %s 失败: %w", change.target, err)
			}
			change.backedUp = true
		}
		if err := os.Rename(change.staged, change.target); err != nil {
			return fmt.Errorf("提交 %s 失败: %w", change.target, err)
		}
		change.installed = true
	}
	return nil
}

func rollbackFileChanges(changes []preparedFileChange) error {
	var rollbackErr error
	for index := len(changes) - 1; index >= 0; index-- {
		change := changes[index]
		if change.installed {
			if err := os.Remove(change.target); err != nil && !errors.Is(err, os.ErrNotExist) {
				rollbackErr = errors.Join(rollbackErr, fmt.Errorf("移除未完成文件 %s 失败: %w", change.target, err))
			}
		}
		if change.backedUp {
			if err := os.Rename(change.backup, change.target); err != nil {
				rollbackErr = errors.Join(rollbackErr, fmt.Errorf("恢复文件 %s 失败: %w", change.target, err))
			}
		}
	}
	return rollbackErr
}
