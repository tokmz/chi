package logger

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestLogger_ConcurrentAccess 测试Logger的并发访问安全性
func TestLogger_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "concurrent.log")

	config := &Config{
		Level:  "info",
		Format: "json",
		Output: OutputConfig{
			Console: ConsoleConfig{Enabled: false},
			File: FileConfig{
				Enabled:  true,
				Filename: logFile,
			},
		},
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	const numGoroutines = 50
	const messagesPerGoroutine = 100

	var wg sync.WaitGroup
	errorChan := make(chan error, numGoroutines)

	// 启动多个goroutine并发写入日志
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					errorChan <- fmt.Errorf("goroutine %d panicked: %v", id, r)
				}
			}()

			for j := 0; j < messagesPerGoroutine; j++ {
				logger.Info("Concurrent message",
					Int("goroutine", id),
					Int("message", j),
					String("timestamp", time.Now().Format(time.RFC3339)),
				)
				// 随机延迟，增加竞争条件
				if j%10 == 0 {
					time.Sleep(1 * time.Millisecond)
				}
			}
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()
	close(errorChan)

	// 检查是否有错误
	for err := range errorChan {
		t.Errorf("Concurrent access error: %v", err)
	}

	t.Logf("Successfully completed %d concurrent goroutines with %d messages each",
		numGoroutines, messagesPerGoroutine)
}

// TestRotationWriter_ConcurrentRotation 测试RotationWriter的并发轮转安全性
func TestRotationWriter_ConcurrentRotation(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "rotation_concurrent.log")

	config := FileConfig{
		Enabled:  true,
		Filename: logFile,
		MaxSize:  1, // 1MB，容易触发轮转
	}

	rotationConfig := RotationConfig{
		Size: SizeRotationConfig{
			Enabled: true,
			MaxSize: 1,
		},
	}

	writer, err := NewRotationWriter(config, rotationConfig)
	if err != nil {
		t.Fatalf("Failed to create RotationWriter: %v", err)
	}
	defer writer.Close()

	const numGoroutines = 20
	const writesPerGoroutine = 50

	var wg sync.WaitGroup
	errorChan := make(chan error, numGoroutines)

	// 启动多个goroutine并发写入，触发轮转
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					errorChan <- fmt.Errorf("goroutine %d panicked: %v", id, r)
				}
			}()

			for j := 0; j < writesPerGoroutine; j++ {
				// 写入较大的数据块，容易触发轮转
				data := fmt.Sprintf("Goroutine %d, Write %d: %s\n",
					id, j, string(make([]byte, 1024))) // 1KB数据

				_, writeErr := writer.Write([]byte(data))
				if writeErr != nil {
					errorChan <- fmt.Errorf("write error in goroutine %d: %v", id, writeErr)
					return
				}

				// 偶尔手动触发轮转
				if j%20 == 0 {
					rotateErr := writer.Rotate()
					if rotateErr != nil {
						errorChan <- fmt.Errorf("rotation error in goroutine %d: %v", id, rotateErr)
						return
					}
				}
			}
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()
	close(errorChan)

	// 检查是否有错误
	for err := range errorChan {
		t.Errorf("Concurrent rotation error: %v", err)
	}

	t.Logf("Successfully completed concurrent rotation test with %d goroutines", numGoroutines)
}

// TestTimeRotationWriter_ConcurrentAccess 测试TimeRotationWriter的并发访问
func TestTimeRotationWriter_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "time_rotation_concurrent.log")

	config := TimeRotationConfig{
		Enabled:    true,
		Interval:   "hour",
		RotateTime: "00:00",
	}

	writer := NewTimeRotationWriter(logFile, config)
	if writer == nil {
		t.Fatal("Failed to create TimeRotationWriter")
	}
	defer writer.Close()

	const numGoroutines = 30
	const writesPerGoroutine = 100

	var wg sync.WaitGroup
	errorChan := make(chan error, numGoroutines)

	// 启动多个goroutine并发写入
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					errorChan <- fmt.Errorf("goroutine %d panicked: %v", id, r)
				}
			}()

			for j := 0; j < writesPerGoroutine; j++ {
				data := fmt.Sprintf("TimeRotation Goroutine %d, Write %d: %s\n",
					id, j, time.Now().Format(time.RFC3339))

				_, writeErr := writer.Write([]byte(data))
				if writeErr != nil {
					errorChan <- fmt.Errorf("write error in goroutine %d: %v", id, writeErr)
					return
				}

				// 随机延迟
				if j%25 == 0 {
					time.Sleep(2 * time.Millisecond)
				}
			}
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()
	close(errorChan)

	// 检查是否有错误
	for err := range errorChan {
		t.Errorf("Concurrent time rotation error: %v", err)
	}

	t.Logf("Successfully completed concurrent time rotation test with %d goroutines", numGoroutines)
}

// TestManager_ConcurrentSafety 测试Manager的并发操作安全性
func TestManager_ConcurrentSafety(t *testing.T) {
	tempDir := t.TempDir()
	config := ManagementConfig{
		Cleanup: CleanupConfig{
			Enabled: true,
			MaxAge:  1,
			Interval: 1 * time.Hour,
		},
		Compression: CompressionConfig{
			Enabled:   true,
			Delay:     1,
			Algorithm: "gzip",
		},
	}

	manager := NewManager(config, tempDir)

	// 启动manager
	err := manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	defer manager.Stop()

	const numGoroutines = 10
	var wg sync.WaitGroup
	errorChan := make(chan error, numGoroutines)

	// 启动多个goroutine并发执行不同操作
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					errorChan <- fmt.Errorf("goroutine %d panicked: %v", id, r)
				}
			}()

			// 根据goroutine ID执行不同操作
			switch id % 4 {
			case 0:
				// 获取统计信息
				for j := 0; j < 20; j++ {
					stats, statsErr := manager.GetStats()
					if statsErr != nil {
						errorChan <- fmt.Errorf("GetStats error in goroutine %d: %v", id, statsErr)
						return
					}
					if stats == nil {
						errorChan <- fmt.Errorf("GetStats returned nil in goroutine %d", id)
						return
					}
					time.Sleep(5 * time.Millisecond)
				}
			case 1:
				// 强制清理
				for j := 0; j < 5; j++ {
					cleanupErr := manager.ForceCleanup()
					if cleanupErr != nil {
						errorChan <- fmt.Errorf("ForceCleanup error in goroutine %d: %v", id, cleanupErr)
						return
					}
					time.Sleep(10 * time.Millisecond)
				}
			case 2:
				// 强制压缩
				for j := 0; j < 5; j++ {
					compressErr := manager.ForceCompression()
					if compressErr != nil {
						errorChan <- fmt.Errorf("ForceCompression error in goroutine %d: %v", id, compressErr)
						return
					}
					time.Sleep(10 * time.Millisecond)
				}
			case 3:
				// 检查运行状态
				for j := 0; j < 30; j++ {
					isRunning := manager.IsRunning()
					if !isRunning {
						errorChan <- fmt.Errorf("Manager not running in goroutine %d", id)
						return
					}
					time.Sleep(3 * time.Millisecond)
				}
			}
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()
	close(errorChan)

	// 检查是否有错误
	for err := range errorChan {
		t.Errorf("Concurrent manager operation error: %v", err)
	}

	t.Logf("Successfully completed concurrent manager operations test with %d goroutines", numGoroutines)
}

// TestGlobalLogger_ConcurrentAccess 测试全局Logger的并发访问安全性
func TestGlobalLogger_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "global_concurrent.log")

	config := &Config{
		Level:  "info",
		Format: "json",
		Output: OutputConfig{
			Console: ConsoleConfig{Enabled: false},
			File: FileConfig{
				Enabled:  true,
				Filename: logFile,
			},
		},
	}

	// 初始化全局logger
	err := InitGlobal(config)
	if err != nil {
		t.Fatalf("Failed to init global logger: %v", err)
	}

	const numGoroutines = 40
	const messagesPerGoroutine = 50

	var wg sync.WaitGroup
	errorChan := make(chan error, numGoroutines)

	// 启动多个goroutine并发使用全局logger
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					errorChan <- fmt.Errorf("goroutine %d panicked: %v", id, r)
				}
			}()

			globalLogger := GetGlobal()
			if globalLogger == nil {
				errorChan <- fmt.Errorf("GetGlobal returned nil in goroutine %d", id)
				return
			}

			for j := 0; j < messagesPerGoroutine; j++ {
				// 使用不同的日志方法
				switch j % 4 {
				case 0:
					Info("Global info message", Int("goroutine", id), Int("message", j))
				case 1:
					Debug("Global debug message", Int("goroutine", id), Int("message", j))
				case 2:
					Warn("Global warn message", Int("goroutine", id), Int("message", j))
				case 3:
					Error("Global error message", Int("goroutine", id), Int("message", j))
				}

				if j%15 == 0 {
					time.Sleep(1 * time.Millisecond)
				}
			}
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()
	close(errorChan)

	// 检查是否有错误
	for err := range errorChan {
		t.Errorf("Concurrent global logger error: %v", err)
	}

	t.Logf("Successfully completed global logger concurrent test with %d goroutines", numGoroutines)
}

// TestLogger_ConcurrentLevelChange 测试并发修改日志级别的安全性
func TestLogger_ConcurrentLevelChange(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "level_change_concurrent.log")

	config := &Config{
		Level:  "info",
		Format: "json",
		Output: OutputConfig{
			Console: ConsoleConfig{Enabled: false},
			File: FileConfig{
				Enabled:  true,
				Filename: logFile,
			},
		},
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	const numGoroutines = 20
	levels := []string{"debug", "info", "warn", "error"}

	var wg sync.WaitGroup
	errorChan := make(chan error, numGoroutines)

	// 启动多个goroutine并发修改级别和写入日志
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					errorChan <- fmt.Errorf("goroutine %d panicked: %v", id, r)
				}
			}()

			for j := 0; j < 50; j++ {
				// 偶尔修改日志级别
				if j%10 == 0 {
					newLevel := levels[j%len(levels)]
					logger.SetLevel(newLevel)
				}

				// 获取当前级别
				currentLevel := logger.GetLevel()
				if currentLevel == "" {
					errorChan <- fmt.Errorf("GetLevel returned empty string in goroutine %d", id)
					return
				}

				// 写入日志
				logger.Info("Level change test",
					Int("goroutine", id),
					Int("iteration", j),
					String("current_level", currentLevel),
				)

				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()
	close(errorChan)

	// 检查是否有错误
	for err := range errorChan {
		t.Errorf("Concurrent level change error: %v", err)
	}

	t.Logf("Successfully completed concurrent level change test with %d goroutines", numGoroutines)
}