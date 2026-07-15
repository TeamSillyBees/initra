package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/mod/modfile"
)

func TestNormalizeAppSlug(t *testing.T) {
	tests := map[string]string{
		"My API_Service": "my-api-service",
		"123 Console":    "app-123-console",
		"demo---api":     "demo-api",
		" demo/服务.api ":  "demo-api",
	}
	for input, expected := range tests {
		t.Run(input, func(t *testing.T) {
			actual, err := normalizeAppSlug(input)
			require.NoError(t, err)
			require.Equal(t, expected, actual)
			require.LessOrEqual(t, len(actual), maxAppSlugLength)
		})
	}
	_, err := normalizeAppSlug("中文应用")
	require.Error(t, err)
}

func TestProjectModuleAndFrameworkVersionValidation(t *testing.T) {
	modulePath, err := normalizeProjectModulePath("", "demo-api")
	require.NoError(t, err)
	require.Equal(t, "demo-api", modulePath)

	_, err = normalizeProjectModulePath("not a module", "demo-api")
	require.ErrorContains(t, err, "module path")
	_, err = resolveFrameworkVersion("latest", "dev", "")
	require.ErrorContains(t, err, "框架版本")
}

func TestGenerateTemplateSecretsUsesIndependentURLSafeValues(t *testing.T) {
	secrets, err := generateTemplateSecrets()
	require.NoError(t, err)

	jwtValues := []string{secrets.local, secrets.dev, secrets.test}
	require.Len(t, map[string]struct{}{jwtValues[0]: {}, jwtValues[1]: {}, jwtValues[2]: {}}, 3)
	for _, value := range jwtValues {
		decoded, decodeErr := base64.RawURLEncoding.DecodeString(value)
		require.NoError(t, decodeErr)
		require.GreaterOrEqual(t, len(decoded), generatedSecretBytes)
	}
	require.GreaterOrEqual(t, len(secrets.adminPassword), 20)
	_, err = base64.RawURLEncoding.DecodeString(secrets.adminPassword)
	require.NoError(t, err)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(secrets.adminPasswordHash), []byte(secrets.adminPassword)))
}

func TestGenerateURLSafeSecretPropagatesRandomSourceFailure(t *testing.T) {
	_, err := generateURLSafeSecret(errorReader{})
	require.ErrorContains(t, err, "random source failed")
	_, err = generateTemplateSecretsFrom(errorReader{})
	require.ErrorContains(t, err, "生成 local JWT secret 失败")
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) {
	return 0, errors.New("random source failed")
}

func TestFrameworkReplaceUsesModfileForPathWithSpaces(t *testing.T) {
	workspace := t.TempDir()
	frameworkDir := filepath.Join(workspace, "framework checkout with spaces")
	require.NoError(t, os.MkdirAll(frameworkDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(frameworkDir, "go.mod"), []byte("module "+frameworkModule+"\n\ngo 1.26.0\n"), 0o644))

	replacePath, err := normalizeReplacePath(frameworkDir)
	require.NoError(t, err)
	projectGoMod := filepath.Join(workspace, "go.mod")
	require.NoError(t, os.WriteFile(projectGoMod, []byte("module example.com/demo\n\ngo 1.26.0\n\nrequire "+frameworkModule+" v0.0.0\n"), 0o644))
	require.NoError(t, applyFrameworkReplace(projectGoMod, replacePath))

	content, err := os.ReadFile(projectGoMod)
	require.NoError(t, err)
	parsed, err := modfile.Parse(projectGoMod, content, nil)
	require.NoError(t, err)
	require.Len(t, parsed.Replace, 1)
	require.Equal(t, frameworkModule, parsed.Replace[0].Old.Path)
	require.Equal(t, filepath.ToSlash(frameworkDir), parsed.Replace[0].New.Path)
	require.Contains(t, string(content), `"`+filepath.ToSlash(frameworkDir)+`"`)
}

func TestNormalizeReplacePathRejectsUnexpectedModule(t *testing.T) {
	target := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(target, "go.mod"), []byte("module example.com/not-initra\n"), 0o644))

	_, err := normalizeReplacePath(target)

	require.ErrorContains(t, err, "replace 目标模块必须是")
}

func TestCreateProjectWritesSpacedReplaceThroughModfile(t *testing.T) {
	workspace := t.TempDir()
	frameworkDir := filepath.Join(workspace, "initra checkout with spaces")
	require.NoError(t, os.MkdirAll(frameworkDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(frameworkDir, "go.mod"), []byte("module "+frameworkModule+"\n\ngo 1.26.0\n"), 0o644))
	target := filepath.Join(workspace, "demo")
	runner := func(dir string, name string, args ...string) ([]byte, error) {
		if name == "git" {
			require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0o755))
		}
		return nil, nil
	}

	err := createProjectWithRunner(target, ioDiscard{}, "dev", newOptions{
		modulePath:       "example.com/demo",
		projectType:      "api",
		localReplacePath: frameworkDir,
	}, runner)
	require.NoError(t, err)
	goModPath := filepath.Join(target, "go.mod")
	content, err := os.ReadFile(goModPath)
	require.NoError(t, err)
	parsed, err := modfile.Parse(goModPath, content, nil)
	require.NoError(t, err)
	require.Len(t, parsed.Replace, 1)
	require.Equal(t, filepath.ToSlash(frameworkDir), parsed.Replace[0].New.Path)
}

func TestCreateProjectPrintsOneTimeAdminPasswordWithoutWritingPlaintext(t *testing.T) {
	target := filepath.Join(t.TempDir(), "demo")
	runner := func(dir string, name string, args ...string) ([]byte, error) {
		if name == "git" {
			require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0o755))
		}
		return nil, nil
	}
	var stdout bytes.Buffer
	require.NoError(t, createProjectWithRunner(target, &stdout, "v1.2.3", newOptions{
		modulePath:  "example.com/demo",
		projectType: "api",
	}, runner))

	passwordMatch := regexp.MustCompile(`initial admin password \(shown once\): ([A-Za-z0-9_-]{20,})`).FindStringSubmatch(stdout.String())
	require.Len(t, passwordMatch, 2, stdout.String())
	require.Equal(t, 1, strings.Count(stdout.String(), "initial admin password (shown once):"))
	password := passwordMatch[1]
	seedPath := filepath.Join(target, "db", "seeds", "001_seed_admin.sql")
	seed := readFile(t, seedPath)
	require.NotContains(t, seed, "admin123")
	hashMatch := regexp.MustCompile(`\$2[aby]\$[0-9]{2}\$[A-Za-z0-9./]{53}`).FindString(seed)
	require.NotEmpty(t, hashMatch)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(hashMatch), []byte(password)))

	require.NoError(t, filepath.WalkDir(target, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		require.NotContains(t, string(content), password, path)
		return nil
	}))
}

func TestCreateProjectDoesNotPrintAdminPasswordBeforeSuccessfulCommit(t *testing.T) {
	target := filepath.Join(t.TempDir(), "demo")
	var stdout bytes.Buffer
	runner := func(string, string, ...string) ([]byte, error) {
		return []byte("failed"), errors.New("command failed")
	}

	err := createProjectWithRunner(target, &stdout, "v1.2.3", newOptions{
		modulePath:  "example.com/demo",
		projectType: "api",
	}, runner)

	require.Error(t, err)
	require.NotContains(t, stdout.String(), "admin password")
	require.NoDirExists(t, target)
}

func TestCreateProjectReportsOneTimePasswordOutputFailureAfterCommit(t *testing.T) {
	target := filepath.Join(t.TempDir(), "demo")
	writeErr := errors.New("broken output")
	err := createProjectWithRunner(target, errorWriter{err: writeErr}, "dev", newOptions{
		modulePath:       "example.com/demo",
		localReplacePath: repoRoot(t),
	}, func(string, string, ...string) ([]byte, error) {
		return nil, nil
	})

	require.ErrorIs(t, err, writeErr)
	require.ErrorContains(t, err, "项目已创建")
	require.FileExists(t, filepath.Join(target, "go.mod"))
}

func TestCreateProjectRendersDisplayNameAndSafeAppSlug(t *testing.T) {
	target := filepath.Join(t.TempDir(), "demo")
	runner := func(dir string, name string, args ...string) ([]byte, error) {
		if name == "git" {
			require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0o755))
		}
		return nil, nil
	}
	require.NoError(t, createProjectWithRunner(target, ioDiscard{}, "v1.2.3", newOptions{
		modulePath:  "example.com/demo",
		appName:     "Acme Billing API",
		projectType: "api",
	}, runner))

	config := readFile(t, filepath.Join(target, "configs", "config.yaml"))
	require.Contains(t, config, `name: "Acme Billing API"`)
	require.Contains(t, config, "slug: acme-billing-api")
	require.Contains(t, config, `dbname: "acme-billing-api"`)
	compose := readFile(t, filepath.Join(target, "docker-compose.yml"))
	require.Contains(t, compose, "name: acme-billing-api")
}

func TestModuleAddRejectsExistingDirectoryWithoutPartialFiles(t *testing.T) {
	target := t.TempDir()
	moduleDir := filepath.Join(target, "internal", "modules", "order")
	require.NoError(t, os.MkdirAll(moduleDir, 0o755))
	marker := filepath.Join(moduleDir, "keep.txt")
	require.NoError(t, os.WriteFile(marker, []byte("keep"), 0o644))
	t.Chdir(target)

	err := run([]string{"module", "add", "order"}, ioDiscard{}, "dev")

	require.Error(t, err)
	entries, readErr := os.ReadDir(moduleDir)
	require.NoError(t, readErr)
	require.Len(t, entries, 1)
	require.Equal(t, "keep.txt", entries[0].Name())
}

func TestConfigAddPreflightKeepsAggregateUnchangedOnConflict(t *testing.T) {
	target := prepareMinimalConfigProject(t)
	generatedPath := filepath.Join(target, "internal", "boot", "feature_flag.config.go")
	require.NoError(t, os.WriteFile(generatedPath, []byte("package boot\n// user file\n"), 0o644))
	configPath := filepath.Join(target, "internal", "boot", "config.go")
	yamlPath := filepath.Join(target, "configs", "config.yaml")
	beforeConfig := readFile(t, configPath)
	beforeYAML := readFile(t, yamlPath)
	t.Chdir(target)

	err := run([]string{"config", "add", "feature_flag"}, ioDiscard{}, "dev")

	require.Error(t, err)
	require.Equal(t, beforeConfig, readFile(t, configPath))
	require.Equal(t, beforeYAML, readFile(t, yamlPath))
	require.Contains(t, readFile(t, generatedPath), "user file")
}

func TestConfigAddGeneratedFilesCompile(t *testing.T) {
	target := prepareMinimalConfigProject(t)
	require.NoError(t, os.WriteFile(filepath.Join(target, "go.mod"), []byte("module example.com/demo\n\ngo 1.26.0\n"), 0o644))
	t.Chdir(target)
	require.NoError(t, run([]string{"config", "add", "feature_flag"}, ioDiscard{}, "dev"))

	command := exec.Command("go", "test", "./internal/boot")
	command.Dir = target
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	require.NoError(t, err, string(output))
}

func prepareMinimalConfigProject(t *testing.T) string {
	t.Helper()
	target := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(target, "internal", "boot"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(target, "configs"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(target, "internal", "boot", "config.go"), []byte("package boot\n\ntype Config struct{}\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(target, "configs", "config.yaml"), []byte("app:\n  name: demo\n"), 0o644))
	return target
}

func TestDoctorJSONIsStableAndRequiredFailureReturnsError(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("PATH", t.TempDir())
	var stdout bytes.Buffer

	err := run([]string{"doctor", "--json"}, &stdout, "dev")

	require.ErrorContains(t, err, "必需项不可用")
	var report doctorReport
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &report))
	require.Equal(t, 1, report.SchemaVersion)
	require.False(t, report.OK)
	names := make([]string, 0, len(report.Checks))
	for _, check := range report.Checks {
		names = append(names, check.Name)
	}
	require.Equal(t, []string{"Go", "Atlas", "Ent", "golangci-lint", "config.yaml", "config.dev.yaml", "Atlas config"}, names)
	require.Equal(t, doctorStatusMissing, report.Checks[0].Status)
	require.Equal(t, doctorStatusOptional, report.Checks[4].Status)
}

func TestDoctorToolCheckHasDeadline(t *testing.T) {
	check := checkDoctorToolWithTimeout("Go", true, 0, "go", "version")
	require.Equal(t, doctorStatusBroken, check.Status)
	require.Contains(t, check.Detail, "timeout")
}

func TestSkillSupportsCheckRepeatAndForceWhilePreservingExtraFiles(t *testing.T) {
	target := t.TempDir()
	t.Chdir(target)
	var stdout bytes.Buffer
	require.NoError(t, run([]string{"skill"}, &stdout, "dev"))
	skillDir := filepath.Join(target, ".agents", "skills", "initra-framework")
	require.FileExists(t, filepath.Join(skillDir, skillManifestName))
	require.NoFileExists(t, filepath.Join(skillDir, "scripts", "check_initra_usage_test.go"))

	stdout.Reset()
	require.NoError(t, run([]string{"skill", "--check"}, &stdout, "dev"))
	require.Contains(t, stdout.String(), "up to date")
	require.NoError(t, run([]string{"skill", "codex", "--check"}, ioDiscard{}, "dev"))
	stdout.Reset()
	require.NoError(t, run([]string{"skill"}, &stdout, "dev"))
	require.Contains(t, stdout.String(), "up to date")

	skillPath := filepath.Join(skillDir, "SKILL.md")
	require.NoError(t, os.WriteFile(skillPath, []byte("local edit\n"), 0o644))
	extraPath := filepath.Join(skillDir, "notes.local.md")
	require.NoError(t, os.WriteFile(extraPath, []byte("keep\n"), 0o644))
	require.Error(t, run([]string{"skill", "--check"}, ioDiscard{}, "dev"))
	require.ErrorContains(t, run([]string{"skill"}, ioDiscard{}, "dev"), "本地修改")
	require.NoError(t, run([]string{"skill", "--force"}, ioDiscard{}, "dev"))
	require.NotEqual(t, "local edit\n", readFile(t, skillPath))
	require.Equal(t, "keep\n", readFile(t, extraPath))
}

func TestSkillUpgradesLegacyExactCopyByAddingManifest(t *testing.T) {
	target := t.TempDir()
	t.Chdir(target)
	require.NoError(t, run([]string{"skill"}, ioDiscard{}, "dev"))
	manifestPath := filepath.Join(target, ".agents", "skills", "initra-framework", skillManifestName)
	require.NoError(t, os.Remove(manifestPath))

	require.NoError(t, run([]string{"skill"}, ioDiscard{}, "dev"))
	require.FileExists(t, manifestPath)
}

func TestSkillRejectsManifestPathTraversalEvenWithForce(t *testing.T) {
	target := t.TempDir()
	t.Chdir(target)
	require.NoError(t, run([]string{"skill"}, ioDiscard{}, "dev"))

	skillDir := filepath.Join(target, ".agents", "skills", "initra-framework")
	manifestPath := filepath.Join(skillDir, skillManifestName)
	var manifest skillManifest
	require.NoError(t, json.Unmarshal([]byte(readFile(t, manifestPath)), &manifest))
	outsidePath := filepath.Join(filepath.Dir(skillDir), "outside.txt")
	require.NoError(t, os.WriteFile(outsidePath, []byte("keep\n"), 0o644))
	manifest.Files["../outside.txt"] = hashSkillContent([]byte("keep\n"))
	content, err := json.Marshal(manifest)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(manifestPath, content, 0o644))

	err = run([]string{"skill", "--force"}, ioDiscard{}, "dev")
	require.ErrorContains(t, err, "manifest 路径")
	require.Equal(t, "keep\n", readFile(t, outsidePath))
}

func TestSkillDoesNotOverwriteNewManagedPathThatWasUserOwned(t *testing.T) {
	target := t.TempDir()
	t.Chdir(target)
	require.NoError(t, run([]string{"skill"}, ioDiscard{}, "dev"))

	skillDir := filepath.Join(target, ".agents", "skills", "initra-framework")
	manifestPath := filepath.Join(skillDir, skillManifestName)
	var manifest skillManifest
	require.NoError(t, json.Unmarshal([]byte(readFile(t, manifestPath)), &manifest))
	const newlyManagedPath = "scripts/check_initra_usage.go"
	delete(manifest.Files, newlyManagedPath)
	content, err := json.Marshal(manifest)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(manifestPath, content, 0o644))
	localPath := filepath.Join(skillDir, filepath.FromSlash(newlyManagedPath))
	require.NoError(t, os.WriteFile(localPath, []byte("user-owned extension\n"), 0o644))

	err = run([]string{"skill"}, ioDiscard{}, "dev")
	require.ErrorContains(t, err, "本地修改")
	require.Equal(t, "user-owned extension\n", readFile(t, localPath))
}

func TestSkillRejectsCheckAndForceTogether(t *testing.T) {
	t.Chdir(t.TempDir())
	err := run([]string{"skill", "--check", "--force"}, ioDiscard{}, "dev")
	require.ErrorContains(t, err, "不能同时使用")
}

func TestRootHelpPublishesSnippetInsteadOfCRUD(t *testing.T) {
	var stdout bytes.Buffer
	require.NoError(t, run([]string{"--help"}, &stdout, "dev"))
	require.Contains(t, stdout.String(), "snippet")
	require.NotContains(t, stdout.String(), "\n  crud ")
	require.ErrorContains(t, run([]string{"crud"}, ioDiscard{}, "dev"), "unknown command")
}

type errorWriter struct {
	err error
}

func (w errorWriter) Write([]byte) (int, error) {
	return 0, w.err
}
