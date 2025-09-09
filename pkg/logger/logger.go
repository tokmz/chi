package logger

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 日志记录器（优化版本：减少内存分配）
type Logger struct {
	zap    *zap.Logger
	sugar  *zap.SugaredLogger
	config *Config
	mu     sync.RWMutex
	// 性能优化：缓冲池
	bufferPool sync.Pool
}

// Field 日志字段
type Field = zap.Field

// 全局日志实例
var (
	globalLogger *Logger
	once         sync.Once
)

// NewLogger 创建新的日志记录器
func NewLogger(config *Config) (*Logger, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	logger := &Logger{
		config: config,
		bufferPool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, 1024) // 1KB 初始缓冲区
			},
		},
	}

	// 构建zap配置
	zapConfig := logger.buildZapConfig()

	// 创建zap logger
	zapLogger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build zap logger: %w", err)
	}

	logger.zap = zapLogger
	logger.sugar = zapLogger.Sugar()

	return logger, nil
}

// InitGlobal 初始化全局日志记录器
func InitGlobal(config *Config) error {
	var err error
	once.Do(func() {
		globalLogger, err = NewLogger(config)
	})
	return err
}

// GetGlobal 获取全局日志记录器
func GetGlobal() *Logger {
	if globalLogger == nil {
		// 使用默认配置初始化
		_ = InitGlobal(nil)
	}
	return globalLogger
}

// buildZapConfig 构建zap配置
func (l *Logger) buildZapConfig() zap.Config {
	var config zap.Config

	if l.config.Development {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	// 设置日志级别
	config.Level = zap.NewAtomicLevelAt(ParseLevel(l.config.Level))

	// 设置输出格式
	if l.config.Format == "console" {
		config.Encoding = "console"
		config.EncoderConfig = l.buildConsoleEncoderConfig()
	} else {
		config.Encoding = "json"
		config.EncoderConfig = l.buildJSONEncoderConfig()
	}

	// 设置输出路径
	config.OutputPaths = l.buildOutputPaths()
	config.ErrorOutputPaths = l.buildErrorOutputPaths()

	// 设置调用信息
	if l.config.Caller.Enabled {
		config.DisableCaller = false
	} else {
		config.DisableCaller = true
	}

	// 设置堆栈跟踪
	config.DisableStacktrace = false

	// 设置采样
	if l.config.Sampling.Enabled {
		config.Sampling = &zap.SamplingConfig{
			Initial:    l.config.Sampling.Initial,
			Thereafter: l.config.Sampling.Thereafter,
		}
	}

	return config
}

// buildConsoleEncoderConfig 构建控制台编码器配置
func (l *Logger) buildConsoleEncoderConfig() zapcore.EncoderConfig {
	config := zap.NewDevelopmentEncoderConfig()

	// 时间格式
	if l.config.Output.Console.TimeFormat != "" {
		config.EncodeTime = zapcore.TimeEncoderOfLayout(l.config.Output.Console.TimeFormat)
	} else {
		config.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// 日志级别
	if l.config.Output.Console.Colorful {
		config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config.EncodeLevel = zapcore.CapitalLevelEncoder
	}

	// 调用者信息
	if l.config.Caller.FullPath {
		config.EncodeCaller = zapcore.FullCallerEncoder
	} else {
		config.EncodeCaller = zapcore.ShortCallerEncoder
	}

	return config
}

// buildJSONEncoderConfig 构建JSON编码器配置
func (l *Logger) buildJSONEncoderConfig() zapcore.EncoderConfig {
	config := zap.NewProductionEncoderConfig()

	// 时间格式
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	// 日志级别
	config.EncodeLevel = zapcore.LowercaseLevelEncoder

	// 调用者信息
	if l.config.Caller.FullPath {
		config.EncodeCaller = zapcore.FullCallerEncoder
	} else {
		config.EncodeCaller = zapcore.ShortCallerEncoder
	}

	return config
}

// buildOutputPaths 构建输出路径（优化版本：减少字符串分配）
func (l *Logger) buildOutputPaths() []string {
	paths := make([]string, 0, len(l.config.Output.MultiFile)+2) // 预分配容量

	// 控制台输出
	if l.config.Output.Console.Enabled {
		paths = append(paths, "stdout")
	}

	// 文件输出
	if l.config.Output.File.Enabled {
		paths = append(paths, l.config.Output.File.Filename)
	}

	// 多文件输出
	for _, file := range l.config.Output.MultiFile {
		if file.Enabled {
			paths = append(paths, file.Filename)
		}
	}

	if len(paths) == 0 {
		paths = append(paths, "stdout")
	}

	return paths
}

// buildErrorOutputPaths 构建错误输出路径（优化版本：减少字符串分配）
func (l *Logger) buildErrorOutputPaths() []string {
	paths := make([]string, 0, 2) // 预分配容量

	// 控制台输出
	if l.config.Output.Console.Enabled {
		paths = append(paths, "stderr")
	}

	// 文件输出
	if l.config.Output.File.Enabled {
		paths = append(paths, l.config.Output.File.Filename)
	}

	if len(paths) == 0 {
		paths = append(paths, "stderr")
	}

	return paths
}

// Debug 记录调试级别日志
func (l *Logger) Debug(msg string, fields ...Field) {
	l.zap.Debug(msg, fields...)
}

// Info 记录信息级别日志
func (l *Logger) Info(msg string, fields ...Field) {
	l.zap.Info(msg, fields...)
}

// Warn 记录警告级别日志
func (l *Logger) Warn(msg string, fields ...Field) {
	l.zap.Warn(msg, fields...)
}

// Error 记录错误级别日志
func (l *Logger) Error(msg string, fields ...Field) {
	l.zap.Error(msg, fields...)
}

// Panic 记录恐慌级别日志
func (l *Logger) Panic(msg string, fields ...Field) {
	l.zap.Panic(msg, fields...)
}

// Fatal 记录致命级别日志
func (l *Logger) Fatal(msg string, fields ...Field) {
	l.zap.Fatal(msg, fields...)
}

// Debugf 格式化记录调试级别日志
func (l *Logger) Debugf(template string, args ...interface{}) {
	l.sugar.Debugf(template, args...)
}

// Infof 格式化记录信息级别日志
func (l *Logger) Infof(template string, args ...interface{}) {
	l.sugar.Infof(template, args...)
}

// Warnf 格式化记录警告级别日志
func (l *Logger) Warnf(template string, args ...interface{}) {
	l.sugar.Warnf(template, args...)
}

// Errorf 格式化记录错误级别日志
func (l *Logger) Errorf(template string, args ...interface{}) {
	l.sugar.Errorf(template, args...)
}

// Panicf 格式化记录恐慌级别日志
func (l *Logger) Panicf(template string, args ...interface{}) {
	l.sugar.Panicf(template, args...)
}

// Fatalf 格式化记录致命级别日志
func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.sugar.Fatalf(template, args...)
}

// Debugw 键值对记录调试级别日志
func (l *Logger) Debugw(msg string, keysAndValues ...interface{}) {
	l.sugar.Debugw(msg, keysAndValues...)
}

// Infow 键值对记录信息级别日志
func (l *Logger) Infow(msg string, keysAndValues ...interface{}) {
	l.sugar.Infow(msg, keysAndValues...)
}

// Warnw 键值对记录警告级别日志
func (l *Logger) Warnw(msg string, keysAndValues ...interface{}) {
	l.sugar.Warnw(msg, keysAndValues...)
}

// Errorw 键值对记录错误级别日志
func (l *Logger) Errorw(msg string, keysAndValues ...interface{}) {
	l.sugar.Errorw(msg, keysAndValues...)
}

// Panicw 键值对记录恐慌级别日志
func (l *Logger) Panicw(msg string, keysAndValues ...interface{}) {
	l.sugar.Panicw(msg, keysAndValues...)
}

// Fatalw 键值对记录致命级别日志
func (l *Logger) Fatalw(msg string, keysAndValues ...interface{}) {
	l.sugar.Fatalw(msg, keysAndValues...)
}

// With 添加字段
func (l *Logger) With(fields ...Field) *Logger {
	newLogger := &Logger{
		zap:    l.zap.With(fields...),
		config: l.config,
	}
	newLogger.sugar = newLogger.zap.Sugar()
	return newLogger
}

// Named 创建命名子记录器
func (l *Logger) Named(name string) *Logger {
	newLogger := &Logger{
		zap:    l.zap.Named(name),
		config: l.config,
	}
	newLogger.sugar = newLogger.zap.Sugar()
	return newLogger
}

// Sync 同步日志
func (l *Logger) Sync() error {
	return l.zap.Sync()
}

// Close 关闭日志记录器
func (l *Logger) Close() error {
	return l.Sync()
}

// GetZap 获取原始zap logger
func (l *Logger) GetZap() *zap.Logger {
	return l.zap
}

// GetSugar 获取sugar logger
func (l *Logger) GetSugar() *zap.SugaredLogger {
	return l.sugar
}

// SetLevel 动态设置日志级别
func (l *Logger) SetLevel(level string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.config.Level = level
	// 注意：这里需要重新构建logger才能生效
	// 在生产环境中，建议使用zap.AtomicLevel来动态调整级别
}

// GetLevel 获取当前日志级别
func (l *Logger) GetLevel() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.config.Level
}

// 全局便捷函数

// Debug 全局调试日志
func Debug(msg string, fields ...Field) {
	GetGlobal().Debug(msg, fields...)
}

// Info 全局信息日志
func Info(msg string, fields ...Field) {
	GetGlobal().Info(msg, fields...)
}

// Warn 全局警告日志
func Warn(msg string, fields ...Field) {
	GetGlobal().Warn(msg, fields...)
}

// Error 全局错误日志
func Error(msg string, fields ...Field) {
	GetGlobal().Error(msg, fields...)
}

// Panic 全局恐慌日志
func Panic(msg string, fields ...Field) {
	GetGlobal().Panic(msg, fields...)
}

// Fatal 全局致命日志
func Fatal(msg string, fields ...Field) {
	GetGlobal().Fatal(msg, fields...)
}

// Debugf 全局格式化调试日志
func Debugf(template string, args ...interface{}) {
	GetGlobal().Debugf(template, args...)
}

// Infof 全局格式化信息日志
func Infof(template string, args ...interface{}) {
	GetGlobal().Infof(template, args...)
}

// Warnf 全局格式化警告日志
func Warnf(template string, args ...interface{}) {
	GetGlobal().Warnf(template, args...)
}

// Errorf 全局格式化错误日志
func Errorf(template string, args ...interface{}) {
	GetGlobal().Errorf(template, args...)
}

// Panicf 全局格式化恐慌日志
func Panicf(template string, args ...interface{}) {
	GetGlobal().Panicf(template, args...)
}

// Fatalf 全局格式化致命日志
func Fatalf(template string, args ...interface{}) {
	GetGlobal().Fatalf(template, args...)
}

// Debugw 全局键值对调试日志
func Debugw(msg string, keysAndValues ...interface{}) {
	GetGlobal().Debugw(msg, keysAndValues...)
}

// Infow 全局键值对信息日志
func Infow(msg string, keysAndValues ...interface{}) {
	GetGlobal().Infow(msg, keysAndValues...)
}

// Warnw 全局键值对警告日志
func Warnw(msg string, keysAndValues ...interface{}) {
	GetGlobal().Warnw(msg, keysAndValues...)
}

// Errorw 全局键值对错误日志
func Errorw(msg string, keysAndValues ...interface{}) {
	GetGlobal().Errorw(msg, keysAndValues...)
}

// Panicw 全局键值对恐慌日志
func Panicw(msg string, keysAndValues ...interface{}) {
	GetGlobal().Panicw(msg, keysAndValues...)
}

// Fatalw 全局键值对致命日志
func Fatalw(msg string, keysAndValues ...interface{}) {
	GetGlobal().Fatalw(msg, keysAndValues...)
}

// Sync 全局同步日志
func Sync() error {
	return GetGlobal().Sync()
}

// 字段构造函数

// String 字符串字段
func String(key, val string) Field {
	return zap.String(key, val)
}

// Int 整数字段
func Int(key string, val int) Field {
	return zap.Int(key, val)
}

// Int64 64位整数字段
func Int64(key string, val int64) Field {
	return zap.Int64(key, val)
}

// Float64 浮点数字段
func Float64(key string, val float64) Field {
	return zap.Float64(key, val)
}

// Bool 布尔字段
func Bool(key string, val bool) Field {
	return zap.Bool(key, val)
}

// Time 时间字段
func Time(key string, val time.Time) Field {
	return zap.Time(key, val)
}

// Duration 持续时间字段
func Duration(key string, val time.Duration) Field {
	return zap.Duration(key, val)
}

// Error 错误字段
func Err(err error) Field {
	return zap.Error(err)
}

// Any 任意类型字段
func Any(key string, val interface{}) Field {
	return zap.Any(key, val)
}
