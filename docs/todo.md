- 格式化日志，提供更好的日志输出体验，区别jsonl日志和console日志
- 优化模板项目的 agents.md 和 readme
- 优化模板项目多个子app管理目录结构
- 明确ent事务机制及能力边界
- 重构 Casbin 为从数据库中动态加载权限，并尝试改为权限标识字符串 


- 走查数据结构的字段类型定义，包括接口请求传参、相应参数、dto参数、ent shcema 定义、数据库实体对象定义等，确保符合以下规范：
1. 禁止在 Ent schema 中随意使用 field.Int。
2. ID 字段统一 field.Int64 + GoType(idgen.ID)。
3. 枚举、状态、类型、等级统一 field.Int16。
4. 普通数量、页数、字符数、重试次数、排序号优先 field.Int32。
5. 金额、积分、余额、累计值根据业务上限选择 Int64；默认偏保守用 Int64。
6. Huma DTO 中也禁止裸 int；明确使用 int16 / int32 / idgen.ID。


走查时不要走查 internal/data/ent 下的内容，这个目录下的 ent schema 定义是由 entgen 生成的，禁止修改。数据库层面只走查 internal/data/schema 下的内容，完成修改之后调用 d:\Project\TeamSillyBees\paperlingo-golang\internal\data\generate.go 重新生成 schema 即可。接口层面要走查相关的接口定义、请求参数、响应参数、DTO 定义等，确保所有数值字段都符合上述规范。


- 检查登录之后是否会自动在当前 ctx 中注入用户信息，确保后续业务逻辑能够正确获取用户信息进行权限校验等操作。 避免手动 ctx = entx.WithOperatorID(ctx, update.OperatorID)
- 重构项目结构，删除 model.go 合并 service 和 repo。直接在 service 中通过 ent 方法操作数据库读写