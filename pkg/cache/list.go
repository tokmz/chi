package cache

import (
	"context"
	"time"
)

// ListOperations 列表操作接口
type ListOperations interface {
	LPush(ctx context.Context, key string, values ...interface{}) (int64, error)
	RPush(ctx context.Context, key string, values ...interface{}) (int64, error)
	LPop(ctx context.Context, key string) (string, error)
	RPop(ctx context.Context, key string) (string, error)
	LLen(ctx context.Context, key string) (int64, error)
	LIndex(ctx context.Context, key string, index int64) (string, error)
	LRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	LSet(ctx context.Context, key string, index int64, value interface{}) error
	LRem(ctx context.Context, key string, count int64, value interface{}) (int64, error)
	LTrim(ctx context.Context, key string, start, stop int64) error
	LInsert(ctx context.Context, key, op string, pivot, value interface{}) (int64, error)
	BLPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error)
	BRPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error)
	RPopLPush(ctx context.Context, source, destination string) (string, error)
	BRPopLPush(ctx context.Context, source, destination string, timeout time.Duration) (string, error)
}

// LPush 从列表左侧推入元素
func (c *Client) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return c.rdb.LPush(ctx, key, values...).Result()
}

// RPush 从列表右侧推入元素
func (c *Client) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return c.rdb.RPush(ctx, key, values...).Result()
}

// LPop 从列表左侧弹出元素
func (c *Client) LPop(ctx context.Context, key string) (string, error) {
	return c.rdb.LPop(ctx, key).Result()
}

// RPop 从列表右侧弹出元素
func (c *Client) RPop(ctx context.Context, key string) (string, error) {
	return c.rdb.RPop(ctx, key).Result()
}

// LLen 获取列表长度
func (c *Client) LLen(ctx context.Context, key string) (int64, error) {
	return c.rdb.LLen(ctx, key).Result()
}

// LIndex 获取列表指定位置的元素
func (c *Client) LIndex(ctx context.Context, key string, index int64) (string, error) {
	return c.rdb.LIndex(ctx, key, index).Result()
}

// LRange 获取列表指定范围的元素
func (c *Client) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.rdb.LRange(ctx, key, start, stop).Result()
}

// LSet 设置列表指定位置的元素值
func (c *Client) LSet(ctx context.Context, key string, index int64, value interface{}) error {
	return c.rdb.LSet(ctx, key, index, value).Err()
}

// LRem 移除列表中的元素
func (c *Client) LRem(ctx context.Context, key string, count int64, value interface{}) (int64, error) {
	return c.rdb.LRem(ctx, key, count, value).Result()
}

// LTrim 修剪列表，只保留指定范围的元素
func (c *Client) LTrim(ctx context.Context, key string, start, stop int64) error {
	return c.rdb.LTrim(ctx, key, start, stop).Err()
}

// LInsert 在列表的指定位置插入元素
func (c *Client) LInsert(ctx context.Context, key, op string, pivot, value interface{}) (int64, error) {
	return c.rdb.LInsert(ctx, key, op, pivot, value).Result()
}

// BLPop 阻塞式从列表左侧弹出元素
func (c *Client) BLPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error) {
	return c.rdb.BLPop(ctx, timeout, keys...).Result()
}

// BRPop 阻塞式从列表右侧弹出元素
func (c *Client) BRPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error) {
	return c.rdb.BRPop(ctx, timeout, keys...).Result()
}

// RPopLPush 从源列表右侧弹出元素并推入目标列表左侧
func (c *Client) RPopLPush(ctx context.Context, source, destination string) (string, error) {
	return c.rdb.RPopLPush(ctx, source, destination).Result()
}

// BRPopLPush 阻塞式从源列表右侧弹出元素并推入目标列表左侧
func (c *Client) BRPopLPush(ctx context.Context, source, destination string, timeout time.Duration) (string, error) {
	return c.rdb.BRPopLPush(ctx, source, destination, timeout).Result()
}