package tracing

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport/http"
	"go.opentelemetry.io/otel/trace"
)

// ErrorResponseEnhancer 错误响应增强中间件
// 自动为错误响应添加 traceid 和 spanid 到 metadata 中
func ErrorResponseEnhancer() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 执行正常的业务逻辑
			reply, err := handler(ctx, req)

			// 如果没有错误，直接返回
			if err == nil {
				return reply, nil
			}

			// 获取当前追踪信息
			span := trace.SpanFromContext(ctx)
			if span != nil {
				// 获取 traceid 和 spanid
				traceID := span.SpanContext().TraceID().String()
				spanID := span.SpanContext().SpanID().String()

				// 创建包含追踪信息的 metadata
				metadata := map[string]string{
					"traceid": traceID,
					"spanid":  spanID,
				}

				// 将追踪信息添加到错误中
				// 先转换为 Kratos 错误类型，然后添加 metadata
				kratosErr := errors.FromError(err)
				if kratosErr != nil {
					enhancedErr := kratosErr.WithMetadata(metadata)
					return reply, enhancedErr
				}
			}

			return reply, err
		}
	}
}

// HTTPErrorResponseEnhancer HTTP 错误响应增强中间件
// 专门处理 HTTP 错误响应，确保在 HTTP 响应中包含追踪信息
func HTTPErrorResponseEnhancer() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 执行正常的业务逻辑
			reply, err := handler(ctx, req)

			// 如果没有错误，直接返回
			if err == nil {
				return reply, nil
			}

			// 尝试从 HTTP 请求上下文中获取追踪信息
			if httpReq, ok := http.RequestFromServerContext(ctx); ok {
				// 从请求头中获取追踪信息（如果存在）
				traceID := httpReq.Header.Get("X-Trace-ID")
				spanID := httpReq.Header.Get("X-Span-ID")

				if traceID == "" || spanID == "" {
					// 如果头信息中没有，尝试从 OpenTelemetry 上下文中获取
					span := trace.SpanFromContext(ctx)
					if span != nil {
						traceID = span.SpanContext().TraceID().String()
						spanID = span.SpanContext().SpanID().String()
					}
				}

				// 如果获取到有效的追踪信息，添加到错误中
				if traceID != "" && spanID != "" {
					metadata := map[string]string{
						"traceid": traceID,
						"spanid":  spanID,
					}
					// 转换为 Kratos 错误类型，然后添加 metadata
					kratosErr := errors.FromError(err)
					if kratosErr != nil {
						err = kratosErr.WithMetadata(metadata)
					}
				}
			}

			return reply, err
		}
	}
}

// GRPCErrorResponseEnhancer gRPC 错误响应增强中间件
// 专门处理 gRPC 错误响应
func GRPCErrorResponseEnhancer() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 执行正常的业务逻辑
			reply, err := handler(ctx, req)

			// 如果没有错误，直接返回
			if err == nil {
				return reply, nil
			}

			// 从 OpenTelemetry 上下文中获取追踪信息
			span := trace.SpanFromContext(ctx)
			if span != nil {
				traceID := span.SpanContext().TraceID().String()
				spanID := span.SpanContext().SpanID().String()

				// 创建包含追踪信息的 metadata
				metadata := map[string]string{
					"traceid": traceID,
					"spanid":  spanID,
				}

				// 将追踪信息添加到错误中
				// 先转换为 Kratos 错误类型，然后添加 metadata
				kratosErr := errors.FromError(err)
				if kratosErr != nil {
					enhancedErr := kratosErr.WithMetadata(metadata)
					return reply, enhancedErr
				}
			}

			return reply, err
		}
	}
}

// ExtractTraceInfoFromError 从错误中提取追踪信息
func ExtractTraceInfoFromError(err error) (string, string, bool) {
	if err == nil {
		return "", "", false
	}

	// 使用 errors.FromError 解析错误
	e := errors.FromError(err)
	if e == nil {
		return "", "", false
	}

	// 从 metadata 中获取追踪信息
	traceID, traceIDExists := e.Metadata["traceid"]
	spanID, spanIDExists := e.Metadata["spanid"]

	if traceIDExists && spanIDExists {
		return traceID, spanID, true
	}

	return "", "", false
}

// FormatErrorWithTrace 格式化错误信息，包含追踪信息
func FormatErrorWithTrace(err error) string {
	if err == nil {
		return "no error"
	}

	traceID, spanID, hasTrace := ExtractTraceInfoFromError(err)
	if hasTrace {
		return fmt.Sprintf("error: %s, traceid: %s, spanid: %s", err.Error(), traceID, spanID)
	}

	return fmt.Sprintf("error: %s", err.Error())
}
