package chi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RouterGroup 路由组结构体
// 封装了Gin的RouterGroup，提供更友好的API接口
// 支持路由分组、中间件应用、路径前缀等功能
type RouterGroup struct {
	// group Gin的路由组实例
	group *gin.RouterGroup
}

// Group 创建子路由组
// 在当前路由组的基础上创建新的子组，支持嵌套分组
// 参数 relativePath: 相对路径前缀
// 返回值: 新的路由组实例
func (rg *RouterGroup) Group(relativePath string, middleware ...MiddlewareFunc) *RouterGroup {
	subGroup := rg.group.Group(relativePath)
	for _, m := range middleware {
		subGroup.Use(wrapMiddleware(m))
	}
	return &RouterGroup{
		group: subGroup,
	}
}

// Use 为当前路由组添加中间件
// 中间件将应用于该组及其子组的所有路由
// 参数 middleware: 中间件函数列表
func (rg *RouterGroup) Use(middleware ...MiddlewareFunc) {
	for _, m := range middleware {
		rg.group.Use(wrapMiddleware(m))
	}
}

// GET 注册GET方法路由
// 在当前路由组下注册GET请求处理器
// 参数 relativePath: 相对路径
// 参数 handler: 处理函数
func (rg *RouterGroup) GET(relativePath string, handler HandlerFunc) {
	rg.group.GET(relativePath, wrapHandler(handler))
}

// POST 注册POST方法路由
// 在当前路由组下注册POST请求处理器
// 参数 relativePath: 相对路径
// 参数 handler: 处理函数
func (rg *RouterGroup) POST(relativePath string, handler HandlerFunc) {
	rg.group.POST(relativePath, wrapHandler(handler))
}

// PUT 注册PUT方法路由
// 在当前路由组下注册PUT请求处理器
// 参数 relativePath: 相对路径
// 参数 handler: 处理函数
func (rg *RouterGroup) PUT(relativePath string, handler HandlerFunc) {
	rg.group.PUT(relativePath, wrapHandler(handler))
}

// DELETE 注册DELETE方法路由
// 在当前路由组下注册DELETE请求处理器
// 参数 relativePath: 相对路径
// 参数 handler: 处理函数
func (rg *RouterGroup) DELETE(relativePath string, handler HandlerFunc) {
	rg.group.DELETE(relativePath, wrapHandler(handler))
}

// PATCH 注册PATCH方法路由
// 在当前路由组下注册PATCH请求处理器
// 参数 relativePath: 相对路径
// 参数 handler: 处理函数
func (rg *RouterGroup) PATCH(relativePath string, handler HandlerFunc) {
	rg.group.PATCH(relativePath, wrapHandler(handler))
}

// OPTIONS 注册OPTIONS方法路由
// 在当前路由组下注册OPTIONS请求处理器
// 参数 relativePath: 相对路径
// 参数 handler: 处理函数
func (rg *RouterGroup) OPTIONS(relativePath string, handler HandlerFunc) {
	rg.group.OPTIONS(relativePath, wrapHandler(handler))
}

// HEAD 注册HEAD方法路由
// 在当前路由组下注册HEAD请求处理器
// 参数 relativePath: 相对路径
// 参数 handler: 处理函数
func (rg *RouterGroup) HEAD(relativePath string, handler HandlerFunc) {
	rg.group.HEAD(relativePath, wrapHandler(handler))
}

// Any 注册所有HTTP方法的路由
// 为指定路径注册所有常用HTTP方法的处理器
// 参数 relativePath: 相对路径
// 参数 handler: 处理函数
func (rg *RouterGroup) Any(relativePath string, handler HandlerFunc) {
	rg.group.Any(relativePath, wrapHandler(handler))
}

// Handle 注册指定HTTP方法的路由
// 允许手动指定HTTP方法类型
// 参数 httpMethod: HTTP方法名（GET、POST等）
// 参数 relativePath: 相对路径
// 参数 handler: 处理函数
func (rg *RouterGroup) Handle(httpMethod, relativePath string, handler HandlerFunc) {
	rg.group.Handle(httpMethod, relativePath, wrapHandler(handler))
}

// Static 注册静态文件服务路由
// 为指定路径提供静态文件服务
// 参数 relativePath: URL路径前缀
// 参数 root: 文件系统根目录
func (rg *RouterGroup) Static(relativePath, root string) {
	rg.group.Static(relativePath, root)
}

// StaticFile 注册单个静态文件路由
// 为指定路径提供单个静态文件服务
// 参数 relativePath: URL路径
// 参数 filepath: 文件系统路径
func (rg *RouterGroup) StaticFile(relativePath, filepath string) {
	rg.group.StaticFile(relativePath, filepath)
}

// StaticFS 注册自定义文件系统的静态文件服务路由
// 使用自定义的http.FileSystem提供静态文件服务，支持嵌入式文件系统等
// 参数 relativePath: URL路径前缀
// 参数 fs: 自定义文件系统接口
func (rg *RouterGroup) StaticFS(relativePath string, fs http.FileSystem) {
	rg.group.StaticFS(relativePath, fs)
}

// StaticFileFS 注册使用自定义文件系统的单个静态文件路由
// 类似StaticFile但可以使用自定义的http.FileSystem
// 参数 relativePath: URL路径
// 参数 filepath: 文件系统路径
// 参数 fs: 自定义文件系统接口
func (rg *RouterGroup) StaticFileFS(relativePath, filepath string, fs http.FileSystem) {
	rg.group.StaticFileFS(relativePath, filepath, fs)
}

// BasePath 获取当前路由组的基础路径
// 返回当前路由组的完整路径前缀
// 返回值: 基础路径字符串
func (rg *RouterGroup) BasePath() string {
	return rg.group.BasePath()
}
