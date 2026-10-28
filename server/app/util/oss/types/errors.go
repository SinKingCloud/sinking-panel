package types

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidArgument    = errors.New("oss: 参数不合法")
	ErrNotFound           = errors.New("oss: 对象或存储桶不存在")
	ErrPermissionDenied   = errors.New("oss: 没有访问权限")
	ErrConflict           = errors.New("oss: 操作冲突")
	ErrPreconditionFailed = errors.New("oss: 前置条件不满足")
)

// Error 保留厂商错误信息供 errors.As 检查，Error() 不输出可能包含签名和凭据的原始响应。
type Error struct {
	Provider   int
	Op         string
	Key        string
	StatusCode int
	Code       string
	Message    string
	Err        error
}

// Error 返回操作摘要，不包含可能泄漏签名或凭据的原始响应。
func (e *Error) Error() string {
	return fmt.Sprintf("oss: %s %q 失败 (provider=%d, status=%d, code=%s)", e.Op, e.Key, e.Provider, e.StatusCode, e.Code)
}

// Unwrap 返回底层错误，保留取消、超时和 SDK 错误链。
func (e *Error) Unwrap() error {
	return e.Err
}

// Is 按 HTTP 状态匹配统一错误分类。
func (e *Error) Is(target error) bool {
	switch target {
	case ErrNotFound:
		return e.StatusCode == 404
	case ErrPermissionDenied:
		return e.StatusCode == 401 || e.StatusCode == 403
	case ErrConflict:
		return e.StatusCode == 409
	case ErrPreconditionFailed:
		return e.StatusCode == 412
	}
	return false
}

// WrapError 保留已有统一错误，否则包装厂商操作错误。
func WrapError(provider int, op, key string, err error) error {
	if err == nil {
		return nil
	}
	var e *Error
	if errors.As(err, &e) {
		return err
	}
	return &Error{Provider: provider, Op: op, Key: key, Err: err}
}
