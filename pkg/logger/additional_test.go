package logger

import (
	"os"
	"path/filepath"
	"testing"
)

// TestConfig_DefaultValues 测试配置的默认值
func TestConfig_DefaultValues(t *testing.T) {
	config := DefaultConfig()
	if config == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	// 验证默认值
	if config.Level != "info" {
		t.Errorf("Expected default level 'info', got '%s'", config.Level)
	}
	if config.Format != "json" {
		t.Errorf("Expected default format 'json', got '%s'", config.Format)
	}
	if !config.Output.Console.Enabled {
		t.Error("Expected console output to be enabled by default")
	}
}

// TestLogger_SetLevel 测试动态设置日志级别
func TestLogger_SetLevel(t *testing.T) {
	config := &Config{
		Level:  "info",
		Format: "json",
		Output: OutputConfig{
			Console: ConsoleConfig{Enabled: true},
		},
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	// 测试获取当前级别
	currentLevel := logger.GetLevel()
	if currentLevel != "info" {
		t.Errorf("Expected level 'info', got '%s'", currentLevel)
	}

	// 测试设置新级别
	logger.SetLevel("debug")
	newLevel := logger.GetLevel()
	if newLevel != "debug" {
		t.Errorf("Expected level 'debug', got '%s'", newLevel)
	}

	// 测试设置无效级别
	logger.SetLevel("invalid")
	invalidLevel := logger.GetLevel()
	if invalidLevel != "info" { // 应该回退到默认级别
		t.Errorf("Expected level 'info' after invalid level, got '%s'", invalidLevel)
	}
}

// TestLogger_WithAndNamed 测试Logger的With和Named方法
func TestLogger_WithAndNamed(t *testing.T) {
	config := &Config{
		Level:  "info",
		Format: "json",
		Output: OutputConfig{
			Console: ConsoleConfig{Enabled: true},
		},
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	// 测试With方法
	loggerWithFields := logger.With(String("component", "test"), Int("version", 1))
	if loggerWithFields == nil {
		t.Error("With() returned nil logger")
	}

	// 测试Named方法
	namedLogger := logger.Named("test-logger")
	if namedLogger == nil {
		t.Error("Named() returned nil logger")
	}

	// 测试链式调用
	chainedLogger := logger.With(String("key", "value")).Named("chained")
	if chainedLogger == nil {
		t.Error("Chained With().Named() returned nil logger")
	}

	// 使用这些logger记录日志
	loggerWithFields.Info("Test message with fields")
	namedLogger.Info("Test message from named logger")
	chainedLogger.Info("Test message from chained logger")
}

// TestLogger_SyncAndClose 测试Logger的Sync和Close方法
func TestLogger_SyncAndClose(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "sync_test.log")

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

	// 写入一些日志
	logger.Info("Test message before sync")

	// 测试Sync方法
	err = logger.Sync()
	if err != nil {
		t.Errorf("Sync() error = %v", err)
	}

	// 测试Close方法
	err = logger.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// 验证文件是否存在
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

// TestLogger_GetZapAndSugar 测试获取底层zap logger
func TestLogger_GetZapAndSugar(t *testing.T) {
	config := &Config{
		Level:  "info",
		Format: "json",
		Output: OutputConfig{
			Console: ConsoleConfig{Enabled: true},
		},
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	// 测试GetZap方法
	zapLogger := logger.GetZap()
	if zapLogger == nil {
		t.Error("GetZap() returned nil")
	}

	// 测试GetSugar方法
	sugarLogger := logger.GetSugar()
	if sugarLogger == nil {
		t.Error("GetSugar() returned nil")
	}

	// 使用底层logger记录日志
	zapLogger.Info("Test message from zap logger")
	sugarLogger.Info("Test message from sugar logger")
}

// TestTimeRotationWriter_Rotation 测试时间轮转功能
func TestTimeRotationWriter_Rotation(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "time_rotation_test.log")

	config := TimeRotationConfig{
		Enabled:  true,
		Interval: "hour",
	}

	writer := NewTimeRotationWriter(logFile, config)
	if writer == nil {
		t.Fatal("NewTimeRotationWriter returned nil")
	}

	// 写入一些数据
	testData := "Time rotation test data\n"
	for i := 0; i < 5; i++ {
		n, err := writer.Write([]byte(testData))
		if err != nil {
			t.Errorf("Write error: %v", err)
		}
		if n != len(testData) {
			t.Errorf("Expected to write %d bytes, wrote %d", len(testData), n)
		}
	}

	// 测试Close方法
	err := writer.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// 验证文件是否存在
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

// TestGlobalSync 测试全局Sync函数
func TestGlobalSync(t *testing.T) {
	// 初始化全局logger
	config := &Config{
		Level:  "info",
		Format: "json",
		Output: OutputConfig{
			Console: ConsoleConfig{Enabled: true},
		},
	}

	err := InitGlobal(config)
	if err != nil {
		t.Fatalf("InitGlobal() error = %v", err)
	}

	// 写入一些全局日志
	Info("Global info message")
	Warn("Global warn message")
	Error("Global error message")

	// 测试全局Sync
	err = Sync()
	if err != nil {
		t.Errorf("Global Sync() error = %v", err)
	}
}

// TestConfig_Validation 测试配置验证的更多场景
func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name: "空输出配置",
			config: &Config{
				Level:  "info",
				Format: "json",
				// 没有输出配置
			},
		},
		{
			name: "开发模式配置",
			config: &Config{
				Level:       "debug",
				Format:      "console",
				Development: true,
				Output: OutputConfig{
					Console: ConsoleConfig{Enabled: true, Colorful: true},
				},
			},
		},
		{
			name: "采样配置",
			config: &Config{
				Level:  "info",
				Format: "json",
				Output: OutputConfig{
					Console: ConsoleConfig{Enabled: true},
				},
				Sampling: SamplingConfig{
					Enabled:    true,
					Initial:    100,
					Thereafter: 100,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证配置
			err := tt.config.Validate()
			if err != nil {
				t.Errorf("Config.Validate() error = %v", err)
			}

			// 尝试创建logger
			logger, err := NewLogger(tt.config)
			if err != nil {
				t.Errorf("NewLogger() error = %v", err)
				return
			}
			if logger == nil {
				t.Error("NewLogger() returned nil logger")
			}

			// 测试日志记录
			logger.Info("Test message", String("test", tt.name))
		})
	}
}