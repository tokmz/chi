package cache

import (
	"fmt"
	"time"
)

// InitConfig 从配置结构体转换为cache包的Config
func InitConfig(addr, password string, db int, poolSize, minIdleConns int, dialTimeout, readTimeout, writeTimeout, defaultTTL time.Duration) *Config {
	return &Config{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
		DialTimeout:  dialTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		DefaultTTL:   defaultTTL,
	}
}



// ParseDuration 解析时间字符串
func ParseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}
	return time.ParseDuration(s)
}

// MustParseDuration 解析时间字符串，失败时panic
func MustParseDuration(s string) time.Duration {
	d, err := ParseDuration(s)
	if err != nil {
		panic(fmt.Sprintf("parse duration failed: %v", err))
	}
	return d
}