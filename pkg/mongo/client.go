package mongo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Client MongoDB客户端
type Client struct {
	client           *mongo.Client
	database         *mongo.Database
	config           *Config
	logger           Logger
	slowQueryMonitor *SlowQueryMonitor
	mu               sync.RWMutex
	closed           bool
}

// NewClient 创建MongoDB客户端
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 创建客户端选项
	clientOptions := config.ToClientOptions()

	// 连接MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// 创建日志记录器
	loggerConfig := config.GetLoggerConfig()
	logger, err := NewMongoLoggerFromConfig(loggerConfig)
	if err != nil {
		// 如果创建新日志记录器失败，回退到默认日志记录器
		logger = NewLogger(config.Log)
	}

	// 创建慢查询监控器
	slowQueryMonitor := NewSlowQueryMonitor(logger, loggerConfig.Mongo.SlowQuery.Threshold)
	if !loggerConfig.Mongo.SlowQuery.Enabled {
		slowQueryMonitor.Disable()
	}

	// 创建客户端实例
	c := &Client{
		client:           client,
		database:         client.Database(config.Database),
		config:           config,
		logger:           logger,
		slowQueryMonitor: slowQueryMonitor,
		closed:           false,
	}// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := c.Ping(ctx); err != nil {
		c.Close()
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	c.logger.Info("MongoDB client connected successfully")
	return c, nil
}

// NewClientWithURI 使用URI创建MongoDB客户端
func NewClientWithURI(uri, database string) (*Client, error) {
	config := DefaultConfig()
	config.URI = uri
	config.Database = database
	return NewClient(config)
}

// Database 获取数据库实例
func (c *Client) Database(name ...string) *mongo.Database {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(name) > 0 && name[0] != "" {
		return c.client.Database(name[0])
	}
	return c.database
}

// Collection 获取集合实例
func (c *Client) Collection(name string, database ...string) *mongo.Collection {
	db := c.Database(database...)
	return db.Collection(name)
}

// Client 获取原生MongoDB客户端
func (c *Client) Client() *mongo.Client {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.client
}

// Ping 测试连接
func (c *Client) Ping(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrConnectionClosed
	}

	return c.client.Ping(ctx, readpref.Primary())
}

// Close 关闭连接
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := c.client.Disconnect(ctx); err != nil {
		c.logger.Error("Failed to disconnect MongoDB client", "error", err)
		return err
	}

	c.logger.Info("MongoDB client disconnected successfully")
	return nil
}

// IsClosed 检查连接是否已关闭
func (c *Client) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

// Stats 获取连接统计信息
func (c *Client) Stats(ctx context.Context) (map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrConnectionClosed
	}

	// 获取服务器状态
	result := c.database.RunCommand(ctx, map[string]interface{}{"serverStatus": 1})
	if result.Err() != nil {
		return nil, result.Err()
	}

	var stats map[string]interface{}
	if err := result.Decode(&stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// HealthCheck 健康检查
func (c *Client) HealthCheck(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrConnectionClosed
	}

	// 执行ping
	if err := c.client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	// 检查数据库连接
	result := c.database.RunCommand(ctx, map[string]interface{}{"ping": 1})
	if result.Err() != nil {
		return fmt.Errorf("database ping failed: %w", result.Err())
	}

	return nil
}

// GetConfig 获取配置
func (c *Client) GetConfig() *Config {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config
}

// SetLogger 设置日志记录器
func (c *Client) SetLogger(logger Logger) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger = logger
}

// GetLogger 获取日志记录器
func (c *Client) GetLogger() Logger {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.logger
}

// UpdateLoggerConfig 更新日志配置
func (c *Client) UpdateLoggerConfig(loggerConfig *MongoLoggerConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return ErrConnectionClosed
	}

	// 创建新的日志记录器
	newLogger, err := NewMongoLoggerFromConfig(loggerConfig)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	// 更新配置中的日志配置
	c.config.Logger = loggerConfig

	// 更新日志记录器
	c.logger = newLogger

	c.logger.Info("Logger configuration updated successfully")
	return nil
}

// GetLoggerConfig 获取当前日志配置
func (c *Client) GetLoggerConfig() *MongoLoggerConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config.GetLoggerConfig()
}

// GetSlowQueryMonitor 获取慢查询监控器
func (c *Client) GetSlowQueryMonitor() *SlowQueryMonitor {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.slowQueryMonitor
}

// GetSlowQueryStats 获取慢查询统计信息
func (c *Client) GetSlowQueryStats() SlowQueryStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.slowQueryMonitor == nil {
		return SlowQueryStats{}
	}
	return c.slowQueryMonitor.GetStats()
}

// ResetSlowQueryStats 重置慢查询统计信息
func (c *Client) ResetSlowQueryStats() {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.slowQueryMonitor != nil {
		c.slowQueryMonitor.ResetStats()
	}
}

// SetSlowQueryThreshold 设置慢查询阈值
func (c *Client) SetSlowQueryThreshold(threshold time.Duration) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.slowQueryMonitor != nil {
		c.slowQueryMonitor.SetThreshold(threshold)
	}
}

// WithContext 创建带上下文的操作
func (c *Client) WithContext(ctx context.Context) *ContextClient {
	return &ContextClient{
		client: c,
		ctx:    ctx,
	}
}

// ContextClient 带上下文的客户端
type ContextClient struct {
	client *Client
	ctx    context.Context
}

// Collection 获取集合实例
func (cc *ContextClient) Collection(name string, database ...string) *mongo.Collection {
	return cc.client.Collection(name, database...)
}

// Database 获取数据库实例
func (cc *ContextClient) Database(name ...string) *mongo.Database {
	return cc.client.Database(name...)
}

// Context 获取上下文
func (cc *ContextClient) Context() context.Context {
	return cc.ctx
}

// ListDatabases 列出所有数据库
func (c *Client) ListDatabases(ctx context.Context, filter interface{}) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrConnectionClosed
	}

	result, err := c.client.ListDatabaseNames(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}

	return result, nil
}

// ListCollections 列出指定数据库的所有集合
func (c *Client) ListCollections(ctx context.Context, database string, filter interface{}) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrConnectionClosed
	}

	db := c.client.Database(database)
	result, err := db.ListCollectionNames(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	return result, nil
}

// CreateCollection 创建集合
func (c *Client) CreateCollection(ctx context.Context, database, collection string, opts ...*options.CreateCollectionOptions) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrConnectionClosed
	}

	db := c.client.Database(database)
	err := db.CreateCollection(ctx, collection, opts...)
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	c.logger.Info("Collection created successfully", "database", database, "collection", collection)
	return nil
}

// DropCollection 删除集合
func (c *Client) DropCollection(ctx context.Context, database, collection string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrConnectionClosed
	}

	db := c.client.Database(database)
	coll := db.Collection(collection)
	err := coll.Drop(ctx)
	if err != nil {
		return fmt.Errorf("failed to drop collection: %w", err)
	}

	c.logger.Info("Collection dropped successfully", "database", database, "collection", collection)
	return nil
}