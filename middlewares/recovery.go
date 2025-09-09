package middlewares

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"chi"
)

// RecoveryConfig panic恢复中间件配置
type RecoveryConfig struct {
	// StackSize 堆栈跟踪的最大字节数
	StackSize int
	// DisableStackAll 是否禁用所有goroutine的堆栈跟踪
	DisableStackAll bool
	// DisablePrintStack 是否禁用打印堆栈信息
	DisablePrintStack bool
	// LogFunc 自定义日志记录函数
	LogFunc func(c *chi.Context, err interface{}, stack []byte)
	// RecoveryHandler 自定义恢复处理函数
	RecoveryHandler func(c *chi.Context, err interface{})
}

// DefaultRecoveryConfig 默认恢复配置
var DefaultRecoveryConfig = RecoveryConfig{
	StackSize:         4 << 10, // 4KB
	DisableStackAll:   false,
	DisablePrintStack: false,
	LogFunc:           defaultLogFunc,
	RecoveryHandler:   defaultRecoveryHandler,
}

// defaultLogFunc 默认日志记录函数
func defaultLogFunc(c *chi.Context, err interface{}, stack []byte) {
	timeFormat := "2006/01/02 - 15:04:05"
	fmt.Printf("[PANIC RECOVERY] %s | %s %s | Error: %v\n%s\n",
		time.Now().Format(timeFormat),
		c.Request().Method,
		c.Request().URL.Path,
		err,
		stack,
	)
}

// defaultRecoveryHandler 默认恢复处理函数
func defaultRecoveryHandler(c *chi.Context, err interface{}) {
	c.JSON(http.StatusInternalServerError, map[string]interface{}{
		"error":   "Internal Server Error",
		"message": "服务器内部错误，请稍后重试",
		"code":    http.StatusInternalServerError,
	})
}

// Recovery 创建panic恢复中间件
// 使用默认配置的恢复中间件
func Recovery() chi.MiddlewareFunc {
	return RecoveryWithConfig(DefaultRecoveryConfig)
}

// RecoveryWithConfig 使用自定义配置创建panic恢复中间件
// config: 恢复配置参数
// 返回值: 配置好的恢复中间件函数
func RecoveryWithConfig(config RecoveryConfig) chi.MiddlewareFunc {
	// 设置默认值
	if config.StackSize <= 0 {
		config.StackSize = DefaultRecoveryConfig.StackSize
	}
	if config.LogFunc == nil {
		config.LogFunc = DefaultRecoveryConfig.LogFunc
	}
	if config.RecoveryHandler == nil {
		config.RecoveryHandler = DefaultRecoveryConfig.RecoveryHandler
	}

	return func(c *chi.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈跟踪信息
				var stack []byte
				if !config.DisablePrintStack {
					stack = make([]byte, config.StackSize)
					if config.DisableStackAll {
						// 只获取当前goroutine的堆栈
						length := runtime.Stack(stack, false)
						stack = stack[:length]
					} else {
						// 获取所有goroutine的堆栈
						length := runtime.Stack(stack, true)
						stack = stack[:length]
					}
				}

				// 记录日志
				config.LogFunc(c, err, stack)

				// 处理恢复
				config.RecoveryHandler(c, err)
			}
		}()

		// 继续处理请求
		c.Next()
	}
}

// RecoveryWithWriter 使用自定义writer的恢复中间件
// 将panic信息写入指定的writer
func RecoveryWithWriter(out func(string)) chi.MiddlewareFunc {
	return RecoveryWithConfig(RecoveryConfig{
		StackSize:         4 << 10,
		DisableStackAll:   false,
		DisablePrintStack: false,
		LogFunc: func(c *chi.Context, err interface{}, stack []byte) {
			timeFormat := "2006/01/02 - 15:04:05"
			msg := fmt.Sprintf("[PANIC RECOVERY] %s | %s %s | Error: %v\n%s\n",
				time.Now().Format(timeFormat),
				c.Request().Method,
				c.Request().URL.Path,
				err,
				stack,
			)
			out(msg)
		},
		RecoveryHandler: defaultRecoveryHandler,
	})
}

// RecoveryWithLogger 使用自定义日志记录器的恢复中间件
// logger: 自定义日志记录函数
func RecoveryWithLogger(logger func(c *chi.Context, err interface{}, stack []byte)) chi.MiddlewareFunc {
	return RecoveryWithConfig(RecoveryConfig{
		StackSize:         4 << 10,
		DisableStackAll:   false,
		DisablePrintStack: false,
		LogFunc:           logger,
		RecoveryHandler:   defaultRecoveryHandler,
	})
}

// RecoveryWithHandler 使用自定义处理器的恢复中间件
// handler: 自定义恢复处理函数
func RecoveryWithHandler(handler func(c *chi.Context, err interface{})) chi.MiddlewareFunc {
	return RecoveryWithConfig(RecoveryConfig{
		StackSize:         4 << 10,
		DisableStackAll:   false,
		DisablePrintStack: false,
		LogFunc:           defaultLogFunc,
		RecoveryHandler:   handler,
	})
}

// RecoveryForProduction 生产环境恢复中间件
// 不打印详细的堆栈信息，只记录错误
func RecoveryForProduction() chi.MiddlewareFunc {
	return RecoveryWithConfig(RecoveryConfig{
		StackSize:         1 << 10, // 1KB
		DisableStackAll:   true,    // 只获取当前goroutine堆栈
		DisablePrintStack: false,
		LogFunc: func(c *chi.Context, err interface{}, stack []byte) {
			// 生产环境简化日志
			timeFormat := "2006/01/02 - 15:04:05"
			fmt.Printf("[PANIC] %s | %s %s | Error: %v\n",
				time.Now().Format(timeFormat),
				c.Request().Method,
				c.Request().URL.Path,
				err,
			)
			// 可以在这里集成日志系统，如logrus、zap等
		},
		RecoveryHandler: func(c *chi.Context, err interface{}) {
			// 生产环境不暴露详细错误信息
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error":   "Internal Server Error",
				"message": "服务暂时不可用，请稍后重试",
				"code":    http.StatusInternalServerError,
			})
		},
	})
}

// RecoveryForDevelopment 开发环境恢复中间件
// 打印详细的堆栈信息，便于调试
func RecoveryForDevelopment() chi.MiddlewareFunc {
	return RecoveryWithConfig(RecoveryConfig{
		StackSize:         8 << 10, // 8KB
		DisableStackAll:   false,   // 获取所有goroutine堆栈
		DisablePrintStack: false,
		LogFunc: func(c *chi.Context, err interface{}, stack []byte) {
			// 开发环境详细日志
			timeFormat := "2006/01/02 - 15:04:05"
			fmt.Printf("\n=== PANIC RECOVERY ===\n")
			fmt.Printf("Time: %s\n", time.Now().Format(timeFormat))
			fmt.Printf("Method: %s\n", c.Request().Method)
			fmt.Printf("Path: %s\n", c.Request().URL.Path)
			fmt.Printf("Error: %v\n", err)
			fmt.Printf("Stack Trace:\n%s\n", stack)
			fmt.Printf("======================\n\n")
		},
		RecoveryHandler: func(c *chi.Context, err interface{}) {
			// 开发环境返回详细错误信息
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error":   "Internal Server Error",
				"message": fmt.Sprintf("服务器发生panic: %v", err),
				"code":    http.StatusInternalServerError,
				"debug":   true,
			})
		},
	})
}

// RecoveryWithMetrics 带指标统计的恢复中间件
// 可以集成Prometheus等监控系统
func RecoveryWithMetrics() chi.MiddlewareFunc {
	// 这里可以集成指标收集逻辑
	var panicCount int64

	return RecoveryWithConfig(RecoveryConfig{
		StackSize:         4 << 10,
		DisableStackAll:   false,
		DisablePrintStack: false,
		LogFunc: func(c *chi.Context, err interface{}, stack []byte) {
			// 增加panic计数
			panicCount++

			// 记录日志
			defaultLogFunc(c, err, stack)

			// 这里可以发送指标到监控系统
			// metrics.IncrementPanicCounter()
			// metrics.RecordPanicPath(c.Request().URL.Path)
		},
		RecoveryHandler: defaultRecoveryHandler,
	})
}

// getStackTrace 获取格式化的堆栈跟踪信息
func getStackTrace(stackSize int, all bool) string {
	stack := make([]byte, stackSize)
	length := runtime.Stack(stack, all)
	return string(stack[:length])
}

// formatPanicError 格式化panic错误信息
func formatPanicError(err interface{}) string {
	switch e := err.(type) {
	case error:
		return e.Error()
	case string:
		return e
	default:
		return fmt.Sprintf("%v", e)
	}
}

// extractPanicInfo 提取panic相关信息
func extractPanicInfo(stack []byte) (file string, line int, function string) {
	lines := strings.Split(string(stack), "\n")
	if len(lines) >= 3 {
		// 通常第二行包含文件和行号信息
		if len(lines) > 1 {
			function = strings.TrimSpace(lines[1])
		}
		if len(lines) > 2 {
			fileInfo := strings.TrimSpace(lines[2])
			parts := strings.Split(fileInfo, ":")
			if len(parts) >= 2 {
				file = parts[0]
				fmt.Sscanf(parts[1], "%d", &line)
			}
		}
	}
	return
}