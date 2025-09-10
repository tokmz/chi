package database

import (
	"fmt"
	"time"

	chilogger "chi/pkg/logger"
	gormlogger "gorm.io/gorm/logger"
)

// DatabaseLoggerConfig 数据库日志配置
type DatabaseLoggerConfig struct {
	// 基础配置
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	// 日志级别: debug, info, warn, error
	Level string `json:"level" yaml:"level" mapstructure:"level"`
	// 输出配置
	Output OutputConfig `json:"output" yaml:"output" mapstructure:"output"`
	// GORM日志配置
	GORM GORMLogConfig `json:"gorm" yaml:"gorm" mapstructure:"gorm"`
	// 慢查询配置
	SlowQuery SlowQueryConfig `json:"slow_query" yaml:"slow_query" mapstructure:"slow_query"`
	// 性能监控配置
	Performance PerformanceConfig `json:"performance" yaml:"performance" mapstructure:"performance"`
	// 错误追踪配置
	ErrorTracking ErrorTrackingConfig `json:"error_tracking" yaml:"error_tracking" mapstructure:"error_tracking"`
}

// OutputConfig 输出配置
type OutputConfig struct {
	// 控制台输出
	Console ConsoleOutputConfig `json:"console" yaml:"console" mapstructure:"console"`
	// 文件输出
	File FileOutputConfig `json:"file" yaml:"file" mapstructure:"file"`
	// 结构化日志
	Structured StructuredOutputConfig `json:"structured" yaml:"structured" mapstructure:"structured"`
}

// ConsoleOutputConfig 控制台输出配置
type ConsoleOutputConfig struct {
	// 是否启用
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	// 是否启用彩色输出
	Colorful bool `json:"colorful" yaml:"colorful" mapstructure:"colorful"`
	// 时间格式
	TimeFormat string `json:"time_format" yaml:"time_format" mapstructure:"time_format"`
	// 是否显示调用者信息
	ShowCaller bool `json:"show_caller" yaml:"show_caller" mapstructure:"show_caller"`
}

// FileOutputConfig 文件输出配置
type FileOutputConfig struct {
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
}

// StructuredOutputConfig 结构化输出配置
type StructuredOutputConfig struct {
	// 是否启用
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	// 输出格式: json, logfmt
	Format string `json:"format" yaml:"format" mapstructure:"format"`
	// 字段映射
	FieldMapping map[string]string `json:"field_mapping" yaml:"field_mapping" mapstructure:"field_mapping"`
}

// GORMLogConfig GORM日志配置
type GORMLogConfig struct {
	// 日志级别
	Level gormlogger.LogLevel `json:"level" yaml:"level" mapstructure:"level"`
	// 是否启用彩色输出
	Colorful bool `json:"colorful" yaml:"colorful" mapstructure:"colorful"`
	// 是否忽略记录未找到的错误
	IgnoreRecordNotFoundError bool `json:"ignore_record_not_found_error" yaml:"ignore_record_not_found_error" mapstructure:"ignore_record_not_found_error"`
	// 参数化查询
	ParameterizedQueries bool `json:"parameterized_queries" yaml:"parameterized_queries" mapstructure:"parameterized_queries"`
	// 慢查询阈值
	SlowThreshold time.Duration `json:"slow_threshold" yaml:"slow_threshold" mapstructure:"slow_threshold"`
}



// ErrorTrackingConfig 错误追踪配置
type ErrorTrackingConfig struct {
	// 是否启用
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	// 是否记录堆栈信息
	LogStackTrace bool `json:"log_stack_trace" yaml:"log_stack_trace" mapstructure:"log_stack_trace"`
	// 错误采样率 (0.0-1.0)
	SampleRate float64 `json:"sample_rate" yaml:"sample_rate" mapstructure:"sample_rate"`
	// 错误分类
	ErrorCategories []string `json:"error_categories" yaml:"error_categories" mapstructure:"error_categories"`
}

// DefaultDatabaseLoggerConfig 返回默认的数据库日志配置
func DefaultDatabaseLoggerConfig() *DatabaseLoggerConfig {
	return &DatabaseLoggerConfig{
		Enabled: true,
		Level:   "info",
		Output: OutputConfig{
			Console: ConsoleOutputConfig{
				Enabled:    true,
				Colorful:   true,
				TimeFormat: "2006-01-02 15:04:05",
				ShowCaller: true,
			},
			File: FileOutputConfig{
				Enabled:    false,
				Filename:   "logs/database.log",
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     7,
				Compress:   true,
			},
			Structured: StructuredOutputConfig{
				Enabled: false,
				Format:  "json",
				FieldMapping: map[string]string{
					"timestamp": "@timestamp",
					"level":     "@level",
					"message":   "@message",
				},
			},
		},
		GORM: GORMLogConfig{
			Level:                     gormlogger.Info,
			Colorful:                  true,
			IgnoreRecordNotFoundError: false,
			ParameterizedQueries:      false,
			SlowThreshold:             time.Millisecond * 200,
		},
		SlowQuery: SlowQueryConfig{
			Enabled:   true,
			Threshold: time.Millisecond * 200,
		},
		Performance: PerformanceConfig{
			Enabled:           true,
			Interval:          time.Minute * 5,
			LogConnectionPool: true,
			StatsWindow:       time.Hour,
		},
		ErrorTracking: ErrorTrackingConfig{
			Enabled:         true,
			LogStackTrace:   true,
			SampleRate:      1.0,
			ErrorCategories: []string{"connection", "query", "transaction", "migration"},
		},
	}
}

// Validate 验证配置
func (c *DatabaseLoggerConfig) Validate() error {
	// 验证日志级别
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[c.Level] {
		return fmt.Errorf("invalid log level: %s", c.Level)
	}

	if c.Performance.Enabled {
		if c.Performance.Interval <= 0 {
			c.Performance.Interval = time.Minute * 5
		}
		if c.Performance.StatsWindow <= 0 {
			c.Performance.StatsWindow = time.Minute * 10
		}
	}

	if c.ErrorTracking.Enabled {
		if c.ErrorTracking.SampleRate < 0 || c.ErrorTracking.SampleRate > 1 {
			c.ErrorTracking.SampleRate = 1.0
		}
	}

	return nil
}

// ToLoggerConfig 转换为logger包的配置
func (c *DatabaseLoggerConfig) ToLoggerConfig() *chilogger.Config {
	loggerConfig := chilogger.DefaultConfig()

	// 设置基础配置
	loggerConfig.Level = c.Level
	loggerConfig.Development = c.Level == "debug"

	// 设置输出配置
	loggerConfig.Output.Console.Enabled = c.Output.Console.Enabled
	loggerConfig.Output.Console.Colorful = c.Output.Console.Colorful
	loggerConfig.Output.Console.TimeFormat = c.Output.Console.TimeFormat

	loggerConfig.Output.File.Enabled = c.Output.File.Enabled
	loggerConfig.Output.File.Filename = c.Output.File.Filename
	loggerConfig.Output.File.MaxSize = c.Output.File.MaxSize
	loggerConfig.Output.File.MaxBackups = c.Output.File.MaxBackups
	loggerConfig.Output.File.MaxAge = c.Output.File.MaxAge
	loggerConfig.Output.File.Compress = c.Output.File.Compress

	// 设置调用者信息
	loggerConfig.Caller.Enabled = c.Output.Console.ShowCaller

	return loggerConfig
}