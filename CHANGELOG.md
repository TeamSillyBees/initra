# 变更记录

本文件记录面向使用者和维护者的重要变化。新增条目先写入“未发布”，正式发布时再由维护者确定版本号、发布日期和迁移说明；本文件不自行承诺语义化版本兼容规则。

## 未发布

### 新增

- 增加安全报告、贡献流程和变更记录基础文档。
- 增加 golangci-lint v2 基础配置。
- 补齐存量导出声明注释，并增加基于 Go AST 的架构测试，阻止新增不规范注释。
- CLI 增加 `snippet add`、可加载的事务化 `config add`、`doctor --json`，以及可检查和安全升级的 Codex skill。
- API 模板增加随机 JWT secret、一次性管理员密码、安全 `app.slug`、数据库外键和 nonroot 容器基线。

### 变更

- 移除语义失真的 `crud add` 和未生效的 `allow_multiple_devices` 配置。
- 配置加载改为拒绝未知字段；JWT 要求至少 32 字节密钥并拒绝无效用户 ID。
- refresh token 改为查询用户成功后原子轮转；Redis token store 初始化或访问异常时 fail-closed。
- HTTP Client 默认不记录错误响应正文，5xx 全量重试改为显式 opt-in。
- local storage 不再返回不会过期的伪 presign URL，改为明确返回不支持。
- `logx.Logger.Sync` 只负责刷新，资源关闭统一使用共享且幂等的 `Close`/`Shutdown`。

### 修复

- 修复本地上传 overwrite 竞态、代理头无条件信任、Redis SCAN glob 扩散、数据库 Ping 无超时和结构化日志脱敏不完整等问题。
- 修复 Windows 空格 replace、模块/配置半写入、模板换行假漂移和 checker 漏报根模块 `internal` import 等脚手架问题。

## 历史版本

仓库已经存在 `v0.1.0` 至 `v0.1.18` tag，但这些版本早于本变更记录建立。为避免根据提交标题猜测用户影响，此处不回填未经核实的逐版本条目；需要追溯时以对应 tag、提交历史和当时的可执行代码为准。

## 待所有者决定

- 版本兼容承诺和长期支持范围；
- 正式 release note 的审批人和发布渠道；
- GoReleaser、checksums、SBOM、签名及 provenance 的实现方案。
