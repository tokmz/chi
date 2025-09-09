# Redis Cache 包

这是一个基于 `github.com/redis/go-redis/v9` 的 Redis 缓存封装包，提供了完整的 Redis 操作功能，支持设置默认过期时间。

## 功能特性

### 🚀 核心功能
- **字符串操作**: 支持基本的字符串读写、批量操作、计数器等
- **哈希表操作**: 完整的哈希表 CRUD 操作
- **列表操作**: 支持队列、栈等列表操作
- **集合操作**: 集合的增删查改及集合运算
- **有序集合操作**: 排行榜、范围查询等有序集合功能
- **计数器操作**: 原子性计数器操作，支持批量操作
- **Lua脚本执行**: 支持自定义 Lua 脚本执行

### 📊 监控与追踪
- **错误处理**: 统一的错误处理机制
- **性能监控**: 支持记录操作类型、键名、参数等详细信息

### ⚙️ 配置管理
- **默认过期时间**: 支持设置全局默认过期时间，避免缓存无限增长
- **连接池管理**: 可配置连接池大小、超时时间等
- **灵活配置**: 支持自定义 Redis 连接参数

## 安装依赖

在项目根目录的 `go.mod` 中添加以下依赖：

```bash
go get github.com/redis/go-redis/v9
```

## 快速开始

### 1. 创建客户端

```go
package main

import (
    "context"
    "time"
    "your-project/pkg/cache"
)

func main() {
    // 使用默认配置
    client := cache.NewClient(nil)
    defer client.Close()
    
    // 或者使用自定义配置
    config := &cache.Config{
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
    client = cache.NewClient(config)
    
    ctx := context.Background()
    
    // 测试连接
    if err := client.Ping(ctx); err != nil {
        panic(err)
    }
}
```

### 2. 字符串操作

```go
// 设置字符串（使用默认过期时间）
err := client.Set(ctx, "user:1:name", "张三")

// 设置字符串（指定过期时间）
err = client.Set(ctx, "user:1:name", "张三", 1*time.Hour)

// 获取字符串
name, err := client.Get(ctx, "user:1:name")

// 设置复杂对象（自动JSON序列化）
user := map[string]interface{}{
    "id":   1,
    "name": "张三",
    "age":  25,
}
err = client.Set(ctx, "user:1:info", user)

// 获取复杂对象（自动JSON反序列化）
var userInfo map[string]interface{}
err = client.GetObject(ctx, "user:1:info", &userInfo)

// 计数器操作
count, err := client.Incr(ctx, "page:views")
count, err = client.IncrBy(ctx, "page:views", 5)
```

### 3. 哈希表操作

```go
// 设置哈希字段
_, err := client.HSet(ctx, "user:2", "name", "李四", "age", 30, "city", "北京")

// 获取哈希字段
name, err := client.HGet(ctx, "user:2", "name")

// 获取所有哈希字段
userData, err := client.HGetAll(ctx, "user:2")

// 递增哈希字段
newAge, err := client.HIncrBy(ctx, "user:2", "age", 1)

// 设置复杂对象到哈希字段
profile := map[string]interface{}{"bio": "软件工程师", "skills": []string{"Go", "Redis"}}
_, err = client.HSet(ctx, "user:2", "profile", profile)

// 获取复杂对象
var userProfile map[string]interface{}
err = client.HGetObject(ctx, "user:2", "profile", &userProfile)
```

### 4. 列表操作

```go
// 从左侧推入元素
_, err := client.LPush(ctx, "tasks", "任务1", "任务2", "任务3")

// 从右侧推入元素
_, err = client.RPush(ctx, "tasks", "任务4")

// 获取列表长度
length, err := client.LLen(ctx, "tasks")

// 获取列表范围
tasks, err := client.LRange(ctx, "tasks", 0, -1)

// 弹出元素
task, err := client.LPop(ctx, "tasks") // 从左侧弹出
task, err = client.RPop(ctx, "tasks") // 从右侧弹出

// 阻塞式弹出（用于队列）
result, err := client.BLPop(ctx, 5*time.Second, "tasks")
```

### 5. 集合操作

```go
// 添加集合成员
_, err := client.SAdd(ctx, "tags", "Go", "Redis", "缓存", "数据库")

// 获取集合成员数量
count, err := client.SCard(ctx, "tags")

// 获取所有集合成员
tags, err := client.SMembers(ctx, "tags")

// 检查成员是否存在
exists, err := client.SIsMember(ctx, "tags", "Go")

// 随机获取成员
tag, err := client.SRandMember(ctx, "tags")

// 集合运算
unionTags, err := client.SUnion(ctx, "tags1", "tags2")
interTags, err := client.SInter(ctx, "tags1", "tags2")
diffTags, err := client.SDiff(ctx, "tags1", "tags2")
```

### 6. 有序集合操作

```go
import "github.com/redis/go-redis/v9"

// 添加有序集合成员
_, err := client.ZAdd(ctx, "leaderboard", 
    redis.Z{Score: 100, Member: "玩家1"},
    redis.Z{Score: 200, Member: "玩家2"},
    redis.Z{Score: 150, Member: "玩家3"},
)

// 获取排行榜前3名（按分数降序）
top3, err := client.ZRevRangeWithScores(ctx, "leaderboard", 0, 2)

// 获取玩家排名
rank, err := client.ZRevRank(ctx, "leaderboard", "玩家1")

// 按分数范围查询
players, err := client.ZRangeByScore(ctx, "leaderboard", &redis.ZRangeBy{
    Min: "100",
    Max: "200",
})

// 增加分数
newScore, err := client.ZIncrBy(ctx, "leaderboard", 10, "玩家1")
```

### 7. 计数器操作

```go
// 基础计数器操作
count, err := client.Increment(ctx, "api:calls", 1*time.Hour)
count, err = client.IncrementBy(ctx, "downloads", 5)
count, err = client.Decrement(ctx, "inventory")

// 浮点数计数器
floatCount, err := client.IncrementFloat(ctx, "temperature", 0.5)

// 获取计数器值
count, err = client.GetCounter(ctx, "api:calls")

// 设置计数器值
err = client.SetCounter(ctx, "api:calls", 100)

// 重置计数器
err = client.ResetCounter(ctx, "api:calls")

// 批量操作
counters, err := client.IncrementMultiple(ctx, []string{"page1:views", "page2:views"})
counters, err = client.GetMultipleCounters(ctx, []string{"page1:views", "page2:views"})
```

### 8. Lua脚本执行

```go
// 执行Lua脚本
script := `
    local key = KEYS[1]
    local increment = tonumber(ARGV[1])
    local ttl = tonumber(ARGV[2])
    
    local current = redis.call('GET', key)
    if current == false then
        current = 0
    else
        current = tonumber(current)
    end
    
    local new_value = current + increment
    redis.call('SET', key, new_value)
    redis.call('EXPIRE', key, ttl)
    
    return new_value
`

result, err := client.Eval(ctx, script, []string{"counter:key"}, 1, 3600)

// 加载脚本并通过SHA1执行
sha1, err := client.ScriptLoad(ctx, script)
result, err = client.EvalSha(ctx, sha1, []string{"counter:key"}, 2, 3600)
```

### 9. 通用操作

```go
// 检查键是否存在
count, err := client.Exists(ctx, "key1", "key2")

// 删除键
count, err = client.Del(ctx, "key1", "key2")

// 设置过期时间
success, err := client.Expire(ctx, "key1", 1*time.Hour)

// 获取剩余生存时间
ttl, err := client.TTL(ctx, "key1")

// 获取键的类型
keyType, err := client.Type(ctx, "key1")

// 查找匹配模式的键
keys, err := client.Keys(ctx, "user:*")

// 迭代键（推荐用于大量键的场景）
keys, cursor, err := client.Scan(ctx, 0, "user:*", 10)
```

## 配置说明

### Config 结构体

```go
type Config struct {
    Addr         string        // Redis地址，默认: "localhost:6379"
    Password     string        // 密码，默认: ""
    DB           int           // 数据库编号，默认: 0
    PoolSize     int           // 连接池大小，默认: 10
    MinIdleConns int           // 最小空闲连接数，默认: 5
    DialTimeout  time.Duration // 连接超时，默认: 5秒
    ReadTimeout  time.Duration // 读取超时，默认: 3秒
    WriteTimeout time.Duration // 写入超时，默认: 3秒
    DefaultTTL   time.Duration // 默认过期时间，默认: 24小时
}
```

### 默认配置

```go
config := cache.DefaultConfig()
// 等同于:
config := &cache.Config{
    Addr:         "localhost:6379",
    Password:     "",
    DB:           0,
    PoolSize:     10,
    MinIdleConns: 5,
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
    DefaultTTL:   24 * time.Hour,
}
```



## 最佳实践

### 1. 连接管理

```go
// 在应用启动时创建客户端
var redisClient *cache.Client

func init() {
    config := &cache.Config{
        Addr:     "localhost:6379",
        PoolSize: 20, // 根据并发需求调整
        DefaultTTL: 2 * time.Hour, // 根据业务需求设置
    }
    redisClient = cache.NewClient(config)
}

// 在应用关闭时清理
func cleanup() {
    redisClient.Close()
}
```

### 2. 错误处理

```go
value, err := client.Get(ctx, "key")
if err != nil {
    if err == redis.Nil {
        // 键不存在
        log.Println("Key not found")
    } else {
        // 其他错误
        log.Printf("Redis error: %v", err)
    }
}
```

### 3. 过期时间管理

```go
// 为不同类型的数据设置不同的过期时间
err := client.Set(ctx, "session:token", token, 30*time.Minute)  // 会话30分钟
err = client.Set(ctx, "cache:data", data, 1*time.Hour)         // 缓存1小时
err = client.Set(ctx, "config:app", config, 24*time.Hour)      // 配置24小时
```

### 4. 批量操作

```go
// 使用批量操作提高性能
keys := []string{"counter1", "counter2", "counter3"}
counters, err := client.IncrementMultiple(ctx, keys)

// 批量获取
values, err := client.MGet(ctx, "key1", "key2", "key3")
```

## 注意事项

1. **内存管理**: 设置合理的默认过期时间，避免缓存无限增长
2. **连接池**: 根据应用并发量调整连接池大小
3. **错误处理**: 妥善处理 Redis 连接错误和键不存在的情况
4. **性能监控**: 监控 Redis 操作性能
5. **键命名**: 使用有意义的键命名规范，如 `user:1:profile`
6. **数据序列化**: 复杂对象会自动序列化为 JSON，注意性能影响

## 示例代码

完整的使用示例请参考 `example.go` 文件，其中包含了所有功能的详细使用方法。

```bash
# 运行示例（需要先启动 Redis 服务）
go run pkg/cache/example.go
```