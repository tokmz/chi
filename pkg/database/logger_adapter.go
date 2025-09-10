package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	chilogger "chi/pkg/logger"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"gorm.io/gorm"
)

// DatabaseLoggerAdapter 数据库日志适配器
type DatabaseLoggerAdapter struct {
	logger       *chilogger.Logger
	config       *DatabaseLoggerConfig
	gormConfig   gormlogger.Config
	slowMonitor  *DatabaseSlowQueryMonitor
	perfMonitor  *DatabasePerformanceMonitor
}

// NewDatabaseLoggerAdapter 创建数据库日志适配器
func NewDatabaseLoggerAdapter(config *DatabaseLoggerConfig) (*DatabaseLoggerAdapter, error) {
	if config == nil {
		config = DefaultDatabaseLoggerConfig()
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 创建logger实例
	loggerConfig := config.ToLoggerConfig()
	logger, err := chilogger.NewLogger(loggerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	adapter := &DatabaseLoggerAdapter{
		logger: logger,
		config: config,
		gormConfig: gormlogger.Config{
			SlowThreshold:             config.GORM.SlowThreshold,
			LogLevel:                  config.GORM.Level,
			IgnoreRecordNotFoundError: config.GORM.IgnoreRecordNotFoundError,
			ParameterizedQueries:      config.GORM.ParameterizedQueries,
			Colorful:                  config.GORM.Colorful,
		},
	}

	// 初始化慢查询监控器
	if config.SlowQuery.Enabled {
		adapter.slowMonitor = NewDatabaseSlowQueryMonitor(config.SlowQuery, logger)
	}

	// 初始化性能监控器
	if config.Performance.Enabled {
		adapter.perfMonitor = NewDatabasePerformanceMonitor(config.Performance, logger)
	}

	return adapter, nil
}

// LogMode 设置日志模式
func (a *DatabaseLoggerAdapter) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newAdapter := *a
	newAdapter.gormConfig.LogLevel = level
	return &newAdapter
}

// Info 输出信息日志
func (a *DatabaseLoggerAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	if a.gormConfig.LogLevel >= gormlogger.Info {
		fields := []chilogger.Field{
			chilogger.String("caller", utils.FileWithLineNum()),
			chilogger.String("component", "database"),
		}
		
		if len(data) > 0 {
			fields = append(fields, chilogger.Any("data", data))
		}
		
		a.logger.Info(msg, fields...)
	}
}

// Warn 输出警告日志
func (a *DatabaseLoggerAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	if a.gormConfig.LogLevel >= gormlogger.Warn {
		fields := []chilogger.Field{
			chilogger.String("caller", utils.FileWithLineNum()),
			chilogger.String("component", "database"),
		}
		
		if len(data) > 0 {
			fields = append(fields, chilogger.Any("data", data))
		}
		
		a.logger.Warn(msg, fields...)
	}
}

// Error 输出错误日志
func (a *DatabaseLoggerAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	if a.gormConfig.LogLevel >= gormlogger.Error {
		fields := []chilogger.Field{
			chilogger.String("caller", utils.FileWithLineNum()),
			chilogger.String("component", "database"),
		}
		
		if len(data) > 0 {
			fields = append(fields, chilogger.Any("data", data))
		}
		
		a.logger.Error(msg, fields...)
	}
}

// Trace 输出SQL追踪日志
func (a *DatabaseLoggerAdapter) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if a.gormConfig.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// 基础字段
	fields := []chilogger.Field{
		chilogger.String("component", "database"),
		chilogger.String("operation", "query"),
		chilogger.Float64("duration_ms", float64(elapsed.Nanoseconds())/1e6),
		chilogger.Int64("rows_affected", rows),
		chilogger.String("sql", sql),
	}

	// 处理慢查询
	if a.slowMonitor != nil && elapsed >= a.config.SlowQuery.Threshold {
		a.slowMonitor.RecordSlowQuery(sql, elapsed, rows, err)
	}

	// 记录性能数据
	if a.perfMonitor != nil {
		a.perfMonitor.RecordQuery(elapsed, err != nil)
	}

	// 根据情况记录不同级别的日志
	switch {
	case err != nil && a.gormConfig.LogLevel >= gormlogger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !a.gormConfig.IgnoreRecordNotFoundError):
		fields = append(fields, chilogger.Err(err))
		a.logger.Error("Database query error", fields...)
		
		// 错误追踪
		if a.config.ErrorTracking.Enabled {
			a.trackError(ctx, err, sql, elapsed)
		}
		
case elapsed > a.gormConfig.SlowThreshold && a.gormConfig.SlowThreshold != 0 && a.gormConfig.LogLevel >= gormlogger.Warn:
		fields = append(fields, chilogger.String("type", "slow_query"))
		a.logger.Warn("Slow database query detected", fields...)
		
case a.gormConfig.LogLevel == gormlogger.Info:
		a.logger.Info("Database query executed", fields...)
	}
}

// trackError 追踪错误
func (a *DatabaseLoggerAdapter) trackError(ctx context.Context, err error, sql string, duration time.Duration) {
	if !a.config.ErrorTracking.Enabled {
		return
	}

	// 采样控制
	if a.config.ErrorTracking.SampleRate < 1.0 {
		// 简单的采样逻辑，实际项目中可以使用更复杂的采样算法
		if time.Now().UnixNano()%100 >= int64(a.config.ErrorTracking.SampleRate*100) {
			return
		}
	}

	fields := []chilogger.Field{
		chilogger.String("component", "database"),
		chilogger.String("error_type", "database_error"),
		chilogger.String("sql", sql),
		chilogger.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
		chilogger.Err(err),
	}

	// 错误分类
	errorCategory := a.categorizeError(err)
	if errorCategory != "" {
		fields = append(fields, chilogger.String("error_category", errorCategory))
	}

	// 记录堆栈信息
	if a.config.ErrorTracking.LogStackTrace {
		fields = append(fields, chilogger.String("caller", utils.FileWithLineNum()))
	}

	a.logger.Error("Database error tracked", fields...)
}

// categorizeError 错误分类
func (a *DatabaseLoggerAdapter) categorizeError(err error) string {
	if err == nil {
		return ""
	}

	errorMsg := err.Error()
	for _, category := range a.config.ErrorTracking.ErrorCategories {
		switch category {
		case "connection":
			if contains(errorMsg, []string{"connection", "connect", "dial", "timeout"}) {
				return "connection"
			}
		case "query":
			if contains(errorMsg, []string{"syntax", "column", "table", "database"}) {
				return "query"
			}
		case "transaction":
			if contains(errorMsg, []string{"transaction", "commit", "rollback", "deadlock"}) {
				return "transaction"
			}
		case "migration":
			if contains(errorMsg, []string{"migration", "migrate", "schema"}) {
				return "migration"
			}
		}
	}

	return "unknown"
}

// contains 检查字符串是否包含任一关键词
func contains(str string, keywords []string) bool {
	for _, keyword := range keywords {
		if len(str) >= len(keyword) {
			for i := 0; i <= len(str)-len(keyword); i++ {
				if str[i:i+len(keyword)] == keyword {
					return true
				}
			}
		}
	}
	return false
}

// GetLogger 获取底层logger实例
func (a *DatabaseLoggerAdapter) GetLogger() *chilogger.Logger {
	return a.logger
}

// GetSlowQueryMonitor 获取慢查询监控器
func (a *DatabaseLoggerAdapter) GetSlowQueryMonitor() *DatabaseSlowQueryMonitor {
	return a.slowMonitor
}

// GetPerformanceMonitor 获取性能监控器
func (a *DatabaseLoggerAdapter) GetPerformanceMonitor() *DatabasePerformanceMonitor {
	return a.perfMonitor
}

// Close 关闭日志适配器
func (a *DatabaseLoggerAdapter) Close() error {
	if a.slowMonitor != nil {
		a.slowMonitor.Stop()
	}
	
	if a.perfMonitor != nil {
		a.perfMonitor.Stop()
	}
	
	return a.logger.Close()
}