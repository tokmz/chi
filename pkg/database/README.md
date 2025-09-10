# Database Package - 日志功能文档

## 概述

database包提供了增强的数据库日志功能，集成了统一的logger包，支持慢查询监控、性能统计和多种日志输出方式。

## 功能特性

### 🚀 核心功能

- **统一日志接口**: 集成chi/pkg/logger包，提供一致的日志体验
- **慢查询监控**: 自动检测和记录超过阈值的SQL查询
- **性能统计**: 实时统计查询性能指标
- **多级别日志**: 支持debug、info、warn、error等日志级别
- **多输出方式**: 支持控制台、文件等多种输出方式
- **GORM集成**: 深度集成GORM日志系统
- **动态配置**: 支持运行时调整日志配置

### 📊 监控指标

- 总查询数量
- 慢查询数量和详情
- 平均查询时间
- 最大查询时间
- QPS (每秒查询数)
- 错误查询统计

## 快速开始

### 基本使用

```go
package main

import (
    "log"
    "time"
    "chi/pkg/database"
)

func main() {
    // 创建数据库配置
    config := &database.Config{
        DSN: "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local",
        PoolConfig: database.PoolConfig{
            MaxOpenConns:    10,
            MaxIdleConns:    5,
            ConnMaxLifetime: time.Hour,
        },
        LogConfig: database.LogConfig{
            Level:          "info",
            ColorfulOutput: true,
            SlowThreshold:  200 * time.Millisecond,
        },
        SlowQueryConfig: database.SlowQueryConfig{
            Enabled:   true,
            Threshold: 200 * time.Millisecond,
        },
    }

    // 创建客户端
    client, err := database.NewClient(config)
    if err != nil {
        log.Fatalf("Failed to create database client: %v", err)
    }
    defer client.Close()

    // 使用数据库
    db := client.DB()
    // ... 执行数据库操作
}
```

## API 参考

### 客户端方法

#### 日志管理

```go
// 获取日志适配器
func (c *Client) GetLoggerAdapter() *DatabaseLoggerAdapter

// 获取慢查询监控器
func (c *Client) GetSlowQueryMonitor() *DatabaseSlowQueryMonitor

// 获取性能监控器
func (c *Client) GetPerformanceMonitor() *DatabasePerformanceMonitor
```

#### 统计信息

```go
// 获取慢查询统计
func (c *Client) GetSlowQueryStats() *SlowQueryStats

// 获取性能统计
func (c *Client) GetPerformanceStats() *PerformanceStats

// 重置慢查询统计
func (c *Client) ResetSlowQueryStats()

// 重置性能统计
func (c *Client) ResetPerformanceStats()
```

#### 动态配置

```go
// 设置慢查询阈值
func (c *Client) SetSlowQueryThreshold(threshold time.Duration)

// 设置日志级别
func (c *Client) SetLogLevel(level string) error
```

## 监控和统计

### 获取慢查询统计

```go
stats := client.GetSlowQueryStats()
if stats != nil {
    fmt.Printf("总查询数: %d\n", stats.TotalQueries)
    fmt.Printf("慢查询数: %d\n", stats.SlowQueries)
    fmt.Printf("平均耗时: %v\n", stats.AverageDuration)
    fmt.Printf("最大耗时: %v\n", stats.MaxDuration)
}
```

### 获取性能统计

```go
stats := client.GetPerformanceStats()
if stats != nil {
    fmt.Printf("总查询数: %d\n", stats.TotalQueries)
    fmt.Printf("总耗时: %v\n", stats.TotalDuration)
    fmt.Printf("平均耗时: %v\n", stats.AverageDuration)
    fmt.Printf("QPS: %.2f\n", stats.QPS)
    fmt.Printf("错误数量: %d\n", stats.ErrorCount)
}
```

### 动态调整配置

```go
// 调整慢查询阈值
client.SetSlowQueryThreshold(500 * time.Millisecond)

// 调整日志级别
client.SetLogLevel("debug")

// 重置统计信息
client.ResetSlowQueryStats()
client.ResetPerformanceStats()
```

## 最佳实践

### 1. 合理设置慢查询阈值

```go
// 根据业务需求设置合适的阈值
// 一般建议：
// - OLTP系统: 100-200ms
// - OLAP系统: 1-5s
client.SetSlowQueryThreshold(200 * time.Millisecond)
```

### 2. 定期监控统计信息

```go
// 定期检查性能统计
go func() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        stats := client.GetPerformanceStats()
        if stats != nil && stats.QPS > 1000 {
            log.Printf("High QPS detected: %.2f", stats.QPS)
        }
    }
}()
```

### 3. 适当的日志级别

```go
// 生产环境建议使用info级别
// 开发环境可以使用debug级别
if isProduction {
    client.SetLogLevel("info")
} else {
    client.SetLogLevel("debug")
}
```

## 版本兼容性

- Go 1.18+
- GORM v1.25+
- MySQL 5.7+/8.0+
- PostgreSQL 12+

## 许可证

本项目采用 MIT 许可证。