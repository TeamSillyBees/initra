package templates

import "embed"

// FS 包含 CLI 可用的项目模板文件。
//
//go:embed basic
var FS embed.FS
