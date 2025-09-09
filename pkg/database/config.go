package database

import (
	"time"

	"gorm.io/gorm/logger"
)

// Config 数据库配置
type Config struct {
	// 主库配置
	Master string `json:"master" yaml:"master" mapstructure:"master"`
	// 从库配置
	Slaves []string `json:"slaves" yaml:"slaves" mapstructure:"slaves"`
	// 连接池配置
	Pool PoolConfig `json:"pool" yaml:"pool" mapstructure:"pool"`
	// 日志配置
	Log LogConfig `json:"log" yaml:"log" mapstructure:"log"`
	// 慢查询配置
	SlowQuery SlowQueryConfig `json:"slow_query" yaml:"slow_query" mapstructure:"slow_query"`
}

// PoolConfig 连接池配置
type PoolConfig struct {
	// 最大空闲连接数
	MaxIdleConns int `json:"max_idle_conns" yaml:"max_idle_conns" mapstructure:"max_idle_conns"`
	// 最大连接数
	MaxOpenConns int `json:"max_open_conns" yaml:"max_open_conns" mapstructure:"max_open_conns"`
	// 连接最大生命周期
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" yaml:"conn_max_lifetime" mapstructure:"conn_max_lifetime"`
	// 连接最大空闲时间
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time" yaml:"conn_max_idle_time" mapstructure:"conn_max_idle_time"`
}

// LogConfig 日志配置
type LogConfig struct {
	// 日志级别: Silent, Error, Warn, Info
	Level logger.LogLevel `json:"level" yaml:"level" mapstructure:"level"`
	// 是否启用彩色输出
	Colorful bool `json:"colorful" yaml:"colorful" mapstructure:"colorful"`
	// 是否忽略记录未找到的错误
	IgnoreRecordNotFoundError bool `json:"ignore_record_not_found_error" yaml:"ignore_record_not_found_error" mapstructure:"ignore_record_not_found_error"`
	// 参数化查询
	ParameterizedQueries bool `json:"parameterized_queries" yaml:"parameterized_queries" mapstructure:"parameterized_queries"`
}

// SlowQueryConfig 慢查询配置
type SlowQueryConfig struct {
	// 是否启用慢查询监控
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	// 慢查询阈值
	Threshold time.Duration `json:"threshold" yaml:"threshold" mapstructure:"threshold"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Master: "",
		Slaves: []string{},
		Pool: PoolConfig{
			MaxIdleConns:    10,
			MaxOpenConns:    100,
			ConnMaxLifetime: time.Hour,
			ConnMaxIdleTime: time.Minute * 30,
		},
		Log: LogConfig{
			Level:                     logger.Info,
			Colorful:                  true,
			IgnoreRecordNotFoundError: false,
			ParameterizedQueries:      false,
		},
		SlowQuery: SlowQueryConfig{
			Enabled:   true,
			Threshold: time.Millisecond * 200,
		},
	}
}
