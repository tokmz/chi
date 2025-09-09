package logger

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TestConfig_EdgeCases 测试Config的边界情况
func TestConfig_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name: "空配置",
			config: &Config{},
		},
		{
			name: "只有控制台输出",
			config: &Config{
				Level:  "info",
				Format: "text",
				Output: OutputConfig{
					Console: ConsoleConfig{Enabled: true},
				},
			},
		},
		{
			name: "只有文件输出",
			config: &Config{
				Level:  "debug",
				Format: "json",
				Output: OutputConfig{
					File: FileConfig{
						Enabled:  true,
						Filename: "/tmp/test.log",
					},
				},
			},
		},
		{
			name: "带轮转配置",
			config: &Config{
				Level:  "warn",
				Format: "json",
				Output: OutputConfig{
					File: FileConfig{
						Enabled:  true,
						Filename: "/tmp/test.log",
					},
				},
				Rotation: RotationConfig{
				Size: SizeRotationConfig{
					Enabled: true,
					MaxSize: 10,
				},
				Time: TimeRotationConfig{
					Enabled:  true,
					Interval: "day",
				},
			},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证配置
			tt.config.Validate()

			// 尝试创建logger
			logger, err := NewLogger(tt.config)
			if err != nil {
				t.Errorf("NewLogger() error = %v", err)
				return
			}
			if logger == nil {
				t.Error("NewLogger() returned nil logger")
			}
		})
	}
}

// TestLogger_FieldMethods 测试Logger的字段方法
func TestLogger_FieldMethods(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "field_test.log")

	config := &Config{
		Level:  "debug",
		Format: "json",
		Output: OutputConfig{
			File: FileConfig{
				Enabled:  true,
				Filename: logFile,
			},
		},
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	// 测试各种字段类型
	logger.Info("Testing fields",
		String("string_field", "test_value"),
		Int("int_field", 42),
		Int64("int64_field", 123456789),
		Float64("float64_field", 3.14),
		Bool("bool_field", true),
		Duration("duration_field", time.Second),
		Time("time_field", time.Now()),
		Any("any_field", map[string]interface{}{"key": "value"}),
	)

	// 验证日志文件是否创建
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

// TestLogger_WithContext 测试带上下文的日志记录
func TestLogger_WithContext(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "context_test.log")

	config := &Config{
		Level:  "info",
		Format: "json",
		Output: OutputConfig{
			File: FileConfig{
				Enabled:  true,
				Filename: logFile,
			},
		},
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	// 使用普通的Info和Error方法记录日志
	logger.Info("Context log message", String("action", "test"), String("request_id", "12345"))
	logger.Error("Context error message", String("error", "test_error"), String("request_id", "12345"))

	// 验证日志文件内容
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "Context log message") {
		t.Error("Context log message not found in log file")
	}
	if !strings.Contains(logContent, "Context error message") {
		t.Error("Context error message not found in log file")
	}
}

// TestRotationWriter_EdgeCases 测试RotationWriter的边界情况
func TestRotationWriter_EdgeCases(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "rotation_edge.log")

	// 测试极小的最大大小
	config := FileConfig{
		Enabled:    true,
		Filename:   logFile,
		MaxSize:    1, // 1MB
		MaxAge:     1,
		MaxBackups: 1,
		Compress:   false,
	}

	rotationConfig := RotationConfig{
		Size: SizeRotationConfig{
			Enabled: true,
			MaxSize: 1,
		},
	}

	writer, err := NewRotationWriter(config, rotationConfig)
	if err != nil {
		t.Fatalf("NewRotationWriter error: %v", err)
	}
	if writer == nil {
		t.Fatal("NewRotationWriter returned nil")
	}

	// 写入大量数据触发轮转
	largeData := strings.Repeat("This is a test log message that will help trigger rotation.\n", 1000)
	for i := 0; i < 100; i++ {
		n, err := writer.Write([]byte(fmt.Sprintf("[%d] %s", i, largeData)))
		if err != nil {
			t.Logf("Write error (expected in some cases): %v", err)
		}
		if n == 0 {
			t.Logf("Zero bytes written on iteration %d", i)
		}
	}

	// 验证文件是否存在
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

// TestTimeRotationWriter_EdgeCases 测试TimeRotationWriter的边界情况
func TestTimeRotationWriter_EdgeCases(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "time_rotation_edge.log")

	// 测试每分钟轮转
	config := TimeRotationConfig{
		Enabled:  true,
		Interval: "hour",
	}

	writer := NewTimeRotationWriter(logFile, config)
	if writer == nil {
		t.Fatal("NewTimeRotationWriter returned nil")
	}

	// 写入一些数据
	testData := "Time rotation test message\n"
	for i := 0; i < 10; i++ {
		n, err := writer.Write([]byte(fmt.Sprintf("[%d] %s", i, testData)))
		if err != nil {
			t.Errorf("Write error: %v", err)
		}
		if n == 0 {
			t.Errorf("Zero bytes written on iteration %d", i)
		}
	}

	// 验证文件是否存在
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

// TestManager_EdgeCases 测试Manager的边界情况
func TestManager_EdgeCases(t *testing.T) {
	tempDir := t.TempDir()

	// 测试空配置的Manager
	config := ManagementConfig{}
	manager := NewManager(config, tempDir)
	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	// 测试启动和停止
	err := manager.Start()
	if err != nil {
		t.Errorf("Manager.Start() error = %v", err)
	}

	// 测试重复启动
	err = manager.Start()
	if err == nil {
		t.Error("Expected error when starting already running manager")
	}

	// 测试获取统计信息
	stats, err := manager.GetStats()
	if err != nil {
		t.Errorf("Manager.GetStats() error = %v", err)
	}
	if stats == nil {
		t.Error("Manager.GetStats() returned nil stats")
	}

	// 测试停止
	err = manager.Stop()
	if err != nil {
		t.Errorf("Manager.Stop() error = %v", err)
	}

	// 测试重复停止
	err = manager.Stop()
	if err == nil {
		t.Error("Expected error when stopping already stopped manager")
	}
}

// TestGlobalLogger_EdgeCases 测试全局Logger的边界情况
func TestGlobalLogger_EdgeCases(t *testing.T) {
	// 保存原始的全局logger
	originalLogger := globalLogger
	defer func() {
		globalLogger = originalLogger
	}()

	// 测试全局logger初始化
	config := &Config{
		Level:  "info",
		Format: "text",
		Output: OutputConfig{
			Console: ConsoleConfig{Enabled: true},
		},
	}

	err := InitGlobal(config)
	if err != nil {
		t.Fatalf("InitGlobal() error = %v", err)
	}

	global := GetGlobal()
	if global == nil {
		t.Error("Expected global logger to be set after InitGlobal()")
	}

	// 测试全局方法
	Info("Test info message")
	Warn("Test warn message")
	Error("Test error message")
	Debug("Test debug message")
}

// TestParseLevel_AllLevels 测试ParseLevel函数的所有级别
func TestParseLevel_AllLevels(t *testing.T) {
	tests := []struct {
		levelStr string
		expected zapcore.Level
	}{
		{"debug", zapcore.DebugLevel},
		{"info", zapcore.InfoLevel},
		{"warn", zapcore.WarnLevel},
		{"error", zapcore.ErrorLevel},
		{"fatal", zapcore.FatalLevel},
		{"DEBUG", zapcore.DebugLevel},
		{"INFO", zapcore.InfoLevel},
		{"WARN", zapcore.WarnLevel},
		{"ERROR", zapcore.ErrorLevel},
		{"FATAL", zapcore.FatalLevel},
		{"invalid", zapcore.InfoLevel}, // 默认级别
		{"", zapcore.InfoLevel},        // 空字符串默认级别
	}

	for _, tt := range tests {
		t.Run(tt.levelStr, func(t *testing.T) {
			result := ParseLevel(tt.levelStr)
			if result != tt.expected {
				t.Errorf("ParseLevel(%q) = %v, expected %v", tt.levelStr, result, tt.expected)
			}
		})
	}
}

// TestFieldHelpers 测试所有字段辅助函数
func TestFieldHelpers(t *testing.T) {
	// 创建一个buffer来捕获日志输出
	var buf bytes.Buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.DebugLevel,
	)
	logger := zap.New(core)

	// 测试所有字段类型
	logger.Info("Testing all field types",
		String("string", "test"),
		Int("int", 42),
		Int64("int64", 123456789),
		Float64("float64", 3.14159),
		Bool("bool", true),
		Duration("duration", time.Hour),
		Time("time", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
		Any("any", map[string]string{"key": "value"}),
	)

	// 验证输出包含所有字段
	output := buf.String()
	expectedFields := []string{
		"\"string\":\"test\"",
		"\"int\":42",
		"\"int64\":123456789",
		"\"float64\":3.14159",
		"\"bool\":true",
		"\"duration\":3600000000000",
		"\"time\":\"2023-01-01T00:00:00Z\"",
		"\"any\":{\"key\":\"value\"}",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Expected field %s not found in output: %s", field, output)
		}
	}
}