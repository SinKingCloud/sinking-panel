package oss

import (
	"fmt"
	"server/app/util/oss/types"
)

// Error 保留厂商错误信息，支持 errors.As 与 errors.Is。
type Error = types.Error

var (
	ErrInvalidArgument    = types.ErrInvalidArgument
	ErrNotFound           = types.ErrNotFound
	ErrPermissionDenied   = types.ErrPermissionDenied
	ErrConflict           = types.ErrConflict
	ErrPreconditionFailed = types.ErrPreconditionFailed
)

// providerError 将底层错误包装为统一的厂商操作错误。
func providerError(provider int, op, key string, err error) error {
	return types.WrapError(provider, op, key, err)
}

// invalid 创建可由 errors.Is 判断的参数错误。
func invalid(message string) error {
	return fmt.Errorf("%w: %s", ErrInvalidArgument, message)
}
