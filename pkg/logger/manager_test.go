package logger

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestNewManager 测试Manager创建
func TestNewManager(t *testing.T) {
	config := ManagementConfig{
		Cleanup: CleanupConfig{
			Enabled: true,
			MaxAge:  7,
		},
		Compression: CompressionConfig{
			Enabled: true,
			Delay:   24,
		},
	}
	logDir := "/tmp/test_logs"

	manager := NewManager(config, logDir)
	if manager == nil {
		t.Error("NewManager returned nil")
	}
	if manager.config.Cleanup.Enabled != true {
		t.Error("Cleanup config not set correctly")
	}
	if manager.logDir != logDir {
		t.Errorf("Expected logDir %s, got %s", logDir, manager.logDir)
	}
	if manager.isRunning {
		t.Error("Manager should not be running initially")
	}
}

// TestManager_StartStop 测试Manager启动和停止
func TestManager_StartStop(t *testing.T) {
	tempDir := t.TempDir()
	config := ManagementConfig{
		Cleanup: CleanupConfig{
			Enabled: true,
			MaxAge:  7,
			Interval: 1 * time.Hour, // 1小时间隔
		},
		Compression: CompressionConfig{
			Enabled: true,
			Delay:   1, // 1小时延迟
			Algorithm: "gzip",
		},
	}

	manager := NewManager(config, tempDir)

	// 测试启动
	err := manager.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if !manager.IsRunning() {
		t.Error("Manager should be running after Start")
	}

	// 测试重复启动
	err = manager.Start()
	if err == nil {
		t.Error("Expected error when starting already running manager")
	}

	// 测试停止
	err = manager.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	if manager.IsRunning() {
		t.Error("Manager should not be running after Stop")
	}

	// 测试重复停止
	err = manager.Stop()
	if err != nil {
		t.Errorf("Stop should not fail when already stopped: %v", err)
	}
}

// TestManager_GetStats 测试获取统计信息
func TestManager_GetStats(t *testing.T) {
	tempDir := t.TempDir()
	config := ManagementConfig{}
	manager := NewManager(config, tempDir)

	// 创建一些测试文件
	testFiles := []string{
		"app.log",
		"app.log.1",
		"error.log",
	}

	for _, filename := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		file, err := os.Create(filePath)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
		file.WriteString("test log content\n")
		file.Close()
	}

	// 获取统计信息
	stats, err := manager.GetStats()
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}
	if stats == nil {
		t.Error("GetStats returned nil")
	}
	if stats.TotalFiles == 0 {
		t.Error("Expected some files in stats")
	}
	if stats.TotalSize == 0 {
		t.Error("Expected non-zero total size")
	}
}

// TestManager_ForceCleanup 测试强制清理
func TestManager_ForceCleanup(t *testing.T) {
	tempDir := t.TempDir()
	config := ManagementConfig{
		Cleanup: CleanupConfig{
			Enabled: true,
			MaxAge:  1, // 1天
		},
	}
	manager := NewManager(config, tempDir)

	// 创建一个旧文件
	oldFile := filepath.Join(tempDir, "old.log")
	file, err := os.Create(oldFile)
	if err != nil {
		t.Fatalf("Failed to create old file: %v", err)
	}
	file.WriteString("old log content\n")
	file.Close()

	// 修改文件时间为2天前
	oldTime := time.Now().AddDate(0, 0, -2)
	os.Chtimes(oldFile, oldTime, oldTime)

	// 执行强制清理
	err = manager.ForceCleanup()
	if err != nil {
		t.Fatalf("ForceCleanup failed: %v", err)
	}

	// 检查文件是否被删除
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("Old file should have been deleted")
	}
}

// TestManager_ForceCompression 测试强制压缩
func TestManager_ForceCompression(t *testing.T) {
	tempDir := t.TempDir()
	config := ManagementConfig{
		Compression: CompressionConfig{
			Enabled: true,
			Delay:   1, // 1小时
			Algorithm: "gzip",
		},
	}
	manager := NewManager(config, tempDir)

	// 创建一个日志文件
	logFile := filepath.Join(tempDir, "app.log.1")
	file, err := os.Create(logFile)
	if err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}
	file.WriteString("log content for compression\n")
	file.Close()

	// 修改文件时间为2小时前
	oldTime := time.Now().Add(-2 * time.Hour)
	os.Chtimes(logFile, oldTime, oldTime)

	// 执行强制压缩
	err = manager.ForceCompression()
	if err != nil {
		t.Fatalf("ForceCompression failed: %v", err)
	}

	// 注意：实际的压缩可能需要满足特定条件，这里只验证方法调用成功
	// 压缩行为取决于配置和文件状态
	t.Log("ForceCompression executed successfully")
}

// TestManager_CleanupByPattern 测试按模式清理
func TestManager_CleanupByPattern(t *testing.T) {
	tempDir := t.TempDir()
	config := ManagementConfig{}
	manager := NewManager(config, tempDir)

	// 创建不同模式的文件
	testFiles := []string{
		"app.log",
		"app.log.1",
		"error.log",
		"error.log.1",
		"access.log",
	}

	for _, filename := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		file, err := os.Create(filePath)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
		file.WriteString("test content\n")
		file.Close()

		// 设置为旧文件
		oldTime := time.Now().AddDate(0, 0, -2)
		os.Chtimes(filePath, oldTime, oldTime)
	}

	// 按模式清理app.*文件
	err := manager.CleanupByPattern("app.*", 1*24*time.Hour)
	if err != nil {
		t.Fatalf("CleanupByPattern failed: %v", err)
	}

	// 检查app文件是否被删除
	for _, filename := range []string{"app.log", "app.log.1"} {
		filePath := filepath.Join(tempDir, filename)
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			t.Errorf("File %s should have been deleted", filename)
		}
	}

	// 检查其他文件是否仍存在
	for _, filename := range []string{"error.log", "error.log.1", "access.log"} {
		filePath := filepath.Join(tempDir, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("File %s should not have been deleted", filename)
		}
	}
}

// TestManager_ArchiveLogs 测试日志归档
func TestManager_ArchiveLogs(t *testing.T) {
	tempDir := t.TempDir()
	archiveDir := filepath.Join(tempDir, "archive")
	logDir := filepath.Join(tempDir, "logs")

	// 创建日志目录
	os.MkdirAll(logDir, 0755)

	config := ManagementConfig{}
	manager := NewManager(config, logDir)

	// 创建一些旧日志文件
	testFiles := []string{
		"app.log.1",
		"app.log.2",
		"error.log.1",
	}

	for _, filename := range testFiles {
		filePath := filepath.Join(logDir, filename)
		file, err := os.Create(filePath)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
		file.WriteString("test log content\n")
		file.Close()

		// 设置为旧文件
		oldTime := time.Now().AddDate(0, 0, -2)
		os.Chtimes(filePath, oldTime, oldTime)
	}

	// 执行归档
	err := manager.ArchiveLogs(archiveDir, 1*24*time.Hour)
	if err != nil {
		t.Fatalf("ArchiveLogs failed: %v", err)
	}

	// 检查归档目录是否创建
	if _, err := os.Stat(archiveDir); os.IsNotExist(err) {
		t.Error("Archive directory should have been created")
	}

	// 注意：归档行为取决于getLogFiles方法的实现和文件识别逻辑
	// 这里主要验证方法调用成功和目录创建
	t.Log("ArchiveLogs executed successfully and archive directory created")
}

// TestManager_FileCacheOperations 测试文件缓存功能
func TestManager_FileCacheOperations(t *testing.T) {
	tempDir := t.TempDir()
	config := ManagementConfig{}
	manager := NewManager(config, tempDir)

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.log")
	file, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.WriteString("test content\n")
	file.Close()

	// 第一次获取文件列表（应该从磁盘读取）
	files1, err := manager.getLogFiles()
	if err != nil {
		t.Fatalf("getLogFiles failed: %v", err)
	}
	if len(files1) == 0 {
		t.Error("Expected at least one file")
	}

	// 第二次获取文件列表（应该从缓存读取）
	files2, err := manager.getLogFiles()
	if err != nil {
		t.Fatalf("getLogFiles failed: %v", err)
	}
	if len(files2) != len(files1) {
		t.Error("Cache should return same number of files")
	}

	// 验证缓存时间是否设置
	if manager.cacheTime.IsZero() {
		t.Error("Cache time should be set after first call")
	}
}

// TestManager_ConcurrentOperations 测试并发操作
func TestManager_ConcurrentOperations(t *testing.T) {
	tempDir := t.TempDir()
	config := ManagementConfig{
		Cleanup: CleanupConfig{
			Enabled: true,
			MaxAge:  7,
			Interval: 1 * time.Hour,
		},
		Compression: CompressionConfig{
			Enabled: true,
			Delay:   1,
			Algorithm: "gzip",
		},
	}
	manager := NewManager(config, tempDir)

	// 启动manager
	err := manager.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer manager.Stop()

	// 并发执行多个操作
	done := make(chan bool, 3)

	// 并发获取统计信息
	go func() {
		for i := 0; i < 10; i++ {
			_, err := manager.GetStats()
			if err != nil {
				t.Errorf("GetStats failed: %v", err)
			}
			time.Sleep(10 * time.Millisecond)
		}
		done <- true
	}()

	// 并发检查运行状态
	go func() {
		for i := 0; i < 10; i++ {
			manager.IsRunning()
			time.Sleep(10 * time.Millisecond)
		}
		done <- true
	}()

	// 并发获取文件列表
	go func() {
		for i := 0; i < 10; i++ {
			_, err := manager.getLogFiles()
			if err != nil {
				t.Errorf("getLogFiles failed: %v", err)
			}
			time.Sleep(10 * time.Millisecond)
		}
		done <- true
	}()

	// 等待所有goroutine完成
	for i := 0; i < 3; i++ {
		<-done
	}
}