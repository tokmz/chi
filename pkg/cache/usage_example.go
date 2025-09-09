package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ExampleUsageWithInit 展示如何使用初始化后的cache客户端
func ExampleUsageWithInit(client *Client) {
	ctx := context.Background()

	// 字符串操作示例
	fmt.Println("=== 字符串操作示例 ===")
	
	// 设置字符串值
	err := client.Set(ctx, "user:1001", "张三", 10*time.Minute)
	if err != nil {
		fmt.Printf("设置字符串失败: %v\n", err)
		return
	}
	fmt.Println("设置用户信息成功")

	// 获取字符串值
	value, err := client.Get(ctx, "user:1001")
	if err != nil {
		fmt.Printf("获取字符串失败: %v\n", err)
		return
	}
	fmt.Printf("获取用户信息: %s\n", value)

	// 哈希操作示例
	fmt.Println("\n=== 哈希操作示例 ===")
	
	// 设置哈希字段
	err = client.HSet(ctx, "user:1002", "name", "李四")
	if err != nil {
		fmt.Printf("设置哈希失败: %v\n", err)
		return
	}
	err = client.HSet(ctx, "user:1002", "age", 25)
	if err != nil {
		fmt.Printf("设置哈希失败: %v\n", err)
		return
	}
	err = client.HSet(ctx, "user:1002", "email", "lisi@example.com")
	if err != nil {
		fmt.Printf("设置哈希失败: %v\n", err)
		return
	}
	fmt.Println("设置用户哈希信息成功")

	// 获取哈希字段
	name, err := client.HGet(ctx, "user:1002", "name")
	if err != nil {
		fmt.Printf("获取哈希字段失败: %v\n", err)
		return
	}
	fmt.Printf("获取用户姓名: %s\n", name)

	// 获取所有哈希字段
	allFields, err := client.HGetAll(ctx, "user:1002")
	if err != nil {
		fmt.Printf("获取所有哈希字段失败: %v\n", err)
		return
	}
	fmt.Printf("获取所有用户信息: %v\n", allFields)

	// 列表操作示例
	fmt.Println("\n=== 列表操作示例 ===")
	
	// 向列表推送元素
	_, err = client.LPush(ctx, "tasks", "任务1", "任务2", "任务3")
	if err != nil {
		fmt.Printf("推送列表元素失败: %v\n", err)
		return
	}
	fmt.Println("推送任务列表成功")

	// 获取列表长度
	length, err := client.LLen(ctx, "tasks")
	if err != nil {
		fmt.Printf("获取列表长度失败: %v\n", err)
		return
	}
	fmt.Printf("任务列表长度: %d\n", length)

	// 获取列表范围
	tasks, err := client.LRange(ctx, "tasks", 0, -1)
	if err != nil {
		fmt.Printf("获取列表范围失败: %v\n", err)
		return
	}
	fmt.Printf("所有任务: %v\n", tasks)

	// 集合操作示例
	fmt.Println("\n=== 集合操作示例 ===")
	
	// 向集合添加成员
	_, err = client.SAdd(ctx, "tags", "golang", "redis", "cache", "database")
	if err != nil {
		fmt.Printf("添加集合成员失败: %v\n", err)
		return
	}
	fmt.Println("添加标签集合成功")

	// 获取集合成员数量
	count, err := client.SCard(ctx, "tags")
	if err != nil {
		fmt.Printf("获取集合成员数量失败: %v\n", err)
		return
	}
	fmt.Printf("标签数量: %d\n", count)

	// 获取所有集合成员
	tags, err := client.SMembers(ctx, "tags")
	if err != nil {
		fmt.Printf("获取集合成员失败: %v\n", err)
		return
	}
	fmt.Printf("所有标签: %v\n", tags)

	// 有序集合操作示例
	fmt.Println("\n=== 有序集合操作示例 ===")
	
	// 向有序集合添加成员
	_, err = client.ZAdd(ctx, "leaderboard", 
		redis.Z{Score: 100.5, Member: "用户A"},
		redis.Z{Score: 95.0, Member: "用户B"},
		redis.Z{Score: 88.5, Member: "用户C"},
		redis.Z{Score: 92.0, Member: "用户D"},
	)
	if err != nil {
		fmt.Printf("添加有序集合成员失败: %v\n", err)
		return
	}
	fmt.Println("添加排行榜成功")

	// 获取排名前3的用户（按分数降序）
	topUsers, err := client.ZRevRange(ctx, "leaderboard", 0, 2)
	if err != nil {
		fmt.Printf("获取排行榜失败: %v\n", err)
		return
	}
	fmt.Printf("排名前3的用户: %v\n", topUsers)

	// 获取用户分数
	score, err := client.ZScore(ctx, "leaderboard", "用户A")
	if err != nil {
		fmt.Printf("获取用户分数失败: %v\n", err)
		return
	}
	fmt.Printf("用户A的分数: %.1f\n", score)

	fmt.Println("\n=== 缓存操作示例完成 ===")
}

// ExampleCacheWithExpiry 展示带过期时间的缓存操作
func ExampleCacheWithExpiry(client *Client) {
	ctx := context.Background()

	fmt.Println("=== 过期时间示例 ===")

	// 设置短期缓存（5秒过期）
	err := client.Set(ctx, "temp:session", "临时会话数据", 5*time.Second)
	if err != nil {
		fmt.Printf("设置临时缓存失败: %v\n", err)
		return
	}
	fmt.Println("设置临时缓存成功（5秒后过期）")

	// 检查TTL
	ttl, err := client.TTL(ctx, "temp:session")
	if err != nil {
		fmt.Printf("获取TTL失败: %v\n", err)
		return
	}
	fmt.Printf("剩余过期时间: %v\n", ttl)

	// 使用默认过期时间
	err = client.Set(ctx, "cache:data", "使用默认过期时间的数据")
	if err != nil {
		fmt.Printf("设置默认过期缓存失败: %v\n", err)
		return
	}
	fmt.Println("设置默认过期缓存成功")

	// 获取默认过期时间
	defaultExpiry := client.GetDefaultExpiry()
	fmt.Printf("默认过期时间: %v\n", defaultExpiry)
}