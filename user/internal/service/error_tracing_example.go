package service

import (
	"context"
	"fmt"

	"user/api/error_reason"
	"user/internal/pkg/tracing"
)

// ErrorTracingExample 演示错误追踪功能的使用示例
type ErrorTracingExample struct{}

// DemoBusinessError 演示在业务逻辑中如何使用错误追踪
func (e *ErrorTracingExample) DemoBusinessError(ctx context.Context) error {
	// 模拟一些业务逻辑
	if err := someBusinessLogic(ctx); err != nil {
		// 这里的错误将被中间件自动增强，包含追踪信息
		return error_reason.ErrorUserInvalidToken("用户认证失败")
	}
	return nil
}

// DemoFormattedErrorLogging 演示如何格式化包含追踪信息的错误日志
func (e *ErrorTracingExample) DemoFormattedErrorLogging(ctx context.Context, err error) string {
	// 使用工具函数格式化错误信息，包含追踪信息
	return tracing.FormatErrorWithTrace(err)
}

// DemoExtractTraceInfo 演示如何从错误中提取追踪信息
func (e *ErrorTracingExample) DemoExtractTraceInfo(err error) (string, string, bool) {
	return tracing.ExtractTraceInfoFromError(err)
}

// DemoMultipleErrorTypes 演示不同类型错误的处理
func (e *ErrorTracingExample) DemoMultipleErrorTypes(ctx context.Context, errorType string) error {
	switch errorType {
	case "invalid_token":
		return error_reason.ErrorUserInvalidToken("Token无效或已过期")
	case "user_not_found":
		return error_reason.ErrorUserNotFound("用户不存在")
	case "database_error":
		return error_reason.ErrorUserDatabaseError("数据库操作失败")
	case "too_many_requests":
		return error_reason.ErrorUserTooManyRequests("请求过于频繁")
	default:
		return error_reason.ErrorUserInternalError("未知错误")
	}
}

// DemoErrorAssertionWithTrace 演示在调试时如何使用错误断言和追踪信息
func (e *ErrorTracingExample) DemoErrorAssertionWithTrace(err error) {
	// 错误断言
	if error_reason.IsUserInvalidToken(err) {
		// 提取追踪信息
		traceID, spanID, hasTrace := e.DemoExtractTraceInfo(err)
		if hasTrace {
			fmt.Printf("Token错误 - 追踪信息: traceid=%s, spanid=%s\n", traceID, spanID)
		} else {
			fmt.Println("Token错误 - 缺少追踪信息")
		}
	}

	if error_reason.IsUserNotFound(err) {
		traceID, spanID, hasTrace := e.DemoExtractTraceInfo(err)
		if hasTrace {
			fmt.Printf("用户不存在 - 追踪信息: traceid=%s, spanid=%s\n", traceID, spanID)
		} else {
			fmt.Println("用户不存在 - 缺少追踪信息")
		}
	}

	// 通用处理
	formattedError := e.DemoFormattedErrorLogging(context.Background(), err)
	fmt.Printf("完整错误信息: %s\n", formattedError)
}

// DemoProductionLogging 演示生产环境中的日志记录最佳实践
func (e *ErrorTracingExample) DemoProductionLogging(ctx context.Context, err error) map[string]interface{} {
	// 构建结构化日志输出
	logData := map[string]interface{}{
		"error_type":   "business_error",
		"error_reason": "",
		"message":      "",
		"has_trace":    false,
		"trace_id":     "",
		"span_id":      "",
	}

	// 提取错误信息
	if err != nil {
		// 这里需要根据实际错误类型进行判断
		logData["error_reason"] = "USER_INVALID_TOKEN"
	}

	logData["message"] = err.Error()

	// 提取追踪信息
	traceID, spanID, hasTrace := tracing.ExtractTraceInfoFromError(err)
	logData["has_trace"] = hasTrace
	if hasTrace {
		logData["trace_id"] = traceID
		logData["span_id"] = spanID
	}

	return logData
}

// DemoAPIVerification 演示如何验证API响应包含追踪信息
func (e *ErrorTracingExample) DemoAPIVerification(response interface{}) bool {
	// 这里可以解析HTTP响应或gRPC响应
	// 检查metadata字段是否包含traceid和spanid

	// 示例：检查是否为Kratos错误响应
	// if err, ok := response.(*errors.Error); ok {
	//     metadata := err.Metadata
	//     _, hasTraceID := metadata["traceid"]
	//     _, hasSpanID := metadata["spanid"]
	//     return hasTraceID && hasSpanID
	// }

	return false
}

// someBusinessLogic 模拟业务逻辑
func someBusinessLogic(ctx context.Context) error {
	// 这里可以模拟实际的业务逻辑
	// 比如数据库查询、第三方API调用等
	return nil
}

/*
使用示例：

1. 在服务中使用：
   errorHandler := &ErrorTracingExample{}
   if err := errorHandler.DemoBusinessError(ctx); err != nil {
       logger.Error(errorHandler.DemoFormattedErrorLogging(ctx, err))
   }

2. 在控制器中处理错误：
   func (s *UserService) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.GetUserResponse, error) {
       // ... 业务逻辑
       if err != nil {
           // 错误将自动包含追踪信息，不需要手动处理
           return nil, err
       }
   }

3. 在客户端处理追踪信息：
   // HTTP响应
   {
     "code": 401,
     "reason": "USER_INVALID_TOKEN",
     "message": "用户认证信息无效",
     "metadata": {
       "traceid": "4bf92f3577b34da6a3ce929d0e0e4736",
       "spanid": "00f0a4777a90e3c10a7c3e5b3e2c3c9d"
     }
   }

4. 在调试时提取追踪信息：
   traceID, spanID, hasTrace := tracing.ExtractTraceInfoFromError(err)
   if hasTrace {
       fmt.Printf("请使用此追踪ID查找日志: traceid=%s, spanid=%s", traceID, spanID)
   }

*/
