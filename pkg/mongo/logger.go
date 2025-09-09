package mongo

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// Logger 日志记录器接口
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
}

// DefaultLogger 默认日志记录器
type DefaultLogger struct {
	level  LogLevel
	logger *log.Logger
}

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

// ParseLogLevel 解析日志级别
func ParseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return LogLevelDebug
	case "INFO":
		return LogLevelInfo
	case "WARN", "WARNING":
		return LogLevelWarn
	case "ERROR":
		return LogLevelError
	case "FATAL":
		return LogLevelFatal
	default:
		return LogLevelInfo
	}
}

// NewLogger 创建新的日志记录器
func NewLogger(config LogConfig) Logger {
	if !config.Enabled {
		return &NoOpLogger{}
	}

	level := ParseLogLevel(config.Level)
	logger := log.New(os.Stdout, "", log.LstdFlags)

	return &DefaultLogger{
		level:  level,
		logger: logger,
	}
}

// Debug 记录调试日志
func (l *DefaultLogger) Debug(msg string, fields ...interface{}) {
	if l.level <= LogLevelDebug {
		l.log(LogLevelDebug, msg, fields...)
	}
}

// Info 记录信息日志
func (l *DefaultLogger) Info(msg string, fields ...interface{}) {
	if l.level <= LogLevelInfo {
		l.log(LogLevelInfo, msg, fields...)
	}
}

// Warn 记录警告日志
func (l *DefaultLogger) Warn(msg string, fields ...interface{}) {
	if l.level <= LogLevelWarn {
		l.log(LogLevelWarn, msg, fields...)
	}
}

// Error 记录错误日志
func (l *DefaultLogger) Error(msg string, fields ...interface{}) {
	if l.level <= LogLevelError {
		l.log(LogLevelError, msg, fields...)
	}
}

// Fatal 记录致命错误日志
func (l *DefaultLogger) Fatal(msg string, fields ...interface{}) {
	l.log(LogLevelFatal, msg, fields...)
	os.Exit(1)
}

// log 内部日志记录方法
func (l *DefaultLogger) log(level LogLevel, msg string, fields ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelStr := level.String()

	// 构建字段字符串
	var fieldStr string
	if len(fields) > 0 {
		var parts []string
		for i := 0; i < len(fields); i += 2 {
			if i+1 < len(fields) {
				parts = append(parts, fmt.Sprintf("%v=%v", fields[i], fields[i+1]))
			}
		}
		if len(parts) > 0 {
			fieldStr = " [" + strings.Join(parts, ", ") + "]"
		}
	}

	logMsg := fmt.Sprintf("[%s] [%s] %s%s", timestamp, levelStr, msg, fieldStr)
	l.logger.Println(logMsg)
}

// NoOpLogger 空操作日志记录器
type NoOpLogger struct{}

// Debug 空操作
func (n *NoOpLogger) Debug(msg string, fields ...interface{}) {}

// Info 空操作
func (n *NoOpLogger) Info(msg string, fields ...interface{}) {}

// Warn 空操作
func (n *NoOpLogger) Warn(msg string, fields ...interface{}) {}

// Error 空操作
func (n *NoOpLogger) Error(msg string, fields ...interface{}) {}

// Fatal 空操作
func (n *NoOpLogger) Fatal(msg string, fields ...interface{}) {}

// SlowQueryLogger 慢查询日志记录器
type SlowQueryLogger struct {
	logger    Logger
	threshold time.Duration
	enabled   bool
}

// NewSlowQueryLogger 创建慢查询日志记录器
func NewSlowQueryLogger(logger Logger, threshold time.Duration, enabled bool) *SlowQueryLogger {
	return &SlowQueryLogger{
		logger:    logger,
		threshold: threshold,
		enabled:   enabled,
	}
}

// LogSlowQuery 记录慢查询
func (s *SlowQueryLogger) LogSlowQuery(operation string, duration time.Duration, collection string, filter interface{}) {
	if !s.enabled || duration < s.threshold {
		return
	}

	s.logger.Warn("Slow query detected",
		"operation", operation,
		"duration", duration.String(),
		"collection", collection,
		"filter", fmt.Sprintf("%+v", filter),
	)
}

// IsEnabled 检查是否启用慢查询日志
func (s *SlowQueryLogger) IsEnabled() bool {
	return s.enabled
}

// GetThreshold 获取慢查询阈值
func (s *SlowQueryLogger) GetThreshold() time.Duration {
	return s.threshold
}

// SetThreshold 设置慢查询阈值
func (s *SlowQueryLogger) SetThreshold(threshold time.Duration) {
	s.threshold = threshold
}

// Enable 启用慢查询日志
func (s *SlowQueryLogger) Enable() {
	s.enabled = true
}

// Disable 禁用慢查询日志
func (s *SlowQueryLogger) Disable() {
	s.enabled = false
}