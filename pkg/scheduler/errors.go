package scheduler

import (
	"fmt"
)

// ErrorType 错误类型
type ErrorType string

const (
	// 配置错误
	ErrInvalidConfig ErrorType = "INVALID_CONFIG"
	// 任务错误
	ErrTaskNotFound    ErrorType = "TASK_NOT_FOUND"
	ErrTaskExists      ErrorType = "TASK_EXISTS"
	ErrTaskRunning     ErrorType = "TASK_RUNNING"
	ErrTaskStopped     ErrorType = "TASK_STOPPED"
	ErrInvalidTaskType ErrorType = "INVALID_TASK_TYPE"
	// 调度器错误
	ErrSchedulerStopped ErrorType = "SCHEDULER_STOPPED"
	ErrSchedulerRunning ErrorType = "SCHEDULER_RUNNING"
	ErrWorkerPoolFull   ErrorType = "WORKER_POOL_FULL"
	ErrInvalidCronExpr  ErrorType = "INVALID_CRON_EXPR"
	// 系统错误
	ErrSystemError ErrorType = "SYSTEM_ERROR"
	ErrTimeout     ErrorType = "TIMEOUT"
	ErrPersistence ErrorType = "PERSISTENCE_ERROR"
	ErrMonitor     ErrorType = "MONITOR_ERROR"
	ErrConcurrency ErrorType = "CONCURRENCY_ERROR"
)

// SchedulerError 调度器错误
type SchedulerError struct {
	Type    ErrorType              `json:"type"`
	Message string                 `json:"message"`
	Cause   error                  `json:"cause,omitempty"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// Error 实现error接口
func (e *SchedulerError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap 返回原始错误
func (e *SchedulerError) Unwrap() error {
	return e.Cause
}

// WithContext 添加上下文信息
func (e *SchedulerError) WithContext(key string, value interface{}) *SchedulerError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// NewSchedulerError 创建新的调度器错误
func NewSchedulerError(errType ErrorType, message string) *SchedulerError {
	return &SchedulerError{
		Type:    errType,
		Message: message,
	}
}

// NewSchedulerErrorWithCause 创建带原因的调度器错误
func NewSchedulerErrorWithCause(errType ErrorType, message string, cause error) *SchedulerError {
	return &SchedulerError{
		Type:    errType,
		Message: message,
		Cause:   cause,
	}
}

// IsSchedulerError 检查是否为调度器错误
func IsSchedulerError(err error) bool {
	_, ok := err.(*SchedulerError)
	return ok
}

// GetErrorType 获取错误类型
func GetErrorType(err error) ErrorType {
	if schedErr, ok := err.(*SchedulerError); ok {
		return schedErr.Type
	}
	return ErrSystemError
}

// IsErrorType 检查错误是否为指定类型
func IsErrorType(err error, errType ErrorType) bool {
	return GetErrorType(err) == errType
}

// WrapError 包装错误
func WrapError(errType ErrorType, message string, cause error) error {
	return NewSchedulerErrorWithCause(errType, message, cause)
}

// ErrorHandler 错误处理器接口
type ErrorHandler interface {
	HandleError(err error, context map[string]interface{})
}

// DefaultErrorHandler 默认错误处理器
type DefaultErrorHandler struct {
	logger Logger
}

// NewDefaultErrorHandler 创建默认错误处理器
func NewDefaultErrorHandler(logger Logger) *DefaultErrorHandler {
	return &DefaultErrorHandler{
		logger: logger,
	}
}

// HandleError 处理错误
func (h *DefaultErrorHandler) HandleError(err error, context map[string]interface{}) {
	if h.logger != nil {
		h.logger.Error("Scheduler error occurred", map[string]interface{}{
			"error":   err.Error(),
			"type":    GetErrorType(err),
			"context": context,
		})
	}
}

// PanicHandler panic处理器
type PanicHandler struct {
	logger Logger
}

// NewPanicHandler 创建panic处理器
func NewPanicHandler(logger Logger) *PanicHandler {
	return &PanicHandler{
		logger: logger,
	}
}

// HandlePanic 处理panic
func (h *PanicHandler) HandlePanic(recovered interface{}, context map[string]interface{}) {
	if h.logger != nil {
		h.logger.Error("Panic recovered in scheduler", map[string]interface{}{
			"panic":   recovered,
			"context": context,
		})
	}
}

// RecoverWithHandler 使用处理器恢复panic
func RecoverWithHandler(handler *PanicHandler, context map[string]interface{}) {
	if r := recover(); r != nil {
		if handler != nil {
			handler.HandlePanic(r, context)
		}
	}
}
