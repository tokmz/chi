package scheduler

import (
	"time"

	"chi/pkg/logger"
)

// LoggerConfig 扩展的日志配置
type LoggerConfig struct {
	// 基础配置
	Level         LogLevel `json:"level" yaml:"level"`
	Output        string   `json:"output" yaml:"output"`
	EnableConsole bool     `json:"enable_console" yaml:"enable_console"`
	UseZapLogger  bool     `json:"use_zap_logger" yaml:"use_zap_logger"`

	// 高级配置
	Format      string `json:"format" yaml:"format"`           // json, console
	Development bool   `json:"development" yaml:"development"` // 开发模式

	// 文件输出配置
	File FileLogConfig `json:"file" yaml:"file"`

	// 控制台输出配置
	Console ConsoleLogConfig `json:"console" yaml:"console"`

	// 调用信息配置
	Caller CallerLogConfig `json:"caller" yaml:"caller"`

	// 日志轮转配置
	Rotation RotationLogConfig `json:"rotation" yaml:"rotation"`

	// 采样配置（用于高频日志场景）
	Sampling SamplingLogConfig `json:"sampling" yaml:"sampling"`
}

// FileLogConfig 文件日志配置
type FileLogConfig struct {
	Enabled    bool   `json:"enabled" yaml:"enabled"`
	Filename   string `json:"filename" yaml:"filename"`
	MaxSize    int    `json:"max_size" yaml:"max_size"`       // MB
	MaxBackups int    `json:"max_backups" yaml:"max_backups"` // 备份文件数
	MaxAge     int    `json:"max_age" yaml:"max_age"`         // 保留天数
	Compress   bool   `json:"compress" yaml:"compress"`       // 是否压缩
	LocalTime  bool   `json:"local_time" yaml:"local_time"`   // 使用本地时间
}

// ConsoleLogConfig 控制台日志配置
type ConsoleLogConfig struct {
	Enabled    bool   `json:"enabled" yaml:"enabled"`
	Colorful   bool   `json:"colorful" yaml:"colorful"`     // 彩色输出
	TimeFormat string `json:"time_format" yaml:"time_format"` // 时间格式
}

// CallerLogConfig 调用信息配置
type CallerLogConfig struct {
	Enabled  bool `json:"enabled" yaml:"enabled"`
	FullPath bool `json:"full_path" yaml:"full_path"` // 显示完整路径
	Skip     int  `json:"skip" yaml:"skip"`           // 跳过的调用层数
}

// RotationLogConfig 日志轮转配置
type RotationLogConfig struct {
	Enabled  bool          `json:"enabled" yaml:"enabled"`
	MaxSize  int           `json:"max_size" yaml:"max_size"`   // MB
	Interval time.Duration `json:"interval" yaml:"interval"`   // 轮转间隔
	Pattern  string        `json:"pattern" yaml:"pattern"`     // 文件名模式
}

// SamplingLogConfig 采样配置
type SamplingLogConfig struct {
	Enabled    bool `json:"enabled" yaml:"enabled"`
	Initial    int  `json:"initial" yaml:"initial"`       // 初始采样数
	Thereafter int  `json:"thereafter" yaml:"thereafter"` // 后续采样数
}

// DefaultLoggerConfig 返回默认日志配置
func DefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		Level:         LogLevelInfo,
		Output:        "",
		EnableConsole: true,
		UseZapLogger:  true,
		Format:        "json",
		Development:   false,
		File: FileLogConfig{
			Enabled:    false,
			Filename:   "scheduler.log",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   true,
			LocalTime:  true,
		},
		Console: ConsoleLogConfig{
			Enabled:    true,
			Colorful:   true,
			TimeFormat: "2006-01-02 15:04:05",
		},
		Caller: CallerLogConfig{
			Enabled:  true,
			FullPath: false,
			Skip:     1,
		},
		Rotation: RotationLogConfig{
			Enabled:  false,
			MaxSize:  100,
			Interval: 24 * time.Hour,
			Pattern:  "scheduler-%Y%m%d.log",
		},
		Sampling: SamplingLogConfig{
			Enabled:    false,
			Initial:    100,
			Thereafter: 100,
		},
	}
}

// ToLoggerConfig 转换为logger包的配置
func (c *LoggerConfig) ToLoggerConfig() *logger.Config {
	config := &logger.Config{
		Level:       c.convertLogLevel(),
		Format:      c.Format,
		Development: c.Development,
		Output: logger.OutputConfig{
			Console: logger.ConsoleConfig{
				Enabled:    c.Console.Enabled,
				Colorful:   c.Console.Colorful,
				TimeFormat: c.Console.TimeFormat,
			},
			File: logger.FileConfig{
				Enabled:     c.File.Enabled,
				Filename:    c.File.Filename,
				MaxSize:     c.File.MaxSize,
				MaxBackups:  c.File.MaxBackups,
				MaxAge:      c.File.MaxAge,
				Compress:    c.File.Compress,
				LocalTime:   c.File.LocalTime,
				LevelFilter: c.convertLogLevel(),
			},
		},
		Caller: logger.CallerConfig{
			Enabled:  c.Caller.Enabled,
			FullPath: c.Caller.FullPath,
			Skip:     c.Caller.Skip,
		},
		Sampling: logger.SamplingConfig{
			Enabled:    c.Sampling.Enabled,
			Initial:    c.Sampling.Initial,
			Thereafter: c.Sampling.Thereafter,
		},
	}

	// 如果指定了输出文件，启用文件输出
	if c.Output != "" {
		config.Output.File.Enabled = true
		config.Output.File.Filename = c.Output
	}

	return config
}

// convertLogLevel 转换日志级别
func (c *LoggerConfig) convertLogLevel() string {
	switch c.Level {
	case LogLevelDebug:
		return "debug"
	case LogLevelInfo:
		return "info"
	case LogLevelWarn:
		return "warn"
	case LogLevelError:
		return "error"
	case LogLevelFatal:
		return "fatal"
	default:
		return "info"
	}
}

// Validate 验证配置
func (c *LoggerConfig) Validate() error {
	if c.File.Enabled && c.File.Filename == "" {
		return NewSchedulerError(ErrInvalidConfig, "file output enabled but filename is empty")
	}

	if c.File.MaxSize <= 0 {
		c.File.MaxSize = 100
	}

	if c.File.MaxBackups < 0 {
		c.File.MaxBackups = 0
	}

	if c.File.MaxAge < 0 {
		c.File.MaxAge = 0
	}

	if c.Caller.Skip < 0 {
		c.Caller.Skip = 0
	}

	if c.Sampling.Initial <= 0 {
		c.Sampling.Initial = 100
	}

	if c.Sampling.Thereafter <= 0 {
		c.Sampling.Thereafter = 100
	}

	return nil
}

// Clone 克隆配置
func (c *LoggerConfig) Clone() *LoggerConfig {
	return &LoggerConfig{
		Level:         c.Level,
		Output:        c.Output,
		EnableConsole: c.EnableConsole,
		UseZapLogger:  c.UseZapLogger,
		Format:        c.Format,
		Development:   c.Development,
		File:          c.File,
		Console:       c.Console,
		Caller:        c.Caller,
		Rotation:      c.Rotation,
		Sampling:      c.Sampling,
	}
}

// NewLoggerFromConfig 从扩展配置创建日志器
func NewLoggerFromConfig(config *LoggerConfig) (Logger, error) {
	if config == nil {
		config = DefaultLoggerConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	if !config.UseZapLogger {
		// 使用旧的日志实现
		if config.Output == "" {
			if config.EnableConsole {
				return NewDefaultLogger(config.Level, nil, true), nil
			}
			return NewNoOpLogger(), nil
		}
		return NewFileLogger(config.Level, config.Output, config.EnableConsole)
	}

	// 使用zap日志实现
	loggerConfig := config.ToLoggerConfig()
	return NewLoggerAdapterWithConfig(loggerConfig, config.Level)
}