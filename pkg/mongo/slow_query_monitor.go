package mongo

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// SlowQueryMonitor 慢查询监控器
type SlowQueryMonitor struct {
	logger    Logger
	threshold time.Duration
	enabled   bool
	mu        sync.RWMutex

	// 统计信息
	stats SlowQueryStats
}

// SlowQueryStats 慢查询统计信息
type SlowQueryStats struct {
	TotalQueries     int64         `json:"total_queries"`
	SlowQueries      int64         `json:"slow_queries"`
	AverageTime      time.Duration `json:"average_time"`
	MaxTime          time.Duration `json:"max_time"`
	MinTime          time.Duration `json:"min_time"`
	LastSlowQuery    time.Time     `json:"last_slow_query"`
	SlowQueryRate    float64       `json:"slow_query_rate"`
	TotalTime        time.Duration `json:"total_time"`
	mu               sync.RWMutex  `json:"-"`
}

// QueryInfo 查询信息
type QueryInfo struct {
	Operation    string                 `json:"operation"`
	Collection   string                 `json:"collection"`
	Database     string                 `json:"database"`
	Filter       interface{}            `json:"filter,omitempty"`
	Update       interface{}            `json:"update,omitempty"`
	Options      interface{}            `json:"options,omitempty"`
	Duration     time.Duration          `json:"duration"`
	Timestamp    time.Time              `json:"timestamp"`
	Context      map[string]interface{} `json:"context,omitempty"`
	Error        error                  `json:"error,omitempty"`
	ResultCount  int64                  `json:"result_count,omitempty"`
	DocsExamined int64                  `json:"docs_examined,omitempty"`
	DocsReturned int64                  `json:"docs_returned,omitempty"`
	IndexUsed    bool                   `json:"index_used"`
	PlanSummary  string                 `json:"plan_summary,omitempty"`
}

// NewSlowQueryMonitor 创建慢查询监控器
func NewSlowQueryMonitor(logger Logger, threshold time.Duration) *SlowQueryMonitor {
	return &SlowQueryMonitor{
		logger:    logger,
		threshold: threshold,
		enabled:   true,
		stats: SlowQueryStats{
			MinTime: time.Hour * 24, // 初始化为一个大值
		},
	}
}

// LogSlowQuery 记录慢查询
func (m *SlowQueryMonitor) LogSlowQuery(queryInfo *QueryInfo) {
	m.mu.RLock()
	enabled := m.enabled
	threshold := m.threshold
	m.mu.RUnlock()

	if !enabled {
		return
	}

	// 更新统计信息
	m.updateStats(queryInfo.Duration, queryInfo.Duration >= threshold)

	// 如果是慢查询，记录详细信息
	if queryInfo.Duration >= threshold {
		m.logSlowQueryDetails(queryInfo)
	}
}

// logSlowQueryDetails 记录慢查询详细信息
func (m *SlowQueryMonitor) logSlowQueryDetails(queryInfo *QueryInfo) {
	// 构建日志字段
	fields := []interface{}{
		"operation", queryInfo.Operation,
		"collection", queryInfo.Collection,
		"database", queryInfo.Database,
		"duration_ms", queryInfo.Duration.Milliseconds(),
		"timestamp", queryInfo.Timestamp.Format(time.RFC3339),
	}

	// 添加查询过滤条件（如果存在且不为空）
	if queryInfo.Filter != nil {
		if filterBytes, err := m.sanitizeAndMarshal(queryInfo.Filter); err == nil {
			fields = append(fields, "filter", string(filterBytes))
		}
	}

	// 添加更新操作（如果存在）
	if queryInfo.Update != nil {
		if updateBytes, err := m.sanitizeAndMarshal(queryInfo.Update); err == nil {
			fields = append(fields, "update", string(updateBytes))
		}
	}

	// 添加性能指标
	if queryInfo.DocsExamined > 0 {
		fields = append(fields, "docs_examined", queryInfo.DocsExamined)
	}
	if queryInfo.DocsReturned > 0 {
		fields = append(fields, "docs_returned", queryInfo.DocsReturned)
	}
	if queryInfo.ResultCount > 0 {
		fields = append(fields, "result_count", queryInfo.ResultCount)
	}

	// 添加索引使用情况
	fields = append(fields, "index_used", queryInfo.IndexUsed)
	if queryInfo.PlanSummary != "" {
		fields = append(fields, "plan_summary", queryInfo.PlanSummary)
	}

	// 添加错误信息（如果存在）
	if queryInfo.Error != nil {
		fields = append(fields, "error", queryInfo.Error.Error())
	}

	// 添加上下文信息
	if len(queryInfo.Context) > 0 {
		for key, value := range queryInfo.Context {
			fields = append(fields, fmt.Sprintf("ctx_%s", key), value)
		}
	}

	m.logger.Warn("Slow query detected", fields...)
}

// sanitizeAndMarshal 安全地序列化查询条件，移除敏感信息
func (m *SlowQueryMonitor) sanitizeAndMarshal(data interface{}) ([]byte, error) {
	// 将数据转换为BSON文档以便处理
	var doc bson.M
	switch v := data.(type) {
	case bson.M:
		doc = v
	case bson.D:
		doc = v.Map()
	case map[string]interface{}:
		doc = bson.M(v)
	default:
		// 尝试通过BSON编解码转换
		bytes, err := bson.Marshal(data)
		if err != nil {
			return nil, err
		}
		if err := bson.Unmarshal(bytes, &doc); err != nil {
			return nil, err
		}
	}

	// 移除敏感字段
	sanitized := m.sanitizeDocument(doc)

	// 序列化为JSON
	return json.Marshal(sanitized)
}

// sanitizeDocument 移除文档中的敏感信息
func (m *SlowQueryMonitor) sanitizeDocument(doc bson.M) bson.M {
	sanitized := make(bson.M)
	sensitiveFields := map[string]bool{
		"password":    true,
		"token":       true,
		"secret":      true,
		"key":         true,
		"credential":  true,
		"auth":        true,
		"private":     true,
		"confidential": true,
	}

	for key, value := range doc {
		// 检查是否为敏感字段
		if sensitiveFields[key] {
			sanitized[key] = "[REDACTED]"
			continue
		}

		// 递归处理嵌套文档
		switch v := value.(type) {
		case bson.M:
			sanitized[key] = m.sanitizeDocument(v)
		case map[string]interface{}:
			sanitized[key] = m.sanitizeDocument(bson.M(v))
		case bson.A:
			sanitized[key] = m.sanitizeArray(v)
		case []interface{}:
			sanitized[key] = m.sanitizeArray(bson.A(v))
		default:
			sanitized[key] = value
		}
	}

	return sanitized
}

// sanitizeArray 处理数组中的敏感信息
func (m *SlowQueryMonitor) sanitizeArray(arr bson.A) bson.A {
	sanitized := make(bson.A, len(arr))
	for i, item := range arr {
		switch v := item.(type) {
		case bson.M:
			sanitized[i] = m.sanitizeDocument(v)
		case map[string]interface{}:
			sanitized[i] = m.sanitizeDocument(bson.M(v))
		default:
			sanitized[i] = item
		}
	}
	return sanitized
}

// updateStats 更新统计信息
func (m *SlowQueryMonitor) updateStats(duration time.Duration, isSlow bool) {
	m.stats.mu.Lock()
	defer m.stats.mu.Unlock()

	m.stats.TotalQueries++
	m.stats.TotalTime += duration

	if isSlow {
		m.stats.SlowQueries++
		m.stats.LastSlowQuery = time.Now()
	}

	// 更新最大最小时间
	if duration > m.stats.MaxTime {
		m.stats.MaxTime = duration
	}
	if duration < m.stats.MinTime {
		m.stats.MinTime = duration
	}

	// 计算平均时间
	if m.stats.TotalQueries > 0 {
		m.stats.AverageTime = m.stats.TotalTime / time.Duration(m.stats.TotalQueries)
	}

	// 计算慢查询率
	if m.stats.TotalQueries > 0 {
		m.stats.SlowQueryRate = float64(m.stats.SlowQueries) / float64(m.stats.TotalQueries) * 100
	}
}

// GetStats 获取统计信息
func (m *SlowQueryMonitor) GetStats() SlowQueryStats {
	m.stats.mu.RLock()
	defer m.stats.mu.RUnlock()

	// 返回统计信息的副本
	return SlowQueryStats{
		TotalQueries:  m.stats.TotalQueries,
		SlowQueries:   m.stats.SlowQueries,
		AverageTime:   m.stats.AverageTime,
		MaxTime:       m.stats.MaxTime,
		MinTime:       m.stats.MinTime,
		LastSlowQuery: m.stats.LastSlowQuery,
		SlowQueryRate: m.stats.SlowQueryRate,
		TotalTime:     m.stats.TotalTime,
	}
}

// ResetStats 重置统计信息
func (m *SlowQueryMonitor) ResetStats() {
	m.stats.mu.Lock()
	defer m.stats.mu.Unlock()

	m.stats = SlowQueryStats{
		MinTime: time.Hour * 24, // 重新初始化为一个大值
	}
}

// SetThreshold 设置慢查询阈值
func (m *SlowQueryMonitor) SetThreshold(threshold time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.threshold = threshold
}

// GetThreshold 获取慢查询阈值
func (m *SlowQueryMonitor) GetThreshold() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.threshold
}

// Enable 启用慢查询监控
func (m *SlowQueryMonitor) Enable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = true
}

// Disable 禁用慢查询监控
func (m *SlowQueryMonitor) Disable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = false
}

// IsEnabled 检查是否启用
func (m *SlowQueryMonitor) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled
}

// MonitorQuery 监控单个查询的执行
func (m *SlowQueryMonitor) MonitorQuery(ctx context.Context, operation, collection, database string, fn func() error) error {
	start := time.Now()
	err := fn()
	duration := time.Since(start)

	// 构建查询信息
	queryInfo := &QueryInfo{
		Operation:  operation,
		Collection: collection,
		Database:   database,
		Duration:   duration,
		Timestamp:  start,
		Error:      err,
	}

	// 从上下文中提取额外信息
	if ctx != nil {
		queryInfo.Context = extractContextInfo(ctx)
	}

	// 记录查询
	m.LogSlowQuery(queryInfo)

	return err
}

// extractContextInfo 从上下文中提取有用的信息
func extractContextInfo(ctx context.Context) map[string]interface{} {
	info := make(map[string]interface{})

	// 提取请求ID（如果存在）
	if requestID := ctx.Value("request_id"); requestID != nil {
		info["request_id"] = requestID
	}

	// 提取用户ID（如果存在）
	if userID := ctx.Value("user_id"); userID != nil {
		info["user_id"] = userID
	}

	// 提取会话ID（如果存在）
	if sessionID := ctx.Value("session_id"); sessionID != nil {
		info["session_id"] = sessionID
	}

	// 检查是否有截止时间
	if deadline, ok := ctx.Deadline(); ok {
		info["deadline"] = deadline.Format(time.RFC3339)
		info["timeout"] = time.Until(deadline).String()
	}

	return info
}