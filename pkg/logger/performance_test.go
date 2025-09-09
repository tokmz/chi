package logger

import (
	"os"
	"strings"
	"testing"
	"time"
)

// BenchmarkTimeRotationWriter_Write 测试时间分割写入器的写入性能
func BenchmarkTimeRotationWriter_Write(b *testing.B) {
	tempDir := "/tmp/logger_bench"
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	tempFile := tempDir + "/test_rotation.log"

	config := TimeRotationConfig{
		Interval: "day",
	}

	writer := NewTimeRotationWriter(tempFile, config)
	defer writer.Close()

	data := []byte("This is a test log message for performance testing\n")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			writer.Write(data)
		}
	})
}

// BenchmarkLogger_Info 测试日志记录器的Info方法性能
func BenchmarkLogger_Info(b *testing.B) {
	config := DefaultConfig()
	config.Output.Console.Enabled = false
	config.Output.File = FileConfig{
		Enabled:     true,
		Filename:    "/tmp/bench_test.log",
		LevelFilter: "info",
	}

	logger, err := NewLogger(config)
	if err != nil {
		b.Fatal(err)
	}
	defer logger.Close()
	defer os.Remove("/tmp/bench_test.log")

	// 确保目录存在
	os.MkdirAll("/tmp", 0755)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("benchmark test message")
		}
	})
}

// BenchmarkManager_GetLogFiles 测试管理器获取日志文件列表的性能
func BenchmarkManager_GetLogFiles(b *testing.B) {
	tempDir := "/tmp/logger_bench"
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	// 创建一些测试文件
	for i := 0; i < 100; i++ {
		file, _ := os.Create(tempDir + "/test_" + string(rune(i)) + ".log")
		file.Close()
	}

	config := ManagementConfig{
		Cleanup: CleanupConfig{
			Enabled: true,
			MaxAge:  7,
		},
	}

	manager := NewManager(config, tempDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.getLogFiles()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestTimeRotationWriter_BufferedWrite 测试缓冲写入功能
func TestTimeRotationWriter_BufferedWrite(t *testing.T) {
	tempDir := "/tmp/logger_test"
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	tempFile := tempDir + "/test_buffered.log"

	config := TimeRotationConfig{
		Interval: "day",
	}

	writer := NewTimeRotationWriter(tempFile, config)
	defer writer.Close()

	// 写入一些数据
	data := []byte("test message\n")
	for i := 0; i < 10; i++ {
		n, err := writer.Write(data)
		if err != nil {
			t.Fatal(err)
		}
		if n != len(data) {
			t.Errorf("expected %d bytes written, got %d", len(data), n)
		}
	}

	// 确保数据被写入
	err := writer.Close()
	if err != nil {
		t.Fatal(err)
	}

	// 检查时间分割生成的文件（带时间戳）
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	// 验证至少有一个日志文件被创建
	logFileFound := false
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "test_buffered") && strings.HasSuffix(file.Name(), ".log") {
			logFileFound = true
			break
		}
	}

	if !logFileFound {
		t.Error("no log file was created with time rotation pattern")
	}
}

// TestManager_FileCache 测试文件缓存功能
func TestManager_FileCache(t *testing.T) {
	tempDir := "/tmp/logger_cache_test"
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	file, _ := os.Create(tempDir + "/test.log")
	file.Close()

	config := ManagementConfig{
		Cleanup: CleanupConfig{
			Enabled: true,
			MaxAge:  7,
		},
	}

	manager := NewManager(config, tempDir)

	// 第一次调用应该扫描文件系统
	start := time.Now()
	files1, err := manager.getLogFiles()
	if err != nil {
		t.Fatal(err)
	}
	firstCallDuration := time.Since(start)

	// 第二次调用应该使用缓存
	start = time.Now()
	files2, err := manager.getLogFiles()
	if err != nil {
		t.Fatal(err)
	}
	secondCallDuration := time.Since(start)

	// 验证结果一致
	if len(files1) != len(files2) {
		t.Errorf("file count mismatch: %d vs %d", len(files1), len(files2))
	}

	// 第二次调用应该更快（使用缓存）
	if secondCallDuration >= firstCallDuration {
		t.Logf("Cache may not be working optimally. First: %v, Second: %v", firstCallDuration, secondCallDuration)
	}
}
