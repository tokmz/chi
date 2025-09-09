package chi

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// =============================================================================
// 类型定义
// =============================================================================

// Server Web服务器结构体，封装了Gin引擎和相关配置
// 提供了完整的HTTP服务器功能，包括路由注册、中间件管理、静态文件服务等
type Server struct {
	// cfg 服务器配置信息
	cfg *Config
	// engine Gin 路由引擎，处理 HTTP 请求路由和中间件
	engine *gin.Engine
	// server 底层 HTTP 服务器实例，用于启动和管理服务
	server *http.Server
	// quit 退出信号通道，用于优雅关闭服务器
	quit chan os.Signal
}

// HandlerFunc 处理函数类型定义
// 封装了自定义的Context，提供更友好的API接口
type HandlerFunc func(*Context)

// MiddlewareFunc 中间件函数类型定义
// 用于处理请求前后的逻辑，如认证、日志、CORS等
type MiddlewareFunc func(*Context)

// =============================================================================
// 构造函数
// =============================================================================

// New 创建新的Server实例
// 初始化Gin引擎和退出信号通道，返回可用的服务器实例
// 返回值: *Server 新创建的服务器实例
func New() *Server {
	engine := gin.New()
	return &Server{
		engine: engine,
		quit:   make(chan os.Signal, 1),
	}
}

// =============================================================================
// 服务器配置方法
// =============================================================================

// SetMode 设置Gin运行模式
// 支持三种模式：gin.DebugMode, gin.ReleaseMode, gin.TestMode
// 参数 mode: 运行模式字符串
func (s *Server) SetMode(mode string) {
	gin.SetMode(mode)
}

// SetTrustedProxies 设置可信代理列表
// 用于安全地获取客户端真实IP地址，防止IP伪造攻击
// 参数 trustedProxies: 可信代理IP地址列表
// 返回值: error 设置过程中的错误信息
func (s *Server) SetTrustedProxies(trustedProxies []string) error {
	return s.engine.SetTrustedProxies(trustedProxies)
}

// RemoteIPHeaders 设置远程IP头信息
// 指定从哪些HTTP头中获取客户端真实IP
// 参数 headers: HTTP头名称列表，如"X-Real-IP", "X-Forwarded-For"
func (s *Server) RemoteIPHeaders(headers ...string) {
	s.engine.RemoteIPHeaders = headers
}

// ForwardedByClientIP 设置是否通过客户端IP转发
// 控制是否信任客户端提供的IP信息
// 参数 value: true表示信任客户端IP，false表示不信任
func (s *Server) ForwardedByClientIP(value bool) {
	s.engine.ForwardedByClientIP = value
}

// UseRawPath 设置是否使用原始路径
// 控制路由匹配时是否使用未解码的原始路径
// 参数 value: true表示使用原始路径，false表示使用解码后的路径
func (s *Server) UseRawPath(value bool) {
	s.engine.UseRawPath = value
}

// UnescapePathValues 设置是否取消转义路径值
// 控制路径参数是否进行URL解码
// 参数 value: true表示进行解码，false表示保持原样
func (s *Server) UnescapePathValues(value bool) {
	s.engine.UnescapePathValues = value
}

// MaxMultipartMemory 设置多部分表单最大内存限制
// 控制文件上传时在内存中缓存的最大字节数，超出部分将写入临时文件
// 参数 value: 最大内存字节数，默认为32MB
func (s *Server) MaxMultipartMemory(value int64) {
	s.engine.MaxMultipartMemory = value
}

// HandleMethodNotAllowed 设置是否处理方法不允许的请求
// 当请求路径存在但HTTP方法不匹配时，是否返回405状态码
// 参数 value: true表示处理并返回405，false表示返回404
func (s *Server) HandleMethodNotAllowed(value bool) {
	s.engine.HandleMethodNotAllowed = value
}

// RedirectTrailingSlash 设置是否重定向尾部斜杠
// 自动处理URL尾部斜杠的重定向，如/users/重定向到/users
// 参数 value: true表示启用重定向，false表示禁用
func (s *Server) RedirectTrailingSlash(value bool) {
	s.engine.RedirectTrailingSlash = value
}

// RedirectFixedPath 设置是否重定向固定路径
// 自动修复URL路径中的错误，如大小写、多余斜杠等
// 参数 value: true表示启用路径修复，false表示禁用
func (s *Server) RedirectFixedPath(value bool) {
	s.engine.RedirectFixedPath = value
}

// =============================================================================
// 中间件管理
// =============================================================================

// Use 添加全局中间件
// 注册的中间件将应用于所有路由，按注册顺序执行
// 参数 middleware: 可变参数，支持同时注册多个中间件
func (s *Server) Use(middleware ...MiddlewareFunc) {
	for _, m := range middleware {
		s.engine.Use(wrapMiddleware(m))
	}
}

// NoRoute 设置404处理函数
// 当请求的路径不存在时调用此处理函数
// 参数 handler: 404错误处理函数
func (s *Server) NoRoute(handler HandlerFunc) {
	s.engine.NoRoute(wrapHandler(handler))
}

// NoMethod 设置405处理函数
// 当请求路径存在但HTTP方法不被允许时调用此处理函数
// 参数 handler: 405错误处理函数
func (s *Server) NoMethod(handler HandlerFunc) {
	s.engine.NoMethod(wrapHandler(handler))
}

// =============================================================================
// 路由注册方法
// =============================================================================

// GET 注册GET请求路由
// 用于处理数据查询和页面展示
// 参数 path: 路由路径，支持参数如/users/:id
// 参数 handler: 处理函数
func (s *Server) GET(path string, handler HandlerFunc) {
	s.engine.GET(path, wrapHandler(handler))
}

// POST 注册POST请求路由
// 用于处理数据创建和表单提交
// 参数 path: 路由路径
// 参数 handler: 处理函数
func (s *Server) POST(path string, handler HandlerFunc) {
	s.engine.POST(path, wrapHandler(handler))
}

// PUT 注册PUT请求路由
// 用于处理数据的完整更新
// 参数 path: 路由路径
// 参数 handler: 处理函数
func (s *Server) PUT(path string, handler HandlerFunc) {
	s.engine.PUT(path, wrapHandler(handler))
}

// DELETE 注册DELETE请求路由
// 用于处理数据删除操作
// 参数 path: 路由路径
// 参数 handler: 处理函数
func (s *Server) DELETE(path string, handler HandlerFunc) {
	s.engine.DELETE(path, wrapHandler(handler))
}

// PATCH 注册PATCH请求路由
// 用于处理数据的部分更新
// 参数 path: 路由路径
// 参数 handler: 处理函数
func (s *Server) PATCH(path string, handler HandlerFunc) {
	s.engine.PATCH(path, wrapHandler(handler))
}

// OPTIONS 注册OPTIONS请求路由
// 用于处理跨域预检请求和API选项查询
// 参数 path: 路由路径
// 参数 handler: 处理函数
func (s *Server) OPTIONS(path string, handler HandlerFunc) {
	s.engine.OPTIONS(path, wrapHandler(handler))
}

// HEAD 注册HEAD请求路由
// 用于获取资源的元信息，不返回响应体
// 参数 path: 路由路径
// 参数 handler: 处理函数
func (s *Server) HEAD(path string, handler HandlerFunc) {
	s.engine.HEAD(path, wrapHandler(handler))
}

// Any 注册所有HTTP方法的路由
// 该路由将响应所有HTTP方法的请求
// 参数 path: 路由路径
// 参数 handler: 处理函数
func (s *Server) Any(path string, handler HandlerFunc) {
	s.engine.Any(path, wrapHandler(handler))
}

// Match 注册指定HTTP方法列表的路由
// 只响应指定方法列表中的请求
// 参数 methods: HTTP方法列表，如["GET", "POST"]
// 参数 path: 路由路径
// 参数 handler: 处理函数
func (s *Server) Match(methods []string, path string, handler HandlerFunc) {
	s.engine.Match(methods, path, wrapHandler(handler))
}

// Handle 通用路由注册方法
// 支持注册任意HTTP方法的路由
// 参数 httpMethod: HTTP方法名称，如"GET", "POST"
// 参数 path: 路由路径
// 参数 handler: 处理函数
func (s *Server) Handle(httpMethod, path string, handler HandlerFunc) {
	s.engine.Handle(httpMethod, path, wrapHandler(handler))
}

// Group 创建路由组
// 创建一个新的路由组，支持路径前缀和独立的中间件管理
// 便于组织相关的路由和应用特定的中间件
// 参数 prefix: 路由组的路径前缀，如"/api/v1"
// 参数 middleware: 可选的中间件列表，仅应用于该路由组
// 返回值: *RouterGroup 新的路由组实例
func (s *Server) Group(prefix string, middleware ...MiddlewareFunc) *RouterGroup {
	group := s.engine.Group(prefix)
	for _, m := range middleware {
		group.Use(wrapMiddleware(m))
	}
	return &RouterGroup{
		group: group,
	}
}

// =============================================================================
// 静态文件服务
// =============================================================================

// Static 注册静态文件目录服务
// 将指定目录下的文件作为静态资源提供访问
// 参数 relativePath: URL路径前缀，如"/static"
// 参数 root: 本地文件系统路径，如"./public"
func (s *Server) Static(relativePath, root string) {
	s.engine.Static(relativePath, root)
}

// StaticFile 注册单个静态文件
// 将单个文件映射到指定的URL路径
// 参数 relativePath: URL路径，如"/favicon.ico"
// 参数 filepath: 本地文件路径，如"./assets/favicon.ico"
func (s *Server) StaticFile(relativePath, filepath string) {
	s.engine.StaticFile(relativePath, filepath)
}

// StaticFS 使用http.FileSystem注册静态文件服务
// 支持自定义文件系统实现，如嵌入式文件系统
// 参数 relativePath: URL路径前缀
// 参数 fs: http.FileSystem接口实现
func (s *Server) StaticFS(relativePath string, fs http.FileSystem) {
	s.engine.StaticFS(relativePath, fs)
}

// =============================================================================
// 模板渲染
// =============================================================================

// LoadHTMLGlob 使用通配符模式加载HTML模板
// 支持使用glob模式匹配多个模板文件
// 参数 pattern: 文件匹配模式，如"templates/*"
func (s *Server) LoadHTMLGlob(pattern string) {
	s.engine.LoadHTMLGlob(pattern)
}

// LoadHTMLFiles 加载指定的HTML模板文件
// 逐个指定需要加载的模板文件
// 参数 files: 模板文件路径列表
func (s *Server) LoadHTMLFiles(files ...string) {
	s.engine.LoadHTMLFiles(files...)
}

// SetFuncMap 设置模板函数映射
// 注册自定义函数供模板使用
// 参数 funcMap: 函数名到函数实现的映射
func (s *Server) SetFuncMap(funcMap map[string]interface{}) {
	s.engine.SetFuncMap(funcMap)
}

// =============================================================================
// 服务器启动方法
// =============================================================================

// Run 启动HTTP服务器
// 在指定地址启动HTTP服务，默认为":8080"
// 参数 addr: 可选的监听地址，如":8080", "localhost:3000"
// 返回值: error 启动过程中的错误信息
func (s *Server) Run(addr ...string) error {
	return s.engine.Run(addr...)
}

// RunTLS 启动HTTPS服务器
// 使用TLS证书启动安全的HTTPS服务
// 参数 addr: 监听地址
// 参数 certFile: TLS证书文件路径
// 参数 keyFile: TLS私钥文件路径
// 返回值: error 启动过程中的错误信息
func (s *Server) RunTLS(addr, certFile, keyFile string) error {
	return s.engine.RunTLS(addr, certFile, keyFile)
}

// RunUnix 启动Unix socket服务器
// 在Unix域套接字上启动服务，适用于本地进程间通信
// 参数 file: Unix socket文件路径
// 返回值: error 启动过程中的错误信息
func (s *Server) RunUnix(file string) error {
	return s.engine.RunUnix(file)
}

// RunFd 在指定文件描述符上启动服务器
// 使用已存在的文件描述符启动服务，适用于特殊部署场景
// 参数 fd: 文件描述符
// 返回值: error 启动过程中的错误信息
func (s *Server) RunFd(fd int) error {
	return s.engine.RunFd(fd)
}

// =============================================================================
// 工具和辅助方法
// =============================================================================

// ServeHTTP 实现http.Handler接口
// 使Server可以作为标准的HTTP处理器使用
// 参数 w: HTTP响应写入器
// 参数 req: HTTP请求对象
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.engine.ServeHTTP(w, req)
}

// Engine 获取底层Gin引擎实例
// 提供对原生Gin引擎的完全访问权限，用于高级定制
// 返回值: *gin.Engine Gin引擎实例
func (s *Server) Engine() *gin.Engine {
	return s.engine
}

// Routes 获取所有已注册的路由信息
// 返回路由列表，包含方法、路径、处理器等信息
// 返回值: gin.RoutesInfo 路由信息列表
func (s *Server) Routes() gin.RoutesInfo {
	return s.engine.Routes()
}

// =============================================================================
// 优雅关机方法
// =============================================================================

// Shutdown 优雅关闭服务器
// 监听系统信号，在接收到关闭信号时优雅地关闭服务器
// 确保正在处理的请求能够完成，新请求被拒绝
// 参数 timeout: 关闭超时时间，超过此时间将强制关闭
// 返回值: error 关闭过程中的错误信息
func (s *Server) Shutdown(timeout time.Duration) error {
	// 监听系统信号
	signal.Notify(s.quit, syscall.SIGINT, syscall.SIGTERM)
	
	// 等待信号
	<-s.quit
	
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	// 优雅关闭服务器
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// RunWithGracefulShutdown 启动服务器并支持优雅关机
// 在独立的goroutine中启动HTTP服务器，同时监听关闭信号
// 当接收到关闭信号时，会优雅地关闭服务器
// 参数 addr: 监听地址，如":8080"
// 参数 timeout: 关闭超时时间，默认30秒
// 返回值: error 启动或关闭过程中的错误信息
func (s *Server) RunWithGracefulShutdown(addr string, timeout ...time.Duration) error {
	// 设置默认超时时间
	shutdownTimeout := 30 * time.Second
	if len(timeout) > 0 {
		shutdownTimeout = timeout[0]
	}
	
	// 创建HTTP服务器实例
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.engine,
	}
	
	// 在goroutine中启动服务器
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// 服务器启动失败，发送信号通知主goroutine
			s.quit <- syscall.SIGTERM
		}
	}()
	
	// 优雅关机
	return s.Shutdown(shutdownTimeout)
}

// RunTLSWithGracefulShutdown 启动HTTPS服务器并支持优雅关机
// 使用TLS证书启动安全的HTTPS服务，同时支持优雅关机
// 参数 addr: 监听地址
// 参数 certFile: TLS证书文件路径
// 参数 keyFile: TLS私钥文件路径
// 参数 timeout: 关闭超时时间，默认30秒
// 返回值: error 启动或关闭过程中的错误信息
func (s *Server) RunTLSWithGracefulShutdown(addr, certFile, keyFile string, timeout ...time.Duration) error {
	// 设置默认超时时间
	shutdownTimeout := 30 * time.Second
	if len(timeout) > 0 {
		shutdownTimeout = timeout[0]
	}
	
	// 创建HTTP服务器实例
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.engine,
	}
	
	// 在goroutine中启动HTTPS服务器
	go func() {
		if err := s.server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			// 服务器启动失败，发送信号通知主goroutine
			s.quit <- syscall.SIGTERM
		}
	}()
	
	// 优雅关机
	return s.Shutdown(shutdownTimeout)
}

// Stop 立即停止服务器
// 强制关闭服务器，不等待正在处理的请求完成
// 返回值: error 停止过程中的错误信息
func (s *Server) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}
