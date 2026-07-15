package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	skillassets "github.com/teamsillybees/initra/tools/skills"
)

const (
	skillSourceRoot   = "initra-framework"
	skillManifestName = ".initra-manifest.json"
)

type skillOptions struct {
	check bool
	force bool
}

type skillManifest struct {
	SchemaVersion int               `json:"schema_version"`
	Files         map[string]string `json:"files"`
}

func newSkillCommand(stdout io.Writer) *cobra.Command {
	opts := skillOptions{}
	cmd := &cobra.Command{
		Use:           "skill",
		Short:         "初始化 initra 框架 skill 文档",
		Long:          "在当前项目初始化 initra 框架相关的 skill 文档，写入 Codex 的 .agents/skills/initra-framework。",
		Example:       "  initra skill\n  initra skill --check\n  initra skill --force",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          requireNoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return syncFrameworkSkill(filepath.Join(".agents", "skills", "initra-framework"), opts, cmd.OutOrStdout())
		},
	}
	configureCommand(cmd, stdout)
	flags := cmd.PersistentFlags()
	flags.BoolVar(&opts.check, "check", false, "只检查 skill 是否为内置最新版本，不修改文件")
	flags.BoolVar(&opts.force, "force", false, "覆盖已修改的内置 skill 文件")
	cmd.AddCommand(newSkillCodexCommand(stdout, &opts))
	return cmd
}

func newSkillCodexCommand(stdout io.Writer, opts *skillOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "codex",
		Short:         "添加 Codex skill 文档",
		Long:          "在当前项目写入 .agents/skills/initra-framework，供 Codex 理解并检查 initra 项目约束。",
		Example:       "  initra skill codex",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          requireNoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return syncFrameworkSkill(filepath.Join(".agents", "skills", "initra-framework"), *opts, cmd.OutOrStdout())
		},
	}
	configureCommand(cmd, stdout)
	return cmd
}

func syncFrameworkSkill(targetRoot string, opts skillOptions, stdout io.Writer) error {
	if opts.check && opts.force {
		return fmt.Errorf("--check 与 --force 不能同时使用")
	}
	sourceFiles, err := embeddedSkillFiles()
	if err != nil {
		return err
	}
	currentManifest := buildSkillManifest(sourceFiles)
	state, err := inspectSkillTarget(targetRoot, sourceFiles, currentManifest)
	if err != nil {
		return err
	}
	if opts.check {
		if len(state.drift) > 0 {
			return fmt.Errorf("skill 需要更新: %s", strings.Join(state.drift, ", "))
		}
		if stdout != nil {
			_, _ = fmt.Fprintf(stdout, "skill %s is up to date\n", targetRoot)
		}
		return nil
	}
	if len(state.drift) == 0 && !opts.force {
		if stdout != nil {
			_, _ = fmt.Fprintf(stdout, "skill %s is up to date\n", targetRoot)
		}
		return nil
	}
	if state.exists && !opts.force && !state.safeToUpgrade(sourceFiles) {
		return fmt.Errorf("skill 包含本地修改，已停止更新；确认覆盖时使用 --force")
	}
	if err := replaceSkillDirectory(targetRoot, sourceFiles, currentManifest, state.previousManifest); err != nil {
		return err
	}
	if stdout != nil {
		action := "updated"
		if !state.exists {
			action = "created"
		}
		_, _ = fmt.Fprintf(stdout, "%s skill %s\n", action, targetRoot)
	}
	return nil
}

type skillTargetState struct {
	exists           bool
	drift            []string
	previousManifest *skillManifest
	manifestValid    bool
	targetRoot       string
}

func (s skillTargetState) safeToUpgrade(sourceFiles map[string][]byte) bool {
	if !s.exists {
		return true
	}
	if s.manifestValid && s.previousManifest != nil {
		for relativePath, expectedHash := range s.previousManifest.Files {
			target, err := skillFilePath(s.targetRoot, relativePath)
			if err != nil {
				return false
			}
			content, err := os.ReadFile(target)
			if err != nil || hashSkillContent(content) != expectedHash {
				return false
			}
		}
		// 新版本可能新增文件。若同名路径原本是用户自己的扩展，默认升级不能覆盖它。
		for relativePath, expected := range sourceFiles {
			if _, managed := s.previousManifest.Files[relativePath]; managed {
				continue
			}
			target, err := skillFilePath(s.targetRoot, relativePath)
			if err != nil {
				return false
			}
			content, err := os.ReadFile(target)
			switch {
			case errors.Is(err, os.ErrNotExist):
				continue
			case err != nil, !bytes.Equal(content, expected):
				return false
			}
		}
		return true
	}
	// 兼容旧版无 manifest 安装：所有当前内置文件均未被修改时可安全补清单。
	for path, expected := range sourceFiles {
		content, err := os.ReadFile(filepath.Join(s.targetRoot, filepath.FromSlash(path)))
		if err != nil || !bytes.Equal(content, expected) {
			return false
		}
	}
	return true
}

func embeddedSkillFiles() (map[string][]byte, error) {
	files := make(map[string][]byte)
	err := fs.WalkDir(skillassets.FS, skillSourceRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}
		relative, err := filepath.Rel(skillSourceRoot, filepath.ToSlash(path))
		if err != nil {
			return err
		}
		content, err := skillassets.FS.ReadFile(path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(relative)] = content
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func buildSkillManifest(files map[string][]byte) skillManifest {
	hashes := make(map[string]string, len(files))
	for path, content := range files {
		hashes[path] = hashSkillContent(content)
	}
	return skillManifest{SchemaVersion: 1, Files: hashes}
}

func hashSkillContent(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func inspectSkillTarget(targetRoot string, sourceFiles map[string][]byte, current skillManifest) (skillTargetState, error) {
	state := skillTargetState{targetRoot: targetRoot}
	info, err := os.Stat(targetRoot)
	switch {
	case errors.Is(err, os.ErrNotExist):
		state.drift = []string{"skill directory missing"}
		return state, nil
	case err != nil:
		return state, err
	case !info.IsDir():
		return state, fmt.Errorf("skill 目标 %s 不是目录", targetRoot)
	}
	state.exists = true

	manifestContent, manifestErr := os.ReadFile(filepath.Join(targetRoot, skillManifestName))
	if manifestErr == nil {
		var previous skillManifest
		if err := json.Unmarshal(manifestContent, &previous); err != nil {
			state.drift = append(state.drift, "invalid manifest")
		} else if err := validateSkillManifest(previous); err != nil {
			return state, fmt.Errorf("skill manifest 无效: %w", err)
		} else {
			state.previousManifest = &previous
			state.manifestValid = true
			if !equalSkillManifest(previous, current) {
				state.drift = append(state.drift, "manifest version")
			}
		}
	} else if errors.Is(manifestErr, os.ErrNotExist) {
		state.drift = append(state.drift, "manifest missing")
	} else {
		return state, manifestErr
	}

	paths := make([]string, 0, len(sourceFiles))
	for path := range sourceFiles {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		content, err := os.ReadFile(filepath.Join(targetRoot, filepath.FromSlash(path)))
		switch {
		case errors.Is(err, os.ErrNotExist):
			state.drift = append(state.drift, path+" missing")
		case err != nil:
			return state, err
		case !bytes.Equal(content, sourceFiles[path]):
			state.drift = append(state.drift, path+" changed")
		}
	}
	sort.Strings(state.drift)
	return state, nil
}

func equalSkillManifest(left skillManifest, right skillManifest) bool {
	if left.SchemaVersion != right.SchemaVersion || len(left.Files) != len(right.Files) {
		return false
	}
	for path, hash := range left.Files {
		if right.Files[path] != hash {
			return false
		}
	}
	return true
}

func validateSkillManifest(manifest skillManifest) error {
	if manifest.SchemaVersion != 1 {
		return fmt.Errorf("不支持 schema_version %d", manifest.SchemaVersion)
	}
	if manifest.Files == nil {
		return fmt.Errorf("files 不能为 null")
	}
	for relativePath, hash := range manifest.Files {
		if _, err := skillFilePath(".", relativePath); err != nil {
			return err
		}
		decoded, err := hex.DecodeString(hash)
		if err != nil || len(decoded) != sha256.Size {
			return fmt.Errorf("文件 %q 的 SHA-256 无效", relativePath)
		}
	}
	return nil
}

func skillFilePath(root string, relativePath string) (string, error) {
	if relativePath == "." || !fs.ValidPath(relativePath) || strings.Contains(relativePath, `\`) {
		return "", fmt.Errorf("manifest 路径 %q 无效", relativePath)
	}
	nativePath := filepath.FromSlash(relativePath)
	if filepath.IsAbs(nativePath) || filepath.VolumeName(nativePath) != "" {
		return "", fmt.Errorf("manifest 路径 %q 必须是相对路径", relativePath)
	}
	return filepath.Join(root, nativePath), nil
}

func replaceSkillDirectory(targetRoot string, sourceFiles map[string][]byte, current skillManifest, previous *skillManifest) (returnErr error) {
	if previous != nil {
		if err := validateSkillManifest(*previous); err != nil {
			return fmt.Errorf("旧 skill manifest 无效: %w", err)
		}
	}
	parent := filepath.Dir(targetRoot)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return err
	}
	stageRoot, err := os.MkdirTemp(parent, ".initra-skill-*")
	if err != nil {
		return fmt.Errorf("创建 skill 临时目录失败: %w", err)
	}
	preserveStage := false
	defer func() {
		if !preserveStage {
			_ = os.RemoveAll(stageRoot)
		}
	}()
	staged := filepath.Join(stageRoot, "skill")
	if _, err := os.Stat(targetRoot); err == nil {
		if err := copyDirectory(targetRoot, staged); err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(staged, 0o755); err != nil {
		return err
	}
	if previous != nil {
		for oldPath := range previous.Files {
			if _, keep := sourceFiles[oldPath]; !keep {
				target, err := skillFilePath(staged, oldPath)
				if err != nil {
					return err
				}
				if err := os.Remove(target); err != nil && !errors.Is(err, os.ErrNotExist) {
					return err
				}
			}
		}
	}
	for path, content := range sourceFiles {
		target := filepath.Join(staged, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, content, 0o644); err != nil {
			return err
		}
	}
	manifestContent, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		return err
	}
	manifestContent = append(manifestContent, '\n')
	if err := os.WriteFile(filepath.Join(staged, skillManifestName), manifestContent, 0o644); err != nil {
		return err
	}

	backup := filepath.Join(stageRoot, "backup")
	existed := false
	if _, err := os.Stat(targetRoot); err == nil {
		existed = true
		if err := os.Rename(targetRoot, backup); err != nil {
			return fmt.Errorf("备份现有 skill 失败: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	defer func() {
		if returnErr == nil || !existed {
			return
		}
		if err := os.RemoveAll(targetRoot); err != nil {
			preserveStage = true
			returnErr = errors.Join(returnErr, fmt.Errorf("清理未完成 skill 失败，原 skill 备份保留在 %s: %w", backup, err))
			return
		}
		if err := os.Rename(backup, targetRoot); err != nil {
			preserveStage = true
			returnErr = errors.Join(returnErr, fmt.Errorf("恢复原 skill 失败，备份保留在 %s: %w", backup, err))
		}
	}()
	if err := os.Rename(staged, targetRoot); err != nil {
		return fmt.Errorf("提交 skill 更新失败: %w", err)
	}
	return nil
}

func copyDirectory(source string, target string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		destination := filepath.Join(target, relative)
		if entry.IsDir() {
			return os.MkdirAll(destination, 0o755)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(destination, content, 0o644)
	})
}
