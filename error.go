package chi

import "net/http"

type Error struct {
	Code    int
	Message string
}

func NewError(code int, message string) *Error {
	if message == "" {
		panic("message cannot be nil")
	}
	return &Error{code, message}
}

var (
	ErrServer  = NewError(http.StatusInternalServerError, "服务异常")
	ErrBinding = NewError(http.StatusBadRequest, "参数错误")
)

func (e *Error) Error() string {
	return e.Message
}
