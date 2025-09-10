package scheduler

import (
	"time"
)

// SchedulerConfig 调度器配置
type SchedulerConfig struct {
	// 基本配置
	MaxWorkers      int           // 最大工作协程数
	QueueSize       int           // 任务队列大小
	TickInterval    time.Duration // 调度检查间隔
	ShutdownTimeout time.Duration // 关闭超时时间

	// 日志配置
	LogLevel      LogLevel // 日志级别
	LogOutput     string   // 日志输出路径
	EnableConsole bool     // 是否启用控制台输出
	UseZapLogger  bool     // 是否使用zap日志库（推荐）
	
	// 扩展日志配置（可选，优先级高于上述基础配置）
	LoggerConfig *LoggerConfig `json:"logger_config,omitempty" yaml:"logger_config,omitempty"`

	// 监控配置
	EnableMonitor   bool          // 是否启用监控
	MonitorInterval time.Duration // 监控检查间隔
	MetricsEnabled  bool          // 是否启用指标收集

	// 持久化配置
	EnablePersistence bool          // 是否启用持久化
	StoragePath       string        // 存储路径
	SaveInterval      time.Duration // 保存间隔

	// 错误处理配置
	PanicRecovery   bool          // 是否启用panic恢复
	ErrorRetryDelay time.Duration // 错误重试延迟
	MaxErrorRetries int           // 最大错误重试次数
}

// DefaultSchedulerConfig 返回默认调度器配置
func DefaultSchedulerConfig() *SchedulerConfig {
	return &SchedulerConfig{
		// 基本配置
		MaxWorkers:      10,
		QueueSize:       1000,
		TickInterval:    1 * time.Second,
		ShutdownTimeout: 30 * time.Second,

		// 日志配置
		LogLevel:      LogLevelInfo,
		LogOutput:     "", // 空字符串表示不输出到文件
		EnableConsole: true,
		UseZapLogger:  true, // 默认使用zap日志库

		// 监控配置
		EnableMonitor:   true,
		MonitorInterval: 10 * time.Second,
		MetricsEnabled:  true,

		// 持久化配置
		EnablePersistence: false,
		StoragePath:       "./scheduler_data",
		SaveInterval:      5 * time.Minute,

		// 错误处理配置
		PanicRecovery:   true,
		ErrorRetryDelay: 5 * time.Second,
		MaxErrorRetries: 3,
	}
}

// Validate 验证配置有效性
func (c *SchedulerConfig) Validate() error {
	if c.MaxWorkers <= 0 {
		return NewSchedulerError(ErrInvalidConfig, "MaxWorkers must be greater than 0")
	}

	if c.QueueSize <= 0 {
		return NewSchedulerError(ErrInvalidConfig, "QueueSize must be greater than 0")
	}

	if c.TickInterval <= 0 {
		return NewSchedulerError(ErrInvalidConfig, "TickInterval must be greater than 0")
	}

	if c.ShutdownTimeout <= 0 {
		return NewSchedulerError(ErrInvalidConfig, "ShutdownTimeout must be greater than 0")
	}

	if c.EnableMonitor && c.MonitorInterval <= 0 {
		return NewSchedulerError(ErrInvalidConfig, "MonitorInterval must be greater than 0 when monitor is enabled")
	}

	if c.EnablePersistence {
		if c.StoragePath == "" {
			return NewSchedulerError(ErrInvalidConfig, "StoragePath cannot be empty when persistence is enabled")
		}
		if c.SaveInterval <= 0 {
			return NewSchedulerError(ErrInvalidConfig, "SaveInterval must be greater than 0 when persistence is enabled")
		}
	}

	if c.MaxErrorRetries < 0 {
		return NewSchedulerError(ErrInvalidConfig, "MaxErrorRetries cannot be negative")
	}

	if c.ErrorRetryDelay < 0 {
		return NewSchedulerError(ErrInvalidConfig, "ErrorRetryDelay cannot be negative")
	}

	return nil
}

// Clone 克隆配置
func (c *SchedulerConfig) Clone() *SchedulerConfig {
	return &SchedulerConfig{
		MaxWorkers:        c.MaxWorkers,
		QueueSize:         c.QueueSize,
		TickInterval:      c.TickInterval,
		ShutdownTimeout:   c.ShutdownTimeout,
		LogLevel:          c.LogLevel,
		LogOutput:         c.LogOutput,
		EnableConsole:     c.EnableConsole,
		UseZapLogger:      c.UseZapLogger,
		EnableMonitor:     c.EnableMonitor,
		MonitorInterval:   c.MonitorInterval,
		MetricsEnabled:    c.MetricsEnabled,
		EnablePersistence: c.EnablePersistence,
		StoragePath:       c.StoragePath,
		SaveInterval:      c.SaveInterval,
		PanicRecovery:     c.PanicRecovery,
		ErrorRetryDelay:   c.ErrorRetryDelay,
		MaxErrorRetries:   c.MaxErrorRetries,
	}
}

// WorkerPoolConfig 工作池配置
type WorkerPoolConfig struct {
	MaxWorkers     int           // 最大工作协程数
	QueueSize      int           // 队列大小
	IdleTimeout    time.Duration // 空闲超时时间
	MaxIdleWorkers int           // 最大空闲工作协程数
}

// DefaultWorkerPoolConfig 返回默认工作池配置
func DefaultWorkerPoolConfig() *WorkerPoolConfig {
	return &WorkerPoolConfig{
		MaxWorkers:     10,
		QueueSize:      100,
		IdleTimeout:    5 * time.Minute,
		MaxIdleWorkers: 2,
	}
}

// MonitorConfig 监控配置
type MonitorConfig struct {
	Enabled        bool          // 是否启用监控
	CheckInterval  time.Duration // 检查间隔
	MetricsEnabled bool          // 是否启用指标收集
	HealthCheck    bool          // 是否启用健康检查
	AlertThreshold int           // 告警阈值
	NotifyCallback func(string)  // 通知回调函数
}

// DefaultMonitorConfig 返回默认监控配置
func DefaultMonitorConfig() *MonitorConfig {
	return &MonitorConfig{
		Enabled:        true,
		CheckInterval:  10 * time.Second,
		MetricsEnabled: true,
		HealthCheck:    true,
		AlertThreshold: 5, // 连续失败5次触发告警
	}
}

// PersistenceConfig 持久化配置
type PersistenceConfig struct {
	Enabled      bool          // 是否启用持久化
	StoragePath  string        // 存储路径
	SaveInterval time.Duration // 保存间隔
	AutoRestore  bool          // 是否自动恢复
	BackupCount  int           // 备份文件数量
}

// DefaultPersistenceConfig 返回默认持久化配置
func DefaultPersistenceConfig() *PersistenceConfig {
	return &PersistenceConfig{
		Enabled:      false,
		StoragePath:  "./scheduler_data",
		SaveInterval: 5 * time.Minute,
		AutoRestore:  true,
		BackupCount:  3,
	}
}
