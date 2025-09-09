# Logger Package

一个基于 zap 的企业级 Go 日志封装包，提供丰富的日志功能、灵活的配置选项和强大的日志管理能力。

## 功能特性

### 🚀 核心功能
- **多种输出格式**: 支持 JSON 结构化格式和可读性强的控制台格式
- **完整日志级别**: Debug、Info、Warn、Error、Panic、Fatal
- **多种记录方式**: 结构化字段、格式化字符串、键值对参数
- **调用信息**: 可选记录文件名、行号和函数名
- **自动堆栈跟踪**: 错误级别自动记录堆栈信息
- **高测试覆盖率**: 测试覆盖率达到 60.7%，确保代码质量和稳定性

### 📁 日志分割
- **按大小分割**: 单文件大小限制，达到阈值自动创建新文件
- **按时间分割**: 支持按小时、天、周、月等时间间隔分割
- **灵活配置**: 可同时启用多种分割策略

### 🗂️ 日志管理
- **自动清理**: 自动清理过期日志文件
- **自动压缩**: 历史日志自动压缩存储（支持 gzip、lz4）
- **统计信息**: 提供详细的日志文件统计数据
- **手动管理**: 支持手动触发清理和压缩操作

### 🎯 输出目标
- **控制台输出**: 开发环境使用彩色输出
- **文件输出**: 生产环境使用文件记录
- **混合输出**: 同时输出到控制台和文件
- **多目标输出**: 支持同时写入多个日志文件，每个文件可配置不同的级别过滤

### ⚡ 性能优化
- **采样机制**: 高频日志采样以减少性能影响
- **异步写入**: 基于 zap 的高性能异步写入
- **连接池**: 高效的文件句柄管理
- **并发安全**: 完整的并发安全测试，支持多线程环境下的稳定运行

## 安装依赖

在项目根目录的 `go.mod` 中添加以下依赖：

```bash
go get go.uber.org/zap
go get gopkg.in/natefinch/lumberjack.v2
```

## 快速开始

### 1. 基本使用

```go
package main

import (
    "chi/pkg/logger"
)

func main() {
    // 使用默认配置初始化全局日志记录器
    if err := logger.InitGlobal(nil); err != nil {
        panic(err)
    }
    defer logger.Sync()

    // 基本日志记录
    logger.Info("应用程序启动")
    logger.Debug("调试信息")
    logger.Warn("警告信息")
    logger.Error("错误信息")

    // 格式化日志
    logger.Infof("用户 %s 登录成功", "张三")
    
    // 键值对日志
    logger.Infow("用户操作",
        "user_id", 12345,
        "action", "login",
        "ip", "192.168.1.100",
    )

    // 结构化字段日志
    logger.Info("订单创建成功",
        logger.String("order_id", "ORD-2024-001"),
        logger.Int64("user_id", 12345),
        logger.Float64("amount", 99.99),
    )
}
```

### 2. 自定义配置

```go
package main

import (
    "time"
    "chi/pkg/logger"
)

func main() {
    // 创建自定义配置
    config := &logger.Config{
        Level:  "debug",
        Format: "json",
        Output: logger.OutputConfig{
            Console: logger.ConsoleConfig{
                Enabled:    true,
                Colorful:   true,
                TimeFormat: "2006-01-02 15:04:05",
            },
            File: logger.FileConfig{
                Enabled:     true,
                Filename:    "logs/app.log",
                MaxSize:     50, // 50MB
                MaxBackups:  5,
                MaxAge:      7, // 7天
                Compress:    true,
                LocalTime:   true,
                LevelFilter: "info", // 文件只记录info及以上级别
            },
        },
        Caller: logger.CallerConfig{
            Enabled:  true,
            FullPath: false,
        },
        Rotation: logger.RotationConfig{
            Size: logger.SizeRotationConfig{
                Enabled: true,
                MaxSize: 50,
            },
            Time: logger.TimeRotationConfig{
                Enabled:    true,
                Interval:   "day",
                RotateTime: "00:00",
            },
        },
        Management: logger.ManagementConfig{
            Cleanup: logger.CleanupConfig{
                Enabled:  true,
                MaxAge:   30,
                Interval: 24 * time.Hour,
            },
            Compression: logger.CompressionConfig{
                Enabled:   true,
                Delay:     24,
                Algorithm: "gzip",
            },
        },
    }

    // 创建日志记录器
    log, err := logger.NewLogger(config)
    if err != nil {
        panic(err)
    }
    defer log.Close()

    log.Info("使用自定义配置的日志记录器")
}
```

### 3. 多文件输出

```go
config := &logger.Config{
    Level:  "debug",
    Format: "json",
    Output: logger.OutputConfig{
        Console: logger.ConsoleConfig{
            Enabled:  true,
            Colorful: true,
        },
        MultiFile: []logger.FileConfig{
            {
                Enabled:     true,
                Filename:    "logs/app.log",
                LevelFilter: "", // 所有级别
            },
            {
                Enabled:     true,
                Filename:    "logs/error.log",
                LevelFilter: "error", // 只记录错误级别
            },
            {
                Enabled:     true,
                Filename:    "logs/access.log",
                LevelFilter: "info", // 记录info及以上级别
            },
        },
    },
}
```

## 配置说明

### 主配置 (Config)

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| Level | string | "info" | 日志级别：debug, info, warn, error, panic, fatal |
| Format | string | "console" | 输出格式：json, console |
| Development | bool | false | 开发模式，影响默认配置 |
| Output | OutputConfig | - | 输出配置 |
| Caller | CallerConfig | - | 调用信息配置 |
| Rotation | RotationConfig | - | 日志分割配置 |
| Management | ManagementConfig | - | 日志管理配置 |
| Sampling | SamplingConfig | - | 采样配置 |

### 输出配置 (OutputConfig)

#### 控制台配置 (ConsoleConfig)

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| Enabled | bool | true | 是否启用控制台输出 |
| Colorful | bool | true | 是否启用彩色输出 |
| TimeFormat | string | "2006-01-02 15:04:05" | 时间格式 |

#### 文件配置 (FileConfig)

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| Enabled | bool | false | 是否启用文件输出 |
| Filename | string | "logs/app.log" | 文件路径 |
| MaxSize | int | 100 | 最大文件大小(MB) |
| MaxBackups | int | 10 | 最大备份数量 |
| MaxAge | int | 30 | 最大保留天数 |
| Compress | bool | true | 是否压缩 |
| LocalTime | bool | true | 使用本地时间 |
| LevelFilter | string | "" | 日志级别过滤 |

### 日志分割配置 (RotationConfig)

#### 按大小分割 (SizeRotationConfig)

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| Enabled | bool | true | 是否启用按大小分割 |
| MaxSize | int | 100 | 最大文件大小(MB) |

#### 按时间分割 (TimeRotationConfig)

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| Enabled | bool | false | 是否启用按时间分割 |
| Interval | string | "day" | 分割间隔：hour, day, week, month |
| RotateTime | string | "00:00" | 分割时间点(小时:分钟) |

### 日志管理配置 (ManagementConfig)

#### 清理配置 (CleanupConfig)

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| Enabled | bool | true | 是否启用自动清理 |
| MaxAge | int | 30 | 保留天数 |
| Interval | time.Duration | 24h | 清理间隔 |

#### 压缩配置 (CompressionConfig)

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| Enabled | bool | true | 是否启用自动压缩 |
| Delay | int | 24 | 压缩延迟(小时) |
| Algorithm | string | "gzip" | 压缩算法：gzip, lz4 |

## API 参考

### 日志记录方法

#### 结构化字段记录
```go
logger.Debug(msg string, fields ...Field)
logger.Info(msg string, fields ...Field)
logger.Warn(msg string, fields ...Field)
logger.Error(msg string, fields ...Field)
logger.Panic(msg string, fields ...Field)
logger.Fatal(msg string, fields ...Field)
```

#### 格式化字符串记录
```go
logger.Debugf(template string, args ...interface{})
logger.Infof(template string, args ...interface{})
logger.Warnf(template string, args ...interface{})
logger.Errorf(template string, args ...interface{})
logger.Panicf(template string, args ...interface{})
logger.Fatalf(template string, args ...interface{})
```

#### 键值对记录
```go
logger.Debugw(msg string, keysAndValues ...interface{})
logger.Infow(msg string, keysAndValues ...interface{})
logger.Warnw(msg string, keysAndValues ...interface{})
logger.Errorw(msg string, keysAndValues ...interface{})
logger.Panicw(msg string, keysAndValues ...interface{})
logger.Fatalw(msg string, keysAndValues ...interface{})
```

### 字段构造函数

```go
logger.String(key, val string) Field
logger.Int(key string, val int) Field
logger.Int64(key string, val int64) Field
logger.Float64(key string, val float64) Field
logger.Bool(key string, val bool) Field
logger.Time(key string, val time.Time) Field
logger.Duration(key string, val time.Duration) Field
logger.Err(err error) Field
logger.Any(key string, val interface{}) Field
```

### 日志记录器操作

```go
// 创建带字段的子记录器
logger.With(fields ...Field) *Logger

// 创建命名子记录器
logger.Named(name string) *Logger

// 同步日志
logger.Sync() error

// 关闭日志记录器
logger.Close() error

// 动态设置日志级别
logger.SetLevel(level string)

// 获取当前日志级别
logger.GetLevel() string
```

### 日志管理

```go
// 创建日志管理器
manager := logger.NewManager(config, logDir)

// 启动管理器
manager.Start() error

// 停止管理器
manager.Stop() error

// 获取统计信息
manager.GetStats() (*LogStats, error)

// 手动清理
manager.ForceCleanup() error

// 手动压缩
manager.ForceCompression() error

// 按模式清理
manager.CleanupByPattern(pattern string, maxAge time.Duration) error

// 归档日志
manager.ArchiveLogs(archiveDir string, maxAge time.Duration) error
```

## 使用示例

### Web 应用日志

```go
package main

import (
    "context"
    "net/http"
    "time"
    "chi/pkg/logger"
)

func main() {
    // 初始化日志
    config := &logger.Config{
        Level:  "info",
        Format: "json",
        Output: logger.OutputConfig{
            Console: logger.ConsoleConfig{
                Enabled:  true,
                Colorful: true,
            },
            File: logger.FileConfig{
                Enabled:  true,
                Filename: "logs/web.log",
            },
        },
    }
    
    log, err := logger.NewLogger(config)
    if err != nil {
        panic(err)
    }
    defer log.Close()

    // HTTP 中间件
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // 创建请求日志记录器
        requestLogger := log.With(
            logger.String("method", r.Method),
            logger.String("path", r.URL.Path),
            logger.String("remote_addr", r.RemoteAddr),
        )
        
        requestLogger.Info("请求开始")
        
        // 处理请求
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Hello, World!"))
        
        // 记录请求完成
        requestLogger.Info("请求完成",
            logger.Int("status", http.StatusOK),
            logger.Duration("duration", time.Since(start)),
        )
    })

    log.Info("服务器启动", logger.String("addr", ":8080"))
    http.ListenAndServe(":8080", nil)
}
```

### 错误处理和恢复

```go
func handlePanic() {
    if r := recover(); r != nil {
        logger.Error("发生恐慌",
            logger.Any("panic", r),
            logger.String("stack", string(debug.Stack())),
        )
    }
}

func riskyOperation() {
    defer handlePanic()
    
    // 可能发生恐慌的操作
    panic("something went wrong")
}
```

### 性能监控

```go
func monitorPerformance(operation string, fn func() error) error {
    start := time.Now()
    
    logger.Debug("操作开始", logger.String("operation", operation))
    
    err := fn()
    
    duration := time.Since(start)
    
    if err != nil {
        logger.Error("操作失败",
            logger.String("operation", operation),
            logger.Duration("duration", duration),
            logger.Err(err),
        )
    } else {
        logger.Info("操作成功",
            logger.String("operation", operation),
            logger.Duration("duration", duration),
        )
    }
    
    return err
}
```

## 最佳实践

### 1. 日志级别使用指南

- **Debug**: 详细的调试信息，仅在开发环境使用
- **Info**: 一般信息，记录程序的正常运行状态
- **Warn**: 警告信息，程序可以继续运行但需要注意
- **Error**: 错误信息，程序遇到错误但可以恢复
- **Panic**: 严重错误，程序无法继续运行
- **Fatal**: 致命错误，程序将退出

### 2. 结构化日志

优先使用结构化字段而不是格式化字符串：

```go
// 推荐
logger.Info("用户登录",
    logger.String("user_id", "12345"),
    logger.String("ip", "192.168.1.100"),
    logger.Duration("duration", 150*time.Millisecond),
)

// 不推荐
logger.Infof("用户 %s 从 %s 登录，耗时 %v", "12345", "192.168.1.100", 150*time.Millisecond)
```

### 3. 上下文传递

在请求处理过程中传递日志记录器：

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    requestLogger := logger.GetGlobal().With(
        logger.String("request_id", generateRequestID()),
        logger.String("user_id", getUserID(r)),
    )
    
    processRequest(r.Context(), requestLogger)
}

func processRequest(ctx context.Context, log *logger.Logger) {
    log.Info("开始处理请求")
    // 处理逻辑
    log.Info("请求处理完成")
}
```

### 4. 错误处理

记录错误时包含足够的上下文信息：

```go
if err := db.Query(sql, args...); err != nil {
    logger.Error("数据库查询失败",
        logger.Err(err),
        logger.String("sql", sql),
        logger.Any("args", args),
        logger.String("operation", "user_query"),
    )
    return err
}
```

### 5. 性能考虑

- 在高频调用的代码中使用适当的日志级别
- 启用采样机制减少日志量
- 避免在日志中记录大量数据
- 使用异步写入提高性能

## 故障排除

### 常见问题

1. **日志文件无法创建**
   - 检查目录权限
   - 确保目录存在或启用自动创建

2. **日志分割不工作**
   - 检查分割配置是否正确
   - 确认文件大小或时间条件是否满足

3. **性能问题**
   - 启用采样机制
   - 调整日志级别
   - 检查磁盘I/O性能

4. **内存占用过高**
   - 检查是否有日志泄漏
   - 调整缓冲区大小
   - 启用日志压缩

### 调试技巧

1. **启用调试模式**
```go
config.Development = true
config.Level = "debug"
```

2. **检查配置**
```go
if err := config.Validate(); err != nil {
    fmt.Printf("配置错误: %v\n", err)
}
```

3. **监控日志统计**
```go
stats, _ := manager.GetStats()
fmt.Printf("日志统计: %+v\n", stats)
```

## 测试覆盖率

本项目具有完善的测试体系，确保代码质量和稳定性：

### 📊 覆盖率统计
- **总体覆盖率**: 60.7%
- **核心功能**: 完全覆盖
- **并发安全**: 全面测试
- **错误处理**: 充分验证

### 🧪 测试文件
- `logger_test.go`: 核心日志功能测试
- `manager_test.go`: 日志管理器测试
- `rotation_test.go`: 日志分割功能测试
- `concurrency_test.go`: 并发安全测试
- `performance_test.go`: 性能基准测试
- `error_handling_test.go`: 错误处理测试
- `coverage_test.go`: 覆盖率补充测试
- `additional_test.go`: 额外功能测试
- `final_test.go`: 边界情况测试

### 🔒 并发安全测试
- **并发访问测试**: 验证多个 goroutine 同时访问日志记录器的安全性
- **并发级别变更**: 测试运行时动态修改日志级别的线程安全性
- **并发写入错误处理**: 验证并发写入时的错误处理机制
- **全局 Logger 并发**: 测试全局日志记录器的并发安全性
- **管理器并发操作**: 验证日志管理器的并发操作安全性

### 🎯 测试运行
```bash
# 运行所有测试
go test -v

# 运行测试并查看覆盖率
go test -cover -v

# 生成详细的覆盖率报告
go test -coverprofile=coverage.out
go tool cover -html=coverage.out

# 运行并发测试
go test -v -run "Test.*Concurrent"

# 运行性能测试
go test -bench=. -benchmem
```

## 版本历史

- **v1.0.0**: 初始版本，包含基本日志功能
- **v1.1.0**: 添加日志分割功能
- **v1.2.0**: 添加日志管理功能
- **v1.3.0**: 添加多文件输出支持
- **v1.4.0**: 性能优化和采样机制
- **v1.5.0**: 完善测试体系，测试覆盖率达到 60.7%，增强并发安全性

## 许可证

本项目采用 MIT 许可证。详见 [LICENSE](../../LICENSE) 文件。

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目。

## 相关链接

- [Zap 官方文档](https://pkg.go.dev/go.uber.org/zap)
- [Lumberjack 文档](https://pkg.go.dev/gopkg.in/natefinch/lumberjack.v2)
- [Go 日志最佳实践](https://dave.cheney.net/2015/11/05/lets-talk-about-logging)