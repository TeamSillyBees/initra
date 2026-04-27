[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [ValidatePattern("^[a-z][a-z0-9_]*$")]
    [string]$Name
)

$ErrorActionPreference = "Stop"

# 模块名会直接作为 Go package 名称，因此只允许小写字母、数字和下划线。
$RepoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$ModuleDir = Join-Path $RepoRoot "internal\app\$Name"

$Dirs = @(
    $ModuleDir,
    (Join-Path $ModuleDir "api"),
    (Join-Path $ModuleDir "domain"),
    (Join-Path $ModuleDir "infra")
)

foreach ($Dir in $Dirs) {
    if (-not (Test-Path $Dir)) {
        New-Item -ItemType Directory -Path $Dir | Out-Null
    }
}

@"
package $Name

// TODO: Module 聚合当前业务模块的 HTTP 注册入口。
// 约定：module.go 只负责暴露模块注册能力，不承载业务逻辑。
"@ | Set-Content -Path (Join-Path $ModuleDir "module.go") -Encoding UTF8

@"
package $Name

// TODO: Provide 按 do 注入规则注册当前模块依赖。
// 约定：新增依赖时先确认 service/repository/cache 的职责边界，避免重复注册同一具体类型。
"@ | Set-Content -Path (Join-Path $ModuleDir "wire.go") -Encoding UTF8

@"
package api

// TODO: Handler 只负责协议层 DTO、上下文解析和调用 domain service。
// 约定：不要在 handler 中编写业务分支或直接访问数据库。
"@ | Set-Content -Path (Join-Path $ModuleDir "api\handler.go") -Encoding UTF8

@"
package api

// TODO: 在这里定义请求 DTO。
// 约定：字段标签同时服务 Huma 文档和运行时校验，命名应贴近 API 语义。
"@ | Set-Content -Path (Join-Path $ModuleDir "api\request.go") -Encoding UTF8

@"
package api

// TODO: 在这里定义响应 DTO。
// 约定：响应对象保持稳定，避免直接暴露数据库模型。
"@ | Set-Content -Path (Join-Path $ModuleDir "api\response.go") -Encoding UTF8

@"
package domain

// TODO: 在这里定义领域实体和值对象。
// 约定：领域对象表达业务含义，不绑定 HTTP、SQL 或缓存实现。
"@ | Set-Content -Path (Join-Path $ModuleDir "domain\entity.go") -Encoding UTF8

@"
package domain

// TODO: Service 承载当前模块的业务编排。
// 约定：service 不接收 gin.Context，外部依赖通过接口注入，便于单元测试。
"@ | Set-Content -Path (Join-Path $ModuleDir "domain\service.go") -Encoding UTF8

@"
package domain

// TODO: 在这里定义当前模块的业务错误。
// 约定：错误最终由 internal/platform/errors 统一转换为 API 响应。
"@ | Set-Content -Path (Join-Path $ModuleDir "domain\errors.go") -Encoding UTF8

@"
package infra

// TODO: Repository 只负责持久化访问。
// 约定：SQL 通过 internal/gen/jet 生成代码构造，不手写字符串拼接 SQL。
"@ | Set-Content -Path (Join-Path $ModuleDir "infra\repository.go") -Encoding UTF8

@"
package infra

// TODO: 在这里封装当前模块缓存访问。
// 约定：缓存键、TTL 和失效策略集中管理，避免散落在 service 中。
"@ | Set-Content -Path (Join-Path $ModuleDir "infra\cache.go") -Encoding UTF8

@"
package infra

// TODO: 在这里定义当前模块鉴权资源常量。
// 约定：资源命名应和 Casbin policy 保持一致，避免魔法字符串散落在 handler 中。
"@ | Set-Content -Path (Join-Path $ModuleDir "infra\policy.go") -Encoding UTF8

Write-Host "created module skeleton: $ModuleDir"
