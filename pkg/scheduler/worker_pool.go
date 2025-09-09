package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// WorkItem 工作项
type WorkItem struct {
	ID          string
	Task        *Task
	ExecuteFunc func() (interface{}, error)
	Callback    func(result interface{}, err error)
	CreatedAt   time.Time
}

// WorkerPool 工作池
type WorkerPool struct {
	mu        sync.RWMutex
	config    *WorkerPoolConfig
	logger    Logger
	workers   []*Worker
	workQueue chan *WorkItem
	quitChan  chan struct{}
	running   bool
	stats     *WorkerPoolStats
}

// Worker 工作协程
type Worker struct {
	id       int
	pool     *WorkerPool
	quitChan chan struct{}
	running  bool
	lastUsed time.Time
}

// WorkerPoolStats 工作池统计信息
type WorkerPoolStats struct {
	TotalWorkers    int           `json:"total_workers"`
	ActiveWorkers   int           `json:"active_workers"`
	IdleWorkers     int           `json:"idle_workers"`
	QueueSize       int           `json:"queue_size"`
	QueueCapacity   int           `json:"queue_capacity"`
	ProcessedTasks  int64         `json:"processed_tasks"`
	FailedTasks     int64         `json:"failed_tasks"`
	AverageWaitTime time.Duration `json:"average_wait_time"`
	LastUpdate      time.Time     `json:"last_update"`
}

// NewWorkerPool 创建新的工作池
func NewWorkerPool(config *WorkerPoolConfig, logger Logger) *WorkerPool {
	if config == nil {
		config = DefaultWorkerPoolConfig()
	}

	return &WorkerPool{
		config:    config,
		logger:    logger,
		workers:   make([]*Worker, 0, config.MaxWorkers),
		workQueue: make(chan *WorkItem, config.QueueSize),
		quitChan:  make(chan struct{}),
		running:   false,
		stats: &WorkerPoolStats{
			QueueCapacity: config.QueueSize,
			LastUpdate:    time.Now(),
		},
	}
}

// Start 启动工作池
func (wp *WorkerPool) Start() error {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.running {
		return NewSchedulerError(ErrSystemError, "worker pool is already running")
	}

	// 启动初始工作协程
	for i := 0; i < wp.config.MaxWorkers; i++ {
		worker := wp.createWorker(i)
		wp.workers = append(wp.workers, worker)
		go worker.start()
	}

	wp.running = true
	wp.stats.TotalWorkers = len(wp.workers)

	// 启动监控协程
	go wp.monitor()

	wp.logger.Info("Worker pool started", map[string]interface{}{
		"max_workers": wp.config.MaxWorkers,
		"queue_size":  wp.config.QueueSize,
	})

	return nil
}

// Stop 停止工作池
func (wp *WorkerPool) Stop() error {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if !wp.running {
		return NewSchedulerError(ErrSystemError, "worker pool is not running")
	}

	// 关闭工作队列
	close(wp.workQueue)

	// 停止所有工作协程
	for _, worker := range wp.workers {
		worker.stop()
	}

	// 等待所有工作协程停止
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		for _, worker := range wp.workers {
			for worker.running {
				time.Sleep(10 * time.Millisecond)
			}
		}
		close(done)
	}()

	select {
	case <-done:
		wp.logger.Info("All workers stopped gracefully", nil)
	case <-ctx.Done():
		wp.logger.Warn("Worker pool stop timeout, some workers may still be running", nil)
	}

	// 发送退出信号
	close(wp.quitChan)

	wp.running = false
	wp.workers = wp.workers[:0]
	wp.stats.TotalWorkers = 0
	wp.stats.ActiveWorkers = 0
	wp.stats.IdleWorkers = 0

	wp.logger.Info("Worker pool stopped", nil)

	return nil
}

// Submit 提交工作项
func (wp *WorkerPool) Submit(item *WorkItem) error {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	if !wp.running {
		return NewSchedulerError(ErrSystemError, "worker pool is not running")
	}

	if item == nil {
		return NewSchedulerError(ErrInvalidConfig, "work item cannot be nil")
	}

	item.CreatedAt = time.Now()

	select {
	case wp.workQueue <- item:
		wp.stats.QueueSize++
		return nil
	default:
		return NewSchedulerError(ErrWorkerPoolFull, "worker pool queue is full")
	}
}

// GetStats 获取工作池统计信息
func (wp *WorkerPool) GetStats() *WorkerPoolStats {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	stats := *wp.stats
	stats.QueueSize = len(wp.workQueue)
	stats.LastUpdate = time.Now()

	return &stats
}

// createWorker 创建工作协程
func (wp *WorkerPool) createWorker(id int) *Worker {
	return &Worker{
		id:       id,
		pool:     wp,
		quitChan: make(chan struct{}),
		running:  false,
		lastUsed: time.Now(),
	}
}

// monitor 监控工作池状态
func (wp *WorkerPool) monitor() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-wp.quitChan:
			return
		case <-ticker.C:
			wp.updateStats()
		}
	}
}

// updateStats 更新统计信息
func (wp *WorkerPool) updateStats() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	activeCount := 0
	idleCount := 0

	for _, worker := range wp.workers {
		if worker.running {
			if time.Since(worker.lastUsed) < wp.config.IdleTimeout {
				activeCount++
			} else {
				idleCount++
			}
		}
	}

	wp.stats.ActiveWorkers = activeCount
	wp.stats.IdleWorkers = idleCount
	wp.stats.QueueSize = len(wp.workQueue)
	wp.stats.LastUpdate = time.Now()
}

// start 启动工作协程
func (w *Worker) start() {
	w.running = true
	w.pool.logger.Debug("Worker started", map[string]interface{}{
		"worker_id": w.id,
	})

	for {
		select {
		case <-w.quitChan:
			w.running = false
			w.pool.logger.Debug("Worker stopped", map[string]interface{}{
				"worker_id": w.id,
			})
			return

		case item := <-w.pool.workQueue:
			if item != nil {
				w.processWorkItem(item)
			}
		}
	}
}

// stop 停止工作协程
func (w *Worker) stop() {
	close(w.quitChan)
}

// processWorkItem 处理工作项
func (w *Worker) processWorkItem(item *WorkItem) {
	w.lastUsed = time.Now()
	start := time.Now()

	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("panic in worker %d: %v", w.id, r)
			w.pool.logger.Error("Worker panic recovered", map[string]interface{}{
				"worker_id": w.id,
				"item_id":   item.ID,
				"panic":     r,
			})

			if item.Callback != nil {
				item.Callback(nil, err)
			}

			w.pool.stats.FailedTasks++
		}

		w.pool.stats.QueueSize--
		w.pool.stats.ProcessedTasks++
	}()

	w.pool.logger.Debug("Processing work item", map[string]interface{}{
		"worker_id": w.id,
		"item_id":   item.ID,
		"wait_time": time.Since(item.CreatedAt),
	})

	// 执行任务
	result, err := item.ExecuteFunc()

	// 调用回调
	if item.Callback != nil {
		item.Callback(result, err)
	}

	processTime := time.Since(start)
	w.pool.logger.Debug("Work item processed", map[string]interface{}{
		"worker_id":    w.id,
		"item_id":      item.ID,
		"process_time": processTime,
		"success":      err == nil,
	})

	if err != nil {
		w.pool.stats.FailedTasks++
	}
}
