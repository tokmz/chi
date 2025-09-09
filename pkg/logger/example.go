package logger

import (
	"context"
	"errors"
	"time"
)

// ExampleBasicUsage 基本使用示例
func ExampleBasicUsage() {
	// 使用默认配置初始化全局日志记录器
	if err := InitGlobal(nil); err != nil {
		panic(err)
	}

	// 基本日志记录
	Info("应用程序启动")
	Debug("调试信息")
	Warn("警告信息")
	Error("错误信息")

	// 格式化日志
	Infof("用户 %s 登录成功", "张三")
	Errorf("连接数据库失败: %v", errors.New("connection timeout"))

	// 键值对日志
	Infow("用户操作",
		"user_id", 12345,
		"action", "login",
		"ip", "192.168.1.100",
		"timestamp", time.Now(),
	)

	// 结构化字段日志
	Info("订单创建成功",
		String("order_id", "ORD-2024-001"),
		Int64("user_id", 12345),
		Float64("amount", 99.99),
		Time("created_at", time.Now()),
	)

	// 同步日志（确保所有日志都被写入）
	Sync()
}

// ExampleCustomConfig 自定义配置示例
func ExampleCustomConfig() {
	// 创建自定义配置
	config := &Config{
		Level:  "debug",
		Format: "json",
		Output: OutputConfig{
			Console: ConsoleConfig{
				Enabled:    true,
				Colorful:   true,
				TimeFormat: "2006-01-02 15:04:05",
			},
			File: FileConfig{
				Enabled:     true,
				Filename:    "logs/app.log",
				MaxSize:     50, // 50MB
				MaxBackups:  5,
				MaxAge:      7, // 7天
				Compress:    true,
				LocalTime:   true,
				LevelFilter: "info", // 文件只记录info及以上级别
			},
		},
		Caller: CallerConfig{
			Enabled:  true,
			FullPath: false,
			Skip:     1,
		},
		Rotation: RotationConfig{
			Size: SizeRotationConfig{
				Enabled: true,
				MaxSize: 50,
			},
			Time: TimeRotationConfig{
				Enabled:    true,
				Interval:   "day",
				RotateTime: "00:00",
			},
		},
		Management: ManagementConfig{
			Cleanup: CleanupConfig{
				Enabled:  true,
				MaxAge:   30,
				Interval: 24 * time.Hour,
			},
			Compression: CompressionConfig{
				Enabled:   true,
				Delay:     24,
				Algorithm: "gzip",
			},
		},
	}

	// 创建日志记录器
	logger, err := NewLogger(config)
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	// 使用自定义日志记录器
	logger.Info("使用自定义配置的日志记录器")
	logger.Debug("这条调试信息会显示在控制台，但不会写入文件")
	logger.Error("这条错误信息会同时显示在控制台和写入文件")
}

// ExampleMultiFileOutput 多文件输出示例
func ExampleMultiFileOutput() {
	config := &Config{
		Level:  "debug",
		Format: "json",
		Output: OutputConfig{
			Console: ConsoleConfig{
				Enabled:  true,
				Colorful: true,
			},
			MultiFile: []FileConfig{
				{
					Enabled:     true,
					Filename:    "logs/app.log",
					MaxSize:     100,
					MaxBackups:  10,
					MaxAge:      30,
					Compress:    true,
					LocalTime:   true,
					LevelFilter: "", // 所有级别
				},
				{
					Enabled:     true,
					Filename:    "logs/error.log",
					MaxSize:     50,
					MaxBackups:  5,
					MaxAge:      30,
					Compress:    true,
					LocalTime:   true,
					LevelFilter: "error", // 只记录错误级别
				},
				{
					Enabled:     true,
					Filename:    "logs/access.log",
					MaxSize:     200,
					MaxBackups:  20,
					MaxAge:      7,
					Compress:    true,
					LocalTime:   true,
					LevelFilter: "info", // 记录info及以上级别
				},
			},
		},
	}

	logger, err := NewLogger(config)
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	// 不同级别的日志会根据配置写入不同的文件
	logger.Debug("调试信息 - 只会写入app.log")
	logger.Info("访问信息 - 会写入app.log和access.log")
	logger.Error("错误信息 - 会写入所有三个文件")
}

// ExampleWithFields 带字段的日志示例
func ExampleWithFields() {
	logger := GetGlobal()

	// 创建带有公共字段的子记录器
	userLogger := logger.With(
		String("service", "user-service"),
		String("version", "v1.0.0"),
	)

	// 使用子记录器记录日志
	userLogger.Info("用户注册",
		String("user_id", "12345"),
		String("email", "user@example.com"),
		Time("registered_at", time.Now()),
	)

	// 创建命名子记录器
	authLogger := logger.Named("auth")
	authLogger.Info("用户认证成功",
		String("user_id", "12345"),
		String("method", "password"),
		Duration("duration", 150*time.Millisecond),
	)
}

// ExampleErrorHandling 错误处理示例
func ExampleErrorHandling() {
	logger := GetGlobal()

	// 记录错误信息
	err := errors.New("数据库连接失败")
	logger.Error("操作失败",
		Err(err),
		String("operation", "user_query"),
		Int("retry_count", 3),
	)

	// 使用错误格式化
	logger.Errorf("处理请求失败: %v", err)

	// 键值对错误记录
	logger.Errorw("API调用失败",
		"error", err,
		"endpoint", "/api/users",
		"status_code", 500,
		"response_time", 2.5,
	)
}

// ExampleContextLogging 上下文日志示例
func ExampleContextLogging() {
	logger := GetGlobal()

	// 模拟HTTP请求处理
	ctx := context.Background()
	requestID := "req-12345"
	userID := "user-67890"

	// 创建带有请求上下文的日志记录器
	requestLogger := logger.With(
		String("request_id", requestID),
		String("user_id", userID),
	)

	// 在整个请求处理过程中使用相同的记录器
	requestLogger.Info("开始处理请求")

	// 模拟业务逻辑
	processOrder(ctx, requestLogger)

	requestLogger.Info("请求处理完成")
}

func processOrder(ctx context.Context, logger *Logger) {
	logger.Info("开始处理订单")

	// 模拟订单处理
	time.Sleep(100 * time.Millisecond)

	logger.Info("订单处理成功",
		String("order_id", "ORD-001"),
		Float64("amount", 99.99),
		Duration("processing_time", 100*time.Millisecond),
	)
}

// ExamplePerformanceLogging 性能日志示例
func ExamplePerformanceLogging() {
	logger := GetGlobal()

	// 记录方法执行时间
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		logger.Info("方法执行完成",
			String("method", "ExamplePerformanceLogging"),
			Duration("duration", duration),
		)
	}()

	// 模拟一些工作
	time.Sleep(50 * time.Millisecond)

	// 记录中间步骤
	logger.Debug("中间步骤完成",
		Duration("elapsed", time.Since(start)),
	)

	time.Sleep(50 * time.Millisecond)
}

// ExampleLogRotation 日志分割示例
func ExampleLogRotation() {
	config := &Config{
		Level:  "info",
		Format: "json",
		Output: OutputConfig{
			File: FileConfig{
				Enabled:   true,
				Filename:  "logs/rotation-test.log",
				MaxSize:   1, // 1MB，便于测试
				Compress:  true,
				LocalTime: true,
			},
		},
		Rotation: RotationConfig{
			Size: SizeRotationConfig{
				Enabled: true,
				MaxSize: 1, // 1MB
			},
			Time: TimeRotationConfig{
				Enabled:  true,
				Interval: "hour", // 每小时分割
			},
		},
	}

	logger, err := NewLogger(config)
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	// 生成大量日志以触发分割
	for i := 0; i < 1000; i++ {
		logger.Info("测试日志分割",
			Int("index", i),
			String("data", "这是一条用于测试日志分割功能的长消息，包含足够的内容来触发文件大小限制"),
			Time("timestamp", time.Now()),
		)
	}
}

// ExampleLogManagement 日志管理示例
func ExampleLogManagement() {
	// 创建日志管理器
	managementConfig := ManagementConfig{
		Cleanup: CleanupConfig{
			Enabled:  true,
			MaxAge:   7,         // 保留7天
			Interval: time.Hour, // 每小时检查一次
		},
		Compression: CompressionConfig{
			Enabled:   true,
			Delay:     1, // 1小时后压缩
			Algorithm: "gzip",
		},
	}

	manager := NewManager(managementConfig, "logs")

	// 启动管理器
	if err := manager.Start(); err != nil {
		panic(err)
	}
	defer manager.Stop()

	// 获取日志统计信息
	stats, err := manager.GetStats()
	if err == nil {
		Info("日志统计信息",
			Int("total_files", stats.TotalFiles),
			Int64("total_size", stats.TotalSize),
			Int("compressed_files", stats.CompressedFiles),
			Int("uncompressed_files", stats.UncompressedFiles),
		)
	}

	// 手动触发清理
	manager.ForceCleanup()

	// 手动触发压缩
	manager.ForceCompression()
}

// ExampleDevelopmentMode 开发模式示例
func ExampleDevelopmentMode() {
	config := &Config{
		Level:       "debug",
		Format:      "console",
		Development: true, // 开发模式
		Output: OutputConfig{
			Console: ConsoleConfig{
				Enabled:    true,
				Colorful:   true,
				TimeFormat: "15:04:05",
			},
		},
		Caller: CallerConfig{
			Enabled:  true,
			FullPath: false,
		},
	}

	logger, err := NewLogger(config)
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	// 开发模式下的日志输出更适合调试
	logger.Debug("调试信息")
	logger.Info("信息")
	logger.Warn("警告")
	logger.Error("错误")
}

// ExampleProductionMode 生产模式示例
func ExampleProductionMode() {
	config := &Config{
		Level:       "info",
		Format:      "json",
		Development: false, // 生产模式
		Output: OutputConfig{
			File: FileConfig{
				Enabled:   true,
				Filename:  "logs/production.log",
				MaxSize:   100,
				Compress:  true,
				LocalTime: true,
			},
		},
		Caller: CallerConfig{
			Enabled: false, // 生产环境可以关闭调用信息以提高性能
		},
		Sampling: SamplingConfig{
			Enabled:    true,
			Initial:    100,
			Thereafter: 100, // 采样以减少日志量
		},
	}

	logger, err := NewLogger(config)
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	// 生产模式下的日志输出更注重性能和结构化
	logger.Info("服务启动",
		String("version", "v1.0.0"),
		String("environment", "production"),
	)

	logger.Error("服务异常",
		String("error", "database connection failed"),
		Int("retry_count", 3),
	)
}
