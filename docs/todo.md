- logx 已区分 console 终端可读日志与 JSONL 文件日志。
- 优化模板项目的 agents.md 和 readme

- 明确ent事务机制及能力边界
- 重构 Casbin 为从数据库中动态加载权限，并尝试改为权限标识字符串 
- 重构 http client 实现和注册机制，当前过于复杂

