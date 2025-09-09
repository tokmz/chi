package chi

import (
	"errors"
	"net/http"
)

// Response
// 响应Model
type Response struct {
	Code    int    `json:"code"`
	Data    any    `json:"data"`
	Message string `json:"message"`
}

// NewResponse
// 初始化响应
func NewResponse(code int, data any, message string) *Response {
	return &Response{
		Code:    code,
		Data:    data,
		Message: message,
	}
}

// NewErrResponse
// 初始化异常响应
func NewErrResponse(code int, message string) *Response {
	return NewResponse(code, nil, message)
}

// NewOkResponse
// 初始化正常响应
func NewOkResponse(data any) *Response {
	return NewResponse(200, data, "success")
}

// PageResp
// 列表响应
type PageResp[T any] struct {
	Total int64 `json:"total"`
	List  T     `json:"list"`
}

// NewPageResp
// 列表响应初始化
func NewPageResp[T any](total int64, list T) *PageResp[T] {
	return &PageResp[T]{
		Total: total,
		List:  list,
	}
}

// Res
// Api响应
func Res(ctx *Context, err error, data ...any) {
	if err != nil {
		var e *Error
		if !errors.As(err, &e) {
			ctx.JSON(http.StatusOK, NewErrResponse(500, "未知异常"))
			return
		}
		ctx.JSON(http.StatusOK, NewErrResponse(e.Code, e.Message))
		return
	}

	if len(data) == 0 {
		ctx.JSON(http.StatusOK, NewOkResponse(nil))
		return
	}

	ctx.JSON(http.StatusOK, NewOkResponse(data[0]))
}

// SuccessRes
// 成功响应
func SuccessRes(ctx *Context, data any) {
	ctx.JSON(http.StatusOK, NewOkResponse(data))
}

// FailRes
// 失败响应
func FailRes(ctx *Context, err error) {
	Res(ctx, err)
}
