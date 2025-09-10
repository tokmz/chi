package mongo

import (
	"context"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// MockLogger 模拟日志记录器用于测试
type MockLogger struct {
	logs []LogEntry
}

type LogEntry struct {
	Level   string
	Message string
	Fields  map[string]interface{}
}

func NewMockLogger() Logger {
	return &MockLogger{
		logs: make([]LogEntry, 0),
	}
}

func (m *MockLogger) Debug(msg string, fields ...interface{}) {
	m.addLog("DEBUG", msg, fields...)
}

func (m *MockLogger) Info(msg string, fields ...interface{}) {
	m.addLog("INFO", msg, fields...)
}

func (m *MockLogger) Warn(msg string, fields ...interface{}) {
	m.addLog("WARN", msg, fields...)
}

func (m *MockLogger) Error(msg string, fields ...interface{}) {
	m.addLog("ERROR", msg, fields...)
}

func (m *MockLogger) Fatal(msg string, fields ...interface{}) {
	m.addLog("FATAL", msg, fields...)
}

func (m *MockLogger) addLog(level, msg string, fields ...interface{}) {
	fieldsMap := make(map[string]interface{})
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key := fields[i].(string)
			fieldsMap[key] = fields[i+1]
		}
	}
	m.logs = append(m.logs, LogEntry{
		Level:   level,
		Message: msg,
		Fields:  fieldsMap,
	})
}

func (m *MockLogger) GetLogs() []LogEntry {
	return m.logs
}

func (m *MockLogger) Clear() {
	m.logs = make([]LogEntry, 0)
}

// TestMongoLoggerConfig 测试MongoDB日志配置
func TestMongoLoggerConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultMongoLoggerConfig()
		if config == nil {
			t.Fatal("DefaultMongoLoggerConfig should not return nil")
		}
		if config.Level != LogLevelInfo {
			t.Errorf("Expected default level to be Info, got %v", config.Level)
		}
		if !config.EnableConsole {
			t.Error("Expected console to be enabled by default")
		}
	})

	t.Run("Validation", func(t *testing.T) {
		config := &MongoLoggerConfig{
			Level:         LogLevelInfo,
			Output:        "",
			EnableConsole: false,
			File: FileLogConfig{
				Enabled: false,
			},
		}
		err := config.Validate()
		if err == nil {
			t.Error("Expected validation error when no output is configured")
		}
	})

	t.Run("ToLoggerConfig", func(t *testing.T) {
		config := DefaultMongoLoggerConfig()
		loggerConfig := config.ToLoggerConfig()
		if loggerConfig == nil {
			t.Fatal("ToLoggerConfig should not return nil")
		}
	})
}

// TestDefaultLogger 测试默认日志记录器
func TestDefaultLogger(t *testing.T) {
	t.Run("CreateLogger", func(t *testing.T) {
		logger := NewLogger(LogConfig{
			Enabled: true,
			Level:   "info",
		})
		if logger == nil {
			t.Fatal("NewLogger should not return nil")
		}
	})

	t.Run("LogLevels", func(t *testing.T) {
		mockLogger := NewMockLogger().(*MockLogger)

		// 测试不同级别的日志
		mockLogger.Debug("debug message", "key", "value")
		mockLogger.Info("info message", "key", "value")
		mockLogger.Warn("warn message", "key", "value")
		mockLogger.Error("error message", "key", "value")

		logs := mockLogger.GetLogs()
		if len(logs) != 4 {
			t.Errorf("Expected 4 logs, got %d", len(logs))
		}
	})

	t.Run("FieldHandling", func(t *testing.T) {
		mockLogger := NewMockLogger().(*MockLogger)

		mockLogger.Info("test message", "string_key", "string_value", "int_key", 123)

		logs := mockLogger.GetLogs()
		if len(logs) != 1 {
			t.Fatalf("Expected 1 log, got %d", len(logs))
		}

		log := logs[0]
		if log.Fields["string_key"] != "string_value" {
			t.Errorf("Expected string_key to be 'string_value', got %v", log.Fields["string_key"])
		}
		if log.Fields["int_key"] != 123 {
			t.Errorf("Expected int_key to be 123, got %v", log.Fields["int_key"])
		}
	})
}

// TestSlowQueryMonitor 测试慢查询监控器
func TestSlowQueryMonitor(t *testing.T) {
	t.Run("CreateMonitor", func(t *testing.T) {
		mockLogger := NewMockLogger().(*MockLogger)
		monitor := NewSlowQueryMonitor(mockLogger, 100*time.Millisecond)
		if monitor == nil {
			t.Fatal("NewSlowQueryMonitor should not return nil")
		}
		if !monitor.IsEnabled() {
			t.Error("Monitor should be enabled by default")
		}
	})

	t.Run("SlowQueryDetection", func(t *testing.T) {
		mockLogger := NewMockLogger().(*MockLogger)
		monitor := NewSlowQueryMonitor(mockLogger, 50*time.Millisecond)

		// 模拟快查询
		fastQuery := &QueryInfo{
			Operation:  "find",
			Collection: "users",
			Database:   "test",
			Duration:   30 * time.Millisecond,
			Timestamp:  time.Now(),
		}
		monitor.LogSlowQuery(fastQuery)

		// 模拟慢查询
		slowQuery := &QueryInfo{
			Operation:  "find",
			Collection: "users",
			Database:   "test",
			Duration:   100 * time.Millisecond,
			Timestamp:  time.Now(),
			Filter:     bson.M{"name": "test"},
		}
		monitor.LogSlowQuery(slowQuery)

		logs := mockLogger.GetLogs()
		// 只有慢查询应该被记录为WARN级别
		slowQueryLogs := 0
		for _, log := range logs {
			if log.Level == "WARN" && strings.Contains(log.Message, "Slow query detected") {
				slowQueryLogs++
			}
		}
		if slowQueryLogs != 1 {
			t.Errorf("Expected 1 slow query log, got %d", slowQueryLogs)
		}
	})

	t.Run("Statistics", func(t *testing.T) {
		mockLogger := NewMockLogger().(*MockLogger)
		monitor := NewSlowQueryMonitor(mockLogger, 50*time.Millisecond)

		// 添加一些查询
		queries := []*QueryInfo{
			{Duration: 30 * time.Millisecond, Timestamp: time.Now()},
			{Duration: 60 * time.Millisecond, Timestamp: time.Now()}, // 慢查询
			{Duration: 40 * time.Millisecond, Timestamp: time.Now()},
			{Duration: 80 * time.Millisecond, Timestamp: time.Now()}, // 慢查询
		}

		for _, query := range queries {
			monitor.LogSlowQuery(query)
		}

		stats := monitor.GetStats()
		if stats.TotalQueries != 4 {
			t.Errorf("Expected 4 total queries, got %d", stats.TotalQueries)
		}
		if stats.SlowQueries != 2 {
			t.Errorf("Expected 2 slow queries, got %d", stats.SlowQueries)
		}
		if stats.SlowQueryRate != 50.0 {
			t.Errorf("Expected 50%% slow query rate, got %.2f%%", stats.SlowQueryRate)
		}
		if stats.MaxTime != 80*time.Millisecond {
			t.Errorf("Expected max time to be 80ms, got %v", stats.MaxTime)
		}
	})

	t.Run("SensitiveDataSanitization", func(t *testing.T) {
		mockLogger := NewMockLogger().(*MockLogger)
		monitor := NewSlowQueryMonitor(mockLogger, 10*time.Millisecond)

		// 包含敏感数据的查询
		sensitiveQuery := &QueryInfo{
			Operation:  "find",
			Collection: "users",
			Database:   "test",
			Duration:   50 * time.Millisecond,
			Timestamp:  time.Now(),
			Filter: bson.M{
				"username": "testuser",
				"password": "secret123",
				"token":    "abc123",
			},
		}
		monitor.LogSlowQuery(sensitiveQuery)

		logs := mockLogger.GetLogs()
		if len(logs) == 0 {
			t.Fatal("Expected at least one log entry")
		}

		// 检查敏感数据是否被清理
		for _, log := range logs {
			if log.Level == "WARN" {
				if filterStr, ok := log.Fields["filter"].(string); ok {
					if strings.Contains(filterStr, "secret123") || strings.Contains(filterStr, "abc123") {
						t.Error("Sensitive data should be redacted from logs")
					}
					if !strings.Contains(filterStr, "[REDACTED]") {
						t.Error("Expected sensitive fields to be marked as [REDACTED]")
					}
				}
			}
		}
	})

	t.Run("MonitorQuery", func(t *testing.T) {
		mockLogger := NewMockLogger().(*MockLogger)
		monitor := NewSlowQueryMonitor(mockLogger, 50*time.Millisecond)

		ctx := context.Background()
		err := monitor.MonitorQuery(ctx, "find", "users", "test", func() error {
			time.Sleep(60 * time.Millisecond) // 模拟慢查询
			return nil
		})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		logs := mockLogger.GetLogs()
		slowQueryFound := false
		for _, log := range logs {
			if log.Level == "WARN" && strings.Contains(log.Message, "Slow query detected") {
				slowQueryFound = true
				break
			}
		}
		if !slowQueryFound {
			t.Error("Expected slow query to be detected and logged")
		}
	})
}

// TestConfigIntegration 测试配置集成
func TestConfigIntegration(t *testing.T) {
	t.Run("GetLoggerConfig", func(t *testing.T) {
		config := DefaultConfig()
		
		// 测试使用旧配置
		loggerConfig := config.GetLoggerConfig()
		if loggerConfig == nil {
			t.Fatal("GetLoggerConfig should not return nil")
		}

		// 测试使用新配置
		config.Logger = DefaultMongoLoggerConfig()
		config.Logger.Level = LogLevelDebug
		newLoggerConfig := config.GetLoggerConfig()
		if newLoggerConfig.Level != LogLevelDebug {
			t.Errorf("Expected debug level, got %v", newLoggerConfig.Level)
		}
	})

	t.Run("NewMongoLoggerFromConfig", func(t *testing.T) {
		config := DefaultMongoLoggerConfig()
		config.UseZapLogger = false // 使用默认logger

		logger, err := NewMongoLoggerFromConfig(config)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if logger == nil {
			t.Fatal("Expected logger to be created")
		}
	})
}

// TestLogLevel 测试日志级别
func TestLogLevel(t *testing.T) {
	t.Run("ParseLogLevel", func(t *testing.T) {
		tests := []struct {
			input    string
			expected LogLevel
		}{
			{"debug", LogLevelDebug},
			{"DEBUG", LogLevelDebug},
			{"info", LogLevelInfo},
			{"INFO", LogLevelInfo},
			{"warn", LogLevelWarn},
			{"WARN", LogLevelWarn},
			{"warning", LogLevelWarn},
			{"error", LogLevelError},
			{"ERROR", LogLevelError},
			{"fatal", LogLevelFatal},
			{"FATAL", LogLevelFatal},
			{"invalid", LogLevelInfo}, // 默认值
		}

		for _, test := range tests {
			result := ParseLogLevel(test.input)
			if result != test.expected {
				t.Errorf("ParseLogLevel(%s) = %v, expected %v", test.input, result, test.expected)
			}
		}
	})

	t.Run("LogLevelString", func(t *testing.T) {
		tests := []struct {
			level    LogLevel
			expected string
		}{
			{LogLevelDebug, "DEBUG"},
			{LogLevelInfo, "INFO"},
			{LogLevelWarn, "WARN"},
			{LogLevelError, "ERROR"},
			{LogLevelFatal, "FATAL"},
		}

		for _, test := range tests {
			result := test.level.String()
			if result != test.expected {
				t.Errorf("LogLevel(%d).String() = %s, expected %s", test.level, result, test.expected)
			}
		}
	})
}

// BenchmarkSlowQueryMonitor 性能测试
func BenchmarkSlowQueryMonitor(b *testing.B) {
	mockLogger := NewMockLogger().(*MockLogger)
	monitor := NewSlowQueryMonitor(mockLogger, 100*time.Millisecond)

	queryInfo := &QueryInfo{
		Operation:  "find",
		Collection: "users",
		Database:   "test",
		Duration:   50 * time.Millisecond,
		Timestamp:  time.Now(),
		Filter:     bson.M{"name": "test"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.LogSlowQuery(queryInfo)
	}
}

// BenchmarkMockLogger Mock日志记录器性能测试
func BenchmarkMockLogger(b *testing.B) {
	mockLogger := NewMockLogger().(*MockLogger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockLogger.Info("test message", "key1", "value1", "key2", "value2")
	}
}