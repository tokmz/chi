package scheduler

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"chi/pkg/logger"
)

// LogLevel 日志级别
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// String 返回日志级别字符串
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger 日志接口
type Logger interface {
	Debug(message string, fields map[string]interface{})
	Info(message string, fields map[string]interface{})
	Warn(message string, fields map[string]interface{})
	Error(message string, fields map[string]interface{})
	Fatal(message string, fields map[string]interface{})
	SetLevel(level LogLevel)
	GetLevel() LogLevel
}

// DefaultLogger 默认日志实现
type DefaultLogger struct {
	level         LogLevel
	logger        *log.Logger
	output        io.Writer
	enableConsole bool
}

// NewDefaultLogger 创建默认日志器
func NewDefaultLogger(level LogLevel, output io.Writer, enableConsole bool) *DefaultLogger {
	if output == nil {
		output = os.Stdout
	}

	return &DefaultLogger{
		level:         level,
		logger:        log.New(output, "", 0),
		output:        output,
		enableConsole: enableConsole,
	}
}

// NewFileLogger 创建文件日志器
func NewFileLogger(level LogLevel, filePath string, enableConsole bool) (*DefaultLogger, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	var output io.Writer = file
	if enableConsole {
		output = io.MultiWriter(file, os.Stdout)
	}

	return NewDefaultLogger(level, output, enableConsole), nil
}

// formatMessage 格式化日志消息
func (l *DefaultLogger) formatMessage(level LogLevel, message string, fields map[string]interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMsg := fmt.Sprintf("[%s] %s %s", timestamp, level.String(), message)

	if len(fields) > 0 {
		logMsg += " |"
		for key, value := range fields {
			logMsg += fmt.Sprintf(" %s=%v", key, value)
		}
	}

	return logMsg
}

// shouldLog 检查是否应该记录日志
func (l *DefaultLogger) shouldLog(level LogLevel) bool {
	return level >= l.level
}

// log 记录日志
func (l *DefaultLogger) log(level LogLevel, message string, fields map[string]interface{}) {
	if !l.shouldLog(level) {
		return
	}

	formattedMsg := l.formatMessage(level, message, fields)
	l.logger.Println(formattedMsg)
}

// Debug 记录调试日志
func (l *DefaultLogger) Debug(message string, fields map[string]interface{}) {
	l.log(LogLevelDebug, message, fields)
}

// Info 记录信息日志
func (l *DefaultLogger) Info(message string, fields map[string]interface{}) {
	l.log(LogLevelInfo, message, fields)
}

// Warn 记录警告日志
func (l *DefaultLogger) Warn(message string, fields map[string]interface{}) {
	l.log(LogLevelWarn, message, fields)
}

// Error 记录错误日志
func (l *DefaultLogger) Error(message string, fields map[string]interface{}) {
	l.log(LogLevelError, message, fields)
}

// Fatal 记录致命错误日志
func (l *DefaultLogger) Fatal(message string, fields map[string]interface{}) {
	l.log(LogLevelFatal, message, fields)
	os.Exit(1)
}

// SetLevel 设置日志级别
func (l *DefaultLogger) SetLevel(level LogLevel) {
	l.level = level
}

// GetLevel 获取日志级别
func (l *DefaultLogger) GetLevel() LogLevel {
	return l.level
}

// NoOpLogger 空操作日志器（用于禁用日志）
type NoOpLogger struct{}

// NewNoOpLogger 创建空操作日志器
func NewNoOpLogger() *NoOpLogger {
	return &NoOpLogger{}
}

// Debug 空操作
func (l *NoOpLogger) Debug(message string, fields map[string]interface{}) {}

// Info 空操作
func (l *NoOpLogger) Info(message string, fields map[string]interface{}) {}

// Warn 空操作
func (l *NoOpLogger) Warn(message string, fields map[string]interface{}) {}

// Error 空操作
func (l *NoOpLogger) Error(message string, fields map[string]interface{}) {}

// Fatal 空操作
func (l *NoOpLogger) Fatal(message string, fields map[string]interface{}) {}

// SetLevel 空操作
func (l *NoOpLogger) SetLevel(level LogLevel) {}

// GetLevel 返回调试级别
func (l *NoOpLogger) GetLevel() LogLevel {
	return LogLevelDebug
}

// LoggerFactory 日志工厂
type LoggerFactory struct{}

// CreateLogger 创建日志器
func (f *LoggerFactory) CreateLogger(config *SchedulerConfig) (Logger, error) {
	// 优先使用新的zap logger
	if config.UseZapLogger {
		return f.createZapLogger(config)
	}

	// 兼容旧的日志实现
	if config.LogOutput == "" {
		// 只输出到控制台或不输出
		if config.EnableConsole {
			return NewDefaultLogger(config.LogLevel, os.Stdout, true), nil
		}
		return NewNoOpLogger(), nil
	}

	// 输出到文件
	return NewFileLogger(config.LogLevel, config.LogOutput, config.EnableConsole)
}

// createZapLogger 创建基于zap的日志器
func (f *LoggerFactory) createZapLogger(config *SchedulerConfig) (Logger, error) {
	// 如果有扩展日志配置，优先使用
	if config.LoggerConfig != nil {
		return NewLoggerFromConfig(config.LoggerConfig)
	}

	// 使用基础配置创建logger配置
	loggerConfig := &logger.Config{
		Level:       f.convertLogLevel(config.LogLevel),
		Format:      "json",
		Development: false,
		Output: logger.OutputConfig{
			Console: logger.ConsoleConfig{
				Enabled:    config.EnableConsole,
				Colorful:   true,
				TimeFormat: "2006-01-02 15:04:05",
			},
			File: logger.FileConfig{
				Enabled:  config.LogOutput != "",
				Filename: config.LogOutput,
				MaxSize:  100, // 100MB
				MaxBackups: 3,
				MaxAge:   7, // 7天
				Compress: true,
				LocalTime: true,
			},
		},
		Caller: logger.CallerConfig{
			Enabled:  true,
			FullPath: false,
			Skip:     1,
		},
	}

	// 创建适配器
	return NewLoggerAdapterWithConfig(loggerConfig, config.LogLevel)
}

// convertLogLevel 转换日志级别
func (f *LoggerFactory) convertLogLevel(level LogLevel) string {
	switch level {
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
