package chi

import (
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/gin-gonic/gin/render"
)

// Context 封装了Gin的Context，提供更友好的API接口
// 通过组合的方式继承gin.Context的所有功能，同时可以扩展自定义方法
type Context struct {
	*gin.Context
}

// =============================================================================
// 中间件控制方法
// =============================================================================

// Next 调用下一个中间件或处理函数
// 在中间件中使用，将控制权传递给下一个处理器
// 通常在中间件的前置处理完成后调用
func (c *Context) Next() {
	c.Context.Next()
}

// IsAborted 检查请求是否已中止
// 返回true表示请求处理已被中止，不会继续执行后续中间件
// 常用于在中间件中检查请求是否应该继续处理
func (c *Context) IsAborted() bool {
	return c.Context.IsAborted()
}

// Abort 中止请求处理
// 阻止调用挂起的处理程序，但不会停止当前处理程序
// 通常在认证失败或其他错误情况下使用
func (c *Context) Abort() {
	c.Context.Abort()
}

// AbortWithStatus 中止请求并设置HTTP状态码
// 设置状态码并中止请求处理，不会执行后续中间件
// 参数 code: HTTP状态码，如400、401、500等
func (c *Context) AbortWithStatus(code int) {
	c.Context.AbortWithStatus(code)
}

// AbortWithStatusJSON 中止请求并返回JSON格式的错误响应
// 设置状态码、返回JSON数据并中止请求处理
// 参数 code: HTTP状态码
// 参数 jsonObj: 要返回的JSON对象
func (c *Context) AbortWithStatusJSON(code int, jsonObj interface{}) {
	c.Context.AbortWithStatusJSON(code, jsonObj)
}

// AbortWithJSON 中止请求并返回JSON响应（AbortWithStatusJSON的别名）
// 提供更简洁的方法名，功能与AbortWithStatusJSON相同
func (c *Context) AbortWithJSON(code int, jsonObj interface{}) {
	c.Context.AbortWithStatusJSON(code, jsonObj)
}

// AbortWithError 中止请求并记录错误信息
// 设置状态码、记录错误并中止请求处理
// 参数 code: HTTP状态码
// 参数 err: 错误对象
// 返回值: gin.Error指针，包含错误详情
func (c *Context) AbortWithError(code int, err error) *gin.Error {
	return c.Context.AbortWithError(code, err)
}

// =============================================================================
// 错误处理方法
// =============================================================================

// Error 记录错误到上下文的错误列表中
// 不会中止请求处理，只是记录错误供后续处理
// 参数 err: 要记录的错误对象
// 返回值: gin.Error指针，包含错误详情和元数据
func (c *Context) Error(err error) *gin.Error {
	return c.Context.Error(err)
}

// Errors 获取当前上下文中记录的所有错误
// 返回错误切片，包含请求处理过程中记录的所有错误
// 常用于统一错误处理中间件中
func (c *Context) Errors() []*gin.Error {
	return c.Context.Errors
}

// LastError 获取最后一个记录的错误
// 返回错误列表中的最后一个错误，如果没有错误则返回nil
// 便于快速获取最近发生的错误
func (c *Context) LastError() *gin.Error {
	return c.Context.Errors.Last()
}

// =============================================================================
// 上下文数据存储和获取方法
// =============================================================================

// Set 在上下文中设置键值对数据
// 用于在中间件之间传递数据，数据仅在当前请求生命周期内有效
// 参数 key: 数据键名
// 参数 value: 要存储的数据值
func (c *Context) Set(key string, value interface{}) {
	c.Context.Set(key, value)
}

// Get 从上下文中获取指定键的数据
// 参数 key: 数据键名
// 返回值 value: 存储的数据值
// 返回值 exists: 布尔值，表示键是否存在
func (c *Context) Get(key string) (value interface{}, exists bool) {
	return c.Context.Get(key)
}

// MustGet 从上下文中获取指定键的数据，如果不存在则panic
// 用于获取确定存在的数据，如果键不存在会引发panic
// 参数 key: 数据键名
// 返回值: 存储的数据值
func (c *Context) MustGet(key string) interface{} {
	return c.Context.MustGet(key)
}

// GetString 获取字符串类型的上下文数据
// 自动进行类型转换，如果数据不是字符串类型则返回空字符串
// 参数 key: 数据键名
// 返回值: 字符串类型的数据值
func (c *Context) GetString(key string) (s string) {
	return c.Context.GetString(key)
}

// GetBool 获取布尔类型的上下文数据
// 自动进行类型转换，如果数据不是布尔类型则返回false
// 参数 key: 数据键名
// 返回值: 布尔类型的数据值
func (c *Context) GetBool(key string) (b bool) {
	return c.Context.GetBool(key)
}

// GetInt 获取整数类型的上下文数据
// 自动进行类型转换，如果数据不是整数类型则返回0
// 参数 key: 数据键名
// 返回值: 整数类型的数据值
func (c *Context) GetInt(key string) (i int) {
	return c.Context.GetInt(key)
}

// GetInt64 获取int64类型的上下文数据
// 自动进行类型转换，如果数据不是int64类型则返回0
// 参数 key: 数据键名
// 返回值: int64类型的数据值
func (c *Context) GetInt64(key string) (i64 int64) {
	return c.Context.GetInt64(key)
}

// GetUint 获取无符号整数类型的上下文数据
// 自动进行类型转换，如果数据不是uint类型则返回0
// 参数 key: 数据键名
// 返回值: uint类型的数据值
func (c *Context) GetUint(key string) (ui uint) {
	return c.Context.GetUint(key)
}

// GetUint64 获取uint64类型的上下文数据
// 自动进行类型转换，如果数据不是uint64类型则返回0
// 参数 key: 数据键名
// 返回值: uint64类型的数据值
func (c *Context) GetUint64(key string) (ui64 uint64) {
	return c.Context.GetUint64(key)
}

// GetFloat64 获取float64类型的上下文数据
// 自动进行类型转换，如果数据不是float64类型则返回0.0
// 参数 key: 数据键名
// 返回值: float64类型的数据值
func (c *Context) GetFloat64(key string) (f64 float64) {
	return c.Context.GetFloat64(key)
}

// GetTime 获取时间类型的上下文数据
// 自动进行类型转换，如果数据不是time.Time类型则返回零值时间
// 参数 key: 数据键名
// 返回值: time.Time类型的数据值
func (c *Context) GetTime(key string) (t time.Time) {
	return c.Context.GetTime(key)
}

// GetDuration 获取时间间隔类型的上下文数据
// 自动进行类型转换，如果数据不是time.Duration类型则返回0
// 参数 key: 数据键名
// 返回值: time.Duration类型的数据值
func (c *Context) GetDuration(key string) (d time.Duration) {
	return c.Context.GetDuration(key)
}

// GetStringSlice 获取字符串切片类型的上下文数据
// 自动进行类型转换，如果数据不是[]string类型则返回nil
// 参数 key: 数据键名
// 返回值: []string类型的数据值
func (c *Context) GetStringSlice(key string) (ss []string) {
	return c.Context.GetStringSlice(key)
}

// GetStringMap 获取字符串映射类型的上下文数据
// 自动进行类型转换，如果数据不是map[string]interface{}类型则返回nil
// 参数 key: 数据键名
// 返回值: map[string]interface{}类型的数据值
func (c *Context) GetStringMap(key string) (sm map[string]interface{}) {
	return c.Context.GetStringMap(key)
}

// GetStringMapString 获取字符串到字符串映射类型的上下文数据
// 自动进行类型转换，如果数据不是map[string]string类型则返回nil
// 参数 key: 数据键名
// 返回值: map[string]string类型的数据值
func (c *Context) GetStringMapString(key string) (sms map[string]string) {
	return c.Context.GetStringMapString(key)
}

// GetStringMapStringSlice 获取字符串到字符串切片映射类型的上下文数据
// 自动进行类型转换，如果数据不是map[string][]string类型则返回nil
// 参数 key: 数据键名
// 返回值: map[string][]string类型的数据值
func (c *Context) GetStringMapStringSlice(key string) (smss map[string][]string) {
	return c.Context.GetStringMapStringSlice(key)
}

// Keys 获取上下文中存储的所有键值对数据
// 返回包含所有存储数据的map，键为字符串，值为interface{}
// 主要用于调试或批量处理上下文数据
func (c *Context) Keys() map[string]interface{} {
	return c.Context.Keys
}

// =============================================================================
// 路径参数获取方法
// =============================================================================

// Param 获取URL路径中的参数值
// 用于获取路由定义中的路径参数，如 /user/:id 中的 id 参数
// 参数 key: 路径参数名
// 返回值: 参数值字符串，如果参数不存在则返回空字符串
func (c *Context) Param(key string) string {
	return c.Context.Param(key)
}

// =============================================================================
// 查询参数获取方法
// =============================================================================

// Query 获取URL查询参数的值
// 用于获取URL中?后面的查询参数，如 /search?q=golang 中的 q 参数
// 参数 key: 查询参数名
// 返回值: 参数值字符串，如果参数不存在则返回空字符串
func (c *Context) Query(key string) string {
	return c.Context.Query(key)
}

// DefaultQuery 获取查询参数值，如果参数不存在则返回默认值
// 提供默认值机制，避免参数缺失时的空值问题
// 参数 key: 查询参数名
// 参数 defaultValue: 参数不存在时的默认值
// 返回值: 参数值或默认值
func (c *Context) DefaultQuery(key, defaultValue string) string {
	return c.Context.DefaultQuery(key, defaultValue)
}

// GetQuery 获取查询参数值，同时返回参数是否存在的标志
// 可以区分参数值为空字符串和参数不存在的情况
// 参数 key: 查询参数名
// 返回值 value: 参数值字符串
// 返回值 exists: 布尔值，表示参数是否存在
func (c *Context) GetQuery(key string) (string, bool) {
	return c.Context.GetQuery(key)
}

// QueryArray 获取查询参数的数组值
// 用于处理同名的多个查询参数，如 /search?tag=go&tag=web
// 参数 key: 查询参数名
// 返回值: 参数值的字符串切片
func (c *Context) QueryArray(key string) []string {
	return c.Context.QueryArray(key)
}

// GetQueryArray 获取查询参数的数组值，同时返回参数是否存在的标志
// 可以区分参数不存在和参数值为空数组的情况
// 参数 key: 查询参数名
// 返回值 values: 参数值的字符串切片
// 返回值 exists: 布尔值，表示参数是否存在
func (c *Context) GetQueryArray(key string) ([]string, bool) {
	return c.Context.GetQueryArray(key)
}

// QueryMap 获取查询参数的映射值
// 用于处理结构化的查询参数，如 /search?filter[name]=john&filter[age]=25
// 参数 key: 查询参数的前缀名
// 返回值: 参数值的字符串映射
func (c *Context) QueryMap(key string) map[string]string {
	return c.Context.QueryMap(key)
}

// GetQueryMap 获取查询参数的映射值，同时返回参数是否存在的标志
// 可以区分参数不存在和参数值为空映射的情况
// 参数 key: 查询参数的前缀名
// 返回值 values: 参数值的字符串映射
// 返回值 exists: 布尔值，表示参数是否存在
func (c *Context) GetQueryMap(key string) (map[string]string, bool) {
	return c.Context.GetQueryMap(key)
}

// =============================================================================
// POST表单数据获取方法
// =============================================================================

// PostForm 获取POST表单中的参数值
// 用于获取application/x-www-form-urlencoded格式的表单数据
// 参数 key: 表单字段名
// 返回值: 字段值字符串，如果字段不存在则返回空字符串
func (c *Context) PostForm(key string) string {
	return c.Context.PostForm(key)
}

// DefaultPostForm 获取POST表单参数值，如果参数不存在则返回默认值
// 提供默认值机制，避免表单字段缺失时的空值问题
// 参数 key: 表单字段名
// 参数 defaultValue: 字段不存在时的默认值
// 返回值: 字段值或默认值
func (c *Context) DefaultPostForm(key, defaultValue string) string {
	return c.Context.DefaultPostForm(key, defaultValue)
}

// GetPostForm 获取POST表单参数值，同时返回参数是否存在的标志
// 可以区分字段值为空字符串和字段不存在的情况
// 参数 key: 表单字段名
// 返回值 value: 字段值字符串
// 返回值 exists: 布尔值，表示字段是否存在
func (c *Context) GetPostForm(key string) (string, bool) {
	return c.Context.GetPostForm(key)
}

// PostFormArray 获取POST表单中同名字段的数组值
// 用于处理同名的多个表单字段，如多选框的值
// 参数 key: 表单字段名
// 返回值: 字段值的字符串切片
func (c *Context) PostFormArray(key string) []string {
	return c.Context.PostFormArray(key)
}

// GetPostFormArray 获取POST表单同名字段的数组值，同时返回字段是否存在的标志
// 可以区分字段不存在和字段值为空数组的情况
// 参数 key: 表单字段名
// 返回值 values: 字段值的字符串切片
// 返回值 exists: 布尔值，表示字段是否存在
func (c *Context) GetPostFormArray(key string) ([]string, bool) {
	return c.Context.GetPostFormArray(key)
}

// PostFormMap 获取POST表单中结构化字段的映射值
// 用于处理结构化的表单数据，如 user[name]=john&user[age]=25
// 参数 key: 表单字段的前缀名
// 返回值: 字段值的字符串映射
func (c *Context) PostFormMap(key string) map[string]string {
	return c.Context.PostFormMap(key)
}

// GetPostFormMap 获取POST表单结构化字段的映射值，同时返回字段是否存在的标志
// 可以区分字段不存在和字段值为空映射的情况
// 参数 key: 表单字段的前缀名
// 返回值 values: 字段值的字符串映射
// 返回值 exists: 布尔值，表示字段是否存在
func (c *Context) GetPostFormMap(key string) (map[string]string, bool) {
	return c.Context.GetPostFormMap(key)
}

// =============================================================================
// 文件上传处理方法
// =============================================================================

// FormFile 获取上传的单个文件
// 用于处理multipart/form-data格式的文件上传
// 参数 name: 文件字段名
// 返回值 file: 文件头信息，包含文件名、大小等
// 返回值 err: 错误信息，如果获取失败
func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	return c.Context.FormFile(name)
}

// MultipartForm 获取完整的多部分表单数据
// 返回包含所有表单字段和文件的multipart.Form结构
// 返回值 form: 多部分表单数据
// 返回值 err: 错误信息，如果解析失败
func (c *Context) MultipartForm() (*multipart.Form, error) {
	return c.Context.MultipartForm()
}

// SaveUploadedFile 保存上传的文件到指定路径
// 便捷方法，直接将上传的文件保存到服务器文件系统
// 参数 file: 文件头信息，通常来自FormFile方法
// 参数 dst: 目标文件路径
// 返回值: 错误信息，如果保存失败
func (c *Context) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	return c.Context.SaveUploadedFile(file, dst)
}

// =============================================================================
// 数据绑定方法
// =============================================================================

// Bind 自动绑定请求数据到结构体（推荐使用ShouldBind）
// 根据Content-Type自动选择绑定方式，如果绑定失败会返回400错误
// 参数 obj: 要绑定数据的结构体指针
// 返回值: 错误信息，如果绑定失败
func (c *Context) Bind(obj interface{}) error {
	return c.Context.ShouldBind(obj)
}

// ShouldBind 尝试绑定请求数据到结构体
// 根据Content-Type自动选择绑定方式，绑定失败只返回错误不设置状态码
// 支持JSON、XML、YAML、Form等多种格式
// 参数 obj: 要绑定数据的结构体指针
// 返回值: 错误信息，如果绑定失败
func (c *Context) ShouldBind(obj interface{}) error {
	return c.Context.ShouldBind(obj)
}

// ShouldBindJSON 绑定JSON格式的请求数据到结构体
// 专门用于处理application/json格式的请求体
// 参数 obj: 要绑定数据的结构体指针
// 返回值: 错误信息，如果绑定失败
func (c *Context) ShouldBindJSON(obj interface{}) error {
	return c.Context.ShouldBindJSON(obj)
}

// ShouldBindXML 绑定XML格式的请求数据到结构体
// 专门用于处理application/xml格式的请求体
// 参数 obj: 要绑定数据的结构体指针
// 返回值: 错误信息，如果绑定失败
func (c *Context) ShouldBindXML(obj interface{}) error {
	return c.Context.ShouldBindXML(obj)
}

// ShouldBindYAML 绑定YAML格式的请求数据到结构体
// 专门用于处理application/yaml格式的请求体
// 参数 obj: 要绑定数据的结构体指针
// 返回值: 错误信息，如果绑定失败
func (c *Context) ShouldBindYAML(obj interface{}) error {
	return c.Context.ShouldBindYAML(obj)
}

// ShouldBindTOML 绑定TOML格式的请求数据到结构体
// 专门用于处理application/toml格式的请求体
// 参数 obj: 要绑定数据的结构体指针
// 返回值: 错误信息，如果绑定失败
func (c *Context) ShouldBindTOML(obj interface{}) error {
	return c.Context.ShouldBindTOML(obj)
}

// ShouldBindQuery 绑定查询参数到结构体
// 将URL查询参数绑定到结构体字段，支持tag标签映射
// 参数 obj: 要绑定数据的结构体指针
// 返回值: 错误信息，如果绑定失败
func (c *Context) ShouldBindQuery(obj interface{}) error {
	return c.Context.ShouldBindQuery(obj)
}

// ShouldBindUri 绑定URI路径参数到结构体
// 将路由路径中的参数绑定到结构体字段
// 参数 obj: 要绑定数据的结构体指针
// 返回值: 错误信息，如果绑定失败
func (c *Context) ShouldBindUri(obj interface{}) error {
	return c.Context.ShouldBindUri(obj)
}

// ShouldBindHeader 绑定请求头到结构体
// 将HTTP请求头绑定到结构体字段，支持tag标签映射
// 参数 obj: 要绑定数据的结构体指针
// 返回值: 错误信息，如果绑定失败
func (c *Context) ShouldBindHeader(obj interface{}) error {
	return c.Context.ShouldBindHeader(obj)
}

// ShouldBindWith 使用指定的绑定器绑定数据
// 允许使用自定义的绑定器进行数据绑定
// 参数 obj: 要绑定数据的结构体指针
// 参数 b: 绑定器接口实现
// 返回值: 错误信息，如果绑定失败
func (c *Context) ShouldBindWith(obj interface{}, b binding.Binding) error {
	return c.Context.ShouldBindWith(obj, b)
}

// ShouldBindBodyWith 使用指定的绑定器绑定请求体数据
// 专门用于绑定请求体数据，支持缓存已读取的请求体
// 参数 obj: 要绑定数据的结构体指针
// 参数 bb: 请求体绑定器接口实现
// 返回值: 错误信息，如果绑定失败
func (c *Context) ShouldBindBodyWith(obj interface{}, bb binding.BindingBody) error {
	return c.Context.ShouldBindBodyWith(obj, bb)
}

// BindJSON 绑定JSON数据（如果失败会设置400状态码）
// 专门用于JSON数据绑定，失败时自动设置HTTP状态码
// 参数 obj: 要绑定数据的结构体指针
// 返回值: 错误信息，如果绑定失败
func (c *Context) BindJSON(obj interface{}) error {
	return c.Context.BindJSON(obj)
}

// BindXML 绑定XML数据（如果失败会设置400状态码）
// 专门用于XML数据绑定，失败时自动设置HTTP状态码
// 参数 obj: 要绑定数据的结构体指针
// 返回值: 错误信息，如果绑定失败
func (c *Context) BindXML(obj interface{}) error {
	return c.Context.BindXML(obj)
}

// BindYAML 绑定YAML数据（如果失败会设置400状态码）
// 专门用于YAML数据绑定，失败时自动设置HTTP状态码
// 参数 obj: 要绑定数据的结构体指针
// 返回值: 错误信息，如果绑定失败
func (c *Context) BindYAML(obj interface{}) error {
	return c.Context.BindYAML(obj)
}

// BindTOML 绑定TOML数据（如果失败会设置400状态码）
// 专门用于TOML数据绑定，失败时自动设置HTTP状态码
// 参数 obj: 要绑定数据的结构体指针
// 返回值: 错误信息，如果绑定失败
func (c *Context) BindTOML(obj interface{}) error {
	return c.Context.BindTOML(obj)
}

// BindQuery 绑定查询参数（如果失败会设置400状态码）
// 专门用于查询参数绑定，失败时自动设置HTTP状态码
// 参数 obj: 要绑定数据的结构体指针
// 返回值: 错误信息，如果绑定失败
func (c *Context) BindQuery(obj interface{}) error {
	return c.Context.BindQuery(obj)
}

// BindUri 绑定URI参数（如果失败会设置400状态码）
// 专门用于URI路径参数绑定，失败时自动设置HTTP状态码
// 参数 obj: 要绑定数据的结构体指针
// 返回值: 错误信息，如果绑定失败
func (c *Context) BindUri(obj interface{}) error {
	return c.Context.BindUri(obj)
}

// BindHeader 绑定请求头（如果失败会设置400状态码）
// 专门用于请求头绑定，失败时自动设置HTTP状态码
// 参数 obj: 要绑定数据的结构体指针
// 返回值: 错误信息，如果绑定失败
func (c *Context) BindHeader(obj interface{}) error {
	return c.Context.BindHeader(obj)
}

// BindWith 使用指定绑定器绑定数据（如果失败会设置400状态码）
// 允许使用自定义绑定器，失败时自动设置HTTP状态码
// 参数 obj: 要绑定数据的结构体指针
// 参数 b: 绑定器接口实现
// 返回值: 错误信息，如果绑定失败
// 注意：此方法已弃用，建议使用ShouldBindWith
func (c *Context) BindWith(obj interface{}, b binding.Binding) error {
	return c.Context.ShouldBindWith(obj, b)
}

// =============================================================================
// 请求信息获取方法
// =============================================================================

// ClientIP 获取客户端真实IP地址
// 会依次检查X-Forwarded-For、X-Real-Ip等头部信息
// 返回客户端的真实IP地址字符串
func (c *Context) ClientIP() string {
	return c.Context.ClientIP()
}

// ContentType 获取请求的Content-Type头部值
// 返回请求的内容类型，如application/json、text/html等
func (c *Context) ContentType() string {
	return c.Context.ContentType()
}

// IsWebsocket 检查当前请求是否为WebSocket升级请求
// 通过检查Connection和Upgrade头部判断
// 返回布尔值，true表示是WebSocket请求
func (c *Context) IsWebsocket() bool {
	return c.Context.IsWebsocket()
}

// GetHeader 获取指定名称的请求头值
// 参数 key: 请求头名称，不区分大小写
// 返回值: 请求头的值，如果不存在则返回空字符串
func (c *Context) GetHeader(key string) string {
	return c.Context.GetHeader(key)
}

// GetRawData 获取原始的请求体数据
// 返回请求体的原始字节数据，适用于处理二进制数据或自定义格式
// 返回值 data: 请求体的字节切片
// 返回值 err: 错误信息，如果读取失败
func (c *Context) GetRawData() ([]byte, error) {
	return c.Context.GetRawData()
}

// Request 获取底层的HTTP请求对象
// 返回*http.Request对象，可以访问所有原始请求信息
func (c *Context) Request() *http.Request {
	return c.Context.Request
}

// =============================================================================
// 响应设置方法
// =============================================================================

// Status 设置HTTP响应状态码
// 参数 code: HTTP状态码，如200、404、500等
func (c *Context) Status(code int) {
	c.Context.Status(code)
}

// Header 设置响应头
// 参数 key: 响应头名称
// 参数 value: 响应头值
func (c *Context) Header(key, value string) {
	c.Context.Header(key, value)
}

// Writer 获取HTTP响应写入器
// 返回gin.ResponseWriter接口，可以直接写入响应数据
func (c *Context) Writer() gin.ResponseWriter {
	return c.Context.Writer
}

// =============================================================================
// Cookie操作方法
// =============================================================================

// Cookie 获取指定名称的Cookie值
// 参数 name: Cookie名称
// 返回值 value: Cookie值
// 返回值 err: 错误信息，如果Cookie不存在或格式错误
func (c *Context) Cookie(name string) (string, error) {
	return c.Context.Cookie(name)
}

// SetCookie 设置Cookie
// 参数 name: Cookie名称
// 参数 value: Cookie值
// 参数 maxAge: 最大存活时间（秒），0表示会话Cookie，负数表示删除
// 参数 path: Cookie路径
// 参数 domain: Cookie域名
// 参数 secure: 是否仅在HTTPS下传输
// 参数 httpOnly: 是否仅HTTP访问（防止XSS攻击）
func (c *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	c.Context.SetCookie(name, value, maxAge, path, domain, secure, httpOnly)
}

// SetSameSite 设置Cookie的SameSite属性
// 用于防止CSRF攻击，控制Cookie在跨站请求中的发送行为
// 参数 samesite: SameSite属性值（Strict、Lax、None）
func (c *Context) SetSameSite(samesite http.SameSite) {
	c.Context.SetSameSite(samesite)
}

// =============================================================================
// 响应渲染方法
// =============================================================================

// JSON 返回JSON格式的响应
// 最常用的API响应方法，自动设置Content-Type为application/json
// 参数 code: HTTP状态码
// 参数 obj: 要序列化为JSON的对象
func (c *Context) JSON(code int, obj interface{}) {
	c.Context.JSON(code, obj)
}

// IndentedJSON 返回格式化的JSON响应
// 返回带有缩进的JSON，便于调试和阅读
// 参数 code: HTTP状态码
// 参数 obj: 要序列化为JSON的对象
func (c *Context) IndentedJSON(code int, obj interface{}) {
	c.Context.IndentedJSON(code, obj)
}

// SecureJSON 返回安全的JSON响应
// 在JSON前添加特殊前缀，防止JSON劫持攻击
// 参数 code: HTTP状态码
// 参数 obj: 要序列化为JSON的对象
func (c *Context) SecureJSON(code int, obj interface{}) {
	c.Context.SecureJSON(code, obj)
}

// PureJSON 返回纯JSON响应
// 不转义HTML字符，保持JSON的原始格式
// 参数 code: HTTP状态码
// 参数 obj: 要序列化为JSON的对象
func (c *Context) PureJSON(code int, obj interface{}) {
	c.Context.PureJSON(code, obj)
}

// AsciiJSON 返回ASCII编码的JSON响应
// 将非ASCII字符转义为Unicode序列
// 参数 code: HTTP状态码
// 参数 obj: 要序列化为JSON的对象
func (c *Context) AsciiJSON(code int, obj interface{}) {
	c.Context.AsciiJSON(code, obj)
}

// JSONP 返回JSONP格式的响应
// 支持跨域请求的JSON响应格式
// 参数 code: HTTP状态码
// 参数 obj: 要序列化为JSON的对象
func (c *Context) JSONP(code int, obj interface{}) {
	c.Context.JSONP(code, obj)
}

// XML 返回XML格式的响应
// 自动设置Content-Type为application/xml
// 参数 code: HTTP状态码
// 参数 obj: 要序列化为XML的对象
func (c *Context) XML(code int, obj interface{}) {
	c.Context.XML(code, obj)
}

// YAML 返回YAML格式的响应
// 自动设置Content-Type为application/yaml
// 参数 code: HTTP状态码
// 参数 obj: 要序列化为YAML的对象
func (c *Context) YAML(code int, obj interface{}) {
	c.Context.YAML(code, obj)
}

// TOML 返回TOML格式的响应
// 自动设置Content-Type为application/toml
// 参数 code: HTTP状态码
// 参数 obj: 要序列化为TOML的对象
func (c *Context) TOML(code int, obj interface{}) {
	c.Context.TOML(code, obj)
}

// ProtoBuf 返回Protocol Buffers格式的响应
// 用于高效的二进制数据传输
// 参数 code: HTTP状态码
// 参数 obj: 要序列化为ProtoBuf的对象
func (c *Context) ProtoBuf(code int, obj interface{}) {
	c.Context.ProtoBuf(code, obj)
}

// String 返回字符串格式的响应
// 支持格式化字符串，类似fmt.Sprintf
// 参数 code: HTTP状态码
// 参数 format: 格式化字符串
// 参数 values: 格式化参数
func (c *Context) String(code int, format string, values ...interface{}) {
	c.Context.String(code, format, values...)
}

// HTML 返回HTML格式的响应
// 使用模板引擎渲染HTML页面
// 参数 code: HTTP状态码
// 参数 name: 模板名称
// 参数 obj: 传递给模板的数据对象
func (c *Context) HTML(code int, name string, obj interface{}) {
	c.Context.HTML(code, name, obj)
}

// Data 返回原始数据响应
// 直接返回字节数据，需要手动设置Content-Type
// 参数 code: HTTP状态码
// 参数 contentType: 内容类型
// 参数 data: 响应数据的字节切片
func (c *Context) Data(code int, contentType string, data []byte) {
	c.Context.Data(code, contentType, data)
}

// DataFromReader 从Reader返回数据响应
// 从io.Reader流式传输数据，适用于大文件或实时数据
// 参数 code: HTTP状态码
// 参数 contentLength: 内容长度，-1表示未知
// 参数 contentType: 内容类型
// 参数 reader: 数据读取器
// 参数 extraHeaders: 额外的响应头
func (c *Context) DataFromReader(code int, contentLength int64, contentType string, reader io.Reader, extraHeaders map[string]string) {
	c.Context.DataFromReader(code, contentLength, contentType, reader, extraHeaders)
}

// Redirect 执行HTTP重定向
// 参数 code: 重定向状态码（301、302、303、307、308）
// 参数 location: 重定向目标URL
func (c *Context) Redirect(code int, location string) {
	c.Context.Redirect(code, location)
}

// Render 使用自定义渲染器渲染响应
// 允许使用自定义的渲染器进行响应渲染
// 参数 code: HTTP状态码
// 参数 r: 渲染器接口实现
func (c *Context) Render(code int, r render.Render) {
	c.Context.Render(code, r)
}

// =============================================================================
// 文件响应方法
// =============================================================================

// File 返回文件响应
// 直接返回服务器文件系统中的文件
// 参数 filepath: 文件路径
func (c *Context) File(filepath string) {
	c.Context.File(filepath)
}

// FileFromFS 从指定文件系统返回文件响应
// 从自定义文件系统（如嵌入式文件系统）返回文件
// 参数 filepath: 文件路径
// 参数 fs: 文件系统接口
func (c *Context) FileFromFS(filepath string, fs http.FileSystem) {
	c.Context.FileFromFS(filepath, fs)
}

// FileAttachment 返回文件附件响应
// 强制浏览器下载文件而不是在浏览器中打开
// 参数 filepath: 服务器文件路径
// 参数 filename: 下载时显示的文件名
func (c *Context) FileAttachment(filepath, filename string) {
	c.Context.FileAttachment(filepath, filename)
}

// =============================================================================
// 流式响应方法
// =============================================================================

// Stream 返回流式响应
// 用于服务器发送事件（SSE）或实时数据传输
// 参数 step: 流式写入函数，返回false时停止流式传输
// 返回值: 布尔值，表示流式传输是否正常结束
func (c *Context) Stream(step func(w io.Writer) bool) bool {
	return c.Context.Stream(step)
}

// SSEvent 发送服务器发送事件（Server-Sent Events）
// 用于实现服务器主动推送数据到客户端
// 参数 name: 事件名称
// 参数 message: 事件消息内容
func (c *Context) SSEvent(name, message string) {
	c.Context.SSEvent(name, message)
}

// =============================================================================
// 内容协商方法
// =============================================================================

// NegotiateFormat 协商客户端接受的内容格式
// 根据Accept头部选择最合适的响应格式
// 参数 offered: 服务器支持的格式列表
// 返回值: 协商结果的格式字符串
func (c *Context) NegotiateFormat(offered ...string) string {
	return c.Context.NegotiateFormat(offered...)
}

// Negotiate 根据协商结果返回相应格式的响应
// 自动根据客户端Accept头部选择响应格式
// 参数 code: HTTP状态码
// 参数 config: 协商配置，包含不同格式的数据
func (c *Context) Negotiate(code int, config gin.Negotiate) {
	c.Context.Negotiate(code, config)
}

// SetAccepted 设置当前上下文接受的内容类型
// 用于内容协商过程中设置接受的格式
// 参数 formats: 接受的格式列表
func (c *Context) SetAccepted(formats ...string) {
	c.Context.SetAccepted(formats...)
}

// Accepted 获取当前上下文接受的内容类型列表
// 返回客户端Accept头部解析后的格式列表
func (c *Context) Accepted() []string {
	return c.Context.Accepted
}

// =============================================================================
// 路由信息方法
// =============================================================================

// FullPath 获取当前路由的完整路径模式
// 返回路由定义时的路径模式，如"/user/:id"
func (c *Context) FullPath() string {
	return c.Context.FullPath()
}

// HandlerName 获取当前处理器的函数名
// 返回处理器函数的完整名称，包含包路径
func (c *Context) HandlerName() string {
	return c.Context.HandlerName()
}

// HandlerNames 获取当前请求处理链中所有处理器的名称
// 返回包含所有中间件和最终处理器名称的切片
func (c *Context) HandlerNames() []string {
	return c.Context.HandlerNames()
}

// Handler 获取当前的处理器函数
// 返回gin.HandlerFunc类型的处理器函数
func (c *Context) Handler() gin.HandlerFunc {
	return c.Context.Handler()
}

// =============================================================================
// 上下文接口实现（context.Context）
// =============================================================================

// Deadline 获取上下文的截止时间
// 实现context.Context接口，用于超时控制
// 返回值 deadline: 截止时间
// 返回值 ok: 是否设置了截止时间
func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.Context.Deadline()
}

// Done 获取上下文完成信号通道
// 实现context.Context接口，用于取消信号传播
// 返回值: 完成信号通道，当上下文被取消时会关闭
func (c *Context) Done() <-chan struct{} {
	return c.Context.Done()
}

// Err 获取上下文的错误信息
// 实现context.Context接口，返回上下文取消的原因
// 返回值: 错误信息，如果上下文未取消则返回nil
func (c *Context) Err() error {
	return c.Context.Err()
}

// Value 获取上下文中存储的值
// 实现context.Context接口，用于跨函数传递数据
// 参数 key: 数据键，通常使用自定义类型避免冲突
// 返回值: 存储的数据值，如果不存在则返回nil
func (c *Context) Value(key interface{}) interface{} {
	return c.Context.Value(key)
}

// =============================================================================
// 上下文复制方法
// =============================================================================

// Copy 创建当前上下文的副本
// 返回一个新的Context实例，可以安全地在goroutine中使用
// 副本包含当前上下文的所有数据，但与原上下文独立
// 返回值: 新的Context实例指针
func (c *Context) Copy() *Context {
	return &Context{Context: c.Context.Copy()}
}
