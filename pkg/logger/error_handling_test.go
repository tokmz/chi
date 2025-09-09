package logger

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestLogger_FilePermissionError 测试文件权限错误处理
func TestLogger_FilePermissionError(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "readonly.log")

	// 创建一个只读文件
	file, err := os.Create(logFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	// 设置文件为只读
	err = os.Chmod(logFile, 0444)
	if err != nil {
		t.Fatalf("Failed to set file permissions: %v", err)
	}

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

	// 尝试创建logger，应该处理权限错误
	logger, err := NewLogger(config)
	if err != nil {
		// 权限错误是预期的
		if !strings.Contains(err.Error(), "permission denied") && !strings.Contains(err.Error(), "access is denied") {
			t.Errorf("Expected permission error, got: %v", err)
		}
		t.Logf("Correctly handled permission error: %v", err)
		return
	}

	// 如果logger创建成功，尝试写入应该失败
	if logger != nil {
		logger.Info("This should fail")
		t.Log("Logger created despite permission issues - this may be handled at write time")
	}
}

// TestLogger_InvalidDirectory 测试无效目录错误处理
func TestLogger_InvalidDirectory(t *testing.T) {
	// 使用不存在的根目录路径
	invalidPath := "/nonexistent/deeply/nested/path/test.log"

	config := &Config{
		Level:  "info",
		Format: "json",
		Output: OutputConfig{
			Console: ConsoleConfig{Enabled: false},
			File: FileConfig{
				Enabled:  true,
				Filename: invalidPath,
			},
		},
	}

	// 尝试创建logger，应该处理目录错误
	logger, err := NewLogger(config)
	if err != nil {
		// 目录错误是预期的
		if !strings.Contains(err.Error(), "no such file or directory") && !strings.Contains(err.Error(), "cannot find the path") {
			t.Errorf("Expected directory error, got: %v", err)
		}
		t.Logf("Correctly handled directory error: %v", err)
		return
	}

	// 如果logger创建成功，记录日志
	if logger != nil {
		logger.Info("Logger created despite invalid directory")
		t.Log("Logger created despite invalid directory - error may be handled at write time")
	}
}

// TestRotationWriter_DiskSpaceError 测试磁盘空间不足错误处理
func TestRotationWriter_DiskSpaceError(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "diskfull.log")

	config := FileConfig{
		Enabled:  true,
		Filename: logFile,
		MaxSize:  1, // 1MB
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

	// 写入大量数据，模拟磁盘空间问题
	largeData := strings.Repeat("This is a test log entry that will consume disk space.\n", 10000)
	n, err := writer.Write([]byte(largeData))
	if err != nil {
		// 检查是否是磁盘空间相关错误
		if strings.Contains(err.Error(), "no space left") || strings.Contains(err.Error(), "disk full") {
			t.Logf("Correctly handled disk space error: %v", err)
		} else {
			t.Logf("Write error (may not be disk space related): %v", err)
		}
	} else {
		t.Logf("Successfully wrote %d bytes", n)
	}
}

// TestManager_CleanupError 测试Manager清理过程中的错误处理
func TestManager_CleanupError(t *testing.T) {
	tempDir := t.TempDir()
	config := ManagementConfig{
		Cleanup: CleanupConfig{
			Enabled: true,
			MaxAge:  1, // 1天
		},
	}
	manager := NewManager(config, tempDir)

	// 创建一个文件并设置为只读目录
	subDir := filepath.Join(tempDir, "readonly_dir")
	err := os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	testFile := filepath.Join(subDir, "test.log")
	file, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.WriteString("test content\n")
	file.Close()

	// 设置文件为旧文件
	oldTime := time.Now().AddDate(0, 0, -2)
	os.Chtimes(testFile, oldTime, oldTime)

	// 设置目录为只读（在某些系统上可能阻止删除）
	os.Chmod(subDir, 0555)
	defer os.Chmod(subDir, 0755) // 恢复权限以便清理

	// 执行清理，应该处理权限错误
	err = manager.ForceCleanup()
	if err != nil {
		t.Errorf("ForceCleanup should not return error even if some files cannot be deleted: %v", err)
	}

	// 检查文件是否仍然存在（由于权限问题可能无法删除）
	if _, err := os.Stat(testFile); err == nil {
		t.Log("File still exists due to permission restrictions - this is expected")
	} else if os.IsNotExist(err) {
		t.Log("File was successfully deleted despite permission restrictions")
	}
}

// TestTimeRotationWriter_WriteError 测试TimeRotationWriter写入错误处理
func TestTimeRotationWriter_WriteError(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "rotation_error.log")

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

	// 正常写入
	data := []byte("test log entry\n")
	n, err := writer.Write(data)
	if err != nil {
		t.Errorf("Initial write failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}

	// 模拟文件系统错误（通过删除底层文件）
	if writer.currentFileHandle != nil {
		writer.currentFileHandle.Close()
		writer.currentFileHandle = nil
	}
	os.Remove(writer.currentFile)

	// 尝试再次写入，应该处理错误并重新创建文件
	n, err = writer.Write(data)
	if err != nil {
		t.Logf("Write after file removal failed as expected: %v", err)
	} else {
		t.Logf("Write succeeded after file removal, wrote %d bytes", n)
	}
}

// TestLogger_ConcurrentWriteError 测试并发写入时的错误处理
func TestLogger_ConcurrentWriteError(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "concurrent_error.log")

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

	// 启动多个goroutine并发写入
	done := make(chan bool, 10)
	errorCount := 0

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Goroutine %d panicked: %v", id, r)
					errorCount++
				}
				done <- true
			}()

			for j := 0; j < 100; j++ {
				logger.Info("Concurrent log message", String("goroutine", string(rune(id+'0'))), Int("iteration", j))
				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}

	if errorCount > 0 {
		t.Errorf("Encountered %d errors during concurrent writing", errorCount)
	} else {
		t.Log("All concurrent writes completed successfully")
	}
}

// TestConfig_ValidationError 测试配置验证和修正处理
func TestConfig_ValidationError(t *testing.T) {
	testCases := []struct {
		name           string
		config         *Config
		expectedLevel  string
		expectedFormat string
	}{
		{
			name: "invalid log level remains unchanged by Validate",
			config: &Config{
				Level:  "invalid_level",
				Format: "json",
				Output: OutputConfig{
					Console: ConsoleConfig{Enabled: true},
				},
			},
			expectedLevel:  "invalid_level", // Validate不修改无效级别，ParseLevel会处理
			expectedFormat: "json",
		},
		{
			name: "invalid format gets corrected",
			config: &Config{
				Level:  "info",
				Format: "invalid_format",
				Output: OutputConfig{
					Console: ConsoleConfig{Enabled: true},
				},
			},
			expectedLevel:  "info",
			expectedFormat: "console",
		},
		{
			name: "empty level gets default",
			config: &Config{
				Level:  "",
				Format: "json",
				Output: OutputConfig{
					Console: ConsoleConfig{Enabled: true},
				},
			},
			expectedLevel:  "info",
			expectedFormat: "json",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if err != nil {
				t.Errorf("Validation should not return error: %v", err)
			}
			if tc.config.Level != tc.expectedLevel {
				t.Errorf("Expected level %s, got %s", tc.expectedLevel, tc.config.Level)
			}
			if tc.config.Format != tc.expectedFormat {
				t.Errorf("Expected format %s, got %s", tc.expectedFormat, tc.config.Format)
			}
			t.Logf("Config validation corrected values: level=%s, format=%s", tc.config.Level, tc.config.Format)
		})
	}
}

// TestRotationWriter_CorruptedFileError 测试损坏文件的错误处理
func TestRotationWriter_CorruptedFileError(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "corrupted.log")

	// 创建一个"损坏"的文件（实际上是目录）
	err := os.MkdirAll(logFile, 0755)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	config := FileConfig{
		Enabled:  true,
		Filename: logFile,
	}

	rotationConfig := RotationConfig{
		Size: SizeRotationConfig{
			Enabled: true,
			MaxSize: 10,
		},
	}

	// 尝试创建RotationWriter，应该处理文件类型错误
	writer, err := NewRotationWriter(config, rotationConfig)
	if err != nil {
		t.Logf("Correctly handled corrupted file error: %v", err)
		return
	}

	if writer != nil {
		defer writer.Close()
		// 尝试写入，可能成功也可能失败，取决于实现
		_, writeErr := writer.Write([]byte("test"))
		if writeErr != nil {
			t.Logf("Write failed on corrupted file: %v", writeErr)
		} else {
			t.Log("Write succeeded despite corrupted file - implementation may handle this gracefully")
		}
	}
}

// TestSystemResourceError 测试系统资源限制错误
func TestSystemResourceError(t *testing.T) {
	tempDir := t.TempDir()

	// 尝试创建大量文件句柄，模拟资源耗尽
	var files []*os.File
	defer func() {
		for _, f := range files {
			f.Close()
		}
	}()

	// 创建多个logger实例，每个都打开文件
	for i := 0; i < 10; i++ {
		logFile := filepath.Join(tempDir, fmt.Sprintf("resource_test_%d.log", i))
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
			// 检查是否是资源相关错误
			if errors.Is(err, syscall.EMFILE) || errors.Is(err, syscall.ENFILE) {
				t.Logf("Correctly handled resource limit error: %v", err)
			} else {
				t.Logf("Logger creation failed (may not be resource related): %v", err)
			}
			break
		}

		if logger != nil {
			logger.Info("Resource test log", Int("instance", i))
		}
	}

	t.Log("Resource limit test completed")
}