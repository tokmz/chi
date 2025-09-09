package cache

import (
	"context"
	"encoding/json"
	"time"
)

// StringOperations 字符串操作接口
type StringOperations interface {
	Set(ctx context.Context, key string, value interface{}, expiration ...time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	GetSet(ctx context.Context, key string, value interface{}) (string, error)
	SetNX(ctx context.Context, key string, value interface{}, expiration ...time.Duration) (bool, error)
	SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	MSet(ctx context.Context, pairs ...interface{}) error
	MGet(ctx context.Context, keys ...string) ([]interface{}, error)
	Incr(ctx context.Context, key string) (int64, error)
	IncrBy(ctx context.Context, key string, value int64) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
	DecrBy(ctx context.Context, key string, value int64) (int64, error)
	Append(ctx context.Context, key, value string) (int64, error)
	StrLen(ctx context.Context, key string) (int64, error)
	GetRange(ctx context.Context, key string, start, end int64) (string, error)
	SetRange(ctx context.Context, key string, offset int64, value string) (int64, error)
}

// Set 设置字符串值
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration ...time.Duration) error {
	exp := c.defaultExpiry
	if len(expiration) > 0 {
		exp = expiration[0]
	}

	// 如果值是复杂类型，序列化为JSON
	var val interface{}
	switch v := value.(type) {
	case string, []byte, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		val = v
	default:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return err
		}
		val = jsonBytes
	}

	return c.rdb.Set(ctx, key, val, exp).Err()
}

// Get 获取字符串值
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}

// GetObject 获取对象（自动反序列化JSON）
func (c *Client) GetObject(ctx context.Context, key string, dest interface{}) error {
	result, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(result), dest)
}

// GetSet 设置新值并返回旧值
func (c *Client) GetSet(ctx context.Context, key string, value interface{}) (string, error) {
	return c.rdb.GetSet(ctx, key, value).Result()
}

// SetNX 仅当键不存在时设置值
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, expiration ...time.Duration) (bool, error) {
	exp := c.defaultExpiry
	if len(expiration) > 0 {
		exp = expiration[0]
	}

	return c.rdb.SetNX(ctx, key, value, exp).Result()
}

// SetEX 设置值并指定过期时间
func (c *Client) SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.rdb.SetEx(ctx, key, value, expiration).Err()
}

// MSet 批量设置多个键值对
func (c *Client) MSet(ctx context.Context, pairs ...interface{}) error {
	return c.rdb.MSet(ctx, pairs...).Err()
}

// MGet 批量获取多个键的值
func (c *Client) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	return c.rdb.MGet(ctx, keys...).Result()
}

// Incr 递增键的整数值
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.rdb.Incr(ctx, key).Result()
}

// IncrBy 按指定值递增键的整数值
func (c *Client) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.rdb.IncrBy(ctx, key, value).Result()
}

// Decr 递减键的整数值
func (c *Client) Decr(ctx context.Context, key string) (int64, error) {
	return c.rdb.Decr(ctx, key).Result()
}

// DecrBy 按指定值递减键的整数值
func (c *Client) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.rdb.DecrBy(ctx, key, value).Result()
}

// Append 追加字符串到键的值末尾
func (c *Client) Append(ctx context.Context, key, value string) (int64, error) {
	return c.rdb.Append(ctx, key, value).Result()
}

// StrLen 获取字符串值的长度
func (c *Client) StrLen(ctx context.Context, key string) (int64, error) {
	return c.rdb.StrLen(ctx, key).Result()
}

// GetRange 获取字符串的子串
func (c *Client) GetRange(ctx context.Context, key string, start, end int64) (string, error) {
	return c.rdb.GetRange(ctx, key, start, end).Result()
}

// SetRange 覆盖字符串的部分内容
func (c *Client) SetRange(ctx context.Context, key string, offset int64, value string) (int64, error) {
	return c.rdb.SetRange(ctx, key, offset, value).Result()
}