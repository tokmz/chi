package mongo

import (
	"chi/pkg/logger"
)

// MongoLoggerAdapter MongoDB日志适配器，将项目logger包适配到mongo的Logger接口
type MongoLoggerAdapter struct {
	logger *logger.Logger
	level  LogLevel
	config *MongoLoggerConfig
}

// NewMongoLoggerAdapter 创建新的MongoDB日志适配器
func NewMongoLoggerAdapter(l *logger.Logger, level LogLevel) *MongoLoggerAdapter {
	if l == nil {
		l = logger.GetGlobal()
	}
	return &MongoLoggerAdapter{
		logger: l,
		level:  level,
	}
}

// NewMongoLoggerAdapterWithConfig 使用配置创建MongoDB日志适配器
func NewMongoLoggerAdapterWithConfig(config *MongoLoggerConfig) (*MongoLoggerAdapter, error) {
	if config == nil {
		config = DefaultMongoLoggerConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	loggerConfig := config.ToLoggerConfig()
	l, err := logger.NewLogger(loggerConfig)
	if err != nil {
		return nil, err
	}

	return &MongoLoggerAdapter{
		logger: l,
		level:  config.Level,
		config: config,
	}, nil
}

// shouldLog 检查是否应该记录日志
func (a *MongoLoggerAdapter) shouldLog(level LogLevel) bool {
	return level >= a.level
}

// convertFields 将interface{}转换为logger.Field
func (a *MongoLoggerAdapter) convertFields(fields ...interface{}) []logger.Field {
	if len(fields) == 0 {
		return nil
	}

	loggerFields := make([]logger.Field, 0, len(fields)/2)
	for i := 0; i < len(fields)-1; i += 2 {
		if key, ok := fields[i].(string); ok {
			loggerFields = append(loggerFields, logger.Any(key, fields[i+1]))
		}
	}
	return loggerFields
}

// Debug 记录调试日志
func (a *MongoLoggerAdapter) Debug(msg string, fields ...interface{}) {
	if !a.shouldLog(LogLevelDebug) {
		return
	}
	a.logger.Debug(msg, a.convertFields(fields...)...)
}

// Info 记录信息日志
func (a *MongoLoggerAdapter) Info(msg string, fields ...interface{}) {
	if !a.shouldLog(LogLevelInfo) {
		return
	}
	a.logger.Info(msg, a.convertFields(fields...)...)
}

// Warn 记录警告日志
func (a *MongoLoggerAdapter) Warn(msg string, fields ...interface{}) {
	if !a.shouldLog(LogLevelWarn) {
		return
	}
	a.logger.Warn(msg, a.convertFields(fields...)...)
}

// Error 记录错误日志
func (a *MongoLoggerAdapter) Error(msg string, fields ...interface{}) {
	if !a.shouldLog(LogLevelError) {
		return
	}
	a.logger.Error(msg, a.convertFields(fields...)...)
}

// Fatal 记录致命错误日志
func (a *MongoLoggerAdapter) Fatal(msg string, fields ...interface{}) {
	if !a.shouldLog(LogLevelFatal) {
		return
	}
	a.logger.Fatal(msg, a.convertFields(fields...)...)
}

// SetLevel 设置日志级别
func (a *MongoLoggerAdapter) SetLevel(level LogLevel) {
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
func (a *MongoLoggerAdapter) GetLevel() LogLevel {
	return a.level
}

// GetUnderlyingLogger 获取底层logger实例
func (a *MongoLoggerAdapter) GetUnderlyingLogger() *logger.Logger {
	return a.logger
}

// GetConfig 获取配置
func (a *MongoLoggerAdapter) GetConfig() *MongoLoggerConfig {
	return a.config
}

// Sync 同步日志
func (a *MongoLoggerAdapter) Sync() error {
	return a.logger.Sync()
}

// Close 关闭日志记录器
func (a *MongoLoggerAdapter) Close() error {
	return a.logger.Close()
}

// EnhancedSlowQueryLogger 增强的慢查询日志记录器
type EnhancedSlowQueryLogger struct {
	logger    Logger
	config    SlowQueryLogConfig
	enabled   bool
}

// NewEnhancedSlowQueryLogger 创建增强的慢查询日志记录器
func NewEnhancedSlowQueryLogger(logger Logger, config SlowQueryLogConfig) *EnhancedSlowQueryLogger {
	return &EnhancedSlowQueryLogger{
		logger:  logger,
		config:  config,
		enabled: config.Enabled,
	}
}

// LogSlowQuery 记录慢查询
func (s *EnhancedSlowQueryLogger) LogSlowQuery(operation string, duration int64, collection string, filter interface{}, result interface{}) {
	if !s.enabled || duration < s.config.Threshold.Nanoseconds()/1000000 {
		return
	}

	fields := []interface{}{
		"operation", operation,
		"duration_ms", duration,
		"collection", collection,
	}

	if s.config.LogQuery && filter != nil {
		fields = append(fields, "filter", filter)
	}

	if s.config.LogResult && result != nil {
		fields = append(fields, "result", result)
	}

	s.logger.Warn("Slow query detected", fields...)
}

// IsEnabled 检查是否启用
func (s *EnhancedSlowQueryLogger) IsEnabled() bool {
	return s.enabled
}

// Enable 启用慢查询日志
func (s *EnhancedSlowQueryLogger) Enable() {
	s.enabled = true
}

// Disable 禁用慢查询日志
func (s *EnhancedSlowQueryLogger) Disable() {
	s.enabled = false
}

// UpdateConfig 更新配置
func (s *EnhancedSlowQueryLogger) UpdateConfig(config SlowQueryLogConfig) {
	s.config = config
	s.enabled = config.Enabled
}

// GetConfig 获取配置
func (s *EnhancedSlowQueryLogger) GetConfig() SlowQueryLogConfig {
	return s.config
}

// ConnectionLogger 连接日志记录器
type ConnectionLogger struct {
	logger Logger
	config ConnectionLogConfig
}

// NewConnectionLogger 创建连接日志记录器
func NewConnectionLogger(logger Logger, config ConnectionLogConfig) *ConnectionLogger {
	return &ConnectionLogger{
		logger: logger,
		config: config,
	}
}

// LogConnect 记录连接建立
func (c *ConnectionLogger) LogConnect(uri string, database string) {
	if !c.config.Enabled || !c.config.LogConnect {
		return
	}
	c.logger.Info("MongoDB connection established", "uri", uri, "database", database)
}

// LogClose 记录连接关闭
func (c *ConnectionLogger) LogClose() {
	if !c.config.Enabled || !c.config.LogClose {
		return
	}
	c.logger.Info("MongoDB connection closed")
}

// LogPoolInfo 记录连接池信息
func (c *ConnectionLogger) LogPoolInfo(poolSize int, activeConns int, idleConns int) {
	if !c.config.Enabled || !c.config.LogPoolInfo {
		return
	}
	c.logger.Debug("MongoDB connection pool info",
		"pool_size", poolSize,
		"active_connections", activeConns,
		"idle_connections", idleConns,
	)
}

// OperationLogger 操作日志记录器
type OperationLogger struct {
	logger Logger
	config OperationLogConfig
}

// NewOperationLogger 创建操作日志记录器
func NewOperationLogger(logger Logger, config OperationLogConfig) *OperationLogger {
	return &OperationLogger{
		logger: logger,
		config: config,
	}
}

// LogCRUD 记录CRUD操作
func (o *OperationLogger) LogCRUD(operation string, collection string, filter interface{}, document interface{}) {
	if !o.config.Enabled || !o.config.LogCRUD {
		return
	}
	o.logger.Debug("MongoDB CRUD operation",
		"operation", operation,
		"collection", collection,
		"filter", filter,
		"document", document,
	)
}

// LogTransaction 记录事务操作
func (o *OperationLogger) LogTransaction(operation string, sessionID string) {
	if !o.config.Enabled || !o.config.LogTransaction {
		return
	}
	o.logger.Info("MongoDB transaction operation",
		"operation", operation,
		"session_id", sessionID,
	)
}

// LogAggregation 记录聚合操作
func (o *OperationLogger) LogAggregation(collection string, pipeline interface{}) {
	if !o.config.Enabled || !o.config.LogAggregation {
		return
	}
	o.logger.Debug("MongoDB aggregation operation",
		"collection", collection,
		"pipeline", pipeline,
	)
}

// LogIndex 记录索引操作
func (o *OperationLogger) LogIndex(operation string, collection string, index interface{}) {
	if !o.config.Enabled || !o.config.LogIndex {
		return
	}
	o.logger.Info("MongoDB index operation",
		"operation", operation,
		"collection", collection,
		"index", index,
	)
}