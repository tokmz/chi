# Chi Web Framework

一个基于 Gin 框架的轻量级 Go Web 框架封装，提供更简洁友好的 API 接口和完整的 Web 开发功能。

## 📋 目录

- [项目概述](#项目概述)
- [核心特性](#核心特性)
- [快速开始](#快速开始)
- [安装指南](#安装指南)
- [基础使用](#基础使用)
- [高级功能](#高级功能)
- [API 参考](#api-参考)
- [示例代码](#示例代码)
- [最佳实践](#最佳实践)
- [贡献指南](#贡献指南)
- [许可证](#许可证)

## 🚀 项目概述

Chi 是一个基于 [Gin](https://github.com/gin-gonic/gin) 框架的 Go Web 框架封装，旨在提供更加简洁、易用的 API 接口。它保留了 Gin 的高性能特性，同时提供了更友好的开发体验和完整的 Web 开发功能。

### 设计理念

- **简洁性**: 提供简洁直观的 API 接口
- **高性能**: 基于 Gin 框架，保持高性能特性
- **易用性**: 封装常用功能，减少样板代码
- **扩展性**: 支持中间件和插件扩展
- **生产就绪**: 内置优雅关机、错误处理等生产环境必需功能

## ✨ 核心特性

### 🎯 核心功能
- **HTTP 路由**: 支持 GET、POST、PUT、DELETE、PATCH、OPTIONS、HEAD 等所有 HTTP 方法
- **路由分组**: 支持路由分组和嵌套分组，便于 API 版本管理
- **中间件系统**: 完整的中间件支持，包括全局和路由级中间件
- **参数绑定**: 支持 JSON、XML、YAML、Query、Form 等多种数据绑定方式
- **响应处理**: 统一的响应格式和多种响应类型支持

### 🛠️ 高级特性
- **静态文件服务**: 支持静态文件和文件系统服务
- **模板渲染**: 支持 HTML 模板渲染和自定义函数
- **文件上传**: 完整的文件上传和处理功能
- **优雅关机**: 内置优雅关机机制，确保服务平滑停止
- **错误处理**: 统一的错误处理和响应机制
- **安全配置**: 支持可信代理、CORS 等安全配置

### 🔧 开发工具
- **上下文封装**: 封装 Gin Context，提供更友好的 API
- **类型安全**: 完整的类型定义和接口约束
- **配置管理**: 灵活的配置管理系统
- **测试支持**: 便于单元测试和集成测试

## 🚀 快速开始

### 最简示例

```go
package main

import "chi"

func main() {
    // 创建服务器实例
    server := chi.New()
    
    // 注册路由
    server.GET("/hello", func(c *chi.Context) {
        c.JSON(200, map[string]string{
            "message": "Hello, Chi!",
        })
    })
    
    // 启动服务器
    server.Run(":8080")
}
```

### 带中间件的示例

```go
package main

import (
    "log"
    "chi"
)

func main() {
    server := chi.New()
    
    // 添加全局中间件
    server.Use(func(c *chi.Context) {
        log.Printf("Request: %s %s", c.Request().Method, c.Request().URL.Path)
        c.Next()
    })
    
    // API 路由组
    api := server.Group("/api/v1")
    {
        api.GET("/users", getUsersHandler)
        api.POST("/users", createUserHandler)
        api.GET("/users/:id", getUserHandler)
    }
    
    // 优雅启动
    server.RunWithGracefulShutdown(":8080")
}

func getUsersHandler(c *chi.Context) {
    c.JSON(200, []string{"user1", "user2"})
}

func createUserHandler(c *chi.Context) {
    var user struct {
        Name string `json:"name" binding:"required"`
        Age  int    `json:"age" binding:"required,min=1"`
    }
    
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(400, map[string]string{"error": err.Error()})
        return
    }
    
    c.JSON(201, user)
}

func getUserHandler(c *chi.Context) {
    id := c.Param("id")
    c.JSON(200, map[string]string{"id": id})
}
```

## 📦 安装指南

### 环境要求

- Go 1.24 或更高版本
- Git（用于获取依赖）

### 安装步骤

1. **初始化 Go 模块**
   ```bash
   mkdir my-chi-app
   cd my-chi-app
   go mod init my-chi-app
   ```

2. **添加 Chi 依赖**
   ```bash
   # 如果 Chi 已发布到公共仓库
   go get github.com/your-org/chi
   
   # 或者使用本地路径（开发阶段）
   go mod edit -replace chi=/path/to/chi
   ```

3. **创建主程序**
   ```go
   // main.go
   package main
   
   import "chi"
   
   func main() {
       server := chi.New()
       server.GET("/", func(c *chi.Context) {
           c.String(200, "Hello, Chi!")
       })
       server.Run(":8080")
   }
   ```

4. **运行应用**
   ```bash
   go run main.go
   ```

### Docker 部署

```dockerfile
# Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```

## 📖 基础使用

### 路由注册

```go
server := chi.New()

// HTTP 方法路由
server.GET("/users", getUsersHandler)
server.POST("/users", createUserHandler)
server.PUT("/users/:id", updateUserHandler)
server.DELETE("/users/:id", deleteUserHandler)
server.PATCH("/users/:id", patchUserHandler)
server.OPTIONS("/users", optionsHandler)
server.HEAD("/users", headHandler)

// 匹配所有方法
server.Any("/ping", pingHandler)

// 自定义方法匹配
server.Match([]string{"GET", "POST"}, "/custom", customHandler)

// 通用处理器
server.Handle("CUSTOM", "/method", customMethodHandler)
```

### 路由参数

```go
// 路径参数
server.GET("/users/:id", func(c *chi.Context) {
    id := c.Param("id")
    c.JSON(200, map[string]string{"user_id": id})
})

// 查询参数
server.GET("/search", func(c *chi.Context) {
    query := c.Query("q")
    page := c.DefaultQuery("page", "1")
    
    c.JSON(200, map[string]string{
        "query": query,
        "page":  page,
    })
})

// 表单参数
server.POST("/form", func(c *chi.Context) {
    name := c.PostForm("name")
    email := c.DefaultPostForm("email", "unknown@example.com")
    
    c.JSON(200, map[string]string{
        "name":  name,
        "email": email,
    })
})
```

### 数据绑定

```go
type User struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
    Age   int    `json:"age" binding:"required,min=1,max=120"`
}

server.POST("/users", func(c *chi.Context) {
    var user User
    
    // JSON 绑定
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(400, map[string]string{"error": err.Error()})
        return
    }
    
    // 处理用户数据
    c.JSON(201, user)
})

// 其他绑定方式
server.POST("/xml", func(c *chi.Context) {
    var data interface{}
    c.ShouldBindXML(&data)  // XML 绑定
})

server.GET("/query", func(c *chi.Context) {
    var params struct {
        Page int `form:"page"`
        Size int `form:"size"`
    }
    c.ShouldBindQuery(&params)  // 查询参数绑定
})
```

### 响应处理

```go
server.GET("/json", func(c *chi.Context) {
    // JSON 响应
    c.JSON(200, map[string]interface{}{
        "message": "success",
        "data":    []int{1, 2, 3},
    })
})

server.GET("/xml", func(c *chi.Context) {
    // XML 响应
    c.XML(200, map[string]string{"message": "success"})
})

server.GET("/yaml", func(c *chi.Context) {
    // YAML 响应
    c.YAML(200, map[string]string{"message": "success"})
})

server.GET("/string", func(c *chi.Context) {
    // 字符串响应
    c.String(200, "Hello, %s!", "World")
})

server.GET("/html", func(c *chi.Context) {
    // HTML 响应
    c.HTML(200, "index.html", map[string]interface{}{
        "title": "Chi Framework",
    })
})

server.GET("/file", func(c *chi.Context) {
    // 文件响应
    c.File("./static/download.pdf")
})

server.GET("/redirect", func(c *chi.Context) {
    // 重定向
    c.Redirect(302, "https://example.com")
})
```

## 🔧 高级功能

### 中间件系统

```go
// 全局中间件
server.Use(func(c *chi.Context) {
    start := time.Now()
    c.Next()
    duration := time.Since(start)
    log.Printf("Request processed in %v", duration)
})

// 认证中间件
func AuthMiddleware() chi.MiddlewareFunc {
    return func(c *chi.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(401, map[string]string{
                "error": "Authorization header required",
            })
            return
        }
        
        // 验证 token 逻辑
        if !validateToken(token) {
            c.AbortWithStatusJSON(401, map[string]string{
                "error": "Invalid token",
            })
            return
        }
        
        c.Set("user_id", getUserIDFromToken(token))
        c.Next()
    }
}

// 应用中间件到路由组
api := server.Group("/api", AuthMiddleware())
```

### 路由分组

```go
server := chi.New()

// API v1 路由组
v1 := server.Group("/api/v1")
{
    // 用户相关路由
    users := v1.Group("/users")
    {
        users.GET("", getUsersHandler)
        users.POST("", createUserHandler)
        users.GET("/:id", getUserHandler)
        users.PUT("/:id", updateUserHandler)
        users.DELETE("/:id", deleteUserHandler)
    }
    
    // 订单相关路由
    orders := v1.Group("/orders", AuthMiddleware())
    {
        orders.GET("", getOrdersHandler)
        orders.POST("", createOrderHandler)
        orders.GET("/:id", getOrderHandler)
    }
}

// API v2 路由组
v2 := server.Group("/api/v2")
{
    v2.GET("/users", getUsersV2Handler)
    v2.POST("/users", createUserV2Handler)
}

// 管理后台路由组
admin := server.Group("/admin", AdminAuthMiddleware())
{
    admin.GET("/dashboard", dashboardHandler)
    admin.GET("/users", adminUsersHandler)
    admin.POST("/users/:id/ban", banUserHandler)
}
```

### 静态文件服务

```go
// 静态文件目录
server.Static("/static", "./static")
server.Static("/assets", "./public/assets")

// 单个静态文件
server.StaticFile("/favicon.ico", "./static/favicon.ico")

// 使用自定义文件系统
server.StaticFS("/files", http.Dir("./uploads"))

// 路由组中的静态文件
api := server.Group("/api")
api.Static("/docs", "./docs")
```

### 模板渲染

```go
// 加载模板
server.LoadHTMLGlob("templates/*")
// 或加载指定文件
server.LoadHTMLFiles("templates/index.html", "templates/user.html")

// 设置模板函数
server.SetFuncMap(map[string]interface{}{
    "formatDate": func(t time.Time) string {
        return t.Format("2006-01-02")
    },
    "upper": strings.ToUpper,
})

// 渲染模板
server.GET("/", func(c *chi.Context) {
    c.HTML(200, "index.html", map[string]interface{}{
        "title": "Chi Framework",
        "users": []string{"Alice", "Bob", "Charlie"},
        "now":   time.Now(),
    })
})
```

### 文件上传

```go
server.POST("/upload", func(c *chi.Context) {
    // 单文件上传
    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(400, map[string]string{"error": err.Error()})
        return
    }
    
    // 保存文件
    dst := "./uploads/" + file.Filename
    if err := c.SaveUploadedFile(file, dst); err != nil {
        c.JSON(500, map[string]string{"error": err.Error()})
        return
    }
    
    c.JSON(200, map[string]string{
        "message": "File uploaded successfully",
        "file":    file.Filename,
    })
})

server.POST("/upload-multiple", func(c *chi.Context) {
    // 多文件上传
    form, err := c.MultipartForm()
    if err != nil {
        c.JSON(400, map[string]string{"error": err.Error()})
        return
    }
    
    files := form.File["files"]
    var uploadedFiles []string
    
    for _, file := range files {
        dst := "./uploads/" + file.Filename
        if err := c.SaveUploadedFile(file, dst); err != nil {
            c.JSON(500, map[string]string{"error": err.Error()})
            return
        }
        uploadedFiles = append(uploadedFiles, file.Filename)
    }
    
    c.JSON(200, map[string]interface{}{
        "message": "Files uploaded successfully",
        "files":   uploadedFiles,
    })
})
```

### 优雅关机

```go
package main

import (
    "log"
    "time"
    "chi"
)

func main() {
    server := chi.New()
    
    server.GET("/", func(c *chi.Context) {
        c.String(200, "Hello, Chi!")
    })
    
    // 优雅关机（默认30秒超时）
    if err := server.RunWithGracefulShutdown(":8080"); err != nil {
        log.Fatal("Server failed to start:", err)
    }
    
    // 或者自定义超时时间
    // server.RunWithGracefulShutdown(":8080", 60*time.Second)
    
    // HTTPS 优雅关机
    // server.RunTLSWithGracefulShutdown(":8443", "cert.pem", "key.pem")
}
```

### 错误处理

```go
// 自定义错误处理
server.NoRoute(func(c *chi.Context) {
    c.JSON(404, map[string]string{
        "error": "Route not found",
        "path":  c.Request().URL.Path,
    })
})

server.NoMethod(func(c *chi.Context) {
    c.JSON(405, map[string]string{
        "error":  "Method not allowed",
        "method": c.Request().Method,
        "path":   c.Request().URL.Path,
    })
})

// 使用内置错误类型
server.GET("/error", func(c *chi.Context) {
    // 使用预定义错误
    chi.FailRes(c, chi.ErrBinding)
    
    // 或自定义错误
    err := chi.NewError(400, "自定义错误信息")
    chi.FailRes(c, err)
})

// 统一响应格式
server.GET("/success", func(c *chi.Context) {
    data := map[string]string{"message": "success"}
    chi.SuccessRes(c, data)
})
```

## 📚 API 参考

### Server 类型

#### 构造函数

```go
// New 创建新的Server实例
func New() *Server
```

#### 配置方法

```go
// SetMode 设置运行模式 (gin.DebugMode, gin.ReleaseMode, gin.TestMode)
func (s *Server) SetMode(mode string)

// SetTrustedProxies 设置可信代理
func (s *Server) SetTrustedProxies(trustedProxies []string) error

// RemoteIPHeaders 设置远程IP头
func (s *Server) RemoteIPHeaders(headers ...string)

// ForwardedByClientIP 设置是否通过客户端IP转发
func (s *Server) ForwardedByClientIP(value bool)

// UseRawPath 设置是否使用原始路径
func (s *Server) UseRawPath(value bool)

// UnescapePathValues 设置是否取消转义路径值
func (s *Server) UnescapePathValues(value bool)

// MaxMultipartMemory 设置最大多部分内存
func (s *Server) MaxMultipartMemory(value int64)

// HandleMethodNotAllowed 设置是否处理方法不允许
func (s *Server) HandleMethodNotAllowed(value bool)

// RedirectTrailingSlash 设置是否重定向尾部斜杠
func (s *Server) RedirectTrailingSlash(value bool)

// RedirectFixedPath 设置是否重定向固定路径
func (s *Server) RedirectFixedPath(value bool)
```

#### 中间件方法

```go
// Use 添加全局中间件
func (s *Server) Use(middleware ...MiddlewareFunc)
```

#### 路由注册方法

```go
// HTTP 方法路由
func (s *Server) GET(path string, handler HandlerFunc)
func (s *Server) POST(path string, handler HandlerFunc)
func (s *Server) PUT(path string, handler HandlerFunc)
func (s *Server) DELETE(path string, handler HandlerFunc)
func (s *Server) PATCH(path string, handler HandlerFunc)
func (s *Server) OPTIONS(path string, handler HandlerFunc)
func (s *Server) HEAD(path string, handler HandlerFunc)

// 特殊路由
func (s *Server) Any(path string, handler HandlerFunc)
func (s *Server) Match(methods []string, path string, handler HandlerFunc)
func (s *Server) Handle(httpMethod, path string, handler HandlerFunc)

// 错误处理路由
func (s *Server) NoRoute(handler HandlerFunc)
func (s *Server) NoMethod(handler HandlerFunc)
```

#### 路由分组方法

```go
// Group 创建路由组
func (s *Server) Group(prefix string, middleware ...MiddlewareFunc) *RouterGroup
```

#### 静态文件方法

```go
// Static 静态文件目录服务
func (s *Server) Static(relativePath, root string)

// StaticFile 单个静态文件服务
func (s *Server) StaticFile(relativePath, filepath string)

// StaticFS 文件系统服务
func (s *Server) StaticFS(relativePath string, fs http.FileSystem)
```

#### 模板方法

```go
// LoadHTMLGlob 加载HTML模板（通配符）
func (s *Server) LoadHTMLGlob(pattern string)

// LoadHTMLFiles 加载HTML模板（指定文件）
func (s *Server) LoadHTMLFiles(files ...string)

// SetFuncMap 设置模板函数
func (s *Server) SetFuncMap(funcMap map[string]interface{})
```

#### 服务器启动方法

```go
// Run 启动HTTP服务器
func (s *Server) Run(addr ...string) error

// RunTLS 启动HTTPS服务器
func (s *Server) RunTLS(addr, certFile, keyFile string) error

// RunUnix 启动Unix套接字服务器
func (s *Server) RunUnix(file string) error

// RunFd 启动文件描述符服务器
func (s *Server) RunFd(fd int) error

// RunWithGracefulShutdown 启动服务器并支持优雅关机
func (s *Server) RunWithGracefulShutdown(addr string, timeout ...time.Duration) error

// RunTLSWithGracefulShutdown 启动HTTPS服务器并支持优雅关机
func (s *Server) RunTLSWithGracefulShutdown(addr, certFile, keyFile string, timeout ...time.Duration) error
```

#### 优雅关机方法

```go
// Shutdown 优雅关闭服务器
func (s *Server) Shutdown(timeout time.Duration) error

// Stop 停止服务器
func (s *Server) Stop() error
```

#### 工具方法

```go
// ServeHTTP 实现http.Handler接口
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request)

// Engine 获取底层Gin引擎
func (s *Server) Engine() *gin.Engine

// Routes 获取路由信息
func (s *Server) Routes() gin.RoutesInfo
```

### Context 类型

#### 中间件控制

```go
// Next 调用下一个中间件
func (c *Context) Next()

// IsAborted 检查是否已中止
func (c *Context) IsAborted() bool

// Abort 中止请求处理
func (c *Context) Abort()

// AbortWithStatus 中止并设置状态码
func (c *Context) AbortWithStatus(code int)

// AbortWithStatusJSON 中止并返回JSON
func (c *Context) AbortWithStatusJSON(code int, jsonObj interface{})

// AbortWithError 中止并设置错误
func (c *Context) AbortWithError(code int, err error) *gin.Error
```

#### 数据存储与获取

```go
// Set 设置键值对
func (c *Context) Set(key string, value interface{})

// Get 获取值
func (c *Context) Get(key string) (value interface{}, exists bool)

// MustGet 获取值（必须存在）
func (c *Context) MustGet(key string) interface{}

// 类型安全的获取方法
func (c *Context) GetString(key string) string
func (c *Context) GetBool(key string) bool
func (c *Context) GetInt(key string) int
func (c *Context) GetInt64(key string) int64
func (c *Context) GetUint(key string) uint
func (c *Context) GetUint64(key string) uint64
func (c *Context) GetFloat64(key string) float64
func (c *Context) GetTime(key string) time.Time
func (c *Context) GetDuration(key string) time.Duration
func (c *Context) GetStringSlice(key string) []string
func (c *Context) GetStringMap(key string) map[string]interface{}
func (c *Context) GetStringMapString(key string) map[string]string
func (c *Context) GetStringMapStringSlice(key string) map[string][]string
```

#### 请求参数获取

```go
// 路径参数
func (c *Context) Param(key string) string

// 查询参数
func (c *Context) Query(key string) string
func (c *Context) DefaultQuery(key, defaultValue string) string
func (c *Context) GetQuery(key string) (string, bool)
func (c *Context) QueryArray(key string) []string
func (c *Context) GetQueryArray(key string) ([]string, bool)
func (c *Context) QueryMap(key string) map[string]string
func (c *Context) GetQueryMap(key string) (map[string]string, bool)

// 表单参数
func (c *Context) PostForm(key string) string
func (c *Context) DefaultPostForm(key, defaultValue string) string
func (c *Context) GetPostForm(key string) (string, bool)
func (c *Context) PostFormArray(key string) []string
func (c *Context) GetPostFormArray(key string) ([]string, bool)
func (c *Context) PostFormMap(key string) map[string]string
func (c *Context) GetPostFormMap(key string) (map[string]string, bool)
```

#### 文件上传

```go
// FormFile 获取单个上传文件
func (c *Context) FormFile(name string) (*multipart.FileHeader, error)

// MultipartForm 获取多部分表单
func (c *Context) MultipartForm() (*multipart.Form, error)

// SaveUploadedFile 保存上传文件
func (c *Context) SaveUploadedFile(file *multipart.FileHeader, dst string) error
```

#### 数据绑定

```go
// 自动绑定（根据Content-Type）
func (c *Context) Bind(obj interface{}) error
func (c *Context) ShouldBind(obj interface{}) error

// 指定格式绑定
func (c *Context) ShouldBindJSON(obj interface{}) error
func (c *Context) ShouldBindXML(obj interface{}) error
func (c *Context) ShouldBindYAML(obj interface{}) error
func (c *Context) ShouldBindTOML(obj interface{}) error
func (c *Context) ShouldBindQuery(obj interface{}) error
func (c *Context) ShouldBindUri(obj interface{}) error
func (c *Context) ShouldBindHeader(obj interface{}) error
func (c *Context) ShouldBindWith(obj interface{}, b binding.Binding) error

// 强制绑定（失败时中止）
func (c *Context) BindJSON(obj interface{}) error
func (c *Context) BindXML(obj interface{}) error
func (c *Context) BindYAML(obj interface{}) error
func (c *Context) BindTOML(obj interface{}) error
func (c *Context) BindQuery(obj interface{}) error
func (c *Context) BindUri(obj interface{}) error
func (c *Context) BindHeader(obj interface{}) error
func (c *Context) BindWith(obj interface{}, b binding.Binding) error
```

#### 请求信息

```go
// ClientIP 获取客户端IP
func (c *Context) ClientIP() string

// ContentType 获取内容类型
func (c *Context) ContentType() string

// IsWebsocket 检查是否为WebSocket
func (c *Context) IsWebsocket() bool

// GetHeader 获取请求头
func (c *Context) GetHeader(key string) string

// GetRawData 获取原始请求数据
func (c *Context) GetRawData() ([]byte, error)

// Request 获取HTTP请求
func (c *Context) Request() *http.Request
```

#### 响应设置

```go
// Status 设置状态码
func (c *Context) Status(code int)

// Header 设置响应头
func (c *Context) Header(key, value string)

// Writer 获取响应写入器
func (c *Context) Writer() gin.ResponseWriter
```

#### Cookie 操作

```go
// Cookie 获取Cookie
func (c *Context) Cookie(name string) (string, error)

// SetCookie 设置Cookie
func (c *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool)

// SetSameSite 设置SameSite属性
func (c *Context) SetSameSite(samesite http.SameSite)
```

#### 响应输出

```go
// JSON 响应
func (c *Context) JSON(code int, obj interface{})
func (c *Context) IndentedJSON(code int, obj interface{})
func (c *Context) SecureJSON(code int, obj interface{})
func (c *Context) PureJSON(code int, obj interface{})
func (c *Context) AsciiJSON(code int, obj interface{})
func (c *Context) JSONP(code int, obj interface{})

// 其他格式响应
func (c *Context) XML(code int, obj interface{})
func (c *Context) YAML(code int, obj interface{})
func (c *Context) TOML(code int, obj interface{})
func (c *Context) ProtoBuf(code int, obj interface{})

// 文本响应
func (c *Context) String(code int, format string, values ...interface{})

// HTML 响应
func (c *Context) HTML(code int, name string, obj interface{})

// 数据响应
func (c *Context) Data(code int, contentType string, data []byte)
func (c *Context) DataFromReader(code int, contentLength int64, contentType string, reader io.Reader, extraHeaders map[string]string)

// 重定向
func (c *Context) Redirect(code int, location string)

// 文件响应
func (c *Context) File(filepath string)
func (c *Context) FileFromFS(filepath string, fs http.FileSystem)
func (c *Context) FileAttachment(filepath, filename string)

// 流式响应
func (c *Context) Stream(step func(w io.Writer) bool) bool
func (c *Context) SSEvent(name, message string)
```

### RouterGroup 类型

```go
// Group 创建子路由组
func (rg *RouterGroup) Group(relativePath string, middleware ...MiddlewareFunc) *RouterGroup

// Use 添加中间件
func (rg *RouterGroup) Use(middleware ...MiddlewareFunc)

// HTTP 方法路由
func (rg *RouterGroup) GET(relativePath string, handler HandlerFunc)
func (rg *RouterGroup) POST(relativePath string, handler HandlerFunc)
func (rg *RouterGroup) PUT(relativePath string, handler HandlerFunc)
func (rg *RouterGroup) DELETE(relativePath string, handler HandlerFunc)
func (rg *RouterGroup) PATCH(relativePath string, handler HandlerFunc)
func (rg *RouterGroup) OPTIONS(relativePath string, handler HandlerFunc)
func (rg *RouterGroup) HEAD(relativePath string, handler HandlerFunc)
func (rg *RouterGroup) Any(relativePath string, handler HandlerFunc)
func (rg *RouterGroup) Handle(httpMethod, relativePath string, handler HandlerFunc)

// 静态文件
func (rg *RouterGroup) Static(relativePath, root string)
func (rg *RouterGroup) StaticFile(relativePath, filepath string)
func (rg *RouterGroup) StaticFS(relativePath string, fs http.FileSystem)
func (rg *RouterGroup) StaticFileFS(relativePath, filepath string, fs http.FileSystem)

// BasePath 获取基础路径
func (rg *RouterGroup) BasePath() string
```

### 响应类型

```go
// Response 统一响应结构
type Response struct {
    Code    int    `json:"code"`
    Data    any    `json:"data"`
    Message string `json:"message"`
}

// PageResp 分页响应结构
type PageResp[T any] struct {
    Total int64 `json:"total"`
    List  T     `json:"list"`
}

// 响应构造函数
func NewResponse(code int, data any, message string) *Response
func NewErrResponse(code int, message string) *Response
func NewOkResponse(data any) *Response
func NewPageResp[T any](total int64, list T) *PageResp[T]

// 响应辅助函数
func Res(ctx *Context, err error, data ...any)
func SuccessRes(ctx *Context, data any)
func FailRes(ctx *Context, err error)
```

### 错误类型

```go
// Error 错误结构
type Error struct {
    Code    int
    Message string
}

// NewError 创建错误
func NewError(code int, message string) *Error

// 预定义错误
var (
    ErrServer  = NewError(http.StatusInternalServerError, "服务异常")
    ErrBinding = NewError(http.StatusBadRequest, "参数错误")
)
```

## 💡 示例代码

### 完整的 RESTful API 示例

```go
package main

import (
    "fmt"
    "log"
    "strconv"
    "time"
    
    "chi"
)

// User 用户模型
type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name" binding:"required"`
    Email     string    `json:"email" binding:"required,email"`
    Age       int       `json:"age" binding:"required,min=1,max=120"`
    CreatedAt time.Time `json:"created_at"`
}

// 模拟数据库
var (
    users  = make(map[int]*User)
    nextID = 1
)

func main() {
    server := chi.New()
    
    // 全局中间件
    server.Use(LoggerMiddleware())
    server.Use(CORSMiddleware())
    
    // 静态文件
    server.Static("/static", "./static")
    
    // API 路由组
    api := server.Group("/api/v1")
    {
        // 用户相关路由
        users := api.Group("/users")
        {
            users.GET("", getUsersHandler)           // GET /api/v1/users
            users.POST("", createUserHandler)        // POST /api/v1/users
            users.GET("/:id", getUserHandler)        // GET /api/v1/users/:id
            users.PUT("/:id", updateUserHandler)     // PUT /api/v1/users/:id
            users.DELETE("/:id", deleteUserHandler)  // DELETE /api/v1/users/:id
        }
        
        // 健康检查
        api.GET("/health", healthHandler)
    }
    
    // 错误处理
    server.NoRoute(func(c *chi.Context) {
        c.JSON(404, map[string]string{
            "error": "Route not found",
            "path":  c.Request().URL.Path,
        })
    })
    
    // 启动服务器
    log.Println("Server starting on :8080")
    if err := server.RunWithGracefulShutdown(":8080"); err != nil {
        log.Fatal("Server failed to start:", err)
    }
}

// 中间件
func LoggerMiddleware() chi.MiddlewareFunc {
    return func(c *chi.Context) {
        start := time.Now()
        path := c.Request().URL.Path
        method := c.Request().Method
        
        c.Next()
        
        duration := time.Since(start)
        status := c.Writer().Status()
        
        log.Printf("%s %s %d %v", method, path, status, duration)
    }
}

func CORSMiddleware() chi.MiddlewareFunc {
    return func(c *chi.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if c.Request().Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        
        c.Next()
    }
}

// 处理器
func getUsersHandler(c *chi.Context) {
    // 查询参数
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
    
    // 模拟分页
    var userList []*User
    for _, user := range users {
        userList = append(userList, user)
    }
    
    start := (page - 1) * size
    end := start + size
    if start > len(userList) {
        start = len(userList)
    }
    if end > len(userList) {
        end = len(userList)
    }
    
    result := userList[start:end]
    
    c.JSON(200, map[string]interface{}{
        "code": 200,
        "data": map[string]interface{}{
            "total": len(users),
            "page":  page,
            "size":  size,
            "list":  result,
        },
        "message": "success",
    })
}

func createUserHandler(c *chi.Context) {
    var user User
    
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(400, map[string]string{
            "error": err.Error(),
        })
        return
    }
    
    // 检查邮箱是否已存在
    for _, existingUser := range users {
        if existingUser.Email == user.Email {
            c.JSON(400, map[string]string{
                "error": "Email already exists",
            })
            return
        }
    }
    
    // 创建用户
    user.ID = nextID
    user.CreatedAt = time.Now()
    users[nextID] = &user
    nextID++
    
    c.JSON(201, map[string]interface{}{
        "code":    201,
        "data":    user,
        "message": "User created successfully",
    })
}

func getUserHandler(c *chi.Context) {
    idStr := c.Param("id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        c.JSON(400, map[string]string{
            "error": "Invalid user ID",
        })
        return
    }
    
    user, exists := users[id]
    if !exists {
        c.JSON(404, map[string]string{
            "error": "User not found",
        })
        return
    }
    
    c.JSON(200, map[string]interface{}{
        "code":    200,
        "data":    user,
        "message": "success",
    })
}

func updateUserHandler(c *chi.Context) {
    idStr := c.Param("id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        c.JSON(400, map[string]string{
            "error": "Invalid user ID",
        })
        return
    }
    
    user, exists := users[id]
    if !exists {
        c.JSON(404, map[string]string{
            "error": "User not found",
        })
        return
    }
    
    var updateData User
    if err := c.ShouldBindJSON(&updateData); err != nil {
        c.JSON(400, map[string]string{
            "error": err.Error(),
        })
        return
    }
    
    // 检查邮箱是否被其他用户使用
    for uid, existingUser := range users {
        if uid != id && existingUser.Email == updateData.Email {
            c.JSON(400, map[string]string{
                "error": "Email already exists",
            })
            return
        }
    }
    
    // 更新用户信息
    user.Name = updateData.Name
    user.Email = updateData.Email
    user.Age = updateData.Age
    
    c.JSON(200, map[string]interface{}{
        "code":    200,
        "data":    user,
        "message": "User updated successfully",
    })
}

func deleteUserHandler(c *chi.Context) {
    idStr := c.Param("id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        c.JSON(400, map[string]string{
            "error": "Invalid user ID",
        })
        return
    }
    
    _, exists := users[id]
    if !exists {
        c.JSON(404, map[string]string{
            "error": "User not found",
        })
        return
    }
    
    delete(users, id)
    
    c.JSON(200, map[string]interface{}{
        "code":    200,
        "message": "User deleted successfully",
    })
}

func healthHandler(c *chi.Context) {
    c.JSON(200, map[string]interface{}{
        "status":    "ok",
        "timestamp": time.Now().Unix(),
        "version":   "1.0.0",
    })
}
```

### 文件上传示例

```go
package main

import (
    "fmt"
    "log"
    "os"
    "path/filepath"
    "time"
    
    "chi"
)

func main() {
    server := chi.New()
    
    // 创建上传目录
    os.MkdirAll("./uploads", 0755)
    
    // 设置最大上传大小 (32MB)
    server.MaxMultipartMemory(32 << 20)
    
    // 静态文件服务（用于访问上传的文件）
    server.Static("/uploads", "./uploads")
    
    // 上传路由
    server.POST("/upload", uploadHandler)
    server.POST("/upload-multiple", uploadMultipleHandler)
    
    // 上传页面
    server.GET("/", func(c *chi.Context) {
        c.HTML(200, "upload.html", nil)
    })
    
    server.LoadHTMLFiles("templates/upload.html")
    
    log.Println("Server starting on :8080")
    server.Run(":8080")
}

func uploadHandler(c *chi.Context) {
    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(400, map[string]string{
            "error": "No file uploaded",
        })
        return
    }
    
    // 生成唯一文件名
    ext := filepath.Ext(file.Filename)
    filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
    dst := filepath.Join("./uploads", filename)
    
    // 保存文件
    if err := c.SaveUploadedFile(file, dst); err != nil {
        c.JSON(500, map[string]string{
            "error": "Failed to save file",
        })
        return
    }
    
    c.JSON(200, map[string]interface{}{
        "message":  "File uploaded successfully",
        "filename": filename,
        "size":     file.Size,
        "url":      fmt.Sprintf("/uploads/%s", filename),
    })
}

func uploadMultipleHandler(c *chi.Context) {
    form, err := c.MultipartForm()
    if err != nil {
        c.JSON(400, map[string]string{
            "error": "Failed to parse multipart form",
        })
        return
    }
    
    files := form.File["files"]
    if len(files) == 0 {
        c.JSON(400, map[string]string{
            "error": "No files uploaded",
        })
        return
    }
    
    var uploadedFiles []map[string]interface{}
    
    for _, file := range files {
        // 生成唯一文件名
        ext := filepath.Ext(file.Filename)
        filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
        dst := filepath.Join("./uploads", filename)
        
        // 保存文件
        if err := c.SaveUploadedFile(file, dst); err != nil {
            c.JSON(500, map[string]string{
                "error": fmt.Sprintf("Failed to save file: %s", file.Filename),
            })
            return
        }
        
        uploadedFiles = append(uploadedFiles, map[string]interface{}{
            "original": file.Filename,
            "filename": filename,
            "size":     file.Size,
            "url":      fmt.Sprintf("/uploads/%s", filename),
        })
        
        // 避免文件名冲突
        time.Sleep(time.Millisecond)
    }
    
    c.JSON(200, map[string]interface{}{
        "message": "Files uploaded successfully",
        "files":   uploadedFiles,
        "count":   len(uploadedFiles),
    })
}
```

### 中间件示例

```go
package main

import (
    "log"
    "strings"
    "time"
    
    "chi"
)

func main() {
    server := chi.New()
    
    // 全局中间件
    server.Use(LoggerMiddleware())
    server.Use(RecoveryMiddleware())
    server.Use(CORSMiddleware())
    
    // 公开路由
    server.POST("/login", loginHandler)
    server.GET("/public", publicHandler)
    
    // 需要认证的路由组
    auth := server.Group("/api", AuthMiddleware())
    {
        auth.GET("/profile", profileHandler)
        auth.POST("/logout", logoutHandler)
        
        // 需要管理员权限的路由组
        admin := auth.Group("/admin", AdminMiddleware())
        {
            admin.GET("/users", adminUsersHandler)
            admin.DELETE("/users/:id", adminDeleteUserHandler)
        }
    }
    
    server.Run(":8080")
}

// 日志中间件
func LoggerMiddleware() chi.MiddlewareFunc {
    return func(c *chi.Context) {
        start := time.Now()
        path := c.Request().URL.Path
        method := c.Request().Method
        clientIP := c.ClientIP()
        
        c.Next()
        
        duration := time.Since(start)
        status := c.Writer().Status()
        
        log.Printf("%s %s %s %d %v", clientIP, method, path, status, duration)
    }
}

// 恢复中间件
func RecoveryMiddleware() chi.MiddlewareFunc {
    return func(c *chi.Context) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("Panic recovered: %v", err)
                c.JSON(500, map[string]string{
                    "error": "Internal server error",
                })
                c.Abort()
            }
        }()
        c.Next()
    }
}

// CORS 中间件
func CORSMiddleware() chi.MiddlewareFunc {
    return func(c *chi.Context) {
        origin := c.GetHeader("Origin")
        
        // 允许的域名列表
        allowedOrigins := []string{
            "http://localhost:3000",
            "https://example.com",
        }
        
        for _, allowed := range allowedOrigins {
            if origin == allowed {
                c.Header("Access-Control-Allow-Origin", origin)
                break
            }
        }
        
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
        c.Header("Access-Control-Allow-Credentials", "true")
        
        if c.Request().Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        
        c.Next()
    }
}

// 认证中间件
func AuthMiddleware() chi.MiddlewareFunc {
    return func(c *chi.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, map[string]string{
                "error": "Authorization header required",
            })
            c.Abort()
            return
        }
        
        // 检查 Bearer token
        if !strings.HasPrefix(authHeader, "Bearer ") {
            c.JSON(401, map[string]string{
                "error": "Invalid authorization header format",
            })
            c.Abort()
            return
        }
        
        token := strings.TrimPrefix(authHeader, "Bearer ")
        
        // 验证 token（这里简化处理）
        userID, role, err := validateToken(token)
        if err != nil {
            c.JSON(401, map[string]string{
                "error": "Invalid token",
            })
            c.Abort()
            return
        }
        
        // 将用户信息存储到上下文
        c.Set("user_id", userID)
        c.Set("user_role", role)
        
        c.Next()
    }
}

// 管理员中间件
func AdminMiddleware() chi.MiddlewareFunc {
    return func(c *chi.Context) {
        role, exists := c.Get("user_role")
        if !exists || role != "admin" {
            c.JSON(403, map[string]string{
                "error": "Admin access required",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// 模拟 token 验证
func validateToken(token string) (userID int, role string, err error) {
    // 这里应该实现真正的 JWT 验证逻辑
    if token == "valid-user-token" {
        return 1, "user", nil
    }
    if token == "valid-admin-token" {
        return 2, "admin", nil
    }
    return 0, "", fmt.Errorf("invalid token")
}

// 处理器示例
func loginHandler(c *chi.Context) {
    var loginData struct {
        Username string `json:"username" binding:"required"`
        Password string `json:"password" binding:"required"`
    }
    
   if err := c.ShouldBindJSON(&loginData); err != nil {
        c.JSON(400, map[string]string{"error": err.Error()})
        return
    }
    
    // 验证用户名密码（简化处理）
    if loginData.Username == "admin" && loginData.Password == "password" {
        c.JSON(200, map[string]string{
            "token": "valid-admin-token",
            "role":  "admin",
        })
    } else if loginData.Username == "user" && loginData.Password == "password" {
        c.JSON(200, map[string]string{
            "token": "valid-user-token",
            "role":  "user",
        })
    } else {
        c.JSON(401, map[string]string{
            "error": "Invalid credentials",
        })
    }
}

func publicHandler(c *chi.Context) {
    c.JSON(200, map[string]string{
        "message": "This is a public endpoint",
    })
}

func profileHandler(c *chi.Context) {
    userID := c.MustGet("user_id").(int)
    role := c.MustGet("user_role").(string)
    
    c.JSON(200, map[string]interface{}{
        "user_id": userID,
        "role":    role,
        "message": "Profile data",
    })
}

func logoutHandler(c *chi.Context) {
    c.JSON(200, map[string]string{
        "message": "Logged out successfully",
    })
}

func adminUsersHandler(c *chi.Context) {
    c.JSON(200, map[string]interface{}{
        "message": "Admin users list",
        "users":   []string{"user1", "user2", "user3"},
    })
}

func adminDeleteUserHandler(c *chi.Context) {
    userID := c.Param("id")
    c.JSON(200, map[string]string{
        "message": "User deleted",
        "user_id": userID,
    })
}
```

## 🎯 最佳实践

### 项目结构建议

```
my-chi-app/
├── main.go                 # 应用入口
├── config/
│   ├── config.go          # 配置管理
│   └── database.go        # 数据库配置
├── handlers/
│   ├── user.go            # 用户相关处理器
│   ├── auth.go            # 认证相关处理器
│   └── admin.go           # 管理相关处理器
├── middleware/
│   ├── auth.go            # 认证中间件
│   ├── cors.go            # CORS中间件
│   └── logger.go          # 日志中间件
├── models/
│   ├── user.go            # 用户模型
│   └── response.go        # 响应模型
├── services/
│   ├── user.go            # 用户服务
│   └── auth.go            # 认证服务
├── utils/
│   ├── jwt.go             # JWT工具
│   └── validator.go       # 验证工具
├── static/                # 静态文件
├── templates/             # 模板文件
├── uploads/               # 上传文件
├── go.mod
├── go.sum
└── README.md
```

### 错误处理最佳实践

```go
// 统一错误响应
func ErrorHandler() chi.MiddlewareFunc {
    return func(c *chi.Context) {
        c.Next()
        
        // 检查是否有错误
        if len(c.Errors()) > 0 {
            err := c.Errors().Last()
            
            // 根据错误类型返回不同响应
            switch e := err.Err.(type) {
            case *chi.Error:
                c.JSON(e.Code, map[string]string{
                    "error": e.Message,
                })
            default:
                c.JSON(500, map[string]string{
                    "error": "Internal server error",
                })
            }
        }
    }
}

// 在处理器中使用
func someHandler(c *chi.Context) {
    if someCondition {
        c.Error(chi.ErrBinding)
        return
    }
    
    // 正常处理逻辑
    c.JSON(200, data)
}
```

### 配置管理最佳实践

```go
// config/config.go
package config

import (
    "os"
    "strconv"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    JWT      JWTConfig
}

type ServerConfig struct {
    Host string
    Port int
    Mode string
}

type DatabaseConfig struct {
    Host     string
    Port     int
    Username string
    Password string
    Database string
}

type JWTConfig struct {
    Secret     string
    ExpireTime int
}

func Load() *Config {
    return &Config{
        Server: ServerConfig{
            Host: getEnv("SERVER_HOST", "localhost"),
            Port: getEnvInt("SERVER_PORT", 8080),
            Mode: getEnv("GIN_MODE", "debug"),
        },
        Database: DatabaseConfig{
            Host:     getEnv("DB_HOST", "localhost"),
            Port:     getEnvInt("DB_PORT", 5432),
            Username: getEnv("DB_USERNAME", "postgres"),
            Password: getEnv("DB_PASSWORD", ""),
            Database: getEnv("DB_DATABASE", "myapp"),
        },
        JWT: JWTConfig{
            Secret:     getEnv("JWT_SECRET", "your-secret-key"),
            ExpireTime: getEnvInt("JWT_EXPIRE_TIME", 3600),
        },
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}
```

### 测试最佳实践

```go
// handlers/user_test.go
package handlers

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "chi"
)

func TestCreateUser(t *testing.T) {
    server := chi.New()
    server.POST("/users", createUserHandler)
    
    user := map[string]interface{}{
        "name":  "Test User",
        "email": "test@example.com",
        "age":   25,
    }
    
    jsonData, _ := json.Marshal(user)
    req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    server.ServeHTTP(w, req)
    
    if w.Code != http.StatusCreated {
        t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
    }
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    
    if response["code"] != float64(201) {
        t.Errorf("Expected code 201, got %v", response["code"])
    }
}

func TestGetUser(t *testing.T) {
    server := chi.New()
    server.GET("/users/:id", getUserHandler)
    
    req := httptest.NewRequest("GET", "/users/1", nil)
    w := httptest.NewRecorder()
    
    server.ServeHTTP(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
    }
}
```

### 性能优化建议

1. **使用连接池**
   ```go
   // 数据库连接池配置
   db.SetMaxOpenConns(25)
   db.SetMaxIdleConns(5)
   db.SetConnMaxLifetime(5 * time.Minute)
   ```

2. **启用 Gzip 压缩**
   ```go
   import "github.com/gin-contrib/gzip"
   
   server.Use(gzip.Gzip(gzip.DefaultCompression))
   ```

3. **使用缓存**
   ```go
   import "github.com/gin-contrib/cache"
   
   server.GET("/api/data", cache.CachePage(
       store.NewInMemoryStore(time.Minute),
       time.Minute,
       dataHandler,
   ))
   ```

4. **限流**
   ```go
   import "github.com/gin-contrib/limiter"
   
   server.Use(limiter.Limit(
       limiter.Rate{Period: time.Minute, Limit: 100},
   ))
   ```

## 🤝 贡献指南

我们欢迎所有形式的贡献！请遵循以下步骤：

### 开发环境设置

1. **Fork 项目**
   ```bash
   git clone https://github.com/your-username/chi.git
   cd chi
   ```

2. **安装依赖**
   ```bash
   go mod download
   ```

3. **运行测试**
   ```bash
   go test ./...
   ```

### 提交规范

- 使用清晰的提交信息
- 遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范
- 示例：
  ```
  feat: 添加用户认证中间件
  fix: 修复路由参数解析问题
  docs: 更新API文档
  test: 添加用户服务测试
  ```

### 代码规范

- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 使用 `golint` 检查代码质量
- 添加必要的注释和文档

### Pull Request 流程

1. 创建功能分支
2. 实现功能并添加测试
3. 确保所有测试通过
4. 提交 Pull Request
5. 等待代码审查

## 📄 许可证

本项目采用 MIT 许可证。详情请参阅 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- [Gin](https://github.com/gin-gonic/gin) - 高性能的 Go Web 框架
- [Go](https://golang.org/) - 优秀的编程语言
- 所有贡献者和用户的支持

## 📞 联系我们

- 项目主页：[GitHub Repository](https://github.com/your-org/chi)
- 问题反馈：[Issues](https://github.com/your-org/chi/issues)
- 讨论交流：[Discussions](https://github.com/your-org/chi/discussions)

---

**Chi Web Framework** - 让 Go Web 开发更简单、更高效！