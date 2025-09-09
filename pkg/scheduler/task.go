package scheduler

import (
	"context"
	"sync"
	"time"
)

// TaskStatus 任务状态
type TaskStatus int

const (
	TaskStatusPending   TaskStatus = iota // 等待中
	TaskStatusRunning                     // 运行中
	TaskStatusPaused                      // 暂停
	TaskStatusStopped                     // 已停止
	TaskStatusCompleted                   // 已完成
	TaskStatusFailed                      // 失败
)

// String 返回任务状态字符串
func (s TaskStatus) String() string {
	switch s {
	case TaskStatusPending:
		return "PENDING"
	case TaskStatusRunning:
		return "RUNNING"
	case TaskStatusPaused:
		return "PAUSED"
	case TaskStatusStopped:
		return "STOPPED"
	case TaskStatusCompleted:
		return "COMPLETED"
	case TaskStatusFailed:
		return "FAILED"
	default:
		return "UNKNOWN"
	}
}

// TaskType 任务类型
type TaskType int

const (
	TaskTypeCron     TaskType = iota // Cron表达式任务
	TaskTypeInterval                 // 固定间隔任务
	TaskTypeOnce                     // 一次性任务
	TaskTypeDelay                    // 延迟任务
)

// String 返回任务类型字符串
func (t TaskType) String() string {
	switch t {
	case TaskTypeCron:
		return "CRON"
	case TaskTypeInterval:
		return "INTERVAL"
	case TaskTypeOnce:
		return "ONCE"
	case TaskTypeDelay:
		return "DELAY"
	default:
		return "UNKNOWN"
	}
}

// TaskFunc 任务执行函数类型
type TaskFunc func() (interface{}, error)

// TaskCallback 任务回调函数类型
type TaskCallback func(taskID string, result interface{}, err error)

// TaskConfig 任务配置
type TaskConfig struct {
	Name          string                 `json:"name"`            // 任务名称
	Description   string                 `json:"description"`     // 任务描述
	Enabled       bool                   `json:"enabled"`         // 是否启用
	Schedule      string                 `json:"schedule"`        // 调度表达式（cron表达式或间隔时间）
	StartTime     time.Time              `json:"start_time"`      // 开始时间（用于一次性任务）
	Timeout       time.Duration          `json:"timeout"`         // 执行超时时间
	MaxRetries    int                    `json:"max_retries"`     // 最大重试次数
	RetryInterval time.Duration          `json:"retry_interval"`  // 重试间隔
	Concurrency   int                    `json:"concurrency"`     // 并发数限制
	SkipIfRunning bool                   `json:"skip_if_running"` // 如果正在运行则跳过
	Metadata      map[string]interface{} `json:"metadata"`        // 元数据
}

// TaskStats 任务统计信息
type TaskStats struct {
	TotalRuns       int64         `json:"total_runs"`        // 总运行次数
	SuccessRuns     int64         `json:"success_runs"`      // 成功运行次数
	FailedRuns      int64         `json:"failed_runs"`       // 失败运行次数
	LastRunTime     time.Time     `json:"last_run_time"`     // 最后运行时间
	LastRunDuration time.Duration `json:"last_run_duration"` // 最后运行耗时
	AverageRunTime  time.Duration `json:"average_run_time"`  // 平均运行时间
	NextRunTime     time.Time     `json:"next_run_time"`     // 下次运行时间
	CreatedAt       time.Time     `json:"created_at"`        // 创建时间
	UpdatedAt       time.Time     `json:"updated_at"`        // 更新时间
}

// Task 任务结构体
type Task struct {
	mu        sync.RWMutex
	ID        string       `json:"id"`
	Type      TaskType     `json:"type"`
	Config    *TaskConfig  `json:"config"`
	Func      TaskFunc     `json:"-"`
	Callback  TaskCallback `json:"-"`
	status    TaskStatus
	stats     *TaskStats
	createdAt time.Time
	updatedAt time.Time
	context   context.Context    `json:"-"`
	cancel    context.CancelFunc `json:"-"`
	lastError error              `json:"-"`
}

// DefaultTaskConfig 返回默认任务配置
func DefaultTaskConfig() *TaskConfig {
	return &TaskConfig{
		Enabled:       true,
		Timeout:       30 * time.Second,
		MaxRetries:    3,
		RetryInterval: 5 * time.Second,
		Concurrency:   1,
		SkipIfRunning: false,
		Metadata:      make(map[string]interface{}),
	}
}

// NewTask 创建新任务
func NewTask(id, name string, fn TaskFunc) *Task {
	ctx, cancel := context.WithCancel(context.Background())
	config := DefaultTaskConfig()
	config.Name = name

	return &Task{
		ID:        id,
		Type:      TaskTypeOnce,
		status:    TaskStatusPending,
		createdAt: time.Now(),
		updatedAt: time.Now(),
		Func:      fn,
		Config:    config,
		stats:     &TaskStats{CreatedAt: time.Now(), UpdatedAt: time.Now()},
		context:   ctx,
		cancel:    cancel,
	}
}

// SetCron 设置Cron表达式
func (t *Task) SetCron(expr string) *Task {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Type = TaskTypeCron
	t.Config.Schedule = expr
	t.updatedAt = time.Now()
	return t
}

// SetInterval 设置执行间隔
func (t *Task) SetInterval(interval time.Duration) *Task {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Type = TaskTypeInterval
	t.Config.Schedule = interval.String()
	t.updatedAt = time.Now()
	return t
}

// SetDelay 设置延迟执行
func (t *Task) SetDelay(delay time.Duration) *Task {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Type = TaskTypeDelay
	t.Config.StartTime = time.Now().Add(delay)
	t.updatedAt = time.Now()
	return t
}

// SetTimeRange 设置时间范围
func (t *Task) SetTimeRange(start, end time.Time) *Task {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Config.StartTime = start
	t.updatedAt = time.Now()
	return t
}

// SetConfig 设置任务配置
func (t *Task) SetConfig(config *TaskConfig) *Task {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Config = config
	t.updatedAt = time.Now()
	return t
}

// SetCallback 设置回调函数
func (t *Task) SetCallback(callback TaskCallback) *Task {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Callback = callback
	t.updatedAt = time.Now()
	return t
}

// SetDescription 设置任务描述
func (t *Task) SetDescription(desc string) *Task {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Config.Description = desc
	t.updatedAt = time.Now()
	return t
}

// GetStatus 获取任务状态
func (t *Task) GetStatus() TaskStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}

// GetStats 获取任务统计信息
func (t *Task) GetStats() *TaskStats {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// 返回副本以避免并发问题
	stats := *t.stats
	return &stats
}

// IsRunning 检查任务是否正在运行
func (t *Task) IsRunning() bool {
	return t.GetStatus() == TaskStatusRunning
}

// CanRun 检查任务是否可以运行
func (t *Task) CanRun() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// 检查任务状态
	if t.status == TaskStatusStopped || t.status == TaskStatusPaused {
		return false
	}

	// 检查是否启用
	if !t.Config.Enabled {
		return false
	}

	// 检查是否跳过正在运行的任务
	if t.Config.SkipIfRunning && t.status == TaskStatusRunning {
		return false
	}

	return true
}

// Execute 执行任务
func (t *Task) Execute() (interface{}, error) {
	if t.Func == nil {
		return nil, NewSchedulerError(ErrInvalidConfig, "task function is nil")
	}

	return t.Func()
}

// updateStatus 更新任务状态（内部方法）
func (t *Task) updateStatus(status TaskStatus) {
	t.status = status
	t.updatedAt = time.Now()

	// 触发回调
	if t.Callback != nil {
		go t.Callback(t.ID, nil, nil)
	}
}

// updateStats 更新统计信息（内部方法）
func (t *Task) updateStats(success bool, err error) {
	t.stats.TotalRuns++
	t.stats.LastRunTime = time.Now()
	t.stats.UpdatedAt = time.Now()

	if err != nil {
		t.stats.FailedRuns++
		t.lastError = err
	} else {
		t.stats.SuccessRuns++
		t.lastError = nil
	}
}
