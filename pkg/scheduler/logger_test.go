package scheduler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"chi/pkg/logger"
)

// TestLoggerFactory_CreateLogger 测试日志工厂创建日志器
func TestLoggerFactory_CreateLogger(t *testing.T) {
	factory := &LoggerFactory{}

	tests := []struct {
		name   string
		config *SchedulerConfig
		wantErr bool
	}{
		{
			name: "使用默认日志器",
			config: &SchedulerConfig{
				LogLevel:      LogLevelInfo,
				LogOutput:     "",
				EnableConsole: true,
				UseZapLogger:  false,
			},
			wantErr: false,
		},
		{
			name: "使用zap日志器",
			config: &SchedulerConfig{
				LogLevel:      LogLevelDebug,
				LogOutput:     "",
				EnableConsole: true,
				UseZapLogger:  true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := factory.CreateLogger(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if logger == nil {
				t.Error("CreateLogger() returned nil logger")
			}
		})
	}
}

// TestLoggerAdapter_BasicLogging 测试日志适配器基本功能
func TestLoggerAdapter_BasicLogging(t *testing.T) {
	// 创建临时文件用于测试
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	config := &logger.Config{
		Level:  "debug",
		Format: "json",
		Output: logger.OutputConfig{
			Console: logger.ConsoleConfig{
				Enabled: false, // 禁用控制台输出以便测试
			},
			File: logger.FileConfig{
				Enabled:  true,
				Filename: logFile,
			},
		},
	}

	adapter, err := NewLoggerAdapterWithConfig(config, LogLevelDebug)
	if err != nil {
		t.Fatalf("Failed to create logger adapter: %v", err)
	}

	// 测试各种日志级别
	adapter.Debug("Debug message", nil)
	adapter.Info("Info message", nil)
	adapter.Warn("Warning message", nil)
	adapter.Error("Error message", nil)

	// 等待日志写入
	time.Sleep(100 * time.Millisecond)

	// 读取日志文件内容
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "Debug message") {
		t.Error("Debug message not found in log")
	}
	if !strings.Contains(logContent, "Info message") {
		t.Error("Info message not found in log")
	}
	if !strings.Contains(logContent, "Warning message") {
		t.Error("Warning message not found in log")
	}
	if !strings.Contains(logContent, "Error message") {
		t.Error("Error message not found in log")
	}
}

// TestLoggerAdapter_WithFields 测试带字段的日志记录
func TestLoggerAdapter_WithFields(t *testing.T) {
	// 创建一个写入到文件的logger配置
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "fields_test.log")

	config := &logger.Config{
		Level:  "debug",
		Format: "json",
		Output: logger.OutputConfig{
			Console: logger.ConsoleConfig{
				Enabled: false,
			},
			File: logger.FileConfig{
				Enabled:  true,
				Filename: logFile,
			},
		},
	}

	adapter, err := NewLoggerAdapterWithConfig(config, LogLevelDebug)
	if err != nil {
		t.Fatalf("Failed to create logger adapter: %v", err)
	}

	// 测试带字段的日志
	adapter.Info("Test message", map[string]interface{}{"key1": "value1"})
	adapter.Warn("Test warning", map[string]interface{}{
		"key2": "value2",
		"key3": 123,
	})

	// 等待写入
	time.Sleep(100 * time.Millisecond)

	// 验证日志文件
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "Test message") {
		t.Error("Test message not found in log")
	}
	if !strings.Contains(logContent, "key1") {
		t.Error("Field key1 not found in log")
	}
}

// TestZapLoggerAdapter 测试ZapLoggerAdapter
func TestZapLoggerAdapter(t *testing.T) {
	// 创建临时文件
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "zap_test.log")

	config := &logger.Config{
		Level:  "info",
		Format: "json",
		Output: logger.OutputConfig{
			File: logger.FileConfig{
				Enabled:  true,
				Filename: logFile,
			},
			Console: logger.ConsoleConfig{
				Enabled: false,
			},
		},
	}

	loggerInstance, err := logger.NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	adapter := NewZapLoggerAdapter(loggerInstance, LogLevelInfo)

	// 测试日志记录
	adapter.Info("ZapLogger test message", nil)
	adapter.Error("ZapLogger error message", map[string]interface{}{"test_key": "test_value"})

	// 等待写入
	time.Sleep(100 * time.Millisecond)

	// 验证日志文件
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "ZapLogger test message") {
		t.Error("Test message not found in log")
	}
	if !strings.Contains(logContent, "ZapLogger error message") {
		t.Error("Error message not found in log")
	}
}

// TestLogLevel_Conversion 测试日志级别转换
func TestLogLevel_Conversion(t *testing.T) {
	factory := &LoggerFactory{}

	tests := []struct {
		schedulerLevel LogLevel
		expectedLevel  string
	}{
		{LogLevelDebug, "debug"},
		{LogLevelInfo, "info"},
		{LogLevelWarn, "warn"},
		{LogLevelError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.schedulerLevel.String(), func(t *testing.T) {
			result := factory.convertLogLevel(tt.schedulerLevel)
			if result != tt.expectedLevel {
				t.Errorf("convertLogLevel(%v) = %v, want %v", tt.schedulerLevel, result, tt.expectedLevel)
			}
		})
	}
}

// TestBasicIntegration 测试基本集成功能
func TestBasicIntegration(t *testing.T) {
	// 测试使用扩展配置的基本功能
	schedulerConfig := &SchedulerConfig{
		LogLevel:     LogLevelInfo,
		UseZapLogger: true,
		EnableConsole: false,
	}

	// 创建日志器
	factory := &LoggerFactory{}
	logger, err := factory.CreateLogger(schedulerConfig)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// 测试日志记录
	logger.Info("Integration test message", nil)
	logger.Debug("Debug with field", map[string]interface{}{"test": "integration"})

	// 验证日志器不为空
	if logger == nil {
		t.Error("Logger should not be nil")
	}
}

// BenchmarkLoggerAdapter 性能基准测试
func BenchmarkLoggerAdapter(b *testing.B) {
	config := &logger.Config{
		Level:  "info",
		Format: "json",
		Output: logger.OutputConfig{
			Console: logger.ConsoleConfig{
				Enabled: false,
			},
		},
	}

	adapter, err := NewLoggerAdapterWithConfig(config, LogLevelInfo)
	if err != nil {
		b.Fatalf("Failed to create logger adapter: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			adapter.Info("Benchmark test message", nil)
		}
	})
}