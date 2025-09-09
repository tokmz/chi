package scheduler

import (
	"context"
	"sync"
	"time"
)

// Monitor 监控器接口
type Monitor interface {
	Start(ctx context.Context) error
	Stop() error
	GetMetrics() *MonitorMetrics
	AddAlert(alert *Alert) error
	RemoveAlert(alertID string) error
	IsHealthy() bool
}

// DefaultMonitor 默认监控器实现
type DefaultMonitor struct {
	mu        sync.RWMutex
	config    *MonitorConfig
	logger    Logger
	running   bool
	ctx       context.Context
	cancel    context.CancelFunc
	metrics   *MonitorMetrics
	alerts    map[string]*Alert
	callbacks []MonitorCallback
}

// MonitorMetrics 监控指标
type MonitorMetrics struct {
	SystemMetrics    *SystemMetrics    `json:"system_metrics"`
	SchedulerMetrics *SchedulerMetrics `json:"scheduler_metrics"`
	TaskMetrics      *TaskMetrics      `json:"task_metrics"`
	WorkerMetrics    *WorkerMetrics    `json:"worker_metrics"`
	LastUpdate       time.Time         `json:"last_update"`
}

// SystemMetrics 系统指标
type SystemMetrics struct {
	CPUUsage    float64       `json:"cpu_usage"`
	MemoryUsage float64       `json:"memory_usage"`
	Goroutines  int           `json:"goroutines"`
	Uptime      time.Duration `json:"uptime"`
}

// SchedulerMetrics 调度器指标
type SchedulerMetrics struct {
	Status         string    `json:"status"`
	TotalTasks     int       `json:"total_tasks"`
	RunningTasks   int       `json:"running_tasks"`
	PausedTasks    int       `json:"paused_tasks"`
	CompletedTasks int64     `json:"completed_tasks"`
	FailedTasks    int64     `json:"failed_tasks"`
	LastUpdate     time.Time `json:"last_update"`
}

// TaskMetrics 任务指标
type TaskMetrics struct {
	AverageExecutionTime time.Duration `json:"average_execution_time"`
	SuccessRate          float64       `json:"success_rate"`
	ErrorRate            float64       `json:"error_rate"`
	Throughput           float64       `json:"throughput"`
}

// WorkerMetrics 工作池指标
type WorkerMetrics struct {
	TotalWorkers    int           `json:"total_workers"`
	ActiveWorkers   int           `json:"active_workers"`
	IdleWorkers     int           `json:"idle_workers"`
	QueueSize       int           `json:"queue_size"`
	QueueCapacity   int           `json:"queue_capacity"`
	AverageWaitTime time.Duration `json:"average_wait_time"`
}

// Alert 告警
type Alert struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Level       AlertLevel             `json:"level"`
	Condition   AlertCondition         `json:"condition"`
	Threshold   float64                `json:"threshold"`
	Enabled     bool                   `json:"enabled"`
	Triggered   bool                   `json:"triggered"`
	TriggerTime time.Time              `json:"trigger_time"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// AlertLevel 告警级别
type AlertLevel int

const (
	AlertLevelInfo AlertLevel = iota
	AlertLevelWarn
	AlertLevelError
	AlertLevelCritical
)

// String 返回告警级别字符串
func (l AlertLevel) String() string {
	switch l {
	case AlertLevelInfo:
		return "INFO"
	case AlertLevelWarn:
		return "WARN"
	case AlertLevelError:
		return "ERROR"
	case AlertLevelCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// AlertCondition 告警条件
type AlertCondition int

const (
	AlertConditionGreaterThan AlertCondition = iota
	AlertConditionLessThan
	AlertConditionEquals
	AlertConditionNotEquals
)

// MonitorCallback 监控回调接口
type MonitorCallback interface {
	OnMetricsUpdate(metrics *MonitorMetrics)
	OnAlertTriggered(alert *Alert)
	OnAlertResolved(alert *Alert)
	OnHealthStatusChange(healthy bool)
}

// NewMonitor 创建新的监控器
func NewMonitor(config *MonitorConfig, logger Logger) *DefaultMonitor {
	if config == nil {
		config = DefaultMonitorConfig()
	}

	return &DefaultMonitor{
		config:  config,
		logger:  logger,
		running: false,
		metrics: &MonitorMetrics{
			SystemMetrics:    &SystemMetrics{},
			SchedulerMetrics: &SchedulerMetrics{},
			TaskMetrics:      &TaskMetrics{},
			WorkerMetrics:    &WorkerMetrics{},
			LastUpdate:       time.Now(),
		},
		alerts:    make(map[string]*Alert),
		callbacks: make([]MonitorCallback, 0),
	}
}

// Start 启动监控器
func (m *DefaultMonitor) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return NewSchedulerError(ErrMonitor, "monitor is already running")
	}

	if !m.config.Enabled {
		m.logger.Info("Monitor is disabled", nil)
		return nil
	}

	m.ctx, m.cancel = context.WithCancel(ctx)
	m.running = true

	// 启动监控协程
	go m.monitorLoop()

	m.logger.Info("Monitor started", map[string]interface{}{
		"check_interval":  m.config.CheckInterval,
		"metrics_enabled": m.config.MetricsEnabled,
		"health_check":    m.config.HealthCheck,
	})

	return nil
}

// Stop 停止监控器
func (m *DefaultMonitor) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return NewSchedulerError(ErrMonitor, "monitor is not running")
	}

	m.cancel()
	m.running = false

	m.logger.Info("Monitor stopped", nil)

	return nil
}

// GetMetrics 获取监控指标
func (m *DefaultMonitor) GetMetrics() *MonitorMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回指标的副本
	metrics := *m.metrics
	systemMetrics := *m.metrics.SystemMetrics
	schedulerMetrics := *m.metrics.SchedulerMetrics
	taskMetrics := *m.metrics.TaskMetrics
	workerMetrics := *m.metrics.WorkerMetrics

	metrics.SystemMetrics = &systemMetrics
	metrics.SchedulerMetrics = &schedulerMetrics
	metrics.TaskMetrics = &taskMetrics
	metrics.WorkerMetrics = &workerMetrics

	return &metrics
}

// AddAlert 添加告警
func (m *DefaultMonitor) AddAlert(alert *Alert) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if alert == nil {
		return NewSchedulerError(ErrInvalidConfig, "alert cannot be nil")
	}

	if alert.ID == "" {
		return NewSchedulerError(ErrInvalidConfig, "alert ID cannot be empty")
	}

	m.alerts[alert.ID] = alert

	m.logger.Info("Alert added", map[string]interface{}{
		"alert_id":   alert.ID,
		"alert_name": alert.Name,
		"level":      alert.Level.String(),
	})

	return nil
}

// RemoveAlert 移除告警
func (m *DefaultMonitor) RemoveAlert(alertID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.alerts[alertID]; !exists {
		return NewSchedulerError(ErrTaskNotFound, "alert not found")
	}

	delete(m.alerts, alertID)

	m.logger.Info("Alert removed", map[string]interface{}{
		"alert_id": alertID,
	})

	return nil
}

// IsHealthy 检查系统健康状态
func (m *DefaultMonitor) IsHealthy() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.config.HealthCheck {
		return true
	}

	// 检查各项指标是否正常
	if m.metrics.SystemMetrics.CPUUsage > 90.0 {
		return false
	}

	if m.metrics.SystemMetrics.MemoryUsage > 90.0 {
		return false
	}

	if m.metrics.WorkerMetrics.QueueSize >= m.metrics.WorkerMetrics.QueueCapacity {
		return false
	}

	return true
}

// AddCallback 添加监控回调
func (m *DefaultMonitor) AddCallback(callback MonitorCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.callbacks = append(m.callbacks, callback)
}

// monitorLoop 监控循环
func (m *DefaultMonitor) monitorLoop() {
	ticker := time.NewTicker(m.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.collectMetrics()
			m.checkAlerts()
			m.notifyCallbacks()
		}
	}
}

// collectMetrics 收集指标
func (m *DefaultMonitor) collectMetrics() {
	if !m.config.MetricsEnabled {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 收集系统指标
	m.collectSystemMetrics()

	// 更新时间戳
	m.metrics.LastUpdate = time.Now()
}

// collectSystemMetrics 收集系统指标
func (m *DefaultMonitor) collectSystemMetrics() {
	// 这里可以集成系统监控库来获取真实的系统指标
	// 为了简化，这里使用模拟数据
	m.metrics.SystemMetrics.CPUUsage = 0.0    // 实际应该获取真实CPU使用率
	m.metrics.SystemMetrics.MemoryUsage = 0.0 // 实际应该获取真实内存使用率
	m.metrics.SystemMetrics.Goroutines = 0    // 实际应该获取goroutine数量
}

// checkAlerts 检查告警
func (m *DefaultMonitor) checkAlerts() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, alert := range m.alerts {
		if !alert.Enabled {
			continue
		}

		// 检查告警条件
		triggered := m.evaluateAlertCondition(alert)

		if triggered && !alert.Triggered {
			// 告警触发
			alert.Triggered = true
			alert.TriggerTime = time.Now()

			m.logger.Warn("Alert triggered", map[string]interface{}{
				"alert_id":   alert.ID,
				"alert_name": alert.Name,
				"level":      alert.Level.String(),
			})

			// 通知回调
			for _, callback := range m.callbacks {
				go callback.OnAlertTriggered(alert)
			}

			// 调用通知回调
			if m.config.NotifyCallback != nil {
				go m.config.NotifyCallback(alert.Name)
			}

		} else if !triggered && alert.Triggered {
			// 告警恢复
			alert.Triggered = false

			m.logger.Info("Alert resolved", map[string]interface{}{
				"alert_id":   alert.ID,
				"alert_name": alert.Name,
			})

			// 通知回调
			for _, callback := range m.callbacks {
				go callback.OnAlertResolved(alert)
			}
		}
	}
}

// evaluateAlertCondition 评估告警条件
func (m *DefaultMonitor) evaluateAlertCondition(alert *Alert) bool {
	// 这里应该根据告警条件和当前指标来判断是否触发告警
	// 为了简化，这里返回false
	return false
}

// notifyCallbacks 通知回调
func (m *DefaultMonitor) notifyCallbacks() {
	for _, callback := range m.callbacks {
		go callback.OnMetricsUpdate(m.GetMetrics())
	}
}

// UpdateSchedulerMetrics 更新调度器指标
func (m *DefaultMonitor) UpdateSchedulerMetrics(stats *SchedulerStats) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics.SchedulerMetrics.TotalTasks = stats.TotalTasks
	m.metrics.SchedulerMetrics.RunningTasks = stats.RunningTasks
	m.metrics.SchedulerMetrics.PausedTasks = stats.PausedTasks
	m.metrics.SchedulerMetrics.CompletedTasks = stats.CompletedTasks
	m.metrics.SchedulerMetrics.FailedTasks = stats.FailedTasks
	m.metrics.SchedulerMetrics.LastUpdate = stats.LastUpdate
}

// UpdateWorkerMetrics 更新工作池指标
func (m *DefaultMonitor) UpdateWorkerMetrics(stats *WorkerPoolStats) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics.WorkerMetrics.TotalWorkers = stats.TotalWorkers
	m.metrics.WorkerMetrics.ActiveWorkers = stats.ActiveWorkers
	m.metrics.WorkerMetrics.IdleWorkers = stats.IdleWorkers
	m.metrics.WorkerMetrics.QueueSize = stats.QueueSize
	m.metrics.WorkerMetrics.QueueCapacity = stats.QueueCapacity
	m.metrics.WorkerMetrics.AverageWaitTime = stats.AverageWaitTime
}
