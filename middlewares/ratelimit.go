package middlewares

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"chi"
)

// RateLimitConfig 限流中间件配置
type RateLimitConfig struct {
	// Rate 每秒允许的请求数
	Rate int
	// Burst 令牌桶容量，允许的突发请求数
	Burst int
	// KeyFunc 生成限流键的函数，默认使用客户端IP
	KeyFunc func(*chi.Context) string
	// SkipFunc 跳过限流的条件函数
	SkipFunc func(*chi.Context) bool
	// ErrorHandler 限流触发时的错误处理函数
	ErrorHandler func(*chi.Context)
}

// TokenBucket 令牌桶结构
type TokenBucket struct {
	capacity int64     // 桶容量
	tokens   int64     // 当前令牌数
	rate     int64     // 每秒补充的令牌数
	lastTime time.Time // 上次更新时间
	mu       sync.Mutex
}

// NewTokenBucket 创建新的令牌桶
func NewTokenBucket(rate, capacity int) *TokenBucket {
	return &TokenBucket{
		capacity: int64(capacity),
		tokens:   int64(capacity),
		rate:     int64(rate),
		lastTime: time.Now(),
	}
}

// Allow 检查是否允许请求通过
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	// 计算需要补充的令牌数
	elapsed := now.Sub(tb.lastTime)
	tokensToAdd := int64(elapsed.Seconds()) * tb.rate

	// 更新令牌数，不超过桶容量
	tb.tokens += tokensToAdd
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
	tb.lastTime = now

	// 检查是否有可用令牌
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}

// RateLimiter 限流器管理器
type RateLimiter struct {
	buckets map[string]*TokenBucket
	mu      sync.RWMutex
	rate    int
	burst   int
}

// NewRateLimiter 创建新的限流器
func NewRateLimiter(rate, burst int) *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*TokenBucket),
		rate:    rate,
		burst:   burst,
	}
}

// Allow 检查指定键是否允许请求
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.RLock()
	bucket, exists := rl.buckets[key]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		// 双重检查
		if bucket, exists = rl.buckets[key]; !exists {
			bucket = NewTokenBucket(rl.rate, rl.burst)
			rl.buckets[key] = bucket
		}
		rl.mu.Unlock()
	}

	return bucket.Allow()
}

// Cleanup 清理过期的令牌桶
func (rl *RateLimiter) Cleanup(maxAge time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, bucket := range rl.buckets {
		bucket.mu.Lock()
		if now.Sub(bucket.lastTime) > maxAge {
			delete(rl.buckets, key)
		}
		bucket.mu.Unlock()
	}
}

// 全局限流器实例
var globalRateLimiter *RateLimiter
var rateLimiterOnce sync.Once

// getGlobalRateLimiter 获取全局限流器实例
func getGlobalRateLimiter(rate, burst int) *RateLimiter {
	rateLimiterOnce.Do(func() {
		globalRateLimiter = NewRateLimiter(rate, burst)
		// 启动清理协程
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				globalRateLimiter.Cleanup(10 * time.Minute)
			}
		}()
	})
	return globalRateLimiter
}

// defaultKeyFunc 默认的键生成函数，使用客户端IP
func defaultKeyFunc(c *chi.Context) string {
	return c.ClientIP()
}

// defaultErrorHandler 默认的错误处理函数
func defaultErrorHandler(c *chi.Context) {
	c.JSON(http.StatusTooManyRequests, map[string]interface{}{
		"error":   "Too Many Requests",
		"message": "请求过于频繁，请稍后再试",
		"code":    http.StatusTooManyRequests,
	})
}

// RateLimit 创建限流中间件
// rate: 每秒允许的请求数
// burst: 令牌桶容量
func RateLimit(rate, burst int) chi.MiddlewareFunc {
	return RateLimitWithConfig(RateLimitConfig{
		Rate:         rate,
		Burst:        burst,
		KeyFunc:      defaultKeyFunc,
		ErrorHandler: defaultErrorHandler,
	})
}

// RateLimitWithConfig 使用自定义配置创建限流中间件
func RateLimitWithConfig(config RateLimitConfig) chi.MiddlewareFunc {
	// 设置默认值
	if config.Rate <= 0 {
		config.Rate = 100 // 默认每秒100个请求
	}
	if config.Burst <= 0 {
		config.Burst = config.Rate * 2 // 默认突发容量为速率的2倍
	}
	if config.KeyFunc == nil {
		config.KeyFunc = defaultKeyFunc
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultErrorHandler
	}

	rateLimiter := getGlobalRateLimiter(config.Rate, config.Burst)

	return func(c *chi.Context) {
		// 检查是否跳过限流
		if config.SkipFunc != nil && config.SkipFunc(c) {
			c.Next()
			return
		}

		// 生成限流键
		key := config.KeyFunc(c)
		if key == "" {
			// 如果无法生成键，跳过限流
			c.Next()
			return
		}

		// 检查是否允许请求
		if !rateLimiter.Allow(key) {
			config.ErrorHandler(c)
			return
		}

		// 继续处理请求
		c.Next()
	}
}

// RateLimitByIP 基于IP的限流中间件
func RateLimitByIP(rate, burst int) chi.MiddlewareFunc {
	return RateLimitWithConfig(RateLimitConfig{
		Rate:  rate,
		Burst: burst,
		KeyFunc: func(c *chi.Context) string {
			return fmt.Sprintf("ip:%s", c.ClientIP())
		},
		ErrorHandler: defaultErrorHandler,
	})
}

// RateLimitByUser 基于用户ID的限流中间件
// userIDKey: 从Context中获取用户ID的键名
func RateLimitByUser(rate, burst int, userIDKey string) chi.MiddlewareFunc {
	return RateLimitWithConfig(RateLimitConfig{
		Rate:  rate,
		Burst: burst,
		KeyFunc: func(c *chi.Context) string {
			userID, exists := c.Get(userIDKey)
			if !exists {
				// 如果没有用户ID，回退到IP限流
				return fmt.Sprintf("ip:%s", c.ClientIP())
			}
			return fmt.Sprintf("user:%v", userID)
		},
		ErrorHandler: defaultErrorHandler,
	})
}

// RateLimitByPath 基于请求路径的限流中间件
func RateLimitByPath(rate, burst int) chi.MiddlewareFunc {
	return RateLimitWithConfig(RateLimitConfig{
		Rate:  rate,
		Burst: burst,
		KeyFunc: func(c *chi.Context) string {
			return fmt.Sprintf("path:%s:%s", c.Request().Method, c.FullPath())
		},
		ErrorHandler: defaultErrorHandler,
	})
}

// RateLimitGlobal 全局限流中间件
func RateLimitGlobal(rate, burst int) chi.MiddlewareFunc {
	return RateLimitWithConfig(RateLimitConfig{
		Rate:  rate,
		Burst: burst,
		KeyFunc: func(c *chi.Context) string {
			return "global"
		},
		ErrorHandler: defaultErrorHandler,
	})
}