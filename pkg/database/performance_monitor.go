package database

import (
	"sync"
	"time"

	chilogger "chi/pkg/logger"
)

// PerformanceStats 性能统计信息
type PerformanceStats struct {
	// 查询统计
	TotalQueries     int64         `json:"total_queries"`
	SuccessQueries   int64         `json:"success_queries"`
	FailedQueries    int64         `json:"failed_queries"`
	TotalDuration    time.Duration `json:"total_duration"`
	AvgDuration      time.Duration `json:"avg_duration"`
	MaxDuration      time.Duration `json:"max_duration"`
	MinDuration      time.Duration `json:"min_duration"`
	
	// 时间窗口统计
	WindowStart      time.Time     `json:"window_start"`
	WindowEnd        time.Time     `json:"window_end"`
	QueriesPerSecond float64       `json:"queries_per_second"`
	ErrorRate        float64       `json:"error_rate"`
	
	// 连接池统计（如果启用）
	ConnectionPool   *ConnectionPoolStats `json:"connection_pool,omitempty"`
}

// DatabasePerformanceMonitor 数据库性能监控器
type DatabasePerformanceMonitor struct {
	config    PerformanceConfig
	logger    *chilogger.Logger
	stats     *PerformanceStats
	mu        sync.RWMutex
	stopCh    chan struct{}
	running   bool
	ticker    *time.Ticker
}

// ConnectionPoolStats 连接池统计信息
type ConnectionPoolStats struct {
	OpenConnections int `json:"open_connections"`
	IdleConnections int `json:"idle_connections"`
	InUseConnections int `json:"in_use_connections"`
	MaxOpenConnections int `json:"max_open_connections"`
	MaxIdleConnections int `json:"max_idle_connections"`
}

// NewDatabasePerformanceMonitor 创建性能监控器
func NewDatabasePerformanceMonitor(config PerformanceConfig, logger *chilogger.Logger) *DatabasePerformanceMonitor {
	monitor := &DatabasePerformanceMonitor{
		config: config,
		logger: logger,
		stats: &PerformanceStats{
			MinDuration: time.Hour * 24, // 初始化为一个大值
			WindowStart: time.Now(),
		},
		stopCh: make(chan struct{}),
	}

	return monitor
}

// Start 启动性能监控器
func (m *DatabasePerformanceMonitor) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return
	}

	m.running = true
	m.stats.WindowStart = time.Now()

	// 启动定时器进行周期性统计报告
	m.ticker = time.NewTicker(m.config.Interval)
	go m.reportLoop()

	m.logger.Info("Database performance monitor started",
		chilogger.String("component", "database"),
		chilogger.Duration("interval", m.config.Interval),
	)
}

// Stop 停止性能监控器
func (m *DatabasePerformanceMonitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	m.running = false
	if m.ticker != nil {
		m.ticker.Stop()
	}
	close(m.stopCh)

	m.logger.Info("Database performance monitor stopped",
		chilogger.String("component", "database"),
	)
}

// RecordQuery 记录查询性能数据
func (m *DatabasePerformanceMonitor) RecordQuery(duration time.Duration, hasError bool) {
	if !m.config.Enabled {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 更新基础统计
	m.stats.TotalQueries++
	m.stats.TotalDuration += duration

	if hasError {
		m.stats.FailedQueries++
	} else {
		m.stats.SuccessQueries++
	}

	// 更新最大最小时间
	if duration > m.stats.MaxDuration {
		m.stats.MaxDuration = duration
	}

	if duration < m.stats.MinDuration {
		m.stats.MinDuration = duration
	}

	// 计算平均时间
	if m.stats.TotalQueries > 0 {
		m.stats.AvgDuration = m.stats.TotalDuration / time.Duration(m.stats.TotalQueries)
	}

	// 计算错误率
	if m.stats.TotalQueries > 0 {
		m.stats.ErrorRate = float64(m.stats.FailedQueries) / float64(m.stats.TotalQueries)
	}
}

// reportLoop 定期报告性能统计
func (m *DatabasePerformanceMonitor) reportLoop() {
	for {
		select {
		case <-m.ticker.C:
			m.generateReport()
		case <-m.stopCh:
			return
		}
	}
}

// generateReport 生成性能报告
func (m *DatabasePerformanceMonitor) generateReport() {
	m.mu.Lock()
	stats := *m.stats
	
	// 计算时间窗口内的QPS
	now := time.Now()
	windowDuration := now.Sub(stats.WindowStart)
	if windowDuration > 0 {
		stats.QueriesPerSecond = float64(stats.TotalQueries) / windowDuration.Seconds()
	}
	stats.WindowEnd = now
	
	// 重置窗口统计
	if windowDuration >= m.config.StatsWindow {
		m.resetWindowStats()
	}
	m.mu.Unlock()

	// 记录性能报告
	fields := []chilogger.Field{
		chilogger.String("component", "database"),
		chilogger.String("type", "performance_report"),
		chilogger.Int64("total_queries", stats.TotalQueries),
		chilogger.Int64("success_queries", stats.SuccessQueries),
		chilogger.Int64("failed_queries", stats.FailedQueries),
		chilogger.Float64("avg_duration_ms", float64(stats.AvgDuration.Nanoseconds())/1e6),
		chilogger.Float64("max_duration_ms", float64(stats.MaxDuration.Nanoseconds())/1e6),
		chilogger.Float64("min_duration_ms", float64(stats.MinDuration.Nanoseconds())/1e6),
		chilogger.Float64("queries_per_second", stats.QueriesPerSecond),
		chilogger.Float64("error_rate", stats.ErrorRate),
	}

	m.logger.Info("Database performance report", fields...)
}

// resetWindowStats 重置窗口统计
func (m *DatabasePerformanceMonitor) resetWindowStats() {
	m.stats.WindowStart = time.Now()
	m.stats.TotalQueries = 0
	m.stats.SuccessQueries = 0
	m.stats.FailedQueries = 0
	m.stats.TotalDuration = 0
	m.stats.AvgDuration = 0
	m.stats.MaxDuration = 0
	m.stats.MinDuration = time.Hour * 24
	m.stats.QueriesPerSecond = 0
	m.stats.ErrorRate = 0
}

// GetStats 获取性能统计信息
func (m *DatabasePerformanceMonitor) GetStats() PerformanceStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := *m.stats
	
	// 计算实时QPS
	now := time.Now()
	windowDuration := now.Sub(stats.WindowStart)
	if windowDuration > 0 {
		stats.QueriesPerSecond = float64(stats.TotalQueries) / windowDuration.Seconds()
	}
	stats.WindowEnd = now

	return stats
}

// Reset 重置统计信息
func (m *DatabasePerformanceMonitor) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats = &PerformanceStats{
		MinDuration: time.Hour * 24,
		WindowStart: time.Now(),
	}

	m.logger.Info("Performance monitor stats reset",
		chilogger.String("component", "database"),
	)
}

// IsRunning 检查监控器是否运行中
func (m *DatabasePerformanceMonitor) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// UpdateConnectionPoolStats 更新连接池统计信息
func (m *DatabasePerformanceMonitor) UpdateConnectionPoolStats(stats ConnectionPoolStats) {
	if !m.config.Enabled || !m.config.LogConnectionPool {
		return
	}

	m.mu.Lock()
	m.stats.ConnectionPool = &stats
	m.mu.Unlock()

	// 记录连接池状态
	fields := []chilogger.Field{
		chilogger.String("component", "database"),
		chilogger.String("type", "connection_pool_stats"),
		chilogger.Int("open_connections", stats.OpenConnections),
		chilogger.Int("idle_connections", stats.IdleConnections),
		chilogger.Int("in_use_connections", stats.InUseConnections),
		chilogger.Int("max_open_connections", stats.MaxOpenConnections),
		chilogger.Int("max_idle_connections", stats.MaxIdleConnections),
	}

	m.logger.Debug("Connection pool status", fields...)
}