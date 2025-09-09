package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// SchedulerStatus 调度器状态
type SchedulerStatus int

const (
	SchedulerStatusStopped SchedulerStatus = iota
	SchedulerStatusRunning
	SchedulerStatusStopping
)

// String 返回调度器状态字符串
func (s SchedulerStatus) String() string {
	switch s {
	case SchedulerStatusStopped:
		return "STOPPED"
	case SchedulerStatusRunning:
		return "RUNNING"
	case SchedulerStatusStopping:
		return "STOPPING"
	default:
		return "UNKNOWN"
	}
}

// Scheduler 调度器接口
type Scheduler interface {
	// 基本操作
	Start() error
	Stop() error
	Restart() error
	GetStatus() SchedulerStatus

	// 任务管理
	AddTask(task *Task) error
	RemoveTask(taskID string) error
	GetTask(taskID string) (*Task, error)
	ListTasks() []*Task
	UpdateTask(task *Task) error

	// 任务控制
	StartTask(taskID string) error
	StopTask(taskID string) error
	PauseTask(taskID string) error
	ResumeTask(taskID string) error
	RunTaskOnce(taskID string) error

	// 监控和统计
	GetTaskStats(taskID string) (*TaskStats, error)
	GetSchedulerStats() *SchedulerStats

	// 配置管理
	UpdateConfig(config *SchedulerConfig) error
	GetConfig() *SchedulerConfig
}

// DefaultScheduler 默认调度器实现
type DefaultScheduler struct {
	mu           sync.RWMutex
	config       *SchedulerConfig
	status       SchedulerStatus
	tasks        map[string]*Task
	cronJobs     map[string]cron.EntryID
	cron         *cron.Cron
	workerPool   *WorkerPool
	monitor      Monitor
	logger       Logger
	errorHandler ErrorHandler
	panicHandler *PanicHandler
	ctx          context.Context
	cancel       context.CancelFunc
	stats        *SchedulerStats
	callbacks    []SchedulerCallback
}

// SchedulerStats 调度器统计信息
type SchedulerStats struct {
	TotalTasks     int           `json:"total_tasks"`
	RunningTasks   int           `json:"running_tasks"`
	PausedTasks    int           `json:"paused_tasks"`
	CompletedTasks int64         `json:"completed_tasks"`
	FailedTasks    int64         `json:"failed_tasks"`
	Uptime         time.Duration `json:"uptime"`
	StartTime      time.Time     `json:"start_time"`
	LastUpdate     time.Time     `json:"last_update"`
}

// SchedulerCallback 调度器回调接口
type SchedulerCallback interface {
	OnSchedulerStart(scheduler Scheduler)
	OnSchedulerStop(scheduler Scheduler)
	OnTaskAdd(scheduler Scheduler, task *Task)
	OnTaskRemove(scheduler Scheduler, taskID string)
	OnTaskStart(scheduler Scheduler, task *Task)
	OnTaskComplete(scheduler Scheduler, task *Task, result interface{})
	OnTaskError(scheduler Scheduler, task *Task, err error)
}

// NewScheduler 创建新的调度器
func NewScheduler(config *SchedulerConfig) (Scheduler, error) {
	if config == nil {
		config = DefaultSchedulerConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	// 创建日志器
	loggerFactory := &LoggerFactory{}
	logger, err := loggerFactory.CreateLogger(config)
	if err != nil {
		return nil, WrapError(ErrSystemError, "failed to create logger", err)
	}

	// 创建错误处理器
	errorHandler := NewDefaultErrorHandler(logger)
	panicHandler := NewPanicHandler(logger)

	// 创建工作池
	workerPoolConfig := &WorkerPoolConfig{
		MaxWorkers:     config.MaxWorkers,
		QueueSize:      config.QueueSize,
		IdleTimeout:    5 * time.Minute,
		MaxIdleWorkers: 2,
	}
	workerPool := NewWorkerPool(workerPoolConfig, logger)

	// 创建监控器
	monitorConfig := &MonitorConfig{
		Enabled:        config.EnableMonitor,
		CheckInterval:  config.MonitorInterval,
		MetricsEnabled: config.MetricsEnabled,
		HealthCheck:    true,
		AlertThreshold: 5,
	}
	monitor := NewMonitor(monitorConfig, logger)

	// 创建cron调度器
	cronScheduler := cron.New(cron.WithSeconds())

	ctx, cancel := context.WithCancel(context.Background())

	scheduler := &DefaultScheduler{
		config:       config.Clone(),
		status:       SchedulerStatusStopped,
		tasks:        make(map[string]*Task),
		cronJobs:     make(map[string]cron.EntryID),
		cron:         cronScheduler,
		workerPool:   workerPool,
		monitor:      monitor,
		logger:       logger,
		errorHandler: errorHandler,
		panicHandler: panicHandler,
		ctx:          ctx,
		cancel:       cancel,
		stats: &SchedulerStats{
			StartTime:  time.Now(),
			LastUpdate: time.Now(),
		},
		callbacks: make([]SchedulerCallback, 0),
	}

	return scheduler, nil
}

// Start 启动调度器
func (s *DefaultScheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.status == SchedulerStatusRunning {
		return NewSchedulerError(ErrSchedulerRunning, "scheduler is already running")
	}

	defer func() {
		if r := recover(); r != nil {
			s.panicHandler.HandlePanic(r, map[string]interface{}{
				"operation": "start_scheduler",
			})
		}
	}()

	// 启动工作池
	if err := s.workerPool.Start(); err != nil {
		return WrapError(ErrSystemError, "failed to start worker pool", err)
	}

	// 启动监控器
	if s.config.EnableMonitor {
		if err := s.monitor.Start(s.ctx); err != nil {
			s.logger.Warn("Failed to start monitor", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	// 启动cron调度器
	s.cron.Start()

	// 恢复所有任务
	for _, task := range s.tasks {
		if task.GetStatus() == TaskStatusRunning {
			if err := s.scheduleTask(task); err != nil {
				s.logger.Error("Failed to schedule task on startup", map[string]interface{}{
					"task_id": task.ID,
					"error":   err.Error(),
				})
			}
		}
	}

	s.status = SchedulerStatusRunning
	s.stats.StartTime = time.Now()

	// 触发回调
	for _, callback := range s.callbacks {
		go func(cb SchedulerCallback) {
			defer RecoverWithHandler(s.panicHandler, map[string]interface{}{
				"operation": "scheduler_start_callback",
			})
			cb.OnSchedulerStart(s)
		}(callback)
	}

	s.logger.Info("Scheduler started successfully", map[string]interface{}{
		"total_tasks": len(s.tasks),
		"config":      fmt.Sprintf("%+v", s.config),
	})

	return nil
}

// Stop 停止调度器
func (s *DefaultScheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.status != SchedulerStatusRunning {
		return NewSchedulerError(ErrSchedulerStopped, "scheduler is not running")
	}

	s.status = SchedulerStatusStopping

	defer func() {
		if r := recover(); r != nil {
			s.panicHandler.HandlePanic(r, map[string]interface{}{
				"operation": "stop_scheduler",
			})
		}
	}()

	// 停止cron调度器
	ctx := s.cron.Stop()
	select {
	case <-ctx.Done():
	case <-time.After(s.config.ShutdownTimeout):
		s.logger.Warn("Cron scheduler stop timeout", nil)
	}

	// 停止所有任务
	for taskID := range s.tasks {
		if err := s.stopTaskInternal(taskID); err != nil {
			s.logger.Error("Failed to stop task during shutdown", map[string]interface{}{
				"task_id": taskID,
				"error":   err.Error(),
			})
		}
	}

	// 停止工作池
	if err := s.workerPool.Stop(); err != nil {
		s.logger.Error("Failed to stop worker pool", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// 停止监控器
	if err := s.monitor.Stop(); err != nil {
		s.logger.Error("Failed to stop monitor", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// 取消上下文
	s.cancel()

	s.status = SchedulerStatusStopped

	// 触发回调
	for _, callback := range s.callbacks {
		go func(cb SchedulerCallback) {
			defer RecoverWithHandler(s.panicHandler, map[string]interface{}{
				"operation": "scheduler_stop_callback",
			})
			cb.OnSchedulerStop(s)
		}(callback)
	}

	s.logger.Info("Scheduler stopped successfully", nil)

	return nil
}

// Restart 重启调度器
func (s *DefaultScheduler) Restart() error {
	if err := s.Stop(); err != nil {
		return err
	}

	// 等待一小段时间确保完全停止
	time.Sleep(100 * time.Millisecond)

	return s.Start()
}

// GetStatus 获取调度器状态
func (s *DefaultScheduler) GetStatus() SchedulerStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

// GetConfig 获取调度器配置
func (s *DefaultScheduler) GetConfig() *SchedulerConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// GetSchedulerStats 获取调度器统计信息
func (s *DefaultScheduler) GetSchedulerStats() *SchedulerStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stats
}

// GetTask 获取任务
func (s *DefaultScheduler) GetTask(taskID string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if task, exists := s.tasks[taskID]; exists {
		return task, nil
	}

	return nil, NewSchedulerError(ErrTaskNotFound, fmt.Sprintf("task %s not found", taskID))
}

// GetTaskStats 获取任务统计信息
func (s *DefaultScheduler) GetTaskStats(taskID string) (*TaskStats, error) {
	task, err := s.GetTask(taskID)
	if err != nil {
		return nil, err
	}
	return task.GetStats(), nil
}

// ListTasks 列出所有任务
func (s *DefaultScheduler) ListTasks() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// RemoveTask 移除任务
func (s *DefaultScheduler) RemoveTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[taskID]; !exists {
		return NewSchedulerError(ErrTaskNotFound, fmt.Sprintf("task %s not found", taskID))
	}

	// 停止任务
	if err := s.stopTaskInternal(taskID); err != nil {
		return err
	}

	// 从任务列表中删除
	delete(s.tasks, taskID)

	// 触发回调
	for _, cb := range s.callbacks {
		go cb.OnTaskRemove(s, taskID)
	}

	return nil
}

// UpdateTask 更新任务
func (s *DefaultScheduler) UpdateTask(task *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[task.ID]; !exists {
		return NewSchedulerError(ErrTaskNotFound, fmt.Sprintf("task %s not found", task.ID))
	}

	// 重新调度任务
	if err := s.scheduleTask(task); err != nil {
		return err
	}

	s.tasks[task.ID] = task
	return nil
}

// StartTask 启动任务
func (s *DefaultScheduler) StartTask(taskID string) error {
	task, err := s.GetTask(taskID)
	if err != nil {
		return err
	}

	task.updateStatus(TaskStatusRunning)
	return nil
}

// StopTask 停止任务
func (s *DefaultScheduler) StopTask(taskID string) error {
	return s.stopTaskInternal(taskID)
}

// PauseTask 暂停任务
func (s *DefaultScheduler) PauseTask(taskID string) error {
	task, err := s.GetTask(taskID)
	if err != nil {
		return err
	}

	task.updateStatus(TaskStatusPaused)
	return nil
}

// ResumeTask 恢复任务
func (s *DefaultScheduler) ResumeTask(taskID string) error {
	task, err := s.GetTask(taskID)
	if err != nil {
		return err
	}

	task.updateStatus(TaskStatusRunning)
	return nil
}

// RunTaskOnce 立即运行任务一次
func (s *DefaultScheduler) RunTaskOnce(taskID string) error {
	task, err := s.GetTask(taskID)
	if err != nil {
		return err
	}

	// 提交到工作池执行
	workItem := &WorkItem{
		ID:   taskID,
		Task: task,
		ExecuteFunc: func() (interface{}, error) {
			return task.Execute()
		},
		Callback: func(result interface{}, err error) {
			s.handleTaskResult(task, result, err)
		},
		CreatedAt: time.Now(),
	}
	return s.workerPool.Submit(workItem)
}

// UpdateConfig 更新配置
func (s *DefaultScheduler) UpdateConfig(config *SchedulerConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = config
	return nil
}

// stopTaskInternal 内部停止任务方法
func (s *DefaultScheduler) stopTaskInternal(taskID string) error {
	task, exists := s.tasks[taskID]
	if !exists {
		return NewSchedulerError(ErrTaskNotFound, fmt.Sprintf("task %s not found", taskID))
	}

	// 从cron中移除任务
	if entryID, exists := s.cronJobs[taskID]; exists {
		s.cron.Remove(entryID)
		delete(s.cronJobs, taskID)
	}

	// 更新任务状态
	task.updateStatus(TaskStatusStopped)

	return nil
}

// AddTask 添加任务
func (s *DefaultScheduler) AddTask(task *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task == nil {
		return NewSchedulerError(ErrInvalidConfig, "task cannot be nil")
	}

	if _, exists := s.tasks[task.ID]; exists {
		return NewSchedulerError(ErrTaskExists, fmt.Sprintf("task with ID %s already exists", task.ID))
	}

	s.tasks[task.ID] = task
	s.stats.TotalTasks++

	// 如果调度器正在运行且任务状态为运行中，则立即调度
	if s.status == SchedulerStatusRunning && task.GetStatus() == TaskStatusRunning {
		if err := s.scheduleTask(task); err != nil {
			return WrapError(ErrSystemError, "failed to schedule task", err)
		}
	}

	// 触发回调
	for _, callback := range s.callbacks {
		go func(cb SchedulerCallback) {
			defer RecoverWithHandler(s.panicHandler, map[string]interface{}{
				"operation": "task_add_callback",
				"task_id":   task.ID,
			})
			cb.OnTaskAdd(s, task)
		}(callback)
	}

	s.logger.Info("Task added successfully", map[string]interface{}{
		"task_id":   task.ID,
		"task_type": task.Type,
		"schedule":  task.Config.Schedule,
	})

	return nil
}

// scheduleTask 调度任务（内部方法）
func (s *DefaultScheduler) scheduleTask(task *Task) error {
	switch task.Type {
	case TaskTypeCron:
		return s.scheduleCronTask(task)
	case TaskTypeInterval:
		return s.scheduleIntervalTask(task)
	case TaskTypeOnce:
		return s.scheduleOnceTask(task)
	default:
		return NewSchedulerError(ErrInvalidTaskType, fmt.Sprintf("unsupported task type: %s", task.Type))
	}
}

// scheduleCronTask 调度cron任务
func (s *DefaultScheduler) scheduleCronTask(task *Task) error {
	entryID, err := s.cron.AddFunc(task.Config.Schedule, func() {
		s.executeTask(task)
	})
	if err != nil {
		return WrapError(ErrInvalidCronExpr, "invalid cron expression", err)
	}

	s.cronJobs[task.ID] = entryID
	return nil
}

// scheduleIntervalTask 调度间隔任务
func (s *DefaultScheduler) scheduleIntervalTask(task *Task) error {
	interval, err := time.ParseDuration(task.Config.Schedule)
	if err != nil {
		return WrapError(ErrInvalidConfig, "invalid interval format", err)
	}

	go func() {
		defer RecoverWithHandler(s.panicHandler, map[string]interface{}{
			"operation": "interval_task_scheduler",
			"task_id":   task.ID,
		})

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				if task.GetStatus() == TaskStatusRunning {
					s.executeTask(task)
				}
			}
		}
	}()

	return nil
}

// scheduleOnceTask 调度一次性任务
func (s *DefaultScheduler) scheduleOnceTask(task *Task) error {
	if task.Config.StartTime.IsZero() {
		// 立即执行
		go s.executeTask(task)
	} else {
		// 延迟执行
		delay := time.Until(task.Config.StartTime)
		if delay > 0 {
			go func() {
				defer RecoverWithHandler(s.panicHandler, map[string]interface{}{
					"operation": "once_task_scheduler",
					"task_id":   task.ID,
				})

				timer := time.NewTimer(delay)
				defer timer.Stop()

				select {
				case <-s.ctx.Done():
					return
				case <-timer.C:
					if task.GetStatus() == TaskStatusRunning {
						s.executeTask(task)
					}
				}
			}()
		} else {
			// 时间已过，立即执行
			go s.executeTask(task)
		}
	}

	return nil
}

// executeTask 执行任务
func (s *DefaultScheduler) executeTask(task *Task) {
	if !task.CanRun() {
		return
	}

	// 提交到工作池执行
	workItem := &WorkItem{
		ID:   fmt.Sprintf("%s-%d", task.ID, time.Now().UnixNano()),
		Task: task,
		ExecuteFunc: func() (interface{}, error) {
			return task.Execute()
		},
		Callback: func(result interface{}, err error) {
			s.handleTaskResult(task, result, err)
		},
	}

	if err := s.workerPool.Submit(workItem); err != nil {
		s.logger.Error("Failed to submit task to worker pool", map[string]interface{}{
			"task_id": task.ID,
			"error":   err.Error(),
		})
		s.handleTaskResult(task, nil, err)
	}
}

// handleTaskResult 处理任务执行结果
func (s *DefaultScheduler) handleTaskResult(task *Task, result interface{}, err error) {
	if err != nil {
		s.stats.FailedTasks++
		task.updateStats(false, err)

		// 触发错误回调
		for _, callback := range s.callbacks {
			go func(cb SchedulerCallback) {
				defer RecoverWithHandler(s.panicHandler, map[string]interface{}{
					"operation": "task_error_callback",
					"task_id":   task.ID,
				})
				cb.OnTaskError(s, task, err)
			}(callback)
		}

		s.logger.Error("Task execution failed", map[string]interface{}{
			"task_id": task.ID,
			"error":   err.Error(),
		})
	} else {
		s.stats.CompletedTasks++
		task.updateStats(true, nil)

		// 触发完成回调
		for _, callback := range s.callbacks {
			go func(cb SchedulerCallback) {
				defer RecoverWithHandler(s.panicHandler, map[string]interface{}{
					"operation": "task_complete_callback",
					"task_id":   task.ID,
				})
				cb.OnTaskComplete(s, task, result)
			}(callback)
		}

		s.logger.Debug("Task executed successfully", map[string]interface{}{
			"task_id": task.ID,
			"result":  result,
		})
	}

	// 如果是一次性任务，执行完成后停止
	if task.Type == TaskTypeOnce {
		task.updateStatus(TaskStatusStopped)
	}

	s.stats.LastUpdate = time.Now()
}

// 其他方法实现...
// RemoveTask, GetTask, ListTasks, UpdateTask, StartTask, StopTask, PauseTask, ResumeTask, RunTaskOnce
// GetTaskStats, GetSchedulerStats, UpdateConfig, GetConfig
// 这些方法的实现将在后续文件中完成
