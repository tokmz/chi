package database

import (
	"sync"
	"time"

	chilogger "chi/pkg/logger"
)

// SlowQueryStats 慢查询统计信息
type SlowQueryStats struct {
	TotalQueries    int64         `json:"total_queries"`
	SlowQueries     int64         `json:"slow_queries"`
	AverageTime     time.Duration `json:"average_time"`
	MaxTime         time.Duration `json:"max_time"`
	MinTime         time.Duration `json:"min_time"`
	LastSlowQuery   time.Time     `json:"last_slow_query"`
	SlowQueryRate   float64       `json:"slow_query_rate"`
}

// SlowQueryRecord 慢查询记录
type SlowQueryRecord struct {
	Timestamp time.Time     `json:"timestamp"`
	SQL       string        `json:"sql"`
	Duration  time.Duration `json:"duration"`
	Params    []interface{} `json:"params,omitempty"`
	Error     string        `json:"error,omitempty"`
}

// DatabaseSlowQueryMonitor 数据库慢查询监控器
type DatabaseSlowQueryMonitor struct {
	config      SlowQueryConfig
	logger      *chilogger.Logger
	stats       *SlowQueryStats
	mu          sync.RWMutex
	stopCh      chan struct{}
	running     bool
	records     []SlowQueryRecord
	maxRecords  int
}

// NewDatabaseSlowQueryMonitor 创建数据库慢查询监控器
func NewDatabaseSlowQueryMonitor(config SlowQueryConfig, logger *chilogger.Logger) *DatabaseSlowQueryMonitor {
	monitor := &DatabaseSlowQueryMonitor{
		config: config,
		logger: logger,
		stats: &SlowQueryStats{
			MinTime: time.Hour * 24, // 初始化为一个大值
		},
		stopCh: make(chan struct{}),
	}

	return monitor
}

// Start 启动监控器
func (m *DatabaseSlowQueryMonitor) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return
	}

	m.running = true
	m.logger.Info("Database slow query monitor started", 
		chilogger.String("component", "database"),
		chilogger.Duration("threshold", m.config.Threshold),
	)
}

// Stop 停止监控器
func (m *DatabaseSlowQueryMonitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	m.running = false
	close(m.stopCh)

	m.logger.Info("Database slow query monitor stopped",
		chilogger.String("component", "database"),
	)
}

// RecordSlowQuery 记录慢查询
func (m *DatabaseSlowQueryMonitor) RecordSlowQuery(sql string, duration time.Duration, rows int64, err error) {
	if !m.config.Enabled {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 更新统计信息
	m.stats.TotalQueries++
	m.stats.SlowQueries++

	if duration > m.stats.MaxTime {
		m.stats.MaxTime = duration
	}

	if duration < m.stats.MinTime {
		m.stats.MinTime = duration
	}

	m.stats.LastSlowQuery = time.Now()
	m.stats.SlowQueryRate = float64(m.stats.SlowQueries) / float64(m.stats.TotalQueries)

	// 记录慢查询日志
	fields := []chilogger.Field{
		chilogger.String("component", "database"),
		chilogger.String("type", "slow_query"),
		chilogger.String("sql", sql),
		chilogger.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
		chilogger.Int64("rows_affected", rows),
		chilogger.Float64("threshold_ms", float64(m.config.Threshold.Nanoseconds())/1e6),
	}

	if err != nil {
		fields = append(fields, chilogger.Err(err))
	}

	m.logger.Warn("Slow query detected", fields...)
}

// RecordQuery 记录普通查询（用于统计）
func (m *DatabaseSlowQueryMonitor) RecordQuery(duration time.Duration) {
	if !m.config.Enabled {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats.TotalQueries++

	if m.stats.TotalQueries > 0 {
		m.stats.SlowQueryRate = float64(m.stats.SlowQueries) / float64(m.stats.TotalQueries)
	}
}

// GetStats 获取统计信息
func (m *DatabaseSlowQueryMonitor) GetStats() SlowQueryStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回统计信息副本
	stats := *m.stats

	return stats
}

// Reset 重置统计信息
func (m *DatabaseSlowQueryMonitor) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats = &SlowQueryStats{
		MinTime: time.Hour * 24,
	}

	m.logger.Info("Slow query monitor stats reset",
		chilogger.String("component", "database"),
	)
}

// IsRunning 检查监控器是否运行中
func (m *DatabaseSlowQueryMonitor) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// GetThreshold 获取慢查询阈值
func (m *DatabaseSlowQueryMonitor) GetThreshold() time.Duration {
	return m.config.Threshold
}

// SetThreshold 设置慢查询阈值
func (m *DatabaseSlowQueryMonitor) SetThreshold(threshold time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config.Threshold = threshold
	m.logger.Info("Slow query threshold updated",
		chilogger.String("component", "database"),
		chilogger.Duration("new_threshold", threshold),
	)
}