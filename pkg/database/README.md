# Database Package

一个基于 GORM 的企业级 Go 数据库访问包，提供读写分离、连接池管理、慢查询监控等功能。

## 功能特性

- ✅ **读写分离**：自动路由读写操作到不同的数据库实例
- ✅ **连接池管理**：完整的连接池配置和监控
- ✅ **慢查询监控**：自动检测和记录慢查询
- ✅ **自定义日志**：支持彩色输出和多级别日志
- ✅ **健康检查**：主从库健康状态监控
- ✅ **事务支持**：简化的事务操作接口
- ✅ **统计信息**：详细的连接池统计数据
- ✅ **配置灵活**：支持多种配置方式和默认值

## 快速开始

### 基本使用

```go
package main

import (
    "context"
    "log"
    "your-project/pkg/database"
)

func main() {
    // 1. 创建配置
    config := &database.Config{
        Master: "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local",
        Slaves: []string{
            "user:password@tcp(slave1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local",
        },
    }

    // 2. 创建客户端
    client, err := database.NewClient(config)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // 3. 使用数据库
    ctx := context.Background()
    
    // 写操作（自动使用主库）
    user := &User{Name: "张三", Email: "zhangsan@example.com"}
    client.DB().WithContext(ctx).Create(user)
    
    // 读操作（自动使用从库）
    var users []User
    client.DB().WithContext(ctx).Find(&users)
}
```

### 使用默认配置

```go
// 使用默认配置
config := database.DefaultConfig()
config.Master = "your-dsn-here"

client, err := database.NewClient(config)
```

## 配置说明

### 完整配置示例

```go
config := &database.Config{
    // 主库配置（必需）
    Master: "user:password@tcp(master:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local",
    
    // 从库配置（可选）
    Slaves: []string{
        "user:password@tcp(slave1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local",
        "user:password@tcp(slave2:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local",
    },
    
    // 连接池配置
    Pool: database.PoolConfig{
        MaxIdleConns:    10,                    // 最大空闲连接数
        MaxOpenConns:    100,                   // 最大连接数
        ConnMaxLifetime: time.Hour,             // 连接最大生命周期
        ConnMaxIdleTime: time.Minute * 30,      // 连接最大空闲时间
    },
    
    // 日志配置
    Log: database.LogConfig{
        Level:                     logger.Info,  // 日志级别
        Colorful:                  true,         // 彩色输出
        IgnoreRecordNotFoundError: false,        // 忽略记录未找到错误
        ParameterizedQueries:      false,        // 参数化查询
    },
    
    // 慢查询配置
    SlowQuery: database.SlowQueryConfig{
        Enabled:   true,                        // 启用慢查询监控
        Threshold: time.Millisecond * 200,      // 慢查询阈值
    },
}
```

### 配置项说明

#### PoolConfig - 连接池配置

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| MaxIdleConns | int | 10 | 最大空闲连接数 |
| MaxOpenConns | int | 100 | 最大连接数 |
| ConnMaxLifetime | time.Duration | 1h | 连接最大生命周期 |
| ConnMaxIdleTime | time.Duration | 30m | 连接最大空闲时间 |

#### LogConfig - 日志配置

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| Level | logger.LogLevel | Info | 日志级别 (Silent/Error/Warn/Info) |
| Colorful | bool | true | 是否启用彩色输出 |
| IgnoreRecordNotFoundError | bool | false | 是否忽略记录未找到错误 |
| ParameterizedQueries | bool | false | 是否使用参数化查询 |

#### SlowQueryConfig - 慢查询配置

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| Enabled | bool | true | 是否启用慢查询监控 |
| Threshold | time.Duration | 200ms | 慢查询阈值 |

## API 参考

### Client 方法

#### 基础方法

```go
// 获取 GORM 数据库实例
func (c *Client) DB() *gorm.DB

// 关闭数据库连接
func (c *Client) Close() error

// 测试数据库连接
func (c *Client) Ping(ctx context.Context) error

// 健康检查
func (c *Client) HealthCheck(ctx context.Context) error
```

#### 读写分离

```go
// 强制使用主库
func (c *Client) Master() *gorm.DB

// 强制使用从库
func (c *Client) Slave() *gorm.DB

// 示例
client.Master().Create(&user)  // 写操作使用主库
client.Slave().Find(&users)    // 读操作使用从库
```

#### 事务操作

```go
// 执行事务
func (c *Client) Transaction(ctx context.Context, fn func(*gorm.DB) error) error

// 示例
err := client.Transaction(ctx, func(tx *gorm.DB) error {
    if err := tx.Create(&user1).Error; err != nil {
        return err
    }
    if err := tx.Create(&user2).Error; err != nil {
        return err
    }
    return nil
})
```

#### 监控和统计

```go
// 获取连接池统计信息
func (c *Client) Stats() (map[string]interface{}, error)

// 获取慢查询日志
func (c *Client) GetSlowQueries(ctx context.Context, limit int) ([]map[string]interface{}, error)

// 动态设置日志级别
func (c *Client) SetLogLevel(level string) error
```

## 使用示例

### 1. 基本 CRUD 操作

```go
// 定义模型
type User struct {
    ID        uint      `gorm:"primarykey"`
    Name      string    `gorm:"size:100;not null"`
    Email     string    `gorm:"size:100;uniqueIndex"`
    Age       int
    CreatedAt time.Time
    UpdatedAt time.Time
}

// 创建
user := &User{Name: "张三", Email: "zhangsan@example.com", Age: 25}
client.DB().WithContext(ctx).Create(user)

// 查询
var users []User
client.DB().WithContext(ctx).Where("age > ?", 18).Find(&users)

// 更新
client.DB().WithContext(ctx).Model(&user).Update("age", 26)

// 删除
client.DB().WithContext(ctx).Delete(&user)
```

### 2. 读写分离示例

```go
// 写操作 - 自动使用主库
client.DB().WithContext(ctx).Create(&user)
client.DB().WithContext(ctx).Model(&user).Update("name", "新名字")

// 强制使用主库
client.Master().WithContext(ctx).Create(&user)

// 读操作 - 自动使用从库
var users []User
client.DB().WithContext(ctx).Find(&users)

// 强制使用从库
client.Slave().WithContext(ctx).First(&user, 1)
```

### 3. 事务示例

```go
err := client.Transaction(ctx, func(tx *gorm.DB) error {
    // 在事务中执行多个操作
    users := []User{
        {Name: "用户1", Email: "user1@example.com"},
        {Name: "用户2", Email: "user2@example.com"},
    }
    
    for _, u := range users {
        if err := tx.Create(&u).Error; err != nil {
            return err // 自动回滚
        }
    }
    
    return nil // 提交事务
})
```

### 4. 监控示例

```go
// 获取连接池统计
stats, err := client.Stats()
if err == nil {
    fmt.Printf("连接池统计: %+v\n", stats)
}

// 健康检查
if err := client.HealthCheck(ctx); err != nil {
    log.Printf("健康检查失败: %v", err)
}

// 获取慢查询
slowQueries, err := client.GetSlowQueries(ctx, 10)
if err == nil {
    fmt.Printf("发现 %d 条慢查询\n", len(slowQueries))
}
```

## 最佳实践

### 1. 连接管理

```go
// ✅ 推荐：使用单例模式管理数据库连接
var dbClient *database.Client

func InitDB() error {
    config := database.DefaultConfig()
    config.Master = os.Getenv("DB_MASTER_DSN")
    
    var err error
    dbClient, err = database.NewClient(config)
    return err
}

func GetDB() *database.Client {
    return dbClient
}

// ✅ 应用退出时关闭连接
func CloseDB() {
    if dbClient != nil {
        dbClient.Close()
    }
}
```

### 2. 上下文使用

```go
// ✅ 推荐：始终使用 context
func GetUserByID(ctx context.Context, id uint) (*User, error) {
    var user User
    err := dbClient.DB().WithContext(ctx).First(&user, id).Error
    return &user, err
}

// ❌ 不推荐：不使用 context
func GetUserByIDWrong(id uint) (*User, error) {
    var user User
    err := dbClient.DB().First(&user, id).Error
    return &user, err
}
```

### 3. 错误处理

```go
// ✅ 推荐：明确处理不同类型的错误
func GetUser(ctx context.Context, id uint) (*User, error) {
    var user User
    err := dbClient.DB().WithContext(ctx).First(&user, id).Error
    
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, fmt.Errorf("用户不存在: %d", id)
        }
        return nil, fmt.Errorf("查询用户失败: %w", err)
    }
    
    return &user, nil
}
```

### 4. 性能优化

```go
// ✅ 推荐：使用预加载避免 N+1 查询
var users []User
dbClient.DB().WithContext(ctx).Preload("Orders").Find(&users)

// ✅ 推荐：使用批量操作
users := []User{{Name: "用户1"}, {Name: "用户2"}}
dbClient.DB().WithContext(ctx).CreateInBatches(users, 100)

// ✅ 推荐：使用索引优化查询
dbClient.DB().WithContext(ctx).Where("email = ?", email).First(&user)
```

## 故障排除

### 常见问题

#### 1. 连接失败

```bash
# 错误信息
failed to connect to master database: dial tcp: connection refused

# 解决方案
1. 检查数据库服务是否启动
2. 验证 DSN 配置是否正确
3. 检查网络连接和防火墙设置
```

#### 2. 慢查询过多

```bash
# 错误信息
[SLOW QUERY] [1500.000ms] [rows:1000] SELECT * FROM users

# 解决方案
1. 添加适当的索引
2. 优化查询条件
3. 使用分页查询
4. 调整慢查询阈值
```

#### 3. 连接池耗尽

```bash
# 错误信息
sql: database is closed

# 解决方案
1. 增加最大连接数
2. 检查连接泄漏
3. 优化连接生命周期
4. 使用连接池监控
```

### 调试技巧

```go
// 启用详细日志
config.Log.Level = logger.Info

// 监控连接池状态
stats, _ := client.Stats()
log.Printf("连接池状态: %+v", stats)

// 检查慢查询
slowQueries, _ := client.GetSlowQueries(ctx, 10)
for _, query := range slowQueries {
    log.Printf("慢查询: %+v", query)
}
```

## 许可证

本项目采用 MIT 许可证。详见 [LICENSE](../../LICENSE) 文件。

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目。

## 更新日志

### v1.0.0
- 初始版本发布
- 支持读写分离
- 连接池管理
- 慢查询监控
- 自定义日志记录器