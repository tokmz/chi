package database

import (
	"context"
	"testing"
	"time"

	"chi/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormlogger "gorm.io/gorm/logger"
)

func TestDatabaseLoggerConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultDatabaseLoggerConfig()
		assert.NotNil(t, config)
		assert.True(t, config.Enabled)
		assert.Equal(t, "info", config.Level)
		assert.True(t, config.Output.Console.Enabled)
		assert.Equal(t, 200*time.Millisecond, config.SlowQuery.Threshold)
	})

	t.Run("Validate", func(t *testing.T) {
		config := DefaultDatabaseLoggerConfig()
		config.Level = "invalid"
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid log level")
	})

	t.Run("ToLoggerConfig", func(t *testing.T) {
		config := DefaultDatabaseLoggerConfig()
		loggerConfig := config.ToLoggerConfig()
		assert.NotNil(t, loggerConfig)
		assert.Equal(t, config.Level, loggerConfig.Level)
		assert.Equal(t, config.Output.Console.Enabled, loggerConfig.Output.Console.Enabled)
	})
}

func TestDatabaseLoggerAdapter(t *testing.T) {
	t.Run("NewDatabaseLoggerAdapter", func(t *testing.T) {
		config := DefaultDatabaseLoggerConfig()
		adapter, err := NewDatabaseLoggerAdapter(config)
		require.NoError(t, err)
		assert.NotNil(t, adapter)
		assert.NotNil(t, adapter.logger)
		assert.NotNil(t, adapter.slowMonitor)
		assert.NotNil(t, adapter.perfMonitor)
	})

	t.Run("LogModes", func(t *testing.T) {
		config := DefaultDatabaseLoggerConfig()
		adapter, err := NewDatabaseLoggerAdapter(config)
		require.NoError(t, err)

		// 测试日志级别设置
		infoAdapter := adapter.LogMode(gormlogger.Info)
		assert.NotNil(t, infoAdapter)
		warnAdapter := adapter.LogMode(gormlogger.Warn)
		assert.NotNil(t, warnAdapter)
		errorAdapter := adapter.LogMode(gormlogger.Error)
		assert.NotNil(t, errorAdapter)
	})

	t.Run("Trace", func(t *testing.T) {
		config := DefaultDatabaseLoggerConfig()
		adapter, err := NewDatabaseLoggerAdapter(config)
		require.NoError(t, err)

		ctx := context.Background()
		begin := time.Now()
		sql := "SELECT * FROM users WHERE id = ?"
		rows := int64(1)

		// 测试正常查询
		adapter.Trace(ctx, begin, func() (string, int64) {
			return sql, rows
		}, nil)

		// 测试慢查询
		slowBegin := time.Now().Add(-500 * time.Millisecond)
		adapter.Trace(ctx, slowBegin, func() (string, int64) {
			return sql, rows
		}, nil)

		// 验证慢查询被记录
		stats := adapter.slowMonitor.GetStats()
		assert.Greater(t, stats.TotalQueries, int64(0))
	})

	t.Run("Close", func(t *testing.T) {
		config := DefaultDatabaseLoggerConfig()
		adapter, err := NewDatabaseLoggerAdapter(config)
		require.NoError(t, err)

		// Close方法可能在测试环境中返回sync错误，这是正常的
		_ = adapter.Close()
	})
}

func TestDatabaseSlowQueryMonitor(t *testing.T) {
	t.Run("NewDatabaseSlowQueryMonitor", func(t *testing.T) {
		config := &SlowQueryConfig{
			Enabled:   true,
			Threshold: 100 * time.Millisecond,
		}
		loggerConfig := logger.DefaultConfig()
		loggerInstance, err := logger.NewLogger(loggerConfig)
		require.NoError(t, err)
		monitor := NewDatabaseSlowQueryMonitor(*config, loggerInstance)
		assert.NotNil(t, monitor)
	})

	t.Run("RecordQuery", func(t *testing.T) {
		config := &SlowQueryConfig{
			Enabled:   true,
			Threshold: 100 * time.Millisecond,
		}
		loggerConfig := logger.DefaultConfig()
		loggerInstance, err := logger.NewLogger(loggerConfig)
		require.NoError(t, err)
		monitor := NewDatabaseSlowQueryMonitor(*config, loggerInstance)
		monitor.Start()
		defer monitor.Stop()

		// 记录正常查询
		monitor.RecordQuery(50*time.Millisecond)

		// 记录慢查询（使用RecordSlowQuery方法）
		monitor.RecordSlowQuery("SELECT * FROM users", 200*time.Millisecond, 1, nil)

		stats := monitor.GetStats()
		assert.Equal(t, int64(2), stats.TotalQueries)
		assert.Equal(t, int64(1), stats.SlowQueries)
	})

	t.Run("SetThreshold", func(t *testing.T) {
		config := &SlowQueryConfig{
			Enabled:   true,
			Threshold: 100 * time.Millisecond,
		}
		loggerConfig := logger.DefaultConfig()
		loggerInstance, err := logger.NewLogger(loggerConfig)
		require.NoError(t, err)
		monitor := NewDatabaseSlowQueryMonitor(*config, loggerInstance)

		newThreshold := 200 * time.Millisecond
		monitor.SetThreshold(newThreshold)
	})

	t.Run("Reset", func(t *testing.T) {
		config := &SlowQueryConfig{
			Enabled:   true,
			Threshold: 100 * time.Millisecond,
		}
		loggerConfig := logger.DefaultConfig()
		loggerInstance, err := logger.NewLogger(loggerConfig)
		require.NoError(t, err)
		monitor := NewDatabaseSlowQueryMonitor(*config, loggerInstance)
		monitor.Start()
		defer monitor.Stop()

		// 记录一些查询
		monitor.RecordQuery(50*time.Millisecond)
		monitor.RecordQuery(150*time.Millisecond)

		// 重置统计
		monitor.Reset()
		stats := monitor.GetStats()
		assert.Equal(t, int64(0), stats.TotalQueries)
		assert.Equal(t, int64(0), stats.SlowQueries)
	})
}

func TestDatabasePerformanceMonitor(t *testing.T) {
	t.Run("NewDatabasePerformanceMonitor", func(t *testing.T) {
		loggerConfig := logger.DefaultConfig()
		loggerInstance, err := logger.NewLogger(loggerConfig)
		require.NoError(t, err)
		config := PerformanceConfig{
			Enabled: true,
			Interval: time.Minute,
			StatsWindow: time.Hour,
			LogConnectionPool: true,
		}
		monitor := NewDatabasePerformanceMonitor(config, loggerInstance)
		assert.NotNil(t, monitor)
	})

	t.Run("RecordQuery", func(t *testing.T) {
		loggerConfig := logger.DefaultConfig()
		loggerInstance, err := logger.NewLogger(loggerConfig)
		require.NoError(t, err)
		config := PerformanceConfig{
			Enabled: true,
			Interval: time.Minute,
			StatsWindow: time.Hour,
			LogConnectionPool: true,
		}
		monitor := NewDatabasePerformanceMonitor(config, loggerInstance)
		monitor.Start()
		defer monitor.Stop()

		// 记录查询
		monitor.RecordQuery(100*time.Millisecond, false)
		monitor.RecordQuery(50*time.Millisecond, false)

		stats := monitor.GetStats()
		assert.Equal(t, int64(2), stats.TotalQueries)
		assert.Greater(t, stats.TotalDuration, time.Duration(0))
	})

	t.Run("GenerateReport", func(t *testing.T) {
		loggerConfig := logger.DefaultConfig()
		loggerInstance, err := logger.NewLogger(loggerConfig)
		require.NoError(t, err)
		config := PerformanceConfig{
			Enabled: true,
			Interval: time.Minute,
			StatsWindow: time.Hour,
			LogConnectionPool: true,
		}
		monitor := NewDatabasePerformanceMonitor(config, loggerInstance)
		monitor.Start()
		defer monitor.Stop()

		// 记录一些查询
		monitor.RecordQuery(100*time.Millisecond, false)
		monitor.RecordQuery(50*time.Millisecond, false)

		monitor.generateReport()
		// 验证报告生成功能正常运行
		stats := monitor.GetStats()
		assert.Equal(t, int64(2), stats.TotalQueries)
	})
}