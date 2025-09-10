package mongo

import (
	"time"

	"chi/pkg/logger"
)

// MongoLoggerConfig 扩展的MongoDB日志配置
type MongoLoggerConfig struct {
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

	// MongoDB特定配置
	Mongo MongoSpecificLogConfig `json:"mongo" yaml:"mongo"`
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
	Colorful   bool   `json:"colorful" yaml:"colorful"`       // 彩色输出
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

// MongoSpecificLogConfig MongoDB特定日志配置
type MongoSpecificLogConfig struct {
	// 慢查询日志
	SlowQuery SlowQueryLogConfig `json:"slow_query" yaml:"slow_query"`
	// 连接日志
	Connection ConnectionLogConfig `json:"connection" yaml:"connection"`
	// 操作日志
	Operation OperationLogConfig `json:"operation" yaml:"operation"`
	// 错误日志
	Error ErrorLogConfig `json:"error" yaml:"error"`
}

// SlowQueryLogConfig 慢查询日志配置
type SlowQueryLogConfig struct {
	Enabled   bool          `json:"enabled" yaml:"enabled"`
	Threshold time.Duration `json:"threshold" yaml:"threshold"` // 慢查询阈值
	LogQuery  bool          `json:"log_query" yaml:"log_query"` // 是否记录查询语句
	LogResult bool          `json:"log_result" yaml:"log_result"` // 是否记录结果统计
}

// ConnectionLogConfig 连接日志配置
type ConnectionLogConfig struct {
	Enabled     bool `json:"enabled" yaml:"enabled"`
	LogConnect  bool `json:"log_connect" yaml:"log_connect"`   // 记录连接建立
	LogClose    bool `json:"log_close" yaml:"log_close"`       // 记录连接关闭
	LogPoolInfo bool `json:"log_pool_info" yaml:"log_pool_info"` // 记录连接池信息
}

// OperationLogConfig 操作日志配置
type OperationLogConfig struct {
	Enabled       bool `json:"enabled" yaml:"enabled"`
	LogCRUD       bool `json:"log_crud" yaml:"log_crud"`             // 记录CRUD操作
	LogAggregation bool `json:"log_aggregation" yaml:"log_aggregation"` // 记录聚合操作
	LogTransaction bool `json:"log_transaction" yaml:"log_transaction"` // 记录事务操作
	LogIndex      bool `json:"log_index" yaml:"log_index"`           // 记录索引操作
}

// ErrorLogConfig 错误日志配置
type ErrorLogConfig struct {
	Enabled      bool `json:"enabled" yaml:"enabled"`
	LogStackTrace bool `json:"log_stack_trace" yaml:"log_stack_trace"` // 记录堆栈跟踪
	LogContext   bool `json:"log_context" yaml:"log_context"`         // 记录上下文信息
}

// DefaultMongoLoggerConfig 返回默认MongoDB日志配置
func DefaultMongoLoggerConfig() *MongoLoggerConfig {
	return &MongoLoggerConfig{
		Level:         LogLevelInfo,
		Output:        "",
		EnableConsole: true,
		UseZapLogger:  true,
		Format:        "json",
		Development:   false,
		File: FileLogConfig{
			Enabled:    false,
			Filename:   "mongo.log",
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
			Enabled:  false,
			FullPath: false,
			Skip:     1,
		},
		Rotation: RotationLogConfig{
			Enabled:  false,
			MaxSize:  100,
			Interval: 24 * time.Hour,
			Pattern:  "mongo-%Y%m%d.log",
		},
		Sampling: SamplingLogConfig{
			Enabled:    false,
			Initial:    100,
			Thereafter: 100,
		},
		Mongo: MongoSpecificLogConfig{
			SlowQuery: SlowQueryLogConfig{
				Enabled:   true,
				Threshold: 100 * time.Millisecond,
				LogQuery:  true,
				LogResult: true,
			},
			Connection: ConnectionLogConfig{
				Enabled:     true,
				LogConnect:  true,
				LogClose:    true,
				LogPoolInfo: false,
			},
			Operation: OperationLogConfig{
				Enabled:       false,
				LogCRUD:       false,
				LogAggregation: false,
				LogTransaction: true,
				LogIndex:      true,
			},
			Error: ErrorLogConfig{
				Enabled:      true,
				LogStackTrace: true,
				LogContext:   true,
			},
		},
	}
}

// ToLoggerConfig 转换为项目logger包的配置
func (c *MongoLoggerConfig) ToLoggerConfig() *logger.Config {
	config := &logger.Config{
		Level:       c.convertLogLevel(),
		Format:      c.Format,
		Development: c.Development,
		Output: logger.OutputConfig{
			Console: logger.ConsoleConfig{
				Enabled:    c.EnableConsole,
				Colorful:   c.Console.Colorful,
				TimeFormat: c.Console.TimeFormat,
			},
			File: logger.FileConfig{
				Enabled:    c.File.Enabled,
				Filename:   c.File.Filename,
				MaxSize:    c.File.MaxSize,
				MaxBackups: c.File.MaxBackups,
				MaxAge:     c.File.MaxAge,
				Compress:   c.File.Compress,
				LocalTime:  c.File.LocalTime,
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

	return config
}

// convertLogLevel 转换日志级别
func (c *MongoLoggerConfig) convertLogLevel() string {
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
func (c *MongoLoggerConfig) Validate() error {
	// 验证日志级别
	if c.Level < LogLevelDebug || c.Level > LogLevelFatal {
		return ErrInvalidLogLevel
	}

	// 验证输出配置
	if !c.EnableConsole && !c.File.Enabled {
		return ErrNoLogOutput
	}

	// 验证文件配置
	if c.File.Enabled {
		if c.File.Filename == "" {
			return ErrInvalidLogFile
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
	}

	// 验证慢查询配置
	if c.Mongo.SlowQuery.Enabled && c.Mongo.SlowQuery.Threshold <= 0 {
		c.Mongo.SlowQuery.Threshold = 100 * time.Millisecond
	}

	return nil
}

// Clone 克隆配置
func (c *MongoLoggerConfig) Clone() *MongoLoggerConfig {
	cloned := *c
	return &cloned
}

// NewMongoLoggerFromConfig 从配置创建MongoDB日志记录器
func NewMongoLoggerFromConfig(config *MongoLoggerConfig) (Logger, error) {
	if config == nil {
		config = DefaultMongoLoggerConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	if config.UseZapLogger {
		// 使用项目logger包
		return NewMongoLoggerAdapterWithConfig(config)
	} else {
		// 使用默认logger
		return NewLogger(LogConfig{
			Enabled: true,
			Level:   config.convertLogLevel(),
		}), nil
	}
}