package cache

import (
	"context"
	"time"
)

// ScriptOperations Lua脚本操作接口
type ScriptOperations interface {
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error)
	EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) (interface{}, error)
	ScriptExists(ctx context.Context, hashes ...string) ([]bool, error)
	ScriptFlush(ctx context.Context) error
	ScriptKill(ctx context.Context) error
	ScriptLoad(ctx context.Context, script string) (string, error)
}

// CommonOperations 通用操作接口
type CommonOperations interface {
	Exists(ctx context.Context, keys ...string) (int64, error)
	Del(ctx context.Context, keys ...string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) (bool, error)
	ExpireAt(ctx context.Context, key string, tm time.Time) (bool, error)
	TTL(ctx context.Context, key string) (time.Duration, error)
	Persist(ctx context.Context, key string) (bool, error)
	Type(ctx context.Context, key string) (string, error)
	Keys(ctx context.Context, pattern string) ([]string, error)
	Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error)
	RandomKey(ctx context.Context) (string, error)
	Rename(ctx context.Context, key, newkey string) error
	RenameNX(ctx context.Context, key, newkey string) (bool, error)
	FlushDB(ctx context.Context) error
	FlushAll(ctx context.Context) error
	DBSize(ctx context.Context) (int64, error)
}

// Eval 执行Lua脚本
func (c *Client) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return c.rdb.Eval(ctx, script, keys, args...).Result()
}

// EvalSha 通过SHA1执行Lua脚本
func (c *Client) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) (interface{}, error) {
	return c.rdb.EvalSha(ctx, sha1, keys, args...).Result()
}

// ScriptExists 检查脚本是否存在
func (c *Client) ScriptExists(ctx context.Context, hashes ...string) ([]bool, error) {
	return c.rdb.ScriptExists(ctx, hashes...).Result()
}

// ScriptFlush 清空所有脚本缓存
func (c *Client) ScriptFlush(ctx context.Context) error {
	return c.rdb.ScriptFlush(ctx).Err()
}

// ScriptKill 终止正在执行的脚本
func (c *Client) ScriptKill(ctx context.Context) error {
	return c.rdb.ScriptKill(ctx).Err()
}

// ScriptLoad 加载脚本到服务器缓存
func (c *Client) ScriptLoad(ctx context.Context, script string) (string, error) {
	return c.rdb.ScriptLoad(ctx, script).Result()
}

// Exists 检查键是否存在
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.rdb.Exists(ctx, keys...).Result()
}

// Del 删除键
func (c *Client) Del(ctx context.Context, keys ...string) (int64, error) {
	return c.rdb.Del(ctx, keys...).Result()
}

// Expire 设置键的过期时间
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return c.rdb.Expire(ctx, key, expiration).Result()
}

// ExpireAt 设置键在指定时间过期
func (c *Client) ExpireAt(ctx context.Context, key string, tm time.Time) (bool, error) {
	return c.rdb.ExpireAt(ctx, key, tm).Result()
}

// TTL 获取键的剩余生存时间
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.rdb.TTL(ctx, key).Result()
}

// Persist 移除键的过期时间
func (c *Client) Persist(ctx context.Context, key string) (bool, error) {
	return c.rdb.Persist(ctx, key).Result()
}

// Type 获取键的数据类型
func (c *Client) Type(ctx context.Context, key string) (string, error) {
	return c.rdb.Type(ctx, key).Result()
}

// Keys 查找匹配模式的键
func (c *Client) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.rdb.Keys(ctx, pattern).Result()
}

// Scan 迭代数据库中的键
func (c *Client) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	keys, cursor, err := c.rdb.Scan(ctx, cursor, match, count).Result()
	return keys, cursor, err
}

// RandomKey 随机返回一个键
func (c *Client) RandomKey(ctx context.Context) (string, error) {
	return c.rdb.RandomKey(ctx).Result()
}

// Rename 重命名键
func (c *Client) Rename(ctx context.Context, key, newkey string) error {
	return c.rdb.Rename(ctx, key, newkey).Err()
}

// RenameNX 仅当新键名不存在时重命名键
func (c *Client) RenameNX(ctx context.Context, key, newkey string) (bool, error) {
	return c.rdb.RenameNX(ctx, key, newkey).Result()
}

// FlushDB 清空当前数据库
func (c *Client) FlushDB(ctx context.Context) error {
	return c.rdb.FlushDB(ctx).Err()
}

// FlushAll 清空所有数据库
func (c *Client) FlushAll(ctx context.Context) error {
	return c.rdb.FlushAll(ctx).Err()
}

// DBSize 获取数据库键的数量
func (c *Client) DBSize(ctx context.Context) (int64, error) {
	return c.rdb.DBSize(ctx).Result()
}