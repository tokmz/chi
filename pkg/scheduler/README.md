# Scheduler - 定时任务调度包

## 概述

Scheduler是一个功能完善的Go语言定时任务调度包，提供了灵活的任务调度、管理和监控功能。支持多种调度方式，具备完整的任务生命周期管理和实时监控能力。

## 特性

- **多种调度方式**：支持Cron表达式、固定间隔、一次性任务、延迟执行等调度方式
- **任务管理**：提供启动、停止、暂停、恢复、更新等完整的任务生命周期管理
- **工作池管理**：内置工作池，支持并发控制和资源管理
- **实时监控**：提供系统指标、调度器指标、任务指标和工作池指标的实时监控
- **告警机制**：支持自定义告警条件和回调通知
- **日志记录**：完善的日志记录和错误处理机制，支持多级别日志输出
- **配置灵活**：丰富的配置选项，支持持久化、监控、错误处理等配置
- **线程安全**：所有操作都是线程安全的，支持高并发场景
- **回调支持**：支持任务执行回调和调度器生命周期回调

## 架构设计

```
scheduler/
├── README.md           # 文档说明
├── scheduler.go        # 主调度器实现
├── task.go            # 任务定义和管理
├── worker_pool.go     # 工作池实现
├── monitor.go         # 监控和指标收集
├── logger.go          # 日志记录器
├── config.go          # 配置定义
├── errors.go          # 错误定义
└── example_test.go    # 使用示例
```

## 核心组件

### 1. Scheduler（调度器）
- **DefaultScheduler**：主调度器实现，负责任务的调度和执行
- 支持Cron表达式、固定间隔、一次性任务、延迟任务等多种调度策略
- 提供完整的任务生命周期管理（添加、删除、启动、停止、暂停、恢复）
- 集成工作池进行任务执行和并发控制
- 支持调度器级别的回调和统计信息

### 2. Task（任务）
- **Task结构体**：任务的核心定义，包含ID、类型、配置、执行函数等
- **TaskConfig**：任务配置，支持超时、重试、并发控制等设置
- **TaskStats**：任务统计信息，记录运行次数、成功率、执行时间等
- 支持四种任务类型：Cron、Interval、Once、Delay
- 线程安全的状态管理和统计更新

### 3. WorkerPool（工作池）
- **WorkerPool**：工作池实现，管理工作协程和任务队列
- **Worker**：工作协程，负责具体的任务执行
- **WorkItem**：工作项，封装待执行的任务和回调
- 支持动态工作协程管理和队列监控
- 提供详细的工作池统计信息

### 4. Monitor（监控器）
- **DefaultMonitor**：监控器实现，提供系统和业务指标监控
- **MonitorMetrics**：监控指标，包含系统、调度器、任务、工作池四类指标
- **Alert**：告警机制，支持自定义告警条件和阈值
- 实时健康检查和指标收集
- 支持监控回调和告警通知

### 5. Logger（日志记录器）
- 多级别日志记录（Debug、Info、Warn、Error）
- 支持控制台和文件输出
- 结构化日志格式
- 异步日志处理能力

### 6. Config（配置管理）
- **SchedulerConfig**：调度器主配置
- **WorkerPoolConfig**：工作池配置
- **MonitorConfig**：监控配置
- **PersistenceConfig**：持久化配置
- 提供默认配置和配置验证功能

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "log"
    "time"
    "your-project/pkg/scheduler"
)

func main() {
    // 创建调度器配置
    config := scheduler.DefaultSchedulerConfig()
    config.MaxWorkers = 5
    config.QueueSize = 100

    // 创建调度器
    s, err := scheduler.NewScheduler(config)
    if err != nil {
        log.Fatal("Failed to create scheduler:", err)
    }

    // 启动调度器
    if err := s.Start(); err != nil {
        log.Fatal("Failed to start scheduler:", err)
    }
    defer s.Stop()

    // 创建一个简单的间隔任务
    task1 := scheduler.NewTask("hello-task", "Hello World Task", func() (interface{}, error) {
        fmt.Println("Hello, Scheduler!")
        return "success", nil
    })
    task1.SetInterval(5 * time.Second)

    // 创建一个Cron任务
    task2 := scheduler.NewTask("cron-task", "Cron Task", func() (interface{}, error) {
        fmt.Println("Cron task executed at", time.Now().Format("15:04:05"))
        return "cron success", nil
    })
    task2.SetCron("0 * * * * *") // 每分钟执行一次

    // 添加任务到调度器
    if err := s.AddTask(task1); err != nil {
        log.Fatal("Failed to add task1:", err)
    }
    if err := s.AddTask(task2); err != nil {
        log.Fatal("Failed to add task2:", err)
    }

    // 等待执行
    time.Sleep(2 * time.Minute)

    // 查看调度器统计信息
    stats := s.GetSchedulerStats()
    fmt.Printf("Total tasks: %d, Completed: %d, Failed: %d\n", 
        stats.TotalTasks, stats.CompletedTasks, stats.FailedTasks)
}
```

### 高级功能

```go
// 带重试机制的任务
task := scheduler.NewTask("retry-task", "Retry Task", func() (interface{}, error) {
    // 可能失败的操作
    return "result", nil
})

// 配置重试参数
config := scheduler.DefaultTaskConfig()
config.MaxRetries = 3
config.RetryInterval = 10 * time.Second
config.Timeout = 30 * time.Second
task.SetConfig(config)

// 设置任务回调
task.SetCallback(func(taskID string, result interface{}, err error) {
    if err != nil {
        fmt.Printf("Task %s failed: %v\n", taskID, err)
    } else {
        fmt.Printf("Task %s completed: %v\n", taskID, result)
    }
})

// 一次性任务
once := scheduler.NewTask("once-task", "One Time Task", func() (interface{}, error) {
    fmt.Println("This runs only once")
    return nil, nil
})
once.SetDelay(10 * time.Second) // 10秒后执行一次

// 延迟任务
delay := scheduler.NewTask("delay-task", "Delay Task", func() (interface{}, error) {
    fmt.Println("This runs after delay")
    return nil, nil
})
delay.SetDelay(1 * time.Minute)
```

## 使用场景

### 数据处理
- 定期数据备份和清理
- 批量数据同步和迁移
- 数据库维护任务
- 日志文件轮转和清理

### 业务自动化
- 定期报表生成和发送
- 订单状态检查和更新
- 用户积分结算
- 优惠券过期处理

### 系统维护
- 系统健康检查
- 性能指标收集
- 缓存预热和刷新
- 临时文件清理

### 通知推送
- 定时消息推送
- 邮件发送任务
- 短信通知
- 系统告警通知

### 监控告警
- 服务可用性监控
- 资源使用率检查
- 业务指标监控
- 异常情况告警

## API文档

### 调度器接口

```go
type Scheduler interface {
    // 基本操作
    Start() error
    Stop() error
    Restart() error
    GetStatus() SchedulerStatus

    // 任务管理
    AddTask(task *Task) error
    RemoveTask(taskID string) error
    GetTask(taskID string) (*Task, error)
    ListTasks() []*Task
    UpdateTask(task *Task) error

    // 任务控制
    StartTask(taskID string) error
    StopTask(taskID string) error
    PauseTask(taskID string) error
    ResumeTask(taskID string) error
    RunTaskOnce(taskID string) error

    // 统计信息
    GetTaskStats(taskID string) (*TaskStats, error)
    GetSchedulerStats() *SchedulerStats

    // 配置管理
    UpdateConfig(config *SchedulerConfig) error
    GetConfig() *SchedulerConfig
}
```

### 任务类型

- **TaskTypeCron**: Cron表达式任务
- **TaskTypeInterval**: 固定间隔任务
- **TaskTypeOnce**: 一次性任务
- **TaskTypeDelay**: 延迟任务

### 任务状态

- **TaskStatusPending**: 等待中
- **TaskStatusRunning**: 运行中
- **TaskStatusPaused**: 暂停
- **TaskStatusStopped**: 已停止
- **TaskStatusCompleted**: 已完成
- **TaskStatusFailed**: 失败

## 配置选项

### 调度器配置

```go
type SchedulerConfig struct {
    MaxWorkers      int           // 最大工作协程数
    QueueSize       int           // 任务队列大小
    TickInterval    time.Duration // 调度检查间隔
    ShutdownTimeout time.Duration // 关闭超时时间
    LogLevel        LogLevel      // 日志级别
    EnableMonitor   bool          // 是否启用监控
    MonitorInterval time.Duration // 监控检查间隔
    PanicRecovery   bool          // 是否启用panic恢复
    // ... 更多配置选项
}
```

### 任务配置

```go
type TaskConfig struct {
    Name          string        // 任务名称
    Description   string        // 任务描述
    Enabled       bool          // 是否启用
    Schedule      string        // 调度表达式
    Timeout       time.Duration // 执行超时时间
    MaxRetries    int           // 最大重试次数
    RetryInterval time.Duration // 重试间隔
    Concurrency   int           // 并发数限制
    SkipIfRunning bool          // 如果正在运行则跳过
    // ... 更多配置选项
}
```

## 监控和指标

调度器提供丰富的监控指标：

- **系统指标**: CPU使用率、内存使用率、协程数量
- **调度器指标**: 任务总数、运行中任务数、完成任务数、失败任务数
- **任务指标**: 平均执行时间、成功率、错误率、吞吐量
- **工作池指标**: 工作协程数、队列大小、平均等待时间

## 依赖

- `github.com/robfig/cron/v3` - Cron表达式解析和调度
- Go 1.18+ - 支持泛型和最新语言特性

## 注意事项

1. **资源管理**: 合理设置工作协程数和队列大小，避免资源浪费
2. **错误处理**: 任务函数应该妥善处理错误，避免panic导致工作协程退出
3. **超时设置**: 为长时间运行的任务设置合理的超时时间
4. **并发控制**: 对于资源敏感的任务，使用并发数限制
5. **监控告警**: 启用监控功能，及时发现和处理异常情况

## 许可证

MIT License