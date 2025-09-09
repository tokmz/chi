package mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// Config MongoDB配置
type Config struct {
	// 连接URI
	URI string `json:"uri" yaml:"uri" mapstructure:"uri"`
	// 数据库名称
	Database string `json:"database" yaml:"database" mapstructure:"database"`
	// 连接池配置
	Pool PoolConfig `json:"pool" yaml:"pool" mapstructure:"pool"`
	// 日志配置
	Log LogConfig `json:"log" yaml:"log" mapstructure:"log"`
	// 读写配置
	ReadWrite ReadWriteConfig `json:"read_write" yaml:"read_write" mapstructure:"read_write"`
	// 超时配置
	Timeout TimeoutConfig `json:"timeout" yaml:"timeout" mapstructure:"timeout"`
	// 认证配置
	Auth AuthConfig `json:"auth" yaml:"auth" mapstructure:"auth"`
	// TLS配置
	TLS TLSConfig `json:"tls" yaml:"tls" mapstructure:"tls"`
}

// PoolConfig 连接池配置
type PoolConfig struct {
	// 最大连接数
	MaxPoolSize uint64 `json:"max_pool_size" yaml:"max_pool_size" mapstructure:"max_pool_size"`
	// 最小连接数
	MinPoolSize uint64 `json:"min_pool_size" yaml:"min_pool_size" mapstructure:"min_pool_size"`
	// 连接最大空闲时间
	MaxConnIdleTime time.Duration `json:"max_conn_idle_time" yaml:"max_conn_idle_time" mapstructure:"max_conn_idle_time"`
	// 连接最大生命周期
	MaxConnLifetime time.Duration `json:"max_conn_lifetime" yaml:"max_conn_lifetime" mapstructure:"max_conn_lifetime"`
}

// LogConfig 日志配置
type LogConfig struct {
	// 是否启用日志
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	// 日志级别: debug, info, warn, error
	Level string `json:"level" yaml:"level" mapstructure:"level"`
	// 是否记录慢查询
	SlowQuery bool `json:"slow_query" yaml:"slow_query" mapstructure:"slow_query"`
	// 慢查询阈值
	SlowQueryThreshold time.Duration `json:"slow_query_threshold" yaml:"slow_query_threshold" mapstructure:"slow_query_threshold"`
}

// ReadWriteConfig 读写配置
type ReadWriteConfig struct {
	// 读偏好: primary, primaryPreferred, secondary, secondaryPreferred, nearest
	ReadPreference string `json:"read_preference" yaml:"read_preference" mapstructure:"read_preference"`
	// 写关注: majority, acknowledged, unacknowledged
	WriteConcern string `json:"write_concern" yaml:"write_concern" mapstructure:"write_concern"`
	// 读关注: local, available, majority, linearizable, snapshot
	ReadConcern string `json:"read_concern" yaml:"read_concern" mapstructure:"read_concern"`
}

// TimeoutConfig 超时配置
type TimeoutConfig struct {
	// 连接超时
	Connect time.Duration `json:"connect" yaml:"connect" mapstructure:"connect"`
	// 服务器选择超时
	ServerSelection time.Duration `json:"server_selection" yaml:"server_selection" mapstructure:"server_selection"`
	// Socket超时
	Socket time.Duration `json:"socket" yaml:"socket" mapstructure:"socket"`
	// 心跳间隔
	Heartbeat time.Duration `json:"heartbeat" yaml:"heartbeat" mapstructure:"heartbeat"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	// 用户名
	Username string `json:"username" yaml:"username" mapstructure:"username"`
	// 密码
	Password string `json:"password" yaml:"password" mapstructure:"password"`
	// 认证数据库
	AuthSource string `json:"auth_source" yaml:"auth_source" mapstructure:"auth_source"`
	// 认证机制
	AuthMechanism string `json:"auth_mechanism" yaml:"auth_mechanism" mapstructure:"auth_mechanism"`
}

// TLSConfig TLS配置
type TLSConfig struct {
	// 是否启用TLS
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	// 是否跳过证书验证
	InsecureSkipVerify bool `json:"insecure_skip_verify" yaml:"insecure_skip_verify" mapstructure:"insecure_skip_verify"`
	// CA证书文件路径
	CAFile string `json:"ca_file" yaml:"ca_file" mapstructure:"ca_file"`
	// 客户端证书文件路径
	CertFile string `json:"cert_file" yaml:"cert_file" mapstructure:"cert_file"`
	// 客户端私钥文件路径
	KeyFile string `json:"key_file" yaml:"key_file" mapstructure:"key_file"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		URI:      "mongodb://localhost:27017",
		Database: "test",
		Pool: PoolConfig{
			MaxPoolSize:     100,
			MinPoolSize:     5,
			MaxConnIdleTime: 30 * time.Minute,
			MaxConnLifetime: time.Hour,
		},
		Log: LogConfig{
			Enabled:            true,
			Level:              "info",
			SlowQuery:          true,
			SlowQueryThreshold: 100 * time.Millisecond,
		},
		ReadWrite: ReadWriteConfig{
			ReadPreference: "primary",
			WriteConcern:   "majority",
			ReadConcern:    "local",
		},
		Timeout: TimeoutConfig{
			Connect:         10 * time.Second,
			ServerSelection: 30 * time.Second,
			Socket:          30 * time.Second,
			Heartbeat:       10 * time.Second,
		},
		Auth: AuthConfig{
			AuthSource: "admin",
		},
		TLS: TLSConfig{
			Enabled: false,
		},
	}
}

// ToClientOptions 将配置转换为MongoDB客户端选项
func (c *Config) ToClientOptions() *options.ClientOptions {
	opts := options.Client().ApplyURI(c.URI)

	// 设置连接池
	opts.SetMaxPoolSize(c.Pool.MaxPoolSize)
	opts.SetMinPoolSize(c.Pool.MinPoolSize)
	opts.SetMaxConnIdleTime(c.Pool.MaxConnIdleTime)

	// 设置超时
	opts.SetConnectTimeout(c.Timeout.Connect)
	opts.SetServerSelectionTimeout(c.Timeout.ServerSelection)
	opts.SetSocketTimeout(c.Timeout.Socket)
	opts.SetHeartbeatInterval(c.Timeout.Heartbeat)

	// 设置读偏好
	if readPref := c.getReadPreference(); readPref != nil {
		opts.SetReadPreference(readPref)
	}

	// 设置写关注
	if writeConcern := c.getWriteConcern(); writeConcern != nil {
		opts.SetWriteConcern(writeConcern)
	}

	// 设置认证
	if c.Auth.Username != "" {
		credential := options.Credential{
			Username:      c.Auth.Username,
			Password:      c.Auth.Password,
			AuthSource:    c.Auth.AuthSource,
			AuthMechanism: c.Auth.AuthMechanism,
		}
		opts.SetAuth(credential)
	}

	return opts
}

// getReadPreference 获取读偏好
func (c *Config) getReadPreference() *readpref.ReadPref {
	switch c.ReadWrite.ReadPreference {
	case "primary":
		return readpref.Primary()
	case "primaryPreferred":
		return readpref.PrimaryPreferred()
	case "secondary":
		return readpref.Secondary()
	case "secondaryPreferred":
		return readpref.SecondaryPreferred()
	case "nearest":
		return readpref.Nearest()
	default:
		return readpref.Primary()
	}
}

// getWriteConcern 获取写关注
func (c *Config) getWriteConcern() *writeconcern.WriteConcern {
	switch c.ReadWrite.WriteConcern {
	case "majority":
		return writeconcern.Majority()
	case "acknowledged":
		return &writeconcern.WriteConcern{W: 1}
	case "unacknowledged":
		return &writeconcern.WriteConcern{W: 0}
	default:
		return writeconcern.Majority()
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.URI == "" {
		return ErrInvalidURI
	}
	if c.Database == "" {
		return ErrInvalidDatabase
	}
	if c.Pool.MaxPoolSize == 0 {
		c.Pool.MaxPoolSize = 100
	}
	if c.Pool.MinPoolSize == 0 {
		c.Pool.MinPoolSize = 5
	}
	if c.Pool.MaxPoolSize < c.Pool.MinPoolSize {
		return ErrInvalidPoolSize
	}
	return nil
}