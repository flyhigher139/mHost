package errors

import (
	"fmt"
)

// AppError 应用程序错误接口
type AppError interface {
	error
	Code() string
	Type() ErrorType
	Details() map[string]interface{}
	Cause() error
}

// ErrorType 错误类型
type ErrorType string

const (
	ErrorTypeValidation ErrorType = "validation"
	ErrorTypePermission ErrorType = "permission"
	ErrorTypeFileSystem ErrorType = "filesystem"
	ErrorTypeNetwork    ErrorType = "network"
	ErrorTypeSystem     ErrorType = "system"
	ErrorTypeInternal   ErrorType = "internal"
)

// appError 具体错误实现
type appError struct {
	code    string
	errType ErrorType
	message string
	details map[string]interface{}
	cause   error
}

// Error 实现error接口
func (e *appError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.message, e.cause)
	}
	return e.message
}

// Code 返回错误代码
func (e *appError) Code() string {
	return e.code
}

// Type 返回错误类型
func (e *appError) Type() ErrorType {
	return e.errType
}

// Details 返回错误详情
func (e *appError) Details() map[string]interface{} {
	return e.details
}

// Cause 返回原始错误
func (e *appError) Cause() error {
	return e.cause
}

// NewValidationError 创建验证错误
func NewValidationError(code, message string, details map[string]interface{}) AppError {
	return &appError{
		code:    code,
		errType: ErrorTypeValidation,
		message: message,
		details: details,
	}
}

// NewPermissionError 创建权限错误
func NewPermissionError(code, message string) AppError {
	return &appError{
		code:    code,
		errType: ErrorTypePermission,
		message: message,
	}
}

// NewFileSystemError 创建文件系统错误
func NewFileSystemError(code, message string, cause error) AppError {
	return &appError{
		code:    code,
		errType: ErrorTypeFileSystem,
		message: message,
		cause:   cause,
	}
}

// NewNetworkError 创建网络错误
func NewNetworkError(code, message string, cause error) AppError {
	return &appError{
		code:    code,
		errType: ErrorTypeNetwork,
		message: message,
		cause:   cause,
	}
}

// NewSystemError 创建系统错误
func NewSystemError(code, message string, cause error) AppError {
	return &appError{
		code:    code,
		errType: ErrorTypeSystem,
		message: message,
		cause:   cause,
	}
}

// NewInternalError 创建内部错误
func NewInternalError(code, message string, cause error) AppError {
	return &appError{
		code:    code,
		errType: ErrorTypeInternal,
		message: message,
		cause:   cause,
	}
}

// WrapError 包装现有错误为AppError
func WrapError(code string, errType ErrorType, message string, cause error) AppError {
	return &appError{
		code:    code,
		errType: errType,
		message: message,
		cause:   cause,
	}
}

// IsAppError 检查是否为AppError类型
func IsAppError(err error) bool {
	_, ok := err.(AppError)
	return ok
}

// GetAppError 获取AppError，如果不是则返回nil
func GetAppError(err error) AppError {
	if appErr, ok := err.(AppError); ok {
		return appErr
	}
	return nil
}