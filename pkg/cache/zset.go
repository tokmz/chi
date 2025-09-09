package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// ZSetOperations 有序集合操作接口
type ZSetOperations interface {
	ZAdd(ctx context.Context, key string, members ...redis.Z) (int64, error)
	ZRem(ctx context.Context, key string, members ...interface{}) (int64, error)
	ZScore(ctx context.Context, key, member string) (float64, error)
	ZRank(ctx context.Context, key, member string) (int64, error)
	ZRevRank(ctx context.Context, key, member string) (int64, error)
	ZRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error)
	ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error)
	ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error)
	ZRangeByScoreWithScores(ctx context.Context, key string, opt *redis.ZRangeBy) ([]redis.Z, error)
	ZRevRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error)
	ZRevRangeByScoreWithScores(ctx context.Context, key string, opt *redis.ZRangeBy) ([]redis.Z, error)
	ZCard(ctx context.Context, key string) (int64, error)
	ZCount(ctx context.Context, key, min, max string) (int64, error)
	ZIncrBy(ctx context.Context, key string, increment float64, member string) (float64, error)
	ZRemRangeByRank(ctx context.Context, key string, start, stop int64) (int64, error)
	ZRemRangeByScore(ctx context.Context, key, min, max string) (int64, error)
	ZUnionStore(ctx context.Context, dest string, store *redis.ZStore) (int64, error)
	ZInterStore(ctx context.Context, dest string, store *redis.ZStore) (int64, error)
	ZScan(ctx context.Context, key string, cursor uint64, match string, count int64) ([]string, uint64, error)
}

// ZAdd 向有序集合添加成员
func (c *Client) ZAdd(ctx context.Context, key string, members ...redis.Z) (int64, error) {
	return c.rdb.ZAdd(ctx, key, members...).Result()
}

// ZRem 从有序集合删除成员
func (c *Client) ZRem(ctx context.Context, key string, members ...any) (int64, error) {
	return c.rdb.ZRem(ctx, key, members...).Result()
}

// ZScore 获取有序集合成员的分数
func (c *Client) ZScore(ctx context.Context, key, member string) (float64, error) {
	return c.rdb.ZScore(ctx, key, member).Result()
}

// ZRank 获取有序集合成员的排名（从小到大）
func (c *Client) ZRank(ctx context.Context, key, member string) (int64, error) {
	return c.rdb.ZRank(ctx, key, member).Result()
}

// ZRevRank 获取有序集合成员的排名（从大到小）
func (c *Client) ZRevRank(ctx context.Context, key, member string) (int64, error) {
	return c.rdb.ZRevRank(ctx, key, member).Result()
}

// ZRange 按排名范围获取有序集合成员
func (c *Client) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.rdb.ZRange(ctx, key, start, stop).Result()
}

// ZRangeWithScores 按排名范围获取有序集合成员及分数
func (c *Client) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	return c.rdb.ZRangeWithScores(ctx, key, start, stop).Result()
}

// ZRevRange 按排名范围获取有序集合成员（逆序）
func (c *Client) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.rdb.ZRevRange(ctx, key, start, stop).Result()
}

// ZRevRangeWithScores 按排名范围获取有序集合成员及分数（逆序）
func (c *Client) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	return c.rdb.ZRevRangeWithScores(ctx, key, start, stop).Result()
}

// ZRangeByScore 按分数范围获取有序集合成员
func (c *Client) ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error) {
	return c.rdb.ZRangeByScore(ctx, key, opt).Result()
}

// ZRangeByScoreWithScores 按分数范围获取有序集合成员及分数
func (c *Client) ZRangeByScoreWithScores(ctx context.Context, key string, opt *redis.ZRangeBy) ([]redis.Z, error) {
	return c.rdb.ZRangeByScoreWithScores(ctx, key, opt).Result()
}

// ZRevRangeByScore 按分数范围获取有序集合成员（逆序）
func (c *Client) ZRevRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error) {
	return c.rdb.ZRevRangeByScore(ctx, key, opt).Result()
}

// ZRevRangeByScoreWithScores 按分数范围获取有序集合成员及分数（逆序）
func (c *Client) ZRevRangeByScoreWithScores(ctx context.Context, key string, opt *redis.ZRangeBy) ([]redis.Z, error) {
	return c.rdb.ZRevRangeByScoreWithScores(ctx, key, opt).Result()
}

// ZCard 获取有序集合成员数量
func (c *Client) ZCard(ctx context.Context, key string) (int64, error) {
	return c.rdb.ZCard(ctx, key).Result()
}

// ZCount 计算分数范围内的成员数量
func (c *Client) ZCount(ctx context.Context, key, min, max string) (int64, error) {
	return c.rdb.ZCount(ctx, key, min, max).Result()
}

// ZIncrBy 为有序集合成员的分数加上增量
func (c *Client) ZIncrBy(ctx context.Context, key string, increment float64, member string) (float64, error) {
	return c.rdb.ZIncrBy(ctx, key, increment, member).Result()
}

// ZRemRangeByRank 移除排名范围内的成员
func (c *Client) ZRemRangeByRank(ctx context.Context, key string, start, stop int64) (int64, error) {
	return c.rdb.ZRemRangeByRank(ctx, key, start, stop).Result()
}

// ZRemRangeByScore 移除分数范围内的成员
func (c *Client) ZRemRangeByScore(ctx context.Context, key, min, max string) (int64, error) {
	return c.rdb.ZRemRangeByScore(ctx, key, min, max).Result()
}

// ZUnionStore 计算多个有序集合的并集并存储到目标集合
func (c *Client) ZUnionStore(ctx context.Context, dest string, store *redis.ZStore) (int64, error) {
	return c.rdb.ZUnionStore(ctx, dest, store).Result()
}

// ZInterStore 计算多个有序集合的交集并存储到目标集合
func (c *Client) ZInterStore(ctx context.Context, dest string, store *redis.ZStore) (int64, error) {
	return c.rdb.ZInterStore(ctx, dest, store).Result()
}

// ZScan 迭代有序集合中的成员
func (c *Client) ZScan(ctx context.Context, key string, cursor uint64, match string, count int64) ([]string, uint64, error) {
	keys, cursor, err := c.rdb.ZScan(ctx, key, cursor, match, count).Result()
	return keys, cursor, err
}
