package cache

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// CounterOperations 计数器操作接口
type CounterOperations interface {
	// 基础计数器操作
	Increment(ctx context.Context, key string, expiration ...time.Duration) (int64, error)
	Decrement(ctx context.Context, key string, expiration ...time.Duration) (int64, error)
	IncrementBy(ctx context.Context, key string, value int64, expiration ...time.Duration) (int64, error)
	DecrementBy(ctx context.Context, key string, value int64, expiration ...time.Duration) (int64, error)
	IncrementFloat(ctx context.Context, key string, value float64, expiration ...time.Duration) (float64, error)
	
	// 计数器获取和设置
	GetCounter(ctx context.Context, key string) (int64, error)
	SetCounter(ctx context.Context, key string, value int64, expiration ...time.Duration) error
	ResetCounter(ctx context.Context, key string) error
	
	// 批量计数器操作
	IncrementMultiple(ctx context.Context, keys []string, expiration ...time.Duration) (map[string]int64, error)
	GetMultipleCounters(ctx context.Context, keys []string) (map[string]int64, error)
}

// Increment 递增计数器（默认+1）
func (c *Client) Increment(ctx context.Context, key string, expiration ...time.Duration) (int64, error) {
	return c.IncrementBy(ctx, key, 1, expiration...)
}

// Decrement 递减计数器（默认-1）
func (c *Client) Decrement(ctx context.Context, key string, expiration ...time.Duration) (int64, error) {
	return c.DecrementBy(ctx, key, 1, expiration...)
}

// IncrementBy 按指定值递增计数器
func (c *Client) IncrementBy(ctx context.Context, key string, value int64, expiration ...time.Duration) (int64, error) {
	// 使用管道操作确保原子性
	pipe := c.rdb.Pipeline()
	incrCmd := pipe.IncrBy(ctx, key, value)
	
	// 如果指定了过期时间，设置过期
	if len(expiration) > 0 && expiration[0] > 0 {
		pipe.Expire(ctx, key, expiration[0])
	} else if len(expiration) == 0 && c.defaultExpiry > 0 {
		// 使用默认过期时间
		pipe.Expire(ctx, key, c.defaultExpiry)
	}
	
	// 执行管道
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	
	return incrCmd.Result()
}

// DecrementBy 按指定值递减计数器
func (c *Client) DecrementBy(ctx context.Context, key string, value int64, expiration ...time.Duration) (int64, error) {
	// 使用管道操作确保原子性
	pipe := c.rdb.Pipeline()
	decrCmd := pipe.DecrBy(ctx, key, value)
	
	// 如果指定了过期时间，设置过期
	if len(expiration) > 0 && expiration[0] > 0 {
		pipe.Expire(ctx, key, expiration[0])
	} else if len(expiration) == 0 && c.defaultExpiry > 0 {
		// 使用默认过期时间
		pipe.Expire(ctx, key, c.defaultExpiry)
	}
	
	// 执行管道
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	
	return decrCmd.Result()
}

// IncrementFloat 按指定浮点值递增计数器
func (c *Client) IncrementFloat(ctx context.Context, key string, value float64, expiration ...time.Duration) (float64, error) {
	// 使用管道操作确保原子性
	pipe := c.rdb.Pipeline()
	incrCmd := pipe.IncrByFloat(ctx, key, value)
	
	// 如果指定了过期时间，设置过期
	if len(expiration) > 0 && expiration[0] > 0 {
		pipe.Expire(ctx, key, expiration[0])
	} else if len(expiration) == 0 && c.defaultExpiry > 0 {
		// 使用默认过期时间
		pipe.Expire(ctx, key, c.defaultExpiry)
	}
	
	// 执行管道
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	
	return incrCmd.Result()
}

// GetCounter 获取计数器当前值
func (c *Client) GetCounter(ctx context.Context, key string) (int64, error) {
	result, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil // 键不存在时返回0
		}
		return 0, err
	}
	
	return strconv.ParseInt(result, 10, 64)
}

// SetCounter 设置计数器值
func (c *Client) SetCounter(ctx context.Context, key string, value int64, expiration ...time.Duration) error {
	exp := c.defaultExpiry
	if len(expiration) > 0 {
		exp = expiration[0]
	}
	
	return c.rdb.Set(ctx, key, value, exp).Err()
}

// ResetCounter 重置计数器（删除键）
func (c *Client) ResetCounter(ctx context.Context, key string) error {
	return c.rdb.Del(ctx, key).Err()
}

// IncrementMultiple 批量递增多个计数器
func (c *Client) IncrementMultiple(ctx context.Context, keys []string, expiration ...time.Duration) (map[string]int64, error) {
	if len(keys) == 0 {
		return make(map[string]int64), nil
	}
	
	// 使用管道批量操作
	pipe := c.rdb.Pipeline()
	cmds := make(map[string]*redis.IntCmd)
	
	for _, key := range keys {
		cmds[key] = pipe.IncrBy(ctx, key, 1)
		
		// 设置过期时间
		if len(expiration) > 0 && expiration[0] > 0 {
			pipe.Expire(ctx, key, expiration[0])
		} else if len(expiration) == 0 && c.defaultExpiry > 0 {
			pipe.Expire(ctx, key, c.defaultExpiry)
		}
	}
	
	// 执行管道
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}
	
	// 收集结果
	results := make(map[string]int64)
	for key, cmd := range cmds {
		val, err := cmd.Result()
		if err != nil {
			return nil, err
		}
		results[key] = val
	}
	
	return results, nil
}

// GetMultipleCounters 批量获取多个计数器的值
func (c *Client) GetMultipleCounters(ctx context.Context, keys []string) (map[string]int64, error) {
	if len(keys) == 0 {
		return make(map[string]int64), nil
	}
	
	// 使用MGET批量获取
	results, err := c.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	
	// 解析结果
	counters := make(map[string]int64)
	for i, key := range keys {
		if results[i] == nil {
			counters[key] = 0 // 键不存在时返回0
			continue
		}
		
		strVal, ok := results[i].(string)
		if !ok {
			counters[key] = 0
			continue
		}
		
		val, err := strconv.ParseInt(strVal, 10, 64)
		if err != nil {
			counters[key] = 0
			continue
		}
		
		counters[key] = val
	}
	
	return counters, nil
}