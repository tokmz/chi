package scheduler

import (
	"chi/pkg/logger"
)

// LoggerAdapter 日志适配器，将项目logger包适配到scheduler的Logger接口
type LoggerAdapter struct {
	logger *logger.Logger
	level  LogLevel
}

// NewLoggerAdapter 创建新的日志适配器
func NewLoggerAdapter(l *logger.Logger, level LogLevel) *LoggerAdapter {
	if l == nil {
		l = logger.GetGlobal()
	}
	return &LoggerAdapter{
		logger: l,
		level:  level,
	}
}

// NewLoggerAdapterWithConfig 使用配置创建日志适配器
func NewLoggerAdapterWithConfig(config *logger.Config, level LogLevel) (*LoggerAdapter, error) {
	l, err := logger.NewLogger(config)
	if err != nil {
		return nil, err
	}
	return NewLoggerAdapter(l, level), nil
}

// shouldLog 检查是否应该记录日志
func (a *LoggerAdapter) shouldLog(level LogLevel) bool {
	return level >= a.level
}

// convertFields 将map[string]interface{}转换为zap.Field
func (a *LoggerAdapter) convertFields(fields map[string]interface{}) []logger.Field {
	if len(fields) == 0 {
		return nil
	}
	
	zapFields := make([]logger.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, logger.Any(key, value))
	}
	return zapFields
}

// Debug 记录调试日志
func (a *LoggerAdapter) Debug(message string, fields map[string]interface{}) {
	if !a.shouldLog(LogLevelDebug) {
		return
	}
	a.logger.Debug(message, a.convertFields(fields)...)
}

// Info 记录信息日志
func (a *LoggerAdapter) Info(message string, fields map[string]interface{}) {
	if !a.shouldLog(LogLevelInfo) {
		return
	}
	a.logger.Info(message, a.convertFields(fields)...)
}

// Warn 记录警告日志
func (a *LoggerAdapter) Warn(message string, fields map[string]interface{}) {
	if !a.shouldLog(LogLevelWarn) {
		return
	}
	a.logger.Warn(message, a.convertFields(fields)...)
}

// Error 记录错误日志
func (a *LoggerAdapter) Error(message string, fields map[string]interface{}) {
	if !a.shouldLog(LogLevelError) {
		return
	}
	a.logger.Error(message, a.convertFields(fields)...)
}

// Fatal 记录致命错误日志
func (a *LoggerAdapter) Fatal(message string, fields map[string]interface{}) {
	if !a.shouldLog(LogLevelFatal) {
		return
	}
	a.logger.Fatal(message, a.convertFields(fields)...)
}

// SetLevel 设置日志级别
func (a *LoggerAdapter) SetLevel(level LogLevel) {
	a.level = level
	// 同时设置底层logger的级别
	switch level {
	case LogLevelDebug:
		a.logger.SetLevel("debug")
	case LogLevelInfo:
		a.logger.SetLevel("info")
	case LogLevelWarn:
		a.logger.SetLevel("warn")
	case LogLevelError:
		a.logger.SetLevel("error")
	case LogLevelFatal:
		a.logger.SetLevel("fatal")
	}
}

// GetLevel 获取日志级别
func (a *LoggerAdapter) GetLevel() LogLevel {
	return a.level
}

// GetUnderlyingLogger 获取底层的logger实例
func (a *LoggerAdapter) GetUnderlyingLogger() *logger.Logger {
	return a.logger
}

// Sync 同步日志缓冲区
func (a *LoggerAdapter) Sync() error {
	return a.logger.Sync()
}

// Close 关闭日志器
func (a *LoggerAdapter) Close() error {
	return a.logger.Close()
}

// ZapLoggerAdapter 直接使用zap logger的适配器（用于高性能场景）
type ZapLoggerAdapter struct {
	logger *logger.Logger
	level  LogLevel
}

// NewZapLoggerAdapter 创建zap logger适配器
func NewZapLoggerAdapter(l *logger.Logger, level LogLevel) *ZapLoggerAdapter {
	if l == nil {
		l = logger.GetGlobal()
	}
	return &ZapLoggerAdapter{
		logger: l,
		level:  level,
	}
}

// shouldLog 检查是否应该记录日志
func (z *ZapLoggerAdapter) shouldLog(level LogLevel) bool {
	return level >= z.level
}

// Debug 记录调试日志（高性能版本）
func (z *ZapLoggerAdapter) Debug(message string, fields map[string]interface{}) {
	if !z.shouldLog(LogLevelDebug) {
		return
	}
	
	// 直接使用zap logger以获得更好的性能
	zapLogger := z.logger.GetZap()
	if len(fields) == 0 {
		zapLogger.Debug(message)
		return
	}
	
	zapFields := make([]logger.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, logger.Any(key, value))
	}
	zapLogger.Debug(message, zapFields...)
}

// Info 记录信息日志（高性能版本）
func (z *ZapLoggerAdapter) Info(message string, fields map[string]interface{}) {
	if !z.shouldLog(LogLevelInfo) {
		return
	}
	
	zapLogger := z.logger.GetZap()
	if len(fields) == 0 {
		zapLogger.Info(message)
		return
	}
	
	zapFields := make([]logger.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, logger.Any(key, value))
	}
	zapLogger.Info(message, zapFields...)
}

// Warn 记录警告日志（高性能版本）
func (z *ZapLoggerAdapter) Warn(message string, fields map[string]interface{}) {
	if !z.shouldLog(LogLevelWarn) {
		return
	}
	
	zapLogger := z.logger.GetZap()
	if len(fields) == 0 {
		zapLogger.Warn(message)
		return
	}
	
	zapFields := make([]logger.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, logger.Any(key, value))
	}
	zapLogger.Warn(message, zapFields...)
}

// Error 记录错误日志（高性能版本）
func (z *ZapLoggerAdapter) Error(message string, fields map[string]interface{}) {
	if !z.shouldLog(LogLevelError) {
		return
	}
	
	zapLogger := z.logger.GetZap()
	if len(fields) == 0 {
		zapLogger.Error(message)
		return
	}
	
	zapFields := make([]logger.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, logger.Any(key, value))
	}
	zapLogger.Error(message, zapFields...)
}

// Fatal 记录致命错误日志（高性能版本）
func (z *ZapLoggerAdapter) Fatal(message string, fields map[string]interface{}) {
	if !z.shouldLog(LogLevelFatal) {
		return
	}
	
	zapLogger := z.logger.GetZap()
	if len(fields) == 0 {
		zapLogger.Fatal(message)
		return
	}
	
	zapFields := make([]logger.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, logger.Any(key, value))
	}
	zapLogger.Fatal(message, zapFields...)
}

// SetLevel 设置日志级别
func (z *ZapLoggerAdapter) SetLevel(level LogLevel) {
	z.level = level
	// 同时设置底层logger的级别
	switch level {
	case LogLevelDebug:
		z.logger.SetLevel("debug")
	case LogLevelInfo:
		z.logger.SetLevel("info")
	case LogLevelWarn:
		z.logger.SetLevel("warn")
	case LogLevelError:
		z.logger.SetLevel("error")
	case LogLevelFatal:
		z.logger.SetLevel("fatal")
	}
}

// GetLevel 获取日志级别
func (z *ZapLoggerAdapter) GetLevel() LogLevel {
	return z.level
}