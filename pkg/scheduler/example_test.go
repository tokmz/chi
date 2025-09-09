package scheduler

import (
	"fmt"
	"log"
	"time"
)

// ExampleScheduler 展示如何使用调度器
func ExampleScheduler() {
	// 创建调度器配置
	config := DefaultSchedulerConfig()
	config.MaxWorkers = 5
	config.QueueSize = 100

	// 创建调度器
	scheduler, err := NewScheduler(config)
	if err != nil {
		log.Fatal("Failed to create scheduler:", err)
	}

	// 启动调度器
	if err = scheduler.Start(); err != nil {
		log.Fatal("Failed to start scheduler:", err)
	}
	defer scheduler.Stop()

	// 创建一个简单的任务
	task1 := NewTask("task1", "Hello World Task", func() (interface{}, error) {
		fmt.Println("Hello, World!")
		return "success", nil
	})

	// 设置为每5秒执行一次
	task1.SetInterval(5 * time.Second)

	// 添加任务到调度器
	if err = scheduler.AddTask(task1); err != nil {
		log.Fatal("Failed to add task:", err)
	}

	// 创建一个Cron任务
	task2 := NewTask("task2", "Cron Task", func() (interface{}, error) {
		fmt.Println("Cron task executed at", time.Now().Format("15:04:05"))
		return "cron success", nil
	})

	// 设置为每分钟执行一次
	task2.SetCron("0 * * * * *")

	// 添加回调函数
	task2.SetCallback(func(taskID string, result interface{}, err error) {
		if err != nil {
			fmt.Printf("Task %s failed: %v\n", taskID, err)
		} else {
			fmt.Printf("Task %s completed: %v\n", taskID, result)
		}
	})

	// 添加任务到调度器
	if err = scheduler.AddTask(task2); err != nil {
		log.Fatal("Failed to add cron task:", err)
	}

	// 创建一次性任务
	task3 := NewTask("task3", "One-time Task", func() (interface{}, error) {
		fmt.Println("One-time task executed")
		return "one-time success", nil
	})

	// 设置为5秒后执行
	task3.SetDelay(5 * time.Second)

	// 添加任务到调度器
	if err = scheduler.AddTask(task3); err != nil {
		log.Fatal("Failed to add one-time task:", err)
	}

	// 立即执行一次任务
	if err = scheduler.RunTaskOnce("task1"); err != nil {
		log.Printf("Failed to run task once: %v", err)
	}

	// 等待一段时间观察任务执行
	time.Sleep(30 * time.Second)

	// 获取任务统计信息
	stats, err := scheduler.GetTaskStats("task1")
	if err != nil {
		log.Printf("Failed to get task stats: %v", err)
	} else {
		fmt.Printf("Task1 Stats: Total runs: %d, Success: %d, Failed: %d\n",
			stats.TotalRuns, stats.SuccessRuns, stats.FailedRuns)
	}

	// 暂停任务
	if err := scheduler.PauseTask("task1"); err != nil {
		log.Printf("Failed to pause task: %v", err)
	}

	// 等待5秒
	time.Sleep(5 * time.Second)

	// 恢复任务
	if err := scheduler.ResumeTask("task1"); err != nil {
		log.Printf("Failed to resume task: %v", err)
	}

	// 列出所有任务
	tasks := scheduler.ListTasks()
	fmt.Printf("Total tasks: %d\n", len(tasks))
	for _, task := range tasks {
		fmt.Printf("Task: %s, Status: %s, Type: %s\n",
			task.Config.Name, task.GetStatus().String(), task.Type.String())
	}

	// 获取调度器统计信息
	schedulerStats := scheduler.GetSchedulerStats()
	fmt.Printf("Scheduler Stats: Total tasks: %d, Running: %d, Completed: %d\n",
		schedulerStats.TotalTasks, schedulerStats.RunningTasks, schedulerStats.CompletedTasks)

	// 移除任务
	if err := scheduler.RemoveTask("task3"); err != nil {
		log.Printf("Failed to remove task: %v", err)
	}

	fmt.Println("Example completed")
}

// TaskWithRetryExample 展示如何创建带重试机制的任务
func TaskWithRetryExample() {
	config := DefaultSchedulerConfig()
	scheduler, _ := NewScheduler(config)
	scheduler.Start()
	defer scheduler.Stop()

	// 创建一个可能失败的任务
	counter := 0
	task := NewTask("retry-task", "Task with Retry", func() (interface{}, error) {
		counter++
		if counter < 3 {
			return nil, fmt.Errorf("simulated failure %d", counter)
		}
		fmt.Println("Task succeeded on attempt", counter)
		return "success", nil
	})

	// 配置重试
	task.Config.MaxRetries = 5
	task.Config.RetryInterval = 2 * time.Second
	task.Config.Timeout = 10 * time.Second

	// 设置回调来观察重试过程
	task.SetCallback(func(taskID string, result interface{}, err error) {
		if err != nil {
			fmt.Printf("Task %s attempt failed: %v\n", taskID, err)
		} else {
			fmt.Printf("Task %s succeeded: %v\n", taskID, result)
		}
	})

	scheduler.AddTask(task)
	scheduler.RunTaskOnce("retry-task")

	// 等待任务完成
	time.Sleep(15 * time.Second)
}

// CustomLoggerExample 展示如何使用自定义日志器
func CustomLoggerExample() {
	// 创建自定义日志器
	logger := NewDefaultLogger(LogLevelInfo, nil, true)

	config := DefaultSchedulerConfig()
	config.LogLevel = LogLevelInfo

	scheduler, _ := NewScheduler(config)
	scheduler.Start()
	defer scheduler.Stop()

	task := NewTask("log-task", "Logging Task", func() (interface{}, error) {
		logger.Info("Task is executing", map[string]interface{}{
			"timestamp": time.Now(),
			"task_id":   "log-task",
		})
		return "logged", nil
	})

	task.SetInterval(3 * time.Second)
	scheduler.AddTask(task)

	time.Sleep(10 * time.Second)
}
