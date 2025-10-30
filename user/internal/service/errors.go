package service

import (
	"net/http"

	biz "user/internal/biz"
)

// 业务错误码常量
const (
	// 用户相关错误
	USER_ERR_CODE_INVALID    = "USER_40001" // 验证码错误或已过期
	USER_ERR_EMAIL_FORMAT    = "USER_40002" // 邮箱格式不正确
	USER_ERR_INVALID_CREDS   = "USER_40103" // 用户名或密码错误
	USER_ERR_TOKEN_INVALID   = "USER_40101" // Access Token 无效或缺失
	USER_ERR_REFRESH_INVALID = "USER_40102" // Refresh Token 无效或缺失
	USER_ERR_EMAIL_EXISTS    = "USER_40901" // 邮箱已被注册
	USER_ERR_TOO_MANY_REQ    = "USER_42901" // 请求过于频繁
	USER_ERR_NOT_FOUND       = "USER_40401" // 用户不存在

	// 系统错误
	SYS_ERR_DB = "SYS_50001" // 数据库操作失败
)

// ErrorMapping 业务错误到错误码的映射
var ErrorMapping = map[error]string{
	// 业务层错误
	biz.ErrInvalidCredentials:      USER_ERR_INVALID_CREDS,
	biz.ErrEmailAlreadyExists:      USER_ERR_EMAIL_EXISTS,
	biz.ErrInvalidVerificationCode: USER_ERR_CODE_INVALID,
	biz.ErrVerificationCodeExpired: USER_ERR_CODE_INVALID,
	biz.ErrTooManyRequests:         USER_ERR_TOO_MANY_REQ,

	// 认证错误
	biz.ErrInvalidToken: USER_ERR_TOKEN_INVALID,
	biz.ErrTokenExpired: USER_ERR_TOKEN_INVALID,
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// MapErrorToHTTP 将业务错误映射为HTTP状态码和错误码
func MapErrorToHTTP(err error) (int, string, string) {
	if err == nil {
		return http.StatusOK, "0", "success"
	}

	// 查找业务错误码
	if businessCode, ok := ErrorMapping[err]; ok {
		// 根据错误类型返回对应的HTTP状态码
		switch err {
		case biz.ErrInvalidCredentials, biz.ErrInvalidToken, biz.ErrTokenExpired:
			return http.StatusUnauthorized, businessCode, err.Error()
		case biz.ErrEmailAlreadyExists:
			return http.StatusConflict, businessCode, err.Error()
		case biz.ErrTooManyRequests:
			return http.StatusTooManyRequests, businessCode, err.Error()
		default:
			return http.StatusBadRequest, businessCode, err.Error()
		}
	}

	// 未知错误，返回500
	return http.StatusInternalServerError, SYS_ERR_DB, "internal server error"
}

// ToErrorResponse 将错误转换为标准错误响应
func ToErrorResponse(err error) *ErrorResponse {
	httpCode, businessCode, message := MapErrorToHTTP(err)
	if httpCode == http.StatusOK {
		return nil
	}
	return &ErrorResponse{
		Code:    businessCode,
		Message: message,
	}
}

// SuccessResponse 成功响应结构
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}, message ...string) *SuccessResponse {
	msg := "success"
	if len(message) > 0 {
		msg = message[0]
	}
	return &SuccessResponse{
		Success: true,
		Data:    data,
		Message: msg,
	}
}
