# MongoDB 日志功能使用指南

本文档详细介绍了 MongoDB 包中增强的日志功能，包括配置、使用方法和最佳实践。

## 目录

- [功能概述](#功能概述)
- [快速开始](#快速开始)
- [配置详解](#配置详解)
- [慢查询监控](#慢查询监控)
- [运行时配置](#运行时配置)
- [最佳实践](#最佳实践)
- [示例代码](#示例代码)
- [故障排除](#故障排除)

## 功能概述

新的日志功能提供了以下特性：

### 🚀 核心功能
- **多级别日志**: 支持 DEBUG、INFO、WARN、ERROR、FATAL 五个级别
- **多输出方式**: 支持控制台、文件、JSON 格式输出
- **慢查询监控**: 自动检测和记录慢查询，支持统计分析
- **运行时配置**: 支持动态调整日志级别和慢查询阈值
- **敏感数据保护**: 自动清理日志中的敏感信息

### 🔧 高级特性
- **日志轮转**: 支持按大小和时间轮转日志文件
- **采样配置**: 高频场景下的日志采样
- **调用信息**: 可选的调用栈信息记录
- **连接池监控**: 连接建立、关闭和池状态监控
- **操作日志**: CRUD、聚合、事务、索引操作的详细记录

## 快速开始

### 基础使用

```go
package main

import (
    "chi/pkg/mongo"
    "log"
)

func main() {
    // 使用默认配置
    config := mongo.DefaultConfig()
    config.URI = "mongodb://localhost:27017"
    config.Database = "myapp"
    
    client, err := mongo.NewClient(config)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // 客户端会自动记录连接、操作等日志
}
```

### 自定义日志配置

```go
// 创建自定义日志配置
loggerConfig := &mongo.MongoLoggerConfig{
    Level:         mongo.LogLevelDebug,
    EnableConsole: true,
    UseZapLogger:  true,
    Format:        "json",
    File: mongo.FileLogConfig{
        Enabled:  true,
        Filename: "app.log",
        MaxSize:  100, // 100MB
    },
    Mongo: mongo.MongoSpecificLogConfig{
        SlowQuery: mongo.SlowQueryLogConfig{
            Enabled:   true,
            Threshold: 100 * time.Millisecond,
        },
    },
}

config := mongo.DefaultConfig()
config.Logger = loggerConfig
```

## 配置详解

### MongoLoggerConfig 结构

```go
type MongoLoggerConfig struct {
    // 基础配置
    Level         LogLevel `json:"level"`           // 日志级别
    Output        string   `json:"output"`          // 输出路径
    EnableConsole bool     `json:"enable_console"`  // 启用控制台输出
    UseZapLogger  bool     `json:"use_zap_logger"`  // 使用高性能zap logger
    
    // 格式配置
    Format      string `json:"format"`      // json | console
    Development bool   `json:"development"` // 开发模式
    
    // 文件输出配置
    File FileLogConfig `json:"file"`
    
    // 控制台输出配置
    Console ConsoleLogConfig `json:"console"`
    
    // MongoDB特定配置
    Mongo MongoSpecificLogConfig `json:"mongo"`
}
```

### 日志级别

```go
const (
    LogLevelDebug LogLevel = iota // 调试信息
    LogLevelInfo                  // 一般信息
    LogLevelWarn                  // 警告信息
    LogLevelError                 // 错误信息
    LogLevelFatal                 // 致命错误
)
```

### 文件配置

```go
type FileLogConfig struct {
    Enabled    bool   `json:"enabled"`     // 启用文件输出
    Filename   string `json:"filename"`    // 文件名
    MaxSize    int    `json:"max_size"`    // 最大文件大小(MB)
    MaxBackups int    `json:"max_backups"` // 最大备份数
    MaxAge     int    `json:"max_age"`     // 最大保存天数
    Compress   bool   `json:"compress"`    // 压缩旧文件
    LocalTime  bool   `json:"local_time"`  // 使用本地时间
}
```

### MongoDB特定配置

```go
type MongoSpecificLogConfig struct {
    SlowQuery  SlowQueryLogConfig  `json:"slow_query"`  // 慢查询配置
    Connection ConnectionLogConfig `json:"connection"`  // 连接配置
    Operation  OperationLogConfig  `json:"operation"`   // 操作配置
    Error      ErrorLogConfig      `json:"error"`       // 错误配置
}
```

## 慢查询监控

### 基础配置

```go
slowQueryConfig := mongo.SlowQueryLogConfig{
    Enabled:   true,
    Threshold: 100 * time.Millisecond, // 100ms阈值
    LogQuery:  true,                    // 记录查询语句
    LogResult: false,                   // 不记录查询结果(生产环境)
}
```

### 获取统计信息

```go
// 获取慢查询统计
stats := client.GetSlowQueryStats()
fmt.Printf("总查询数: %d\n", stats.TotalQueries)
fmt.Printf("慢查询数: %d\n", stats.SlowQueries)
fmt.Printf("慢查询率: %.2f%%\n", stats.SlowQueryRate)
fmt.Printf("最大耗时: %v\n", stats.MaxTime)

// 重置统计信息
client.ResetSlowQueryStats()
```

### 手动监控查询

```go
monitor := client.GetSlowQueryMonitor()
err := monitor.MonitorQuery(ctx, "find", "users", "mydb", func() error {
    // 执行查询操作
    return collection.Find(ctx, filter)
})
```

## 运行时配置

### 动态更新日志配置

```go
// 获取当前配置
currentConfig := client.GetLoggerConfig()

// 创建新配置
newConfig := mongo.DefaultMongoLoggerConfig()
newConfig.Level = mongo.LogLevelDebug

// 更新配置
err := client.UpdateLoggerConfig(newConfig)
if err != nil {
    log.Printf("Failed to update config: %v", err)
}
```

### 调整慢查询阈值

```go
// 设置新的慢查询阈值
client.SetSlowQueryThreshold(50 * time.Millisecond)

// 启用/禁用慢查询监控
monitor := client.GetSlowQueryMonitor()
monitor.Enable()  // 启用
monitor.Disable() // 禁用
```

## 最佳实践

### 开发环境配置

```go
devConfig := &mongo.MongoLoggerConfig{
    Level:         mongo.LogLevelDebug,
    EnableConsole: true,
    UseZapLogger:  false, // 使用简单logger便于调试
    Format:        "console",
    Development:   true,
    Console: mongo.ConsoleLogConfig{
        Enabled:    true,
        Colorful:   true,
        TimeFormat: "15:04:05",
    },
    Mongo: mongo.MongoSpecificLogConfig{
        SlowQuery: mongo.SlowQueryLogConfig{
            Enabled:   true,
            Threshold: 10 * time.Millisecond, // 敏感阈值
            LogQuery:  true,
            LogResult: true, // 开发环境可以记录结果
        },
        Operation: mongo.OperationLogConfig{
            Enabled:        true,
            LogCRUD:        true, // 记录所有CRUD操作
            LogAggregation: true,
        },
    },
}
```

### 生产环境配置

```go
prodConfig := &mongo.MongoLoggerConfig{
    Level:         mongo.LogLevelInfo,
    EnableConsole: false, // 生产环境关闭控制台输出
    UseZapLogger:  true,  // 使用高性能logger
    Format:        "json",
    Development:   false,
    File: mongo.FileLogConfig{
        Enabled:    true,
        Filename:   "/var/log/app/mongo.log",
        MaxSize:    100,
        MaxBackups: 10,
        MaxAge:     30,
        Compress:   true,
    },
    Rotation: mongo.RotationLogConfig{
        Enabled:  true,
        MaxSize:  100,
        Interval: 24 * time.Hour,
        Pattern:  "mongo-%Y%m%d.log",
    },
    Mongo: mongo.MongoSpecificLogConfig{
        SlowQuery: mongo.SlowQueryLogConfig{
            Enabled:   true,
            Threshold: 100 * time.Millisecond,
            LogQuery:  true,
            LogResult: false, // 生产环境不记录结果
        },
        Operation: mongo.OperationLogConfig{
            Enabled:        false, // 关闭详细操作日志
            LogTransaction: true,  // 只记录事务
            LogIndex:       true,  // 只记录索引操作
        },
        Error: mongo.ErrorLogConfig{
            Enabled:       true,
            LogStackTrace: true,
            LogContext:    true,
        },
    },
}
```

### 性能优化建议

1. **生产环境使用 zap logger**:
   ```go
   config.UseZapLogger = true
   ```

2. **合理设置日志级别**:
   - 开发环境: `LogLevelDebug`
   - 测试环境: `LogLevelInfo`
   - 生产环境: `LogLevelWarn` 或 `LogLevelError`

3. **配置日志采样**:
   ```go
   config.Sampling = mongo.SamplingLogConfig{
       Enabled:    true,
       Initial:    100,  // 前100条记录所有
       Thereafter: 100,  // 之后每100条记录1条
   }
   ```

4. **慢查询阈值设置**:
   - 开发环境: 10-50ms
   - 生产环境: 100-500ms

## 示例代码

完整的示例代码请参考 `examples/logger_example.go` 文件，包含：

- 基础日志使用
- 自定义配置
- 慢查询监控
- 动态配置更新
- 高级用法示例

运行示例：

```bash
cd examples
go run logger_example.go
```

## 故障排除

### 常见问题

1. **日志文件无法创建**
   - 检查文件路径权限
   - 确保目录存在
   - 检查磁盘空间

2. **慢查询未记录**
   - 确认 `SlowQuery.Enabled = true`
   - 检查阈值设置是否合理
   - 验证查询确实超过阈值

3. **日志级别不生效**
   - 确认配置正确传递给客户端
   - 检查是否有运行时配置覆盖
   - 验证日志记录器初始化成功

4. **性能问题**
   - 使用 `UseZapLogger = true`
   - 关闭不必要的详细日志
   - 配置日志采样
   - 使用异步日志输出

### 调试技巧

1. **启用调试模式**:
   ```go
   config.Development = true
   config.Level = mongo.LogLevelDebug
   ```

2. **检查配置**:
   ```go
   currentConfig := client.GetLoggerConfig()
   fmt.Printf("Current config: %+v\n", currentConfig)
   ```

3. **监控日志输出**:
   ```bash
   tail -f /path/to/mongo.log
   ```

### 日志格式示例

**JSON 格式**:
```json
{
  "level": "INFO",
  "timestamp": "2024-01-15T10:30:45Z",
  "message": "MongoDB connection established",
  "uri": "mongodb://localhost:27017",
  "database": "myapp"
}
```

**控制台格式**:
```
2024-01-15 10:30:45 [INFO] MongoDB connection established uri=mongodb://localhost:27017 database=myapp
```

**慢查询日志**:
```json
{
  "level": "WARN",
  "timestamp": "2024-01-15T10:30:46Z",
  "message": "Slow query detected",
  "operation": "find",
  "collection": "users",
  "database": "myapp",
  "duration": "150ms",
  "threshold": "100ms",
  "filter": "{\"age\": {\"$gt\": 18}}"
}
```

## 版本兼容性

- 新的日志配置通过 `Config.Logger` 字段设置
- 旧的日志配置 `Config.Log` 仍然支持，用于向后兼容
- 如果同时设置了新旧配置，优先使用新配置
- 可以通过 `GetLoggerConfig()` 方法获取有效的日志配置

## 更新日志

### v1.1.0
- 新增 MongoLoggerConfig 扩展配置
- 新增慢查询监控功能
- 新增运行时配置更新
- 新增敏感数据保护
- 新增日志轮转和采样功能
- 向后兼容旧的 LogConfig

---

如有问题或建议，请提交 Issue 或 Pull Request。