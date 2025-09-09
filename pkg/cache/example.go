package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// ExampleUsage 展示cache包的使用方法
func ExampleUsage() {
	// 1. 创建配置
	config := &Config{
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

	// 2. 创建客户端
	client := NewClient(config)
	defer client.Close()

	ctx := context.Background()

	// 3. 测试连接
	if err := client.Ping(ctx); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	fmt.Println("Connected to Redis successfully!")

	// 4. 字符串操作示例
	stringExample(ctx, client)

	// 5. 哈希表操作示例
	hashExample(ctx, client)

	// 6. 列表操作示例
	listExample(ctx, client)

	// 7. 集合操作示例
	setExample(ctx, client)

	// 8. 有序集合操作示例
	zsetExample(ctx, client)

	// 9. 计数器操作示例
	counterExample(ctx, client)

	// 10. Lua脚本示例
	scriptExample(ctx, client)
}

// stringExample 字符串操作示例
func stringExample(ctx context.Context, client *Client) {
	fmt.Println("\n=== 字符串操作示例 ===")

	// 设置字符串
	err := client.Set(ctx, "user:1:name", "张三", 1*time.Hour)
	if err != nil {
		log.Printf("Set error: %v", err)
		return
	}

	// 获取字符串
	name, err := client.Get(ctx, "user:1:name")
	if err != nil {
		log.Printf("Get error: %v", err)
		return
	}
	fmt.Printf("用户名: %s\n", name)

	// 设置复杂对象
	user := map[string]interface{}{
		"id":   1,
		"name": "张三",
		"age":  25,
	}
	err = client.Set(ctx, "user:1:info", user)
	if err != nil {
		log.Printf("Set object error: %v", err)
		return
	}

	// 获取复杂对象
	var userInfo map[string]interface{}
	err = client.GetObject(ctx, "user:1:info", &userInfo)
	if err != nil {
		log.Printf("Get object error: %v", err)
		return
	}
	fmt.Printf("用户信息: %+v\n", userInfo)

	// 计数器操作
	count, err := client.Incr(ctx, "page:views")
	if err != nil {
		log.Printf("Incr error: %v", err)
		return
	}
	fmt.Printf("页面访问次数: %d\n", count)
}

// hashExample 哈希表操作示例
func hashExample(ctx context.Context, client *Client) {
	fmt.Println("\n=== 哈希表操作示例 ===")

	// 设置哈希字段
	err := client.HSet(ctx, "user:2", "name", "李四")
	if err != nil {
		log.Printf("HSet error: %v", err)
		return
	}
	err = client.HSet(ctx, "user:2", "age", 30)
	if err != nil {
		log.Printf("HSet error: %v", err)
		return
	}
	err = client.HSet(ctx, "user:2", "city", "北京")
	if err != nil {
		log.Printf("HSet error: %v", err)
		return
	}

	// 获取哈希字段
	name, err := client.HGet(ctx, "user:2", "name")
	if err != nil {
		log.Printf("HGet error: %v", err)
		return
	}
	fmt.Printf("用户名: %s\n", name)

	// 获取所有哈希字段
	userData, err := client.HGetAll(ctx, "user:2")
	if err != nil {
		log.Printf("HGetAll error: %v", err)
		return
	}
	fmt.Printf("用户数据: %+v\n", userData)

	// 递增哈希字段
	newAge, err := client.HIncrBy(ctx, "user:2", "age", 1)
	if err != nil {
		log.Printf("HIncrBy error: %v", err)
		return
	}
	fmt.Printf("新年龄: %d\n", newAge)
}

// listExample 列表操作示例
func listExample(ctx context.Context, client *Client) {
	fmt.Println("\n=== 列表操作示例 ===")

	// 从左侧推入元素
	_, err := client.LPush(ctx, "tasks", "任务1", "任务2", "任务3")
	if err != nil {
		log.Printf("LPush error: %v", err)
		return
	}

	// 获取列表长度
	length, err := client.LLen(ctx, "tasks")
	if err != nil {
		log.Printf("LLen error: %v", err)
		return
	}
	fmt.Printf("任务列表长度: %d\n", length)

	// 获取列表范围
	tasks, err := client.LRange(ctx, "tasks", 0, -1)
	if err != nil {
		log.Printf("LRange error: %v", err)
		return
	}
	fmt.Printf("任务列表: %v\n", tasks)

	// 从右侧弹出元素
	task, err := client.RPop(ctx, "tasks")
	if err != nil {
		log.Printf("RPop error: %v", err)
		return
	}
	fmt.Printf("弹出的任务: %s\n", task)
}

// setExample 集合操作示例
func setExample(ctx context.Context, client *Client) {
	fmt.Println("\n=== 集合操作示例 ===")

	// 添加集合成员
	_, err := client.SAdd(ctx, "tags", "Go", "Redis", "缓存", "数据库")
	if err != nil {
		log.Printf("SAdd error: %v", err)
		return
	}

	// 获取集合成员数量
	count, err := client.SCard(ctx, "tags")
	if err != nil {
		log.Printf("SCard error: %v", err)
		return
	}
	fmt.Printf("标签数量: %d\n", count)

	// 获取所有集合成员
	tags, err := client.SMembers(ctx, "tags")
	if err != nil {
		log.Printf("SMembers error: %v", err)
		return
	}
	fmt.Printf("所有标签: %v\n", tags)

	// 检查成员是否存在
	exists, err := client.SIsMember(ctx, "tags", "Go")
	if err != nil {
		log.Printf("SIsMember error: %v", err)
		return
	}
	fmt.Printf("Go标签存在: %t\n", exists)
}

// zsetExample 有序集合操作示例
func zsetExample(ctx context.Context, client *Client) {
	fmt.Println("\n=== 有序集合操作示例 ===")

	// 添加有序集合成员
	_, err := client.ZAdd(ctx, "leaderboard", 
		redis.Z{Score: 100, Member: "玩家1"},
		redis.Z{Score: 200, Member: "玩家2"},
		redis.Z{Score: 150, Member: "玩家3"},
	)
	if err != nil {
		log.Printf("ZAdd error: %v", err)
		return
	}

	// 获取排行榜前3名
	top3, err := client.ZRevRangeWithScores(ctx, "leaderboard", 0, 2)
	if err != nil {
		log.Printf("ZRevRangeWithScores error: %v", err)
		return
	}
	fmt.Println("排行榜前3名:")
	for i, player := range top3 {
		fmt.Printf("%d. %s - %.0f分\n", i+1, player.Member, player.Score)
	}

	// 获取玩家排名
	rank, err := client.ZRevRank(ctx, "leaderboard", "玩家1")
	if err != nil {
		log.Printf("ZRevRank error: %v", err)
		return
	}
	fmt.Printf("玩家1排名: %d\n", rank+1) // Redis排名从0开始
}

// counterExample 计数器操作示例
func counterExample(ctx context.Context, client *Client) {
	fmt.Println("\n=== 计数器操作示例 ===")

	// 递增计数器
	count, err := client.Increment(ctx, "api:calls", 1*time.Hour)
	if err != nil {
		log.Printf("Increment error: %v", err)
		return
	}
	fmt.Printf("API调用次数: %d\n", count)

	// 按指定值递增
	count, err = client.IncrementBy(ctx, "downloads", 5)
	if err != nil {
		log.Printf("IncrementBy error: %v", err)
		return
	}
	fmt.Printf("下载次数: %d\n", count)

	// 批量递增多个计数器
	counters, err := client.IncrementMultiple(ctx, []string{"page1:views", "page2:views", "page3:views"})
	if err != nil {
		log.Printf("IncrementMultiple error: %v", err)
		return
	}
	fmt.Printf("页面访问统计: %+v\n", counters)
}

// scriptExample Lua脚本示例
func scriptExample(ctx context.Context, client *Client) {
	fmt.Println("\n=== Lua脚本示例 ===")

	// 原子性递增并设置过期时间的Lua脚本
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

	// 执行脚本
	result, err := client.Eval(ctx, script, []string{"script:counter"}, 1, 3600)
	if err != nil {
		log.Printf("Eval error: %v", err)
		return
	}
	fmt.Printf("脚本执行结果: %v\n", result)

	// 加载脚本并获取SHA1
	sha1, err := client.ScriptLoad(ctx, script)
	if err != nil {
		log.Printf("ScriptLoad error: %v", err)
		return
	}
	fmt.Printf("脚本SHA1: %s\n", sha1)

	// 通过SHA1执行脚本
	result, err = client.EvalSha(ctx, sha1, []string{"script:counter"}, 2, 3600)
	if err != nil {
		log.Printf("EvalSha error: %v", err)
		return
	}
	fmt.Printf("通过SHA1执行结果: %v\n", result)
}