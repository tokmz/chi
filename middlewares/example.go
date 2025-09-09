package middlewares

import (
	"time"

	"chi"
)

// ExampleUsage 展示中间件的使用方法
func ExampleUsage() {
	// 创建服务器实例
	server := chi.New()

	// =============================================================================
	// Panic恢复中间件使用示例
	// =============================================================================

	// 1. 使用默认恢复配置（建议放在最前面）
	server.Use(Recovery())

	// 2. 开发环境恢复配置（详细错误信息）
	// server.Use(RecoveryForDevelopment())

	// 3. 生产环境恢复配置（简化错误信息）
	// server.Use(RecoveryForProduction())

	// 4. 自定义恢复配置
	// server.Use(RecoveryWithConfig(RecoveryConfig{
	// 	StackSize: 8 << 10, // 8KB
	// 	DisableStackAll: false,
	// 	LogFunc: func(c *chi.Context, err interface{}, stack []byte) {
	// 		// 自定义日志记录逻辑
	// 		log.Printf("Panic recovered: %v\n%s", err, stack)
	// 	},
	// 	RecoveryHandler: func(c *chi.Context, err interface{}) {
	// 		// 自定义错误响应
	// 		c.JSON(500, map[string]interface{}{
	// 			"error": "系统异常",
	// 			"request_id": c.GetHeader("X-Request-ID"),
	// 		})
	// 	},
	// }))

	// =============================================================================
	// CORS 中间件使用示例
	// =============================================================================

	// 1. 使用默认CORS配置
	server.Use(CORS())

	// 2. 开发环境CORS配置（允许所有源）
	// server.Use(CORSForDevelopment())

	// 3. 生产环境CORS配置（指定允许的源）
	// server.Use(CORSForProduction([]string{
	// 	"https://example.com",
	// 	"https://app.example.com",
	// }))

	// 4. 自定义CORS配置
	// server.Use(CORSWithConfig(CORSConfig{
	// 	AllowOrigins: []string{"https://trusted-domain.com"},
	// 	AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
	// 	AllowHeaders: []string{"Content-Type", "Authorization"},
	// 	AllowCredentials: true,
	// 	MaxAge: 1 * time.Hour,
	// }))

	// =============================================================================
	// 限流中间件使用示例
	// =============================================================================

	// 1. 基本限流（每秒100个请求，突发200个）
	server.Use(RateLimit(100, 200))

	// 2. 基于IP的限流
	// server.Use(RateLimitByIP(50, 100))

	// 3. 基于用户的限流
	// server.Use(RateLimitByUser(200, 400, "user_id"))

	// 4. 基于路径的限流
	// server.Use(RateLimitByPath(10, 20))

	// 5. 全局限流
	// server.Use(RateLimitGlobal(1000, 2000))

	// 6. 自定义限流配置
	// server.Use(RateLimitWithConfig(RateLimitConfig{
	// 	Rate:  50,
	// 	Burst: 100,
	// 	KeyFunc: func(c *chi.Context) string {
	// 		// 自定义键生成逻辑
	// 		return c.GetHeader("X-API-Key")
	// 	},
	// 	SkipFunc: func(c *chi.Context) bool {
	// 		// 跳过某些路径的限流
	// 		return c.FullPath() == "/health"
	// 	},
	// 	ErrorHandler: func(c *chi.Context) {
	// 		// 自定义错误响应
	// 		c.JSON(429, map[string]string{
	// 			"error": "请求过于频繁",
	// 			"retry_after": "60",
	// 		})
	// 	},
	// }))

	// =============================================================================
	// 路由组中使用中间件
	// =============================================================================

	// API路由组，应用特定的限流策略
	apiGroup := server.Group("/api/v1",
		// API接口限流更严格
		RateLimitByIP(30, 60),
		// 允许特定源的CORS
		CORSWithConfig(CORSConfig{
			AllowOrigins:     []string{"https://app.example.com"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
			AllowHeaders:     []string{"Content-Type", "Authorization"},
			AllowCredentials: true,
			MaxAge:           1 * time.Hour,
		}),
	)

	// 用户相关接口
	apiGroup.GET("/users", func(c *chi.Context) {
		c.JSON(200, map[string]string{"message": "用户列表"})
	})

	// 管理员路由组，更严格的限流
	adminGroup := server.Group("/admin",
		// 管理员接口限流更严格
		RateLimitByIP(10, 20),
		// 只允许管理后台域名
		CORSWithConfig(CORSConfig{
			AllowOrigins:     []string{"https://admin.example.com"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
			AllowCredentials: true,
		}),
	)

	adminGroup.GET("/dashboard", func(c *chi.Context) {
		c.JSON(200, map[string]string{"message": "管理面板"})
	})

	// =============================================================================
	// 特殊路径的中间件配置
	// =============================================================================

	// 健康检查接口，不应用限流
	server.GET("/health", func(c *chi.Context) {
		c.JSON(200, map[string]string{"status": "ok"})
	})

	// 文件上传接口，使用更宽松的限流
	uploadGroup := server.Group("/upload",
		// 文件上传限流更宽松
		RateLimitWithConfig(RateLimitConfig{
			Rate:  5,  // 每秒5个请求
			Burst: 10, // 突发10个请求
			KeyFunc: func(c *chi.Context) string {
				return c.ClientIP()
			},
		}),
	)
	uploadGroup.POST("", func(c *chi.Context) {
		c.JSON(200, map[string]string{"message": "文件上传成功"})
	})

	// 启动服务器
	server.Run(":8080")
}

// ProductionMiddlewares 生产环境推荐的中间件配置
func ProductionMiddlewares(allowedOrigins []string) []chi.MiddlewareFunc {
	return []chi.MiddlewareFunc{
		// Panic恢复中间件（必须放在最前面）
		RecoveryForProduction(),
		// 生产环境CORS配置
		CORSForProduction(allowedOrigins),
		// 基于IP的限流，防止单个IP过度请求
		RateLimitByIP(100, 200),
		// 全局限流，防止服务器过载
		RateLimitGlobal(1000, 2000),
	}
}

// DevelopmentMiddlewares 开发环境推荐的中间件配置
func DevelopmentMiddlewares() []chi.MiddlewareFunc {
	return []chi.MiddlewareFunc{
		// Panic恢复中间件（必须放在最前面）
		RecoveryForDevelopment(),
		// 开发环境宽松的CORS配置
		CORSForDevelopment(),
		// 开发环境宽松的限流配置
		RateLimit(1000, 2000),
	}
}