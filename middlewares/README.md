# Chi 框架中间件

本目录包含了为 Chi 框架开发的常用中间件，包括跨域（CORS）和限流（Rate Limiting）功能。

## 目录结构

```
middlewares/
├── cors.go         # CORS跨域中间件
├── ratelimit.go    # 限流中间件
├── recovery.go     # Panic恢复中间件
├── example.go      # 使用示例
└── README.md       # 说明文档
```

## CORS 跨域中间件

### 功能特性

- 支持自定义允许的源（Origin）
- 支持自定义允许的HTTP方法
- 支持自定义允许的请求头
- 支持凭据（Credentials）传递
- 支持预检请求（Preflight）处理
- 提供开发和生产环境的预设配置

### 基本使用

```go
package main

import (
    "chi"
    "chi/middlewares"
)

func main() {
    server := chi.New()
    
    // 使用默认CORS配置
    server.Use(middlewares.CORS())
    
    // 或使用开发环境配置（允许所有源）
    server.Use(middlewares.CORSForDevelopment())
    
    // 或使用生产环境配置（指定允许的源）
    server.Use(middlewares.CORSForProduction([]string{
        "https://example.com",
        "https://app.example.com",
    }))
    
    server.Run(":8080")
}
```

### 自定义配置

```go
server.Use(middlewares.CORSWithConfig(middlewares.CORSConfig{
    AllowOrigins: []string{"https://trusted-domain.com"},
    AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders: []string{"Content-Type", "Authorization"},
    AllowCredentials: true,
    MaxAge: 1 * time.Hour,
}))
```

### 配置参数说明

| 参数 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `AllowOrigins` | `[]string` | 允许的源列表，支持通配符 "*" | `["*"]` |
| `AllowMethods` | `[]string` | 允许的HTTP方法列表 | `["GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"]` |
| `AllowHeaders` | `[]string` | 允许的请求头列表 | `["Origin", "Content-Length", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-CSRF-Token"]` |
| `ExposeHeaders` | `[]string` | 暴露给客户端的响应头列表 | `[]` |
| `AllowCredentials` | `bool` | 是否允许发送凭据 | `false` |
| `MaxAge` | `time.Duration` | 预检请求的缓存时间 | `12小时` |

## Panic恢复中间件

### 功能特性

- 捕获和处理运行时panic，防止服务器崩溃
- 支持自定义堆栈跟踪大小和范围
- 支持自定义日志记录和错误处理
- 提供开发和生产环境的预设配置
- 支持指标统计和监控集成
- 格式化的错误信息和堆栈跟踪

### 基本使用

```go
package main

import (
    "chi"
    "chi/middlewares"
)

func main() {
    server := chi.New()
    
    // 使用默认恢复配置（建议放在最前面）
    server.Use(middlewares.Recovery())
    
    // 或使用开发环境配置（详细错误信息）
    server.Use(middlewares.RecoveryForDevelopment())
    
    // 或使用生产环境配置（简化错误信息）
    server.Use(middlewares.RecoveryForProduction())
    
    server.Run(":8080")
}
```

### 自定义配置

```go
server.Use(middlewares.RecoveryWithConfig(middlewares.RecoveryConfig{
    StackSize: 8 << 10, // 8KB
    DisableStackAll: false,
    LogFunc: func(c *chi.Context, err interface{}, stack []byte) {
        // 自定义日志记录逻辑
        log.Printf("Panic recovered: %v\n%s", err, stack)
    },
    RecoveryHandler: func(c *chi.Context, err interface{}) {
        // 自定义错误响应
        c.JSON(500, map[string]interface{}{
            "error": "系统异常",
            "request_id": c.GetHeader("X-Request-ID"),
        })
    },
}))
```

### 配置参数说明

| 参数 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `StackSize` | `int` | 堆栈跟踪的最大字节数 | `4KB` |
| `DisableStackAll` | `bool` | 是否禁用所有goroutine的堆栈跟踪 | `false` |
| `DisablePrintStack` | `bool` | 是否禁用打印堆栈信息 | `false` |
| `LogFunc` | `func(*chi.Context, interface{}, []byte)` | 自定义日志记录函数 | 默认控制台输出 |
| `RecoveryHandler` | `func(*chi.Context, interface{})` | 自定义恢复处理函数 | 返回500状态码 |

### 预设配置说明

1. **开发环境配置** (`RecoveryForDevelopment`): 详细的错误信息和堆栈跟踪，便于调试
2. **生产环境配置** (`RecoveryForProduction`): 简化的错误信息，不暴露敏感信息
3. **带指标配置** (`RecoveryWithMetrics`): 支持panic统计，便于监控

### 重要提示

⚠️ **Panic恢复中间件必须放在所有其他中间件的最前面**，这样才能捕获到其他中间件中发生的panic。

```go
server := chi.New()

// ✅ 正确：Recovery放在最前面
server.Use(middlewares.Recovery())
server.Use(middlewares.CORS())
server.Use(middlewares.RateLimit(100, 200))

// ❌ 错误：Recovery不在最前面，无法捕获CORS中间件的panic
server.Use(middlewares.CORS())
server.Use(middlewares.Recovery())
```

## 限流中间件

### 功能特性

- 基于令牌桶算法的限流实现
- 支持多种限流策略（IP、用户、路径、全局）
- 支持自定义键生成函数
- 支持跳过限流的条件函数
- 支持自定义错误处理
- 自动清理过期的令牌桶

### 基本使用

```go
package main

import (
    "chi"
    "chi/middlewares"
)

func main() {
    server := chi.New()
    
    // 基本限流：每秒100个请求，突发200个
    server.Use(middlewares.RateLimit(100, 200))
    
    // 基于IP的限流
    server.Use(middlewares.RateLimitByIP(50, 100))
    
    // 基于用户的限流
    server.Use(middlewares.RateLimitByUser(200, 400, "user_id"))
    
    // 基于路径的限流
    server.Use(middlewares.RateLimitByPath(10, 20))
    
    // 全局限流
    server.Use(middlewares.RateLimitGlobal(1000, 2000))
    
    server.Run(":8080")
}
```

### 自定义配置

```go
server.Use(middlewares.RateLimitWithConfig(middlewares.RateLimitConfig{
    Rate:  50,
    Burst: 100,
    KeyFunc: func(c *chi.Context) string {
        // 自定义键生成逻辑
        return c.GetHeader("X-API-Key")
    },
    SkipFunc: func(c *chi.Context) bool {
        // 跳过某些路径的限流
        return c.FullPath() == "/health"
    },
    ErrorHandler: func(c *chi.Context) {
        // 自定义错误响应
        c.JSON(429, map[string]string{
            "error": "请求过于频繁",
            "retry_after": "60",
        })
    },
}))
```

### 配置参数说明

| 参数 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `Rate` | `int` | 每秒允许的请求数 | `100` |
| `Burst` | `int` | 令牌桶容量，允许的突发请求数 | `Rate * 2` |
| `KeyFunc` | `func(*chi.Context) string` | 生成限流键的函数 | 使用客户端IP |
| `SkipFunc` | `func(*chi.Context) bool` | 跳过限流的条件函数 | `nil` |
| `ErrorHandler` | `func(*chi.Context)` | 限流触发时的错误处理函数 | 返回429状态码 |

### 限流策略说明

1. **基于IP限流** (`RateLimitByIP`): 每个IP地址独立计算限流
2. **基于用户限流** (`RateLimitByUser`): 每个用户独立计算限流，需要先设置用户ID到Context
3. **基于路径限流** (`RateLimitByPath`): 每个API路径独立计算限流
4. **全局限流** (`RateLimitGlobal`): 所有请求共享同一个限流计数器

## 组合使用

### 推荐的中间件组合

```go
package main

import (
    "chi"
    "chi/middlewares"
)

func main() {
    server := chi.New()
    
    // 开发环境（Recovery必须放在最前面）
    server.Use(middlewares.RecoveryForDevelopment())
    server.Use(middlewares.CORSForDevelopment())
    server.Use(middlewares.RateLimit(1000, 2000))
    
    // 生产环境
    // server.Use(middlewares.RecoveryForProduction())
    // server.Use(middlewares.CORSForProduction([]string{"https://yourdomain.com"}))
    // server.Use(middlewares.RateLimitByIP(100, 200))
    // server.Use(middlewares.RateLimitGlobal(1000, 2000))
    
    server.Run(":8080")
}
```

### 全局中间件

```go
server := chi.New()

// 全局应用Recovery、CORS和限流
server.Use(
    middlewares.Recovery(),
    middlewares.CORS(),
    middlewares.RateLimit(100, 200),
)
```

### 路由组中间件

```go
// API路由组，应用特定的限流策略
apiGroup := server.Group("/api/v1",
    middlewares.RateLimitByIP(30, 60),
    middlewares.CORSWithConfig(middlewares.CORSConfig{
        AllowOrigins:     []string{"https://app.example.com"},
        AllowCredentials: true,
    }),
)

apiGroup.GET("/users", getUsersHandler)
```

### 环境配置

```go
// 生产环境配置
if env == "production" {
    server.Use(middlewares.ProductionMiddlewares([]string{
        "https://example.com",
        "https://app.example.com",
    })...)
} else {
    // 开发环境配置
    server.Use(middlewares.DevelopmentMiddlewares()...)
}
```

## 性能考虑

### CORS中间件

- 预检请求会被直接处理，不会传递到业务逻辑
- 支持缓存预检请求结果，减少重复处理
- 字符串拼接操作已优化，减少内存分配

### 限流中间件

- 使用令牌桶算法，性能优于滑动窗口
- 支持自动清理过期的令牌桶，防止内存泄漏
- 使用读写锁优化并发性能
- 全局限流器单例模式，减少资源消耗

## 注意事项

1. **CORS配置**:
   - 生产环境不要使用通配符 "*" 作为允许源
   - 当 `AllowCredentials` 为 `true` 时，不能使用通配符源
   - 合理设置 `MaxAge` 以平衡性能和安全性

2. **限流配置**:
   - 根据服务器性能和业务需求合理设置限流参数
   - 考虑使用不同的限流策略组合
   - 重要接口可以设置更严格的限流
   - 健康检查等接口可以跳过限流

3. **监控和调试**:
   - 建议添加日志记录限流触发情况
   - 监控限流中间件的性能影响
   - 根据实际使用情况调整限流参数

## 扩展开发

如需开发其他中间件，可以参考现有中间件的实现模式：

1. 定义配置结构体
2. 提供默认配置
3. 实现配置验证和默认值设置
4. 返回符合 `chi.MiddlewareFunc` 类型的函数
5. 在函数中处理请求并调用 `c.Next()` 继续处理

```go
func CustomMiddleware() chi.MiddlewareFunc {
    return func(c *chi.Context) {
        // 前置处理
        
        c.Next() // 继续处理请求
        
        // 后置处理
    }
}
```