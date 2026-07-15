# 贡献指南

## 开始之前

`initra` 同时包含可复用 Go package、工程化 CLI、API 模板和可运行示例。修改前先阅读 `AGENTS.md`、相关代码和测试；当文档与可执行代码冲突时，以当前代码和测试为事实源，并在同一改动中修正文档。

安全问题请按 `SECURITY.md` 私密报告，不要先创建公开 Issue 或提交包含漏洞细节的 PR。

## 改动范围

- 保持改动聚焦，不重写无关代码，也不覆盖工作区中他人的未提交修改。
- `pkg/*` 承载可复用运行时能力，`cmd/initra` 承载脚手架命令，根仓库 `internal/*` 只服务本仓库自身。
- `examples` 是 API 模板的可运行事实源；修改示例后同步 `templates/api`，并验证生成后的 Go 代码。
- 业务模块遵循 flat package 结构，不机械增加 controller、repository 或 model 分层。
- 修改 Go 文件后运行 `gofmt`。生产 Go 源码的导出声明应有以标识符开头的中文文档注释；生成代码和测试内部 fake 不要求机械补注释。
- 面向使用者的行为变化、兼容性变化和安全修复应记录到 `CHANGELOG.md` 的“未发布”部分。

## 本地验证

在仓库根目录按影响范围执行：

```powershell
# 根模块
go test ./pkg/... ./cmd/initra/... ./internal/... -count=1
go vet ./pkg/... ./cmd/initra/... ./internal/...

# 示例模块
go test ./examples/... -count=1
go vet ./examples/...

# 模板一致性与 CLI
go run ./tools/sync_api_templates.go --check
go build -o $env:TEMP\initra.exe ./cmd/initra

# 可选的聚合 lint；需要使用支持 Go 1.26 的 golangci-lint v2
golangci-lint run ./...
Push-Location examples
golangci-lint run ./...
Pop-Location
```

只改动窄范围时可以先运行相关 package 测试，但合并前应覆盖所有受影响模块。模板、CLI 或公共 package 的改动还应生成临时项目并运行 `go test`、`go vet` 和 `go build`。

## 提交说明

提交或 PR 说明至少应包含：

- 问题和修改目标；
- 关键设计取舍以及明确未处理的事项；
- 实际运行的验证命令和结果；
- 配置、数据库迁移、API 或生成产物的兼容性影响；
- 安全相关改动的脱敏说明。

不要提交真实凭证、生产数据、编辑器缓存、临时生成目录或本机绝对路径。

## 治理边界

当前仓库没有 `LICENSE`、`CODEOWNERS` 或完整的二进制发布信任链。这些文件不能根据技术偏好自动生成：

- 许可证或内部授权声明必须由权利人确认适用主体、使用范围和第三方义务；
- `CODEOWNERS` 必须由仓库所有者指定真实维护团队和审批责任；
- GoReleaser、checksums、SBOM、签名和 provenance 必须先确定发布平台、产物范围、签名身份与密钥托管方案。

在所有者完成上述决策前，贡献者只记录缺口，不擅自指定许可证、人员、发布账号或签名方案。
