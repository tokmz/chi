package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

// Client 数据库客户端
type Client struct {
	db            *gorm.DB
	config        *Config
	loggerAdapter *DatabaseLoggerAdapter
}

// NewClient 创建数据库客户端
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 验证配置
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 创建日志适配器
	loggerConfig := DefaultDatabaseLoggerConfig()
	loggerAdapter, err := NewDatabaseLoggerAdapter(loggerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger adapter: %w", err)
	}

	// 创建GORM配置
	gormConfig := &gorm.Config{
		Logger: loggerAdapter,
	}

	// 连接主库
	db, err := gorm.Open(mysql.Open(config.Master), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to master database: %w", err)
	}

	// 配置读写分离
	if err := configureDBResolver(db, config); err != nil {
		return nil, fmt.Errorf("failed to configure db resolver: %w", err)
	}

	// 配置连接池
	if err := configureConnectionPool(db, config.Pool); err != nil {
		return nil, fmt.Errorf("failed to configure connection pool: %w", err)
	}



	client := &Client{
		db:            db,
		config:        config,
		loggerAdapter: loggerAdapter,
	}

	// 启动监控器
	if loggerAdapter.slowMonitor != nil {
		loggerAdapter.slowMonitor.Start()
	}
	if loggerAdapter.perfMonitor != nil {
		loggerAdapter.perfMonitor.Start()
	}

	return client, nil
}

// validateConfig 验证配置
func validateConfig(config *Config) error {
	if config.Master == "" {
		return fmt.Errorf("master DSN is required")
	}
	return nil
}

// configureDBResolver 配置读写分离
func configureDBResolver(db *gorm.DB, config *Config) error {
	if len(config.Slaves) == 0 {
		// 没有从库配置，跳过读写分离设置
		return nil
	}

	// 准备从库连接
	replicas := make([]gorm.Dialector, 0, len(config.Slaves))
	for _, slaveDSN := range config.Slaves {
		replicas = append(replicas, mysql.Open(slaveDSN))
	}

	// 配置dbresolver插件
	resolverConfig := dbresolver.Config{
		// 从库用于读操作
		Replicas: replicas,
		// 读写分离策略
		Policy: dbresolver.RandomPolicy{},
	}

	// 安装dbresolver插件
	if err := db.Use(dbresolver.Register(resolverConfig)); err != nil {
		return fmt.Errorf("failed to register dbresolver: %w", err)
	}

	return nil
}

// configureConnectionPool 配置连接池
func configureConnectionPool(db *gorm.DB, poolConfig PoolConfig) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(poolConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(poolConfig.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(poolConfig.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(poolConfig.ConnMaxIdleTime)

	return nil
}

// DB 获取GORM数据库实例
func (c *Client) DB() *gorm.DB {
	return c.db
}

// Close 关闭数据库连接
func (c *Client) Close() error {
	// 关闭日志适配器
	if c.loggerAdapter != nil {
		if err := c.loggerAdapter.Close(); err != nil {
			// 记录错误但不阻止关闭数据库连接
			fmt.Printf("Failed to close logger adapter: %v\n", err)
		}
	}

	sqlDB, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	return sqlDB.Close()
}

// Ping 测试数据库连接
func (c *Client) Ping(ctx context.Context) error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	return sqlDB.PingContext(ctx)
}

// Stats 获取数据库连接统计信息
func (c *Client) Stats() (map[string]interface{}, error) {
	sqlDB, err := c.db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections":     stats.MaxOpenConnections,
		"open_connections":         stats.OpenConnections,
		"in_use":                   stats.InUse,
		"idle":                     stats.Idle,
		"wait_count":               stats.WaitCount,
		"wait_duration":            stats.WaitDuration,
		"max_idle_closed":          stats.MaxIdleClosed,
		"max_idle_time_closed":     stats.MaxIdleTimeClosed,
		"max_lifetime_closed":      stats.MaxLifetimeClosed,
	}, nil
}

// Transaction 执行事务
func (c *Client) Transaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return c.db.WithContext(ctx).Transaction(fn)
}

// WithContext 设置上下文
func (c *Client) WithContext(ctx context.Context) *gorm.DB {
	return c.db.WithContext(ctx)
}

// Master 强制使用主库
func (c *Client) Master() *gorm.DB {
	return c.db.Clauses(dbresolver.Write)
}

// Slave 强制使用从库
func (c *Client) Slave() *gorm.DB {
	return c.db.Clauses(dbresolver.Read)
}

// GetConfig 获取配置
func (c *Client) GetConfig() *Config {
	return c.config
}

// SetLogLevel 动态设置日志级别
func (c *Client) SetLogLevel(level string) error {
	var logLevel logger.LogLevel
	switch level {
	case "silent":
		logLevel = logger.Silent
	case "error":
		logLevel = logger.Error
	case "warn":
		logLevel = logger.Warn
	case "info":
		logLevel = logger.Info
	default:
		return fmt.Errorf("invalid log level: %s", level)
	}

	c.db.Logger = c.db.Logger.LogMode(logLevel)
	return nil
}

// HealthCheck 健康检查
func (c *Client) HealthCheck(ctx context.Context) error {
	// 检查主库连接
	if err := c.Ping(ctx); err != nil {
		return fmt.Errorf("master database health check failed: %w", err)
	}

	// 检查从库连接（如果有配置）
	if len(c.config.Slaves) > 0 {
		// 尝试执行一个简单的读操作来验证从库连接
		var count int64
		if err := c.Slave().WithContext(ctx).Raw("SELECT 1").Count(&count).Error; err != nil {
			return fmt.Errorf("slave database health check failed: %w", err)
		}
	}

	return nil
}

// GetSlowQueries 获取慢查询列表
func (c *Client) GetSlowQueries(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	// 这里需要根据具体的数据库实现来获取慢查询
	// MySQL可以查询performance_schema.events_statements_summary_by_digest
	// 这里提供一个示例实现
	var results []map[string]interface{}
	
	query := `
		SELECT 
			DIGEST_TEXT as sql_text,
			COUNT_STAR as exec_count,
			AVG_TIMER_WAIT/1000000000 as avg_time_ms,
			MAX_TIMER_WAIT/1000000000 as max_time_ms,
			SUM_ROWS_EXAMINED as total_rows_examined
		FROM performance_schema.events_statements_summary_by_digest 
		WHERE DIGEST_TEXT IS NOT NULL 
		ORDER BY AVG_TIMER_WAIT DESC 
		LIMIT ?
	`
	
	if err := c.db.WithContext(ctx).Raw(query, limit).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get slow queries: %w", err)
	}
	
	return results, nil
}

// GetLoggerAdapter 获取日志适配器
func (c *Client) GetLoggerAdapter() *DatabaseLoggerAdapter {
	return c.loggerAdapter
}

// GetSlowQueryMonitor 获取慢查询监控器
func (c *Client) GetSlowQueryMonitor() *DatabaseSlowQueryMonitor {
	if c.loggerAdapter != nil {
		return c.loggerAdapter.GetSlowQueryMonitor()
	}
	return nil
}

// GetPerformanceMonitor 获取性能监控器
func (c *Client) GetPerformanceMonitor() *DatabasePerformanceMonitor {
	if c.loggerAdapter != nil {
		return c.loggerAdapter.GetPerformanceMonitor()
	}
	return nil
}

// GetSlowQueryStats 获取慢查询统计信息
func (c *Client) GetSlowQueryStats() *SlowQueryStats {
	monitor := c.GetSlowQueryMonitor()
	if monitor != nil {
		stats := monitor.GetStats()
		return &stats
	}
	return nil
}

// GetPerformanceStats 获取性能统计信息
func (c *Client) GetPerformanceStats() *PerformanceStats {
	monitor := c.GetPerformanceMonitor()
	if monitor != nil {
		stats := monitor.GetStats()
		return &stats
	}
	return nil
}

// ResetSlowQueryStats 重置慢查询统计
func (c *Client) ResetSlowQueryStats() {
	monitor := c.GetSlowQueryMonitor()
	if monitor != nil {
		monitor.Reset()
	}
}

// ResetPerformanceStats 重置性能统计
func (c *Client) ResetPerformanceStats() {
	monitor := c.GetPerformanceMonitor()
	if monitor != nil {
		monitor.Reset()
	}
}

// SetSlowQueryThreshold 设置慢查询阈值
func (c *Client) SetSlowQueryThreshold(threshold time.Duration) {
	monitor := c.GetSlowQueryMonitor()
	if monitor != nil {
		monitor.SetThreshold(threshold)
	}
}