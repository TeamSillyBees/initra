package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

const entPackage = "github.com/teamsillybees/initra/examples/internal/data/ent"

// main 生成 Ent 代码，并把手写 schema 与生成代码分开放置。
func main() {
	root := dataDir()
	target := filepath.Join(root, "ent")
	if err := os.MkdirAll(target, 0o755); err != nil {
		log.Fatalf("create ent target dir: %v", err)
	}
	err := entc.Generate(
		filepath.Join(root, "schema"),
		&gen.Config{
			Target:  target,
			Package: entPackage,
		},
		entc.FeatureNames("sql/versioned-migration"),
	)
	if err != nil {
		log.Fatalf("generate ent code: %v", err)
	}
}

// dataDir 返回 internal/data 目录。
func dataDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("resolve entgenerate source path")
	}
	return filepath.Dir(filepath.Dir(file))
}
