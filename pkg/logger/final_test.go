package logger

import (
	"testing"
)

// TestParseLevel_EdgeCases 测试ParseLevel的边界情况
func TestParseLevel_EdgeCases(t *testing.T) {
	tests := []string{
		"DEBUG", "Info", "WARN", "Error", "FATAL",
		"debug", "info", "warn", "error", "fatal",
		"unknown", "", "trace", "panic",
	}

	for _, level := range tests {
		t.Run(level, func(t *testing.T) {
			result := ParseLevel(level)
			// 只要不panic就算成功
			if result.String() == "" {
				t.Errorf("ParseLevel(%q) returned empty string", level)
			}
		})
	}
}

// TestConfig_ValidateEdgeCases 测试配置验证的边界情况
func TestConfig_ValidateEdgeCases(t *testing.T) {
	configs := []*Config{
		{}, // 空配置
		{Level: ""}, // 空级别
		{Format: ""}, // 空格式
		{Level: "info", Format: "unknown"}, // 未知格式
		{Level: "unknown", Format: "json"}, // 未知级别
	}

	for i, config := range configs {
		t.Run("config_"+string(rune('0'+i)), func(t *testing.T) {
			err := config.Validate()
			// 验证不应该返回错误，而是修正配置
			if err != nil {
				t.Errorf("Validate() returned error: %v", err)
			}
		})
	}
}

// TestGlobalLogger_SafetyChecks 测试全局logger的安全检查
func TestGlobalLogger_SafetyChecks(t *testing.T) {
	// 测试全局logger的安全性
	config := DefaultConfig()
	config.Output.Console.Enabled = true
	config.Output.File.Enabled = false
	InitGlobal(config)
	global := GetGlobal()
	if global == nil {
		t.Error("全局logger不应该为nil")
		return
	}

	// 测试全局便捷函数
	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")

	// 测试格式化函数
	Debugf("test %s", "format")
	Infof("test %s", "format")
	Warnf("test %s", "format")
	Errorf("test %s", "format")

	// 测试结构化日志函数
	Debugw("test", "key", "value")
	Infow("test", "key", "value")
	Warnw("test", "key", "value")
	Errorw("test", "key", "value")

	// 测试Sync
	err := Sync()
	if err != nil {
		t.Errorf("Sync() error = %v", err)
	}
}

// TestLogger_AllMethods 测试Logger的所有方法
func TestLogger_AllMethods(t *testing.T) {
	config := &Config{
		Level:  "debug",
		Format: "json",
		Output: OutputConfig{
			Console: ConsoleConfig{Enabled: true},
		},
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	// 测试所有日志方法
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	logger.Debugf("debug %s", "formatted")
	logger.Infof("info %s", "formatted")
	logger.Warnf("warn %s", "formatted")
	logger.Errorf("error %s", "formatted")

	logger.Debugw("debug with", "key", "value")
	logger.Infow("info with", "key", "value")
	logger.Warnw("warn with", "key", "value")
	logger.Errorw("error with", "key", "value")

	// 测试字段方法
	logger.Info("test fields",
		String("string", "value"),
		Int("int", 42),
		Int64("int64", 123),
		Float64("float", 3.14),
		Bool("bool", true),
		Any("any", map[string]string{"key": "value"}),
	)
}

// TestDefaultConfig_AllFields 测试默认配置的所有字段
func TestDefaultConfig_AllFields(t *testing.T) {
	config := DefaultConfig()

	// 验证所有默认值
	if config.Level == "" {
		t.Error("Default level should not be empty")
	}
	if config.Format == "" {
		t.Error("Default format should not be empty")
	}

	// 验证输出配置
	if !config.Output.Console.Enabled {
		t.Error("Console output should be enabled by default")
	}

	// 验证调用者配置
	if config.Caller.Skip < 0 {
		t.Error("Caller skip should not be negative")
	}

	// 创建logger验证配置有效性
	logger, err := NewLogger(config)
	if err != nil {
		t.Errorf("Default config should create valid logger: %v", err)
	}
	if logger == nil {
		t.Error("Logger should not be nil with default config")
	}
}