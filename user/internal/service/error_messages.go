package service

import (
	"github.com/go-kratos/kratos/v2/errors"
)

// ErrorMessageMap 错误消息映射
// 支持多语言错误消息，这里提供中件友好的错误信息
var ErrorMessageMap = map[string]string{
	// UserService 错误消息
	"USER_INVALID_TOKEN":         "访问令牌无效，请重新登录",
	"USER_TOKEN_EXPIRED":         "访问令牌已过期，请重新登录",
	"USER_INVALID_CREDENTIALS":   "用户名或密码错误",
	"USER_REFRESH_TOKEN_INVALID": "刷新令牌无效，请重新登录",

	"USER_INVALID_EMAIL":             "邮箱格式不正确",
	"USER_INVALID_VERIFICATION_CODE": "验证码错误",
	"USER_VERIFICATION_CODE_EXPIRED": "验证码已过期",
	"USER_INVALID_REQUEST":           "请求参数无效",
	"USER_INVALID_NICKNAME":          "昵称格式不正确",

	"USER_EMAIL_ALREADY_EXISTS":    "该邮箱已被注册",
	"USER_NICKNAME_ALREADY_EXISTS": "该昵称已被使用",

	"USER_NOT_FOUND":         "用户不存在",
	"USER_PROFILE_NOT_FOUND": "用户资料不存在",

	"USER_TOO_MANY_REQUESTS": "请求过于频繁，请稍后再试",
	"USER_LOGIN_TOO_MANY":    "登录尝试次数过多，请稍后再试",

	"USER_DATABASE_ERROR":      "数据库操作失败",
	"USER_INTERNAL_ERROR":      "服务内部错误",
	"USER_SERVICE_UNAVAILABLE": "用户服务暂时不可用",

	// AuthService 错误消息
	"AUTH_INVALID_CREDENTIALS":   "用户名或密码错误",
	"AUTH_TOKEN_INVALID":         "访问令牌无效",
	"AUTH_TOKEN_EXPIRED":         "访问令牌已过期",
	"AUTH_REFRESH_TOKEN_INVALID": "刷新令牌无效",

	"AUTH_INVALID_REQUEST": "认证请求参数无效",
	"AUTH_INVALID_EMAIL":   "邮箱格式不正确",
	"AUTH_INVALID_CODE":    "验证码格式错误",

	"AUTH_EMAIL_EXISTS": "邮箱已被注册",

	"AUTH_TOO_MANY_REQUESTS": "认证请求过于频繁",
	"AUTH_LOGIN_BLOCKED":     "登录已被临时锁定",

	"AUTH_DATABASE_ERROR":      "认证数据库操作失败",
	"AUTH_SERVICE_ERROR":       "认证服务内部错误",
	"AUTH_SERVICE_UNAVAILABLE": "认证服务暂时不可用",

	// 系统级错误消息
	"DATABASE_CONNECTION_ERROR": "数据库连接失败",
	"DATABASE_OPERATION_ERROR":  "数据库操作失败",
	"DATABASE_TIMEOUT_ERROR":    "数据库操作超时",

	"SERVICE_UNAVAILABLE": "服务暂时不可用",
	"SERVICE_OVERLOADED":  "服务过载，请稍后再试",

	"REDIS_CONNECTION_ERROR": "缓存服务连接失败",
	"EXTERNAL_SERVICE_ERROR": "外部服务错误",
}

// GetFriendlyErrorMessage 获取用户友好的错误消息
func GetFriendlyErrorMessage(reason string) string {
	if message, exists := ErrorMessageMap[reason]; exists {
		return message
	}
	// 如果没有找到映射的错误消息，返回通用错误消息
	return "操作失败，请稍后重试"
}

// StandardErrorResponse 标准错误响应结构
type StandardErrorResponse struct {
	Code    int                    `json:"code"`              // HTTP状态码
	Reason  string                 `json:"reason"`            // 错误原因
	Message string                 `json:"message"`           // 用户友好的错误信息
	Details map[string]interface{} `json:"details,omitempty"` // 错误详情
	Meta    map[string]string      `json:"meta,omitempty"`    // 错误元数据
}

// NewStandardErrorResponse 创建标准错误响应
func NewStandardErrorResponse(err error) *StandardErrorResponse {
	if err == nil {
		return nil
	}

	// 使用 Kratos 的错误解析
	e := errors.FromError(err)
	if e == nil {
		return &StandardErrorResponse{
			Code:    500,
			Reason:  "INTERNAL_ERROR",
			Message: GetFriendlyErrorMessage("INTERNAL_ERROR"),
		}
	}

	// 获取友好的错误消息
	message := GetFriendlyErrorMessage(e.Reason)
	if message == "" {
		message = e.Message
	}

	return &StandardErrorResponse{
		Code:    int(e.Code),
		Reason:  e.Reason,
		Message: message,
		Meta: map[string]string{
			"request_id": getRequestID(),
			"timestamp":  getCurrentTime(),
		},
	}
}

// Helper functions (在实际项目中可能需要从上下文或请求中获取)
func getRequestID() string {
	// 这里应该从请求上下文中获取真实的请求ID
	// 现在返回模拟值
	return "req_" + "xxxxxx"
}

func getCurrentTime() string {
	// 这里应该返回当前时间
	// 现在返回模拟值
	return "2023-01-01T00:00:00Z"
}
