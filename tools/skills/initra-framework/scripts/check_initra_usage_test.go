package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRootInternalImportFindingsUsesGoAST(t *testing.T) {
	for _, relative := range []string{
		filepath.Join("pkg", "feature", "providers.go"),
		filepath.Join("internal", "boot", "providers.go"),
	} {
		t.Run(filepath.ToSlash(relative), func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, relative)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			content := `package feature

import bootinternal "github.com/teamsillybees/initra/internal/boot"

var _ = bootinternal.Start
`
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}

			findings, err := scanFile(path, defaultRules())
			if err != nil {
				t.Fatal(err)
			}
			if len(findings) != 1 {
				t.Fatalf("findings = %v, want one root internal import finding", findings)
			}
			if findings[0].ruleID != "root-internal-import" || findings[0].line != 3 {
				t.Fatalf("finding = %+v", findings[0])
			}
		})
	}
}

func TestRootInternalImportFindingsIgnoresCommentsAndStrings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "service.go")
	content := `package service

// github.com/teamsillybees/initra/internal/boot is forbidden.
const example = "github.com/teamsillybees/initra/internal/data"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	findings, err := rootInternalImportFindings(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %v, want none", findings)
	}
}
