package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client Redis缓存客户端
type Client struct {
	rdb           *redis.Client
	defaultExpiry time.Duration
}

// Config Redis配置
type Config struct {
	Addr         string        `json:"addr" mapstructure:"addr"`                     // Redis地址
	Password     string        `json:"password" mapstructure:"password"`             // 密码
	DB           int           `json:"db" mapstructure:"db"`                         // 数据库编号
	PoolSize     int           `json:"pool_size" mapstructure:"pool_size"`           // 连接池大小
	MinIdleConns int           `json:"min_idle_conns" mapstructure:"min_idle_conns"` // 最小空闲连接数
	DialTimeout  time.Duration `json:"dial_timeout" mapstructure:"dial_timeout"`     // 连接超时
	ReadTimeout  time.Duration `json:"read_timeout" mapstructure:"read_timeout"`     // 读取超时
	WriteTimeout time.Duration `json:"write_timeout" mapstructure:"write_timeout"`   // 写入超时
	DefaultTTL   time.Duration `json:"default_ttl" mapstructure:"default_ttl"`       // 默认过期时间
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Addr:         "localhost:6379",
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		DefaultTTL:   24 * time.Hour, // 默认24小时过期
	}
}

// NewClient 创建新的Redis客户端
func NewClient(config *Config) *Client {
	if config == nil {
		config = DefaultConfig()
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	return &Client{
		rdb:           rdb,
		defaultExpiry: config.DefaultTTL,
	}
}

// Ping 测试连接
func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// Close 关闭连接
func (c *Client) Close() error {
	return c.rdb.Close()
}

// GetClient 获取原始Redis客户端
func (c *Client) GetClient() *redis.Client {
	return c.rdb
}

// SetDefaultExpiry 设置默认过期时间
func (c *Client) SetDefaultExpiry(expiry time.Duration) {
	c.defaultExpiry = expiry
}

// GetDefaultExpiry 获取默认过期时间
func (c *Client) GetDefaultExpiry() time.Duration {
	return c.defaultExpiry
}
