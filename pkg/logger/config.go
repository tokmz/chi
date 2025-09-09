package logger

import (
	"time"

	"go.uber.org/zap/zapcore"
)

// Config 日志配置
type Config struct {
	// 日志级别
	Level string `json:"level" yaml:"level" mapstructure:"level"`
	// 输出格式: json, console
	Format string `json:"format" yaml:"format" mapstructure:"format"`
	// 输出目标配置
	Output OutputConfig `json:"output" yaml:"output" mapstructure:"output"`
	// 调用信息配置
	Caller CallerConfig `json:"caller" yaml:"caller" mapstructure:"caller"`
	// 日志分割配置
	Rotation RotationConfig `json:"rotation" yaml:"rotation" mapstructure:"rotation"`
	// 日志管理配置
	Management ManagementConfig `json:"management" yaml:"management" mapstructure:"management"`
	// 开发模式
	Development bool `json:"development" yaml:"development" mapstructure:"development"`
	// 采样配置
	Sampling SamplingConfig `json:"sampling" yaml:"sampling" mapstructure:"sampling"`
}

// OutputConfig 输出配置
type OutputConfig struct {
	// 控制台输出
	Console ConsoleConfig `json:"console" yaml:"console" mapstructure:"console"`
	// 文件输出
	File FileConfig `json:"file" yaml:"file" mapstructure:"file"`
	// 多文件输出
	MultiFile []FileConfig `json:"multi_file" yaml:"multi_file" mapstructure:"multi_file"`
}

// ConsoleConfig 控制台输出配置
type ConsoleConfig struct {
	// 是否启用
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	// 是否启用彩色输出
	Colorful bool `json:"colorful" yaml:"colorful" mapstructure:"colorful"`
	// 时间格式
	TimeFormat string `json:"time_format" yaml:"time_format" mapstructure:"time_format"`
}

// FileConfig 文件输出配置
type FileConfig struct {
	// 是否启用
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	// 文件路径
	Filename string `json:"filename" yaml:"filename" mapstructure:"filename"`
	// 最大文件大小(MB)
	MaxSize int `json:"max_size" yaml:"max_size" mapstructure:"max_size"`
	// 最大备份数量
	MaxBackups int `json:"max_backups" yaml:"max_backups" mapstructure:"max_backups"`
	// 最大保留天数
	MaxAge int `json:"max_age" yaml:"max_age" mapstructure:"max_age"`
	// 是否压缩
	Compress bool `json:"compress" yaml:"compress" mapstructure:"compress"`
	// 本地时间
	LocalTime bool `json:"local_time" yaml:"local_time" mapstructure:"local_time"`
	// 日志级别过滤
	LevelFilter string `json:"level_filter" yaml:"level_filter" mapstructure:"level_filter"`
}

// CallerConfig 调用信息配置
type CallerConfig struct {
	// 是否启用
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	// 是否显示完整路径
	FullPath bool `json:"full_path" yaml:"full_path" mapstructure:"full_path"`
	// 跳过的调用层数
	Skip int `json:"skip" yaml:"skip" mapstructure:"skip"`
}

// RotationConfig 日志分割配置
type RotationConfig struct {
	// 按大小分割
	Size SizeRotationConfig `json:"size" yaml:"size" mapstructure:"size"`
	// 按时间分割
	Time TimeRotationConfig `json:"time" yaml:"time" mapstructure:"time"`
}

// SizeRotationConfig 按大小分割配置
type SizeRotationConfig struct {
	// 是否启用
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	// 最大文件大小(MB)
	MaxSize int `json:"max_size" yaml:"max_size" mapstructure:"max_size"`
}

// TimeRotationConfig 按时间分割配置
type TimeRotationConfig struct {
	// 是否启用
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	// 分割间隔: hour, day, week, month
	Interval string `json:"interval" yaml:"interval" mapstructure:"interval"`
	// 分割时间点(小时:分钟)
	RotateTime string `json:"rotate_time" yaml:"rotate_time" mapstructure:"rotate_time"`
}

// ManagementConfig 日志管理配置
type ManagementConfig struct {
	// 自动清理
	Cleanup CleanupConfig `json:"cleanup" yaml:"cleanup" mapstructure:"cleanup"`
	// 自动压缩
	Compression CompressionConfig `json:"compression" yaml:"compression" mapstructure:"compression"`
}

// CleanupConfig 清理配置
type CleanupConfig struct {
	// 是否启用
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	// 保留天数
	MaxAge int `json:"max_age" yaml:"max_age" mapstructure:"max_age"`
	// 清理间隔
	Interval time.Duration `json:"interval" yaml:"interval" mapstructure:"interval"`
}

// CompressionConfig 压缩配置
type CompressionConfig struct {
	// 是否启用
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	// 压缩延迟(小时)
	Delay int `json:"delay" yaml:"delay" mapstructure:"delay"`
	// 压缩算法: gzip, lz4
	Algorithm string `json:"algorithm" yaml:"algorithm" mapstructure:"algorithm"`
}

// SamplingConfig 采样配置
type SamplingConfig struct {
	// 是否启用
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	// 初始采样数
	Initial int `json:"initial" yaml:"initial" mapstructure:"initial"`
	// 后续采样数
	Thereafter int `json:"thereafter" yaml:"thereafter" mapstructure:"thereafter"`
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Level:       "info",
		Format:      "console",
		Development: false,
		Output: OutputConfig{
			Console: ConsoleConfig{
				Enabled:    true,
				Colorful:   true,
				TimeFormat: "2006-01-02 15:04:05",
			},
			File: FileConfig{
				Enabled:     false,
				Filename:    "logs/app.log",
				MaxSize:     100,
				MaxBackups:  10,
				MaxAge:      30,
				Compress:    true,
				LocalTime:   true,
				LevelFilter: "",
			},
		},
		Caller: CallerConfig{
			Enabled:  true,
			FullPath: false,
			Skip:     1,
		},
		Rotation: RotationConfig{
			Size: SizeRotationConfig{
				Enabled: true,
				MaxSize: 100,
			},
			Time: TimeRotationConfig{
				Enabled:    false,
				Interval:   "day",
				RotateTime: "00:00",
			},
		},
		Management: ManagementConfig{
			Cleanup: CleanupConfig{
				Enabled:  true,
				MaxAge:   30,
				Interval: 24 * time.Hour,
			},
			Compression: CompressionConfig{
				Enabled:   true,
				Delay:     24,
				Algorithm: "gzip",
			},
		},
		Sampling: SamplingConfig{
			Enabled:    false,
			Initial:    100,
			Thereafter: 100,
		},
	}
}

// ParseLevel 解析日志级别
func ParseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	// 验证日志级别
	if c.Level == "" {
		c.Level = "info"
	}

	// 验证输出格式
	if c.Format != "json" && c.Format != "console" {
		c.Format = "console"
	}

	// 验证时间分割间隔
	if c.Rotation.Time.Enabled {
		if c.Rotation.Time.Interval != "hour" &&
			c.Rotation.Time.Interval != "day" &&
			c.Rotation.Time.Interval != "week" &&
			c.Rotation.Time.Interval != "month" {
			c.Rotation.Time.Interval = "day"
		}
	}

	// 验证压缩算法
	if c.Management.Compression.Enabled {
		if c.Management.Compression.Algorithm != "gzip" &&
			c.Management.Compression.Algorithm != "lz4" {
			c.Management.Compression.Algorithm = "gzip"
		}
	}

	return nil
}
