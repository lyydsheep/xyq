package service

import (
	"context"
	"net/http"
	error2 "user/api/error_reason"
	v1 "user/api/user/v1"
)

// UserServiceErrorHandler 用户服务的错误处理示例
type UserServiceErrorHandler struct{}

// GetCurrentUser 获取当前用户（示例）
func (h *UserServiceErrorHandler) GetCurrentUser(ctx context.Context, req *v1.GetCurrentUserRequest) (*v1.GetCurrentUserResponse, error) {
	// 模拟不同类型的错误情况

	// 1. 用户未登录或Token无效
	// return nil, error_reason.ErrorUserInvalidToken("用户未登录或Token无效")

	// 2. 用户不存在
	// return nil, error_reason.ErrorUserNotFound("用户不存在")

	// 3. 数据库错误
	// return nil, error_reason.ErrorUserDatabaseError("数据库连接失败")

	// 4. 服务暂时不可用
	// return nil, error_reason.ErrorUserServiceUnavailable("用户服务维护中")

	// 正常情况
	return &v1.GetCurrentUserResponse{
		Id:        1,
		Email:     "user@example.com",
		Nickname:  "测试用户",
		IsPremium: false,
	}, nil
}

// UpdateCurrentUser 更新当前用户（示例）
func (h *UserServiceErrorHandler) UpdateCurrentUser(ctx context.Context, req *v1.UpdateCurrentUserRequest) (*v1.UpdateCurrentUserResponse, error) {
	// 模拟不同类型的错误情况

	// 1. 昵称格式不正确
	// return nil, error_reason.ErrorUserInvalidNickname("昵称格式不正确")

	// 2. 昵称已被使用
	// return nil, error_reason.ErrorUserNicknameAlreadyExists("该昵称已被使用")

	// 3. 请求过于频繁
	// return nil, error_reason.ErrorUserTooManyRequests("更新操作过于频繁，请稍后再试")

	// 正常情况
	return &v1.UpdateCurrentUserResponse{
		Id:        1,
		Email:     "user@example.com",
		Nickname:  req.Nickname,
		IsPremium: false,
	}, nil
}

// AuthServiceErrorHandler 认证服务的错误处理示例
type AuthServiceErrorHandler struct{}

// Login 登录（示例）
func (h *AuthServiceErrorHandler) Login(ctx context.Context, req interface{}) (interface{}, error) {
	// 模拟不同类型的错误情况

	// 1. 用户名或密码错误
	// return nil, error_reason.ErrorAuthInvalidCredentials("用户名或密码错误")

	// 2. 登录尝试次数过多
	// return nil, error_reason.ErrorAuthLoginBlocked("登录尝试次数过多，请稍后再试")

	// 3. 请求过于频繁
	// return nil, error_reason.ErrorAuthTooManyRequests("登录请求过于频繁")

	// 正常情况
	return map[string]interface{}{
		"access_token":  "mock_access_token",
		"refresh_token": "mock_refresh_token",
	}, nil
}

// Register 注册（示例）
func (h *AuthServiceErrorHandler) Register(ctx context.Context, req interface{}) (interface{}, error) {
	// 模拟不同类型的错误情况

	// 1. 邮箱格式不正确
	// return nil, error_reason.ErrorAuthInvalidEmail("邮箱格式不正确")

	// 2. 邮箱已被注册
	// return nil, error_reason.ErrorAuthEmailExists("该邮箱已被注册")

	// 正常情况
	return map[string]interface{}{
		"message": "注册成功",
	}, nil
}

// HTTPErrorHandler HTTP错误处理中间件示例
func HTTPErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				// 记录恐慌
				// logger.Error("panic", log.Any("panic", r))

				// 返回友好的错误响应
				errorResponse := NewStandardErrorResponse(error2.ErrorUserInternalError("服务内部错误"))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(errorResponse.Code)
				w.Write([]byte("{\"error_reason\":\"服务内部错误\"}"))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// ErrorAssertionExample 错误断言使用示例
func ErrorAssertionExample(err error) string {
	if error2.IsUserInvalidToken(err) {
		return "检测到无效Token错误"
	}

	if error2.IsUserNotFound(err) {
		return "检测到用户不存在错误"
	}

	if error2.IsUserEmailAlreadyExists(err) {
		return "检测到邮箱已存在错误"
	}

	if error2.IsUserDatabaseError(err) {
		return "检测到数据库错误"
	}

	if error2.IsAuthInvalidCredentials(err) {
		return "检测到认证凭证无效错误"
	}

	return "未知错误类型"
}
