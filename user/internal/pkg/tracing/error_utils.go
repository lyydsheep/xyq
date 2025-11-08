package tracing

import (
	"context"
	"fmt"

	errors "github.com/go-kratos/kratos/v2/errors"
)

// NewErrorWithTrace 创建带追踪信息的错误
func NewErrorWithTrace(ctx context.Context, code int, reason, format string, args ...interface{}) *errors.Error {
	traceInfo := ExtractTraceInfo(ctx)

	// 先格式化消息
	message := fmt.Sprintf(format, args...)

	// 创建错误
	err := errors.New(code, reason, message)

	// 如果有追踪信息，添加到 metadata
	if traceInfo.TraceID != "" || traceInfo.SpanID != "" {
		metadata := make(map[string]string)
		if traceInfo.TraceID != "" {
			metadata["traceid"] = traceInfo.TraceID
		}
		if traceInfo.SpanID != "" {
			metadata["spanid"] = traceInfo.SpanID
		}

		// 设置 metadata
		err = err.WithMetadata(metadata)
	}

	return err
}

// NewUserErrorWithTrace 创建用户相关的错误（带追踪信息）
func NewUserErrorWithTrace(ctx context.Context, reason, format string, args ...interface{}) *errors.Error {
	return NewErrorWithTrace(ctx, 0, reason, format, args...)
}

// NewAuthErrorWithTrace 创建认证相关的错误（带追踪信息）
func NewAuthErrorWithTrace(ctx context.Context, reason, format string, args ...interface{}) *errors.Error {
	return NewErrorWithTrace(ctx, 0, reason, format, args...)
}

// NewSystemErrorWithTrace 创建系统相关的错误（带追踪信息）
func NewSystemErrorWithTrace(ctx context.Context, reason, format string, args ...interface{}) *errors.Error {
	return NewErrorWithTrace(ctx, 0, reason, format, args...)
}
