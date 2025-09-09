package cache

import (
	"context"
)

// HashOperations 哈希操作接口
type HashOperations interface {
	HSet(ctx context.Context, key, field string, value interface{}) error
	HGet(ctx context.Context, key, field string) (string, error)
	HMSet(ctx context.Context, key string, values ...interface{}) error
	HMGet(ctx context.Context, key string, fields ...string) ([]interface{}, error)
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HDel(ctx context.Context, key string, fields ...string) (int64, error)
	HExists(ctx context.Context, key, field string) (bool, error)
	HLen(ctx context.Context, key string) (int64, error)
	HKeys(ctx context.Context, key string) ([]string, error)
	HVals(ctx context.Context, key string) ([]string, error)
	HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error)
	HIncrByFloat(ctx context.Context, key, field string, incr float64) (float64, error)
	HSetNX(ctx context.Context, key, field string, value interface{}) (bool, error)
}

// HSet 设置哈希字段值
func (c *Client) HSet(ctx context.Context, key, field string, value interface{}) error {
	return c.rdb.HSet(ctx, key, field, value).Err()
}

// HGet 获取哈希字段值
func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	return c.rdb.HGet(ctx, key, field).Result()
}

// HMSet 批量设置哈希字段
func (c *Client) HMSet(ctx context.Context, key string, values ...interface{}) error {
	return c.rdb.HMSet(ctx, key, values...).Err()
}

// HMGet 批量获取哈希字段值
func (c *Client) HMGet(ctx context.Context, key string, fields ...string) ([]interface{}, error) {
	return c.rdb.HMGet(ctx, key, fields...).Result()
}

// HGetAll 获取哈希的所有字段和值
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.rdb.HGetAll(ctx, key).Result()
}

// HDel 删除哈希字段
func (c *Client) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	return c.rdb.HDel(ctx, key, fields...).Result()
}

// HExists 检查哈希字段是否存在
func (c *Client) HExists(ctx context.Context, key, field string) (bool, error) {
	return c.rdb.HExists(ctx, key, field).Result()
}

// HLen 获取哈希字段数量
func (c *Client) HLen(ctx context.Context, key string) (int64, error) {
	return c.rdb.HLen(ctx, key).Result()
}

// HKeys 获取哈希的所有字段名
func (c *Client) HKeys(ctx context.Context, key string) ([]string, error) {
	return c.rdb.HKeys(ctx, key).Result()
}

// HVals 获取哈希的所有值
func (c *Client) HVals(ctx context.Context, key string) ([]string, error) {
	return c.rdb.HVals(ctx, key).Result()
}

// HIncrBy 递增哈希字段的整数值
func (c *Client) HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error) {
	return c.rdb.HIncrBy(ctx, key, field, incr).Result()
}

// HIncrByFloat 递增哈希字段的浮点数值
func (c *Client) HIncrByFloat(ctx context.Context, key, field string, incr float64) (float64, error) {
	return c.rdb.HIncrByFloat(ctx, key, field, incr).Result()
}

// HSetNX 仅当哈希字段不存在时设置值
func (c *Client) HSetNX(ctx context.Context, key, field string, value interface{}) (bool, error) {
	return c.rdb.HSetNX(ctx, key, field, value).Result()
}