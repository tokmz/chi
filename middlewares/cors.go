package middlewares

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"chi"
)

// CORSConfig CORS中间件配置
type CORSConfig struct {
	// AllowOrigins 允许的源列表，支持通配符 "*"
	AllowOrigins []string
	// AllowMethods 允许的HTTP方法列表
	AllowMethods []string
	// AllowHeaders 允许的请求头列表
	AllowHeaders []string
	// ExposeHeaders 暴露给客户端的响应头列表
	ExposeHeaders []string
	// AllowCredentials 是否允许发送凭据（cookies、授权头等）
	AllowCredentials bool
	// MaxAge 预检请求的缓存时间（秒）
	MaxAge time.Duration
}

// DefaultCORSConfig 默认CORS配置
var DefaultCORSConfig = CORSConfig{
	AllowOrigins: []string{"*"},
	AllowMethods: []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodHead,
		http.MethodOptions,
	},
	AllowHeaders: []string{
		"Origin",
		"Content-Length",
		"Content-Type",
		"Accept",
		"Authorization",
		"X-Requested-With",
		"X-CSRF-Token",
	},
	ExposeHeaders:    []string{},
	AllowCredentials: false,
	MaxAge:           12 * time.Hour,
}

// CORS 创建CORS中间件
// 使用默认配置的CORS中间件
func CORS() chi.MiddlewareFunc {
	return CORSWithConfig(DefaultCORSConfig)
}

// CORSWithConfig 使用自定义配置创建CORS中间件
// config: CORS配置参数
// 返回值: 配置好的CORS中间件函数
func CORSWithConfig(config CORSConfig) chi.MiddlewareFunc {
	// 预处理配置
	if len(config.AllowOrigins) == 0 {
		config.AllowOrigins = DefaultCORSConfig.AllowOrigins
	}
	if len(config.AllowMethods) == 0 {
		config.AllowMethods = DefaultCORSConfig.AllowMethods
	}
	if len(config.AllowHeaders) == 0 {
		config.AllowHeaders = DefaultCORSConfig.AllowHeaders
	}
	if config.MaxAge == 0 {
		config.MaxAge = DefaultCORSConfig.MaxAge
	}

	allowMethods := strings.Join(config.AllowMethods, ",")
	allowHeaders := strings.Join(config.AllowHeaders, ",")
	exposeHeaders := strings.Join(config.ExposeHeaders, ",")
	maxAge := strconv.FormatInt(int64(config.MaxAge.Seconds()), 10)

	return func(c *chi.Context) {
		origin := c.GetHeader("Origin")
		request := c.Request()

		// 检查是否允许该源
		allowedOrigin := ""
		if len(config.AllowOrigins) == 1 && config.AllowOrigins[0] == "*" {
			// 如果允许所有源
			if config.AllowCredentials {
				// 当允许凭据时，不能使用通配符，必须指定具体源
				allowedOrigin = origin
			} else {
				allowedOrigin = "*"
			}
		} else {
			// 检查源是否在允许列表中
			for _, allowOrigin := range config.AllowOrigins {
				if allowOrigin == origin {
					allowedOrigin = origin
					break
				}
			}
		}

		// 设置CORS响应头
		if allowedOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)
		}

		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// 处理预检请求
		if request.Method == http.MethodOptions {
			c.Header("Access-Control-Allow-Methods", allowMethods)
			c.Header("Access-Control-Allow-Headers", allowHeaders)
			c.Header("Access-Control-Max-Age", maxAge)
			c.Status(http.StatusNoContent)
			return
		}

		// 设置暴露的响应头
		if exposeHeaders != "" {
			c.Header("Access-Control-Expose-Headers", exposeHeaders)
		}

		// 继续处理请求
		c.Next()
	}
}

// CORSForDevelopment 开发环境CORS配置
// 允许所有源、所有方法、所有头部，适用于开发调试
func CORSForDevelopment() chi.MiddlewareFunc {
	return CORSWithConfig(CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: false,
		MaxAge:           24 * time.Hour,
	})
}

// CORSForProduction 生产环境CORS配置
// 严格的CORS配置，需要指定具体的允许源
func CORSForProduction(allowedOrigins []string) chi.MiddlewareFunc {
	return CORSWithConfig(CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
		},
		AllowCredentials: true,
		MaxAge:           1 * time.Hour,
	})
}