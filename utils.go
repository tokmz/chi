package chi

import "github.com/gin-gonic/gin"

// wrapHandler 包装处理函数
func wrapHandler(handler HandlerFunc) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		ctx := &Context{
			Context: ginCtx,
		}
		handler(ctx)
	}
}

// wrapMiddleware 包装中间件函数
func wrapMiddleware(middleware MiddlewareFunc) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		ctx := &Context{
			Context: ginCtx,
		}
		middleware(ctx)
	}
}
