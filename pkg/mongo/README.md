# MongoDB 操作模块

一个功能完整、高性能的 MongoDB 操作封装模块，为 Go 应用程序提供简洁易用的 MongoDB 数据库操作接口。

## 🚀 特性

- **🔧 完整配置管理**: 支持连接池、超时、认证、TLS等全面配置
- **⚡ 高性能连接池**: 智能连接池管理，支持连接复用和自动回收
- **📝 丰富的CRUD操作**: 封装常用的增删改查操作，支持批量操作和聚合查询
- **🔒 事务支持**: 完整的事务操作支持，包括会话管理和事务重试
- **✅ 文档验证**: 基于Schema的文档验证和类型检查
- **📊 日志记录**: 完整的操作日志和慢查询监控
- **🛡️ 错误处理**: 统一的错误处理和异常管理
- **🧪 测试友好**: 提供丰富的使用示例和测试用例

## 📦 安装

```bash
go get go.mongodb.org/mongo-driver/mongo
```

## 🏗️ 模块结构

```
mongo/
├── client.go          # MongoDB客户端核心实现
├── config.go          # 配置管理
├── crud.go            # CRUD操作封装
├── errors.go          # 错误定义
├── logger.go          # 日志记录器
├── transaction.go     # 事务操作支持
├── validator.go       # 文档验证器
├── example.go         # 使用示例
└── README.md          # 文档说明
```

## 🚀 快速开始

### 基础使用

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "chi/pkg/mongo"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
    ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Name     string             `bson:"name" json:"name"`
    Email    string             `bson:"email" json:"email"`
    Age      int                `bson:"age" json:"age"`
    CreatedAt time.Time         `bson:"created_at" json:"created_at"`
}

func main() {
    // 1. 创建客户端
    client, err := mongo.NewClientWithURI("mongodb://localhost:27017", "myapp")
    if err != nil {
        log.Fatal("Failed to create client:", err)
    }
    defer client.Close()
    
    // 2. 创建仓储
    userRepo := mongo.NewRepository(client, "myapp", "users")
    
    // 3. 插入文档
    user := &User{
        Name:      "张三",
        Email:     "zhangsan@example.com",
        Age:       25,
        CreatedAt: time.Now(),
    }
    
    ctx := context.Background()
    result, err := userRepo.InsertOne(ctx, user)
    if err != nil {
        log.Fatal("Failed to insert user:", err)
    }
    fmt.Printf("Inserted user with ID: %v\n", result.InsertedID)
    
    // 4. 查询文档
    var foundUser User
    err = userRepo.FindOne(ctx, bson.M{"email": "zhangsan@example.com"}).Decode(&foundUser)
    if err != nil {
        log.Fatal("Failed to find user:", err)
    }
    fmt.Printf("Found user: %+v\n", foundUser)
    
    // 5. 更新文档
    update := bson.M{"$set": bson.M{"age": 26}}
    updateResult, err := userRepo.UpdateOne(ctx, bson.M{"_id": foundUser.ID}, update)
    if err != nil {
        log.Fatal("Failed to update user:", err)
    }
    fmt.Printf("Updated %d document(s)\n", updateResult.ModifiedCount)
}
```

### 自定义配置

```go
config := &mongo.Config{
    URI:      "mongodb://localhost:27017",
    Database: "myapp",
    Pool: mongo.PoolConfig{
        MaxPoolSize:     100,
        MinPoolSize:     10,
        MaxConnIdleTime: 30 * time.Minute,
    },
    Log: mongo.LogConfig{
        Enabled:            true,
        Level:              "info",
        SlowQuery:          true,
        SlowQueryThreshold: 100 * time.Millisecond,
    },
    Timeout: mongo.TimeoutConfig{
        Connect:         10 * time.Second,
        ServerSelection: 30 * time.Second,
        Socket:          30 * time.Second,
    },
}

client, err := mongo.NewClient(config)
if err != nil {
    log.Fatal("Failed to create client:", err)
}
defer client.Close()
```

## 📚 核心功能

### 1. 连接管理

```go
// 创建客户端
client, err := mongo.NewClientWithURI("mongodb://localhost:27017", "database")

// 健康检查
err = client.HealthCheck(context.Background())

// 获取统计信息
stats, err := client.Stats(context.Background())

// 关闭连接
client.Close()
```

### 2. CRUD 操作

```go
repo := mongo.NewRepository(client, "database", "collection")
ctx := context.Background()

// 插入
result, err := repo.InsertOne(ctx, document)
results, err := repo.InsertMany(ctx, documents)

// 查询
singleResult := repo.FindOne(ctx, filter)
cursor, err := repo.Find(ctx, filter, options.Find().SetLimit(10))
var results []Document
err = repo.FindAll(ctx, filter, &results)

// 更新
result, err := repo.UpdateOne(ctx, filter, update)
result, err := repo.UpdateMany(ctx, filter, update)
result, err := repo.ReplaceOne(ctx, filter, replacement)

// 删除
result, err := repo.DeleteOne(ctx, filter)
result, err := repo.DeleteMany(ctx, filter)

// 统计
count, err := repo.CountDocuments(ctx, filter)

// 聚合
cursor, err := repo.Aggregate(ctx, pipeline)
var results []bson.M
err = repo.AggregateAll(ctx, pipeline, &results)
```

### 3. 事务操作

```go
tm := mongo.NewTransactionManager(client)

err = tm.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
    // 在事务中执行操作
    _, err := repo.InsertOne(sc, document1)
    if err != nil {
        return nil, err
    }
    
    _, err = repo.UpdateOne(sc, filter, update)
    if err != nil {
        return nil, err
    }
    
    return nil, nil
})
```

### 4. 文档验证

```go
// 定义Schema
schema := map[string]interface{}{
    "type": "object",
    "required": []interface{}{"name", "email"},
    "properties": map[string]interface{}{
        "name": map[string]interface{}{
            "type": "string",
            "minLength": 2,
            "maxLength": 50,
        },
        "email": map[string]interface{}{
            "type": "string",
            "minLength": 5,
        },
        "age": map[string]interface{}{
            "type": "integer",
            "minimum": 0,
            "maximum": 150,
        },
    },
}

// 创建验证器
validator := mongo.NewSchemaValidator(schema, client.GetLogger())

// 创建带验证的仓储
validatedRepo := mongo.NewValidatedRepository(repo, validator)

// 插入时自动验证
result, err := validatedRepo.InsertOne(ctx, document)
```

### 5. 日志记录

```go
// 自定义日志配置
logConfig := mongo.LogConfig{
    Enabled:            true,
    Level:              "debug",
    SlowQuery:          true,
    SlowQueryThreshold: 50 * time.Millisecond,
}

// 慢查询会自动记录
// [2024-01-15 10:30:45] [WARN] Slow query detected [operation=Find, duration=150ms, collection=users]
```

## ⚙️ 配置选项

### 连接配置

```go
type Config struct {
    URI      string        // MongoDB连接URI
    Database string        // 数据库名称
    Pool     PoolConfig    // 连接池配置
    Log      LogConfig     // 日志配置
    ReadWrite ReadWriteConfig // 读写配置
    Timeout  TimeoutConfig // 超时配置
    Auth     AuthConfig    // 认证配置
    TLS      TLSConfig     // TLS配置
}
```

### 连接池配置

```go
type PoolConfig struct {
    MaxPoolSize     uint64        // 最大连接数 (默认: 100)
    MinPoolSize     uint64        // 最小连接数 (默认: 5)
    MaxConnIdleTime time.Duration // 连接最大空闲时间 (默认: 30分钟)
}
```

### 日志配置

```go
type LogConfig struct {
    Enabled            bool          // 是否启用日志 (默认: true)
    Level              string        // 日志级别: debug, info, warn, error (默认: info)
    SlowQuery          bool          // 是否记录慢查询 (默认: true)
    SlowQueryThreshold time.Duration // 慢查询阈值 (默认: 100ms)
}
```

### 超时配置

```go
type TimeoutConfig struct {
    Connect         time.Duration // 连接超时 (默认: 10秒)
    ServerSelection time.Duration // 服务器选择超时 (默认: 30秒)
    Socket          time.Duration // Socket超时 (默认: 30秒)
    Heartbeat       time.Duration // 心跳间隔 (默认: 10秒)
}
```

## 🔧 高级功能

### 聚合查询

```go
pipeline := []bson.M{
    {"$match": bson.M{"status": "active"}},
    {"$group": bson.M{
        "_id":   "$category",
        "count": bson.M{"$sum": 1},
        "total": bson.M{"$sum": "$amount"},
    }},
    {"$sort": bson.M{"total": -1}},
}

var results []bson.M
err = repo.AggregateAll(ctx, pipeline, &results)
```

### 批量操作

```go
models := []mongo.WriteModel{
    mongo.NewInsertOneModel().SetDocument(doc1),
    mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update),
    mongo.NewDeleteOneModel().SetFilter(deleteFilter),
}

result, err := repo.BulkWrite(ctx, models)
```

### 索引管理

```go
// 创建索引
indexModel := mongo.IndexModel{
    Keys: bson.D{{"email", 1}},
    Options: options.Index().SetUnique(true),
}

collection := client.Collection("users")
_, err = collection.Indexes().CreateOne(ctx, indexModel)
```

## 🛡️ 错误处理

模块定义了完整的错误类型：

```go
// 配置相关错误
ErrInvalidURI
ErrInvalidDatabase
ErrInvalidPoolSize

// 连接相关错误
ErrConnectionFailed
ErrConnectionClosed
ErrPingFailed

// 操作相关错误
ErrDocumentNotFound
ErrInvalidObjectID
ErrInvalidFilter

// 事务相关错误
ErrTransactionFailed
ErrTransactionAborted

// 验证相关错误
ErrValidationFailed
ErrSchemaNotFound
```

## 🧪 测试

运行示例代码：

```go
package main

import "chi/pkg/mongo"

func main() {
    // 运行所有示例
    mongo.RunAllExamples()
}
```

## 📝 最佳实践

### 1. 连接管理

- 在应用启动时创建客户端，在应用关闭时关闭连接
- 使用连接池避免频繁创建连接
- 设置合适的超时时间

### 2. 错误处理

```go
result, err := repo.FindOne(ctx, filter)
if err != nil {
    if err == mongo.ErrDocumentNotFound {
        // 处理文档未找到
        return nil, nil
    }
    // 处理其他错误
    return nil, fmt.Errorf("failed to find document: %w", err)
}
```

### 3. 上下文使用

```go
// 设置超时
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

result, err := repo.FindOne(ctx, filter)
```

### 4. 事务使用

- 只在需要原子性操作时使用事务
- 保持事务尽可能短
- 处理事务重试逻辑

### 5. 性能优化

- 使用索引优化查询性能
- 启用慢查询监控
- 合理设置连接池大小
- 使用聚合管道优化复杂查询

## 🤝 贡献

欢迎提交 Issue 和 Pull Request 来改进这个模块。

## 📄 许可证

本项目采用 MIT 许可证。