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
	// 创建自定义日志配置
	config := DefaultSchedulerConfig()
	config.LogLevel = LogLevelDebug
	config.EnableConsole = true

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

	// 创建任务
	task := NewTask("debug-task", "Debug Task", func() (interface{}, error) {
		fmt.Println("Debug task with custom logger")
		return "debug success", nil
	})
	task.SetInterval(3 * time.Second)

	// 添加任务
	scheduler.AddTask(task)

	// 运行一段时间
	time.Sleep(10 * time.Second)
}

// ZapLoggerExample 演示使用Zap高性能日志器
func ZapLoggerExample() {
	// 使用扩展日志配置
	config := DefaultSchedulerConfig()
	config.LoggerConfig = &LoggerConfig{
		Level:         LogLevelInfo,
		Output:        "logs/scheduler.log",
		EnableConsole: true,
		UseZapLogger:  true,
		Format:        "json",
		Development:   false,
		File: FileLogConfig{
			Enabled:    true,
			Filename:   "logs/scheduler.log",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   true,
			LocalTime:  true,
		},
		Console: ConsoleLogConfig{
			Enabled:    true,
			Colorful:   true,
			TimeFormat: "2006-01-02 15:04:05",
		},
	}

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

	// 创建多个任务演示不同日志级别
	tasks := []*Task{
		NewTask("info-task", "Info Level Task", func() (interface{}, error) {
			fmt.Println("Info level task executed")
			return "info", nil
		}),
		NewTask("warn-task", "Warn Level Task", func() (interface{}, error) {
			fmt.Println("Warn level task executed")
			return "warn", fmt.Errorf("warning: this is a test warning")
		}),
		NewTask("error-task", "Error Level Task", func() (interface{}, error) {
			return nil, fmt.Errorf("error: this is a test error")
		}),
	}

	// 设置不同的执行间隔
	tasks[0].SetInterval(15 * time.Second)
	tasks[1].SetInterval(20 * time.Second)
	tasks[2].SetInterval(25 * time.Second)

	// 添加任务到调度器
	for _, task := range tasks {
		if err := scheduler.AddTask(task); err != nil {
			log.Printf("Failed to add task %s: %v", task.ID, err)
		}
	}

	time.Sleep(1 * time.Minute)
}

// DevelopmentLoggerExample 演示开发环境日志配置
func DevelopmentLoggerExample() {
	// 开发环境：详细日志，仅控制台输出，彩色显示
	config := DefaultSchedulerConfig()
	config.LoggerConfig = &LoggerConfig{
		Level:         LogLevelDebug,
		Output:        "", // 空字符串表示不输出到文件
		EnableConsole: true,
		UseZapLogger:  true,
		Format:        "console",
		Development:   true,
		Console: ConsoleLogConfig{
			Enabled:    true,
			Colorful:   true,
			TimeFormat: "15:04:05",
		},
	}

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

	// 创建调试任务
	task := NewTask("debug-task", "Debug Task", func() (interface{}, error) {
		fmt.Println("Debug task with detailed logging")
		return map[string]interface{}{
			"timestamp": time.Now(),
			"status":    "success",
			"details":   "Task completed successfully",
		}, nil
	})
	task.SetInterval(5 * time.Second)

	if err := scheduler.AddTask(task); err != nil {
		log.Fatal("Failed to add task:", err)
	}

	time.Sleep(30 * time.Second)
}

// LoggerMigrationExample 演示从基础日志配置迁移到扩展配置
func LoggerMigrationExample() {
	fmt.Println("=== 日志配置迁移示例 ===")

	// 步骤1: 原有的基础配置
	fmt.Println("步骤1: 使用基础日志配置")
	oldConfig := DefaultSchedulerConfig()
	oldConfig.LogLevel = LogLevelInfo

	s1, err := NewScheduler(oldConfig)
	if err != nil {
		log.Fatal("Failed to create scheduler with old config:", err)
	}
	s1.Start()
	time.Sleep(5 * time.Second)
	s1.Stop()

	// 步骤2: 迁移到扩展配置
	fmt.Println("步骤2: 迁移到扩展日志配置")
	newConfig := DefaultSchedulerConfig()
	// 保留原有基础配置作为后备
	newConfig.LogLevel = LogLevelInfo
	// 添加扩展配置（优先级更高）
	newConfig.LoggerConfig = &LoggerConfig{
		Level:         LogLevelInfo,
		Output:        "logs/migrated.log",
		EnableConsole: true,
		UseZapLogger:  true,
		Format:        "json",
		Development:   false,
		File: FileLogConfig{
			Enabled:    true,
			Filename:   "logs/migrated.log",
			MaxSize:    50,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   true,
			LocalTime:  true,
		},
	}

	s2, err := NewScheduler(newConfig)
	if err != nil {
		log.Fatal("Failed to create scheduler with new config:", err)
	}
	s2.Start()
	time.Sleep(5 * time.Second)
	s2.Stop()

	fmt.Println("迁移完成！现在使用高性能Zap日志器")
}
