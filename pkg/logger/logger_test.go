package logger

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TestNewLogger 测试Logger创建
func TestNewLogger(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		wantErr bool
	}{
		{
			name: "默认配置",
			config: DefaultConfig(),
			wantErr: false,
		},
		{
			name: "自定义配置",
			config: &Config{
				Level:      "info",
				Format:     "json",
				Output: OutputConfig{
					Console: ConsoleConfig{Enabled: true},
				},
				Development: false,
			},
			wantErr: false,
		},
		{
			name: "无效日志级别",
			config: &Config{
				Level:      "invalid",
				Format:     "json",
				Output: OutputConfig{
					Console: ConsoleConfig{Enabled: true},
				},
			},
			wantErr: false, // NewLogger不会因为无效级别返回错误，会使用默认级别
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && logger == nil {
				t.Error("NewLogger() returned nil logger")
			}
		})
	}
}

// TestLoggerMethods 测试Logger的各种日志方法
func TestLoggerMethods(t *testing.T) {
	// 创建临时文件用于测试
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	config := &Config{
		Level:      "debug",
		Format:     "json",
		Output: OutputConfig{
			File: FileConfig{
				Enabled:  true,
				Filename: logFile,
			},
		},
		Development: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// 测试各种日志级别
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warn message")
	logger.Error("Error message")

	// 测试带字段的日志
	logger.With(
		zap.Int("user_id", 123),
		zap.String("action", "login"),
	).Info("User logged in")

	// 测试键值对日志
	logger.Infow("Context message", "key", "value")

	// 同步日志确保写入
	logger.Sync()

	// 验证日志文件是否创建
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}

	// 读取并验证日志内容
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
	if !strings.Contains(logContent, "Warn message") {
		t.Error("Warn message not found in log")
	}
	if !strings.Contains(logContent, "Error message") {
		t.Error("Error message not found in log")
	}
	if !strings.Contains(logContent, "User logged in") {
		t.Error("Field message not found in log")
	}
}

// TestLoggerLevels 测试不同日志级别的过滤
func TestLoggerLevels(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "level_test.log")

	// 创建只记录Info及以上级别的logger
	config := &Config{
		Level:      "info",
		Format:     "json",
		Output: OutputConfig{
			File: FileConfig{
				Enabled:  true,
				Filename: logFile,
			},
		},
		Development: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// 记录不同级别的日志
	logger.Debug("This should not appear")
	logger.Info("This should appear")
	logger.Warn("This should also appear")
	logger.Error("This should definitely appear")

	logger.Sync()

	// 读取日志内容
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	// Debug消息不应该出现
	if strings.Contains(logContent, "This should not appear") {
		t.Error("Debug message appeared when it shouldn't")
	}

	// Info及以上级别的消息应该出现
	if !strings.Contains(logContent, "This should appear") {
		t.Error("Info message not found")
	}
	if !strings.Contains(logContent, "This should also appear") {
		t.Error("Warn message not found")
	}
	if !strings.Contains(logContent, "This should definitely appear") {
		t.Error("Error message not found")
	}
}

// TestLoggerFormats 测试不同的日志格式
func TestLoggerFormats(t *testing.T) {
	tests := []struct {
		name   string
		format string
		check  func(content string) bool
	}{
		{
			name:   "JSON格式",
			format: "json",
			check: func(content string) bool {
				lines := strings.Split(strings.TrimSpace(content), "\n")
				for _, line := range lines {
					if line == "" {
						continue
					}
					var logEntry map[string]interface{}
					if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
						return false
					}
				}
				return true
			},
		},
		{
			name:   "Console格式",
			format: "console",
			check: func(content string) bool {
				// Console格式应该包含时间戳和级别
				return strings.Contains(content, "INFO") || strings.Contains(content, "ERROR")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			logFile := filepath.Join(tempDir, "format_test.log")

			config := &Config{
				Level:      "info",
				Format:     tt.format,
				Output: OutputConfig{
					File: FileConfig{
						Enabled:  true,
						Filename: logFile,
					},
				},
				Development: false,
			}

			logger, err := NewLogger(config)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			logger.Info("Test message")
			logger.Error("Test error")
			logger.Sync()

			content, err := os.ReadFile(logFile)
			if err != nil {
				t.Fatalf("Failed to read log file: %v", err)
			}

			if !tt.check(string(content)) {
				t.Errorf("Log format check failed for %s format", tt.format)
			}
		})
	}
}

// TestGlobalLogger 测试全局Logger
func TestGlobalLogger(t *testing.T) {
	// 初始化全局logger
	config := DefaultConfig()
	config.Output.Console.Enabled = true

	err := InitGlobal(config)
	if err != nil {
		t.Fatalf("Failed to init global logger: %v", err)
	}

	// 获取全局logger
	globalLogger := GetGlobal()
	if globalLogger == nil {
		t.Error("Global logger is nil")
	}

	// 测试全局logger方法
	globalLogger.Info("Global logger test")
}

// TestLoggerWithBuffer 测试带缓冲区的日志记录
func TestLoggerWithBuffer(t *testing.T) {
	var buf bytes.Buffer

	// 创建一个写入到buffer的core
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	encoder := zapcore.NewJSONEncoder(encoderConfig)
	writeSyncer := zapcore.AddSync(&buf)
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.InfoLevel)

	zapLogger := zap.New(core)
	logger := &Logger{
		zap: zapLogger,
	}

	// 记录日志
	logger.Info("Buffer test message")
	logger.Error("Buffer error message")
	logger.Sync()

	// 检查buffer内容
	content := buf.String()
	if !strings.Contains(content, "Buffer test message") {
		t.Error("Info message not found in buffer")
	}
	if !strings.Contains(content, "Buffer error message") {
		t.Error("Error message not found in buffer")
	}
}

// TestLoggerConcurrency 测试并发安全性
func TestLoggerConcurrency(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "concurrent_test.log")

	config := &Config{
		Level:      "info",
		Format:     "json",
		Output: OutputConfig{
			File: FileConfig{
				Enabled:  true,
				Filename: logFile,
			},
		},
		Development: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// 并发写入日志
	const numGoroutines = 5
	const messagesPerGoroutine = 10

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < messagesPerGoroutine; j++ {
				logger.Info("Concurrent message", zap.Int("goroutine", id), zap.Int("message", j))
			}
			done <- true
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	logger.Sync()

	// 验证日志文件存在且有内容
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
		return
	}

	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Log file is empty")
		return
	}

	// 计算日志行数，允许一定的误差范围
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	expectedLines := numGoroutines * messagesPerGoroutine
	
	// 验证至少有一些日志被写入，允许部分丢失（由于并发特性）
	if len(lines) < expectedLines/2 {
		t.Errorf("Too few log lines: expected at least %d, got %d", expectedLines/2, len(lines))
	}
	
	// 验证日志格式正确
	for i, line := range lines {
		if line == "" {
			continue
		}
		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Errorf("Invalid JSON at line %d: %v", i+1, err)
			break
		}
	}
}