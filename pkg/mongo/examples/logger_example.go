package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"chi/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	// 示例1: 使用默认日志配置
	fmt.Println("=== 示例1: 使用默认日志配置 ===")
	basicExample()

	// 示例2: 使用自定义日志配置
	fmt.Println("\n=== 示例2: 使用自定义日志配置 ===")
	customLoggerExample()

	// 示例3: 慢查询监控
	fmt.Println("\n=== 示例3: 慢查询监控 ===")
	slowQueryExample()

	// 示例4: 运行时更新日志配置
	fmt.Println("\n=== 示例4: 运行时更新日志配置 ===")
	dynamicConfigExample()
}

// basicExample 基础日志使用示例
func basicExample() {
	// 使用默认配置创建客户端
	config := mongo.DefaultConfig()
	config.URI = "mongodb://localhost:27017"
	config.Database = "test"

	client, err := mongo.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// 获取集合
	collection := client.Collection("users")

	// 执行一些操作，观察日志输出
	ctx := context.Background()
	_, err = collection.InsertOne(ctx, bson.M{"name": "Alice", "age": 30})
	if err != nil {
		fmt.Printf("Insert error: %v\n", err)
	}

	err = collection.FindOne(ctx, bson.M{"name": "Alice"}).Err()
	if err != nil {
		fmt.Printf("Find error: %v\n", err)
	}
}

// customLoggerExample 自定义日志配置示例
func customLoggerExample() {
	// 创建自定义日志配置
	loggerConfig := &mongo.MongoLoggerConfig{
		Level:         mongo.LogLevelDebug,
		EnableConsole: true,
		UseZapLogger:  false, // 使用简单的默认logger
		Format:        "console",
		Development:   true,
		File: mongo.FileLogConfig{
			Enabled:  true,
			Filename: "mongo_debug.log",
			MaxSize:  10, // 10MB
		},
		Console: mongo.ConsoleLogConfig{
			Enabled:    true,
			Colorful:   true,
			TimeFormat: "15:04:05",
		},
		Mongo: mongo.MongoSpecificLogConfig{
			SlowQuery: mongo.SlowQueryLogConfig{
				Enabled:   true,
				Threshold: 50 * time.Millisecond, // 50ms阈值
				LogQuery:  true,
				LogResult: true,
			},
			Connection: mongo.ConnectionLogConfig{
				Enabled:     true,
				LogConnect:  true,
				LogClose:    true,
				LogPoolInfo: true,
			},
			Operation: mongo.OperationLogConfig{
				Enabled:        true,
				LogCRUD:        true,
				LogAggregation: true,
				LogTransaction: true,
				LogIndex:       true,
			},
		},
	}

	// 创建配置
	config := mongo.DefaultConfig()
	config.URI = "mongodb://localhost:27017"
	config.Database = "test"
	config.Logger = loggerConfig

	client, err := mongo.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// 执行操作
	collection := client.Collection("products")
	ctx := context.Background()

	// 插入文档
	_, err = collection.InsertOne(ctx, bson.M{
		"name":  "Laptop",
		"price": 999.99,
		"tags":  []string{"electronics", "computer"},
	})
	if err != nil {
		fmt.Printf("Insert error: %v\n", err)
	}

	// 查询文档
	cursor, err := collection.Find(ctx, bson.M{"price": bson.M{"$gt": 500}})
	if err != nil {
		fmt.Printf("Find error: %v\n", err)
	} else {
		cursor.Close(ctx)
	}
}

// slowQueryExample 慢查询监控示例
func slowQueryExample() {
	config := mongo.DefaultConfig()
	config.URI = "mongodb://localhost:27017"
	config.Database = "test"

	// 配置慢查询监控
	loggerConfig := mongo.DefaultMongoLoggerConfig()
	loggerConfig.Mongo.SlowQuery.Enabled = true
	loggerConfig.Mongo.SlowQuery.Threshold = 10 * time.Millisecond // 很低的阈值用于演示
	config.Logger = loggerConfig

	client, err := mongo.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	collection := client.Collection("large_collection")
	ctx := context.Background()

	// 执行一个可能较慢的查询
	monitor := client.GetSlowQueryMonitor()
	err = monitor.MonitorQuery(ctx, "find", "large_collection", "test", func() error {
		// 模拟慢查询
		time.Sleep(50 * time.Millisecond)
		_, err := collection.Find(ctx, bson.M{})
		return err
	})

	if err != nil {
		fmt.Printf("Query error: %v\n", err)
	}

	// 获取慢查询统计信息
	stats := client.GetSlowQueryStats()
	fmt.Printf("慢查询统计:\n")
	fmt.Printf("  总查询数: %d\n", stats.TotalQueries)
	fmt.Printf("  慢查询数: %d\n", stats.SlowQueries)
	fmt.Printf("  慢查询率: %.2f%%\n", stats.SlowQueryRate)
	fmt.Printf("  最大耗时: %v\n", stats.MaxTime)
	if stats.TotalQueries > 0 {
		avgTime := stats.TotalTime / time.Duration(stats.TotalQueries)
		fmt.Printf("  平均耗时: %v\n", avgTime)
	}
}

// dynamicConfigExample 动态配置更新示例
func dynamicConfigExample() {
	config := mongo.DefaultConfig()
	config.URI = "mongodb://localhost:27017"
	config.Database = "test"

	client, err := mongo.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// 获取当前日志配置
	currentConfig := client.GetLoggerConfig()
	fmt.Printf("当前日志级别: %v\n", currentConfig.Level)

	// 更新日志配置为Debug级别
	newConfig := mongo.DefaultMongoLoggerConfig()
	newConfig.Level = mongo.LogLevelDebug
	newConfig.Mongo.SlowQuery.Threshold = 1 * time.Millisecond // 更敏感的阈值

	err = client.UpdateLoggerConfig(newConfig)
	if err != nil {
		fmt.Printf("Failed to update logger config: %v\n", err)
		return
	}

	fmt.Println("日志配置已更新为Debug级别")

	// 执行一些操作来观察新的日志输出
	collection := client.Collection("test_collection")
	ctx := context.Background()

	_, err = collection.InsertOne(ctx, bson.M{"test": "dynamic_config"})
	if err != nil {
		fmt.Printf("Insert error: %v\n", err)
	}

	// 调整慢查询阈值
	client.SetSlowQueryThreshold(5 * time.Millisecond)
	fmt.Println("慢查询阈值已调整为5ms")

	// 重置慢查询统计
	client.ResetSlowQueryStats()
	fmt.Println("慢查询统计已重置")
}

// 高级用法示例
func advancedUsageExample() {
	// 创建具有完整配置的客户端
	loggerConfig := &mongo.MongoLoggerConfig{
		Level:         mongo.LogLevelInfo,
		EnableConsole: true,
		UseZapLogger:  true, // 使用高性能zap logger
		Format:        "json",
		Development:   false,
		File: mongo.FileLogConfig{
			Enabled:    true,
			Filename:   "/var/log/mongo/app.log",
			MaxSize:    100, // 100MB
			MaxBackups: 5,
			MaxAge:     30, // 30天
			Compress:   true,
			LocalTime:  true,
		},
		Console: mongo.ConsoleLogConfig{
			Enabled:    false, // 生产环境关闭控制台输出
			Colorful:   false,
			TimeFormat: "2006-01-02T15:04:05Z07:00",
		},
		Caller: mongo.CallerLogConfig{
			Enabled:  true,
			FullPath: false,
			Skip:     1,
		},
		Rotation: mongo.RotationLogConfig{
			Enabled:  true,
			MaxSize:  100,
			Interval: 24 * time.Hour,
			Pattern:  "mongo-%Y%m%d.log",
		},
		Sampling: mongo.SamplingLogConfig{
			Enabled:    true,
			Initial:    100,
			Thereafter: 100,
		},
		Mongo: mongo.MongoSpecificLogConfig{
			SlowQuery: mongo.SlowQueryLogConfig{
				Enabled:   true,
				Threshold: 100 * time.Millisecond,
				LogQuery:  true,
				LogResult: false, // 生产环境不记录结果
			},
			Connection: mongo.ConnectionLogConfig{
				Enabled:     true,
				LogConnect:  true,
				LogClose:    true,
				LogPoolInfo: false,
			},
			Operation: mongo.OperationLogConfig{
				Enabled:        false, // 生产环境关闭详细操作日志
				LogCRUD:        false,
				LogAggregation: false,
				LogTransaction: true,
				LogIndex:       true,
			},
			Error: mongo.ErrorLogConfig{
				Enabled:       true,
				LogStackTrace: true,
				LogContext:    true,
			},
		},
	}

	config := &mongo.Config{
		URI:      "mongodb://localhost:27017",
		Database: "production",
		Pool: mongo.PoolConfig{
			MaxPoolSize:     100,
			MinPoolSize:     10,
			MaxConnIdleTime: 30 * time.Minute,
			MaxConnLifetime: 1 * time.Hour,
		},
		Logger: loggerConfig,
		Timeout: mongo.TimeoutConfig{
			Connect:         10 * time.Second,
			ServerSelection: 5 * time.Second,
			Socket:          30 * time.Second,
			Heartbeat:       10 * time.Second,
		},
	}

	client, err := mongo.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create production client: %v", err)
	}
	defer client.Close()

	fmt.Println("生产环境客户端创建成功，具有完整的日志配置")
}