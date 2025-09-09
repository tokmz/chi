package cache

import (
	"context"
	"encoding/json"
)

// SetOperations 集合操作接口
type SetOperations interface {
	SAdd(ctx context.Context, key string, members ...interface{}) (int64, error)
	SRem(ctx context.Context, key string, members ...interface{}) (int64, error)
	SMembers(ctx context.Context, key string) ([]string, error)
	SIsMember(ctx context.Context, key string, member interface{}) (bool, error)
	SCard(ctx context.Context, key string) (int64, error)
	SPop(ctx context.Context, key string) (string, error)
	SPopN(ctx context.Context, key string, count int64) ([]string, error)
	SRandMember(ctx context.Context, key string) (string, error)
	SRandMemberN(ctx context.Context, key string, count int64) ([]string, error)
	SMove(ctx context.Context, source, destination string, member interface{}) (bool, error)
	SUnion(ctx context.Context, keys ...string) ([]string, error)
	SUnionStore(ctx context.Context, destination string, keys ...string) (int64, error)
	SInter(ctx context.Context, keys ...string) ([]string, error)
	SInterStore(ctx context.Context, destination string, keys ...string) (int64, error)
	SDiff(ctx context.Context, keys ...string) ([]string, error)
	SDiffStore(ctx context.Context, destination string, keys ...string) (int64, error)
	SScan(ctx context.Context, key string, cursor uint64, match string, count int64) ([]string, uint64, error)
}

// SAdd 向集合添加成员
func (c *Client) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	// 处理复杂类型的序列化
	processedMembers := make([]interface{}, len(members))
	for i, m := range members {
		switch member := m.(type) {
		case string, []byte, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
			processedMembers[i] = member
		default:
			jsonBytes, err := json.Marshal(member)
			if err != nil {
				return 0, err
			}
			processedMembers[i] = jsonBytes
		}
	}

	return c.rdb.SAdd(ctx, key, processedMembers...).Result()
}

// SRem 从集合删除成员
func (c *Client) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return c.rdb.SRem(ctx, key, members...).Result()
}

// SMembers 获取集合所有成员
func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.rdb.SMembers(ctx, key).Result()
}

// SIsMember 检查成员是否在集合中
func (c *Client) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return c.rdb.SIsMember(ctx, key, member).Result()
}

// SCard 获取集合成员数量
func (c *Client) SCard(ctx context.Context, key string) (int64, error) {
	return c.rdb.SCard(ctx, key).Result()
}

// SPop 随机移除并返回集合中的一个成员
func (c *Client) SPop(ctx context.Context, key string) (string, error) {
	return c.rdb.SPop(ctx, key).Result()
}

// SPopN 随机移除并返回集合中的count个成员
func (c *Client) SPopN(ctx context.Context, key string, count int64) ([]string, error) {
	return c.rdb.SPopN(ctx, key, count).Result()
}

// SRandMember 随机返回集合中的一个成员
func (c *Client) SRandMember(ctx context.Context, key string) (string, error) {
	return c.rdb.SRandMember(ctx, key).Result()
}

// SRandMemberN 随机返回集合中的count个成员
func (c *Client) SRandMemberN(ctx context.Context, key string, count int64) ([]string, error) {
	return c.rdb.SRandMemberN(ctx, key, count).Result()
}

// SMove 将成员从源集合移动到目标集合
func (c *Client) SMove(ctx context.Context, source, destination string, member interface{}) (bool, error) {
	return c.rdb.SMove(ctx, source, destination, member).Result()
}

// SUnion 返回多个集合的并集
func (c *Client) SUnion(ctx context.Context, keys ...string) ([]string, error) {
	return c.rdb.SUnion(ctx, keys...).Result()
}

// SUnionStore 计算多个集合的并集并存储到目标集合
func (c *Client) SUnionStore(ctx context.Context, destination string, keys ...string) (int64, error) {
	return c.rdb.SUnionStore(ctx, destination, keys...).Result()
}

// SInter 返回多个集合的交集
func (c *Client) SInter(ctx context.Context, keys ...string) ([]string, error) {
	return c.rdb.SInter(ctx, keys...).Result()
}

// SInterStore 计算多个集合的交集并存储到目标集合
func (c *Client) SInterStore(ctx context.Context, destination string, keys ...string) (int64, error) {
	return c.rdb.SInterStore(ctx, destination, keys...).Result()
}

// SDiff 返回多个集合的差集
func (c *Client) SDiff(ctx context.Context, keys ...string) ([]string, error) {
	return c.rdb.SDiff(ctx, keys...).Result()
}

// SDiffStore 计算多个集合的差集并存储到目标集合
func (c *Client) SDiffStore(ctx context.Context, destination string, keys ...string) (int64, error) {
	return c.rdb.SDiffStore(ctx, destination, keys...).Result()
}

// SScan 迭代集合中的成员
func (c *Client) SScan(ctx context.Context, key string, cursor uint64, match string, count int64) ([]string, uint64, error) {
	keys, cursor, err := c.rdb.SScan(ctx, key, cursor, match, count).Result()
	return keys, cursor, err
}