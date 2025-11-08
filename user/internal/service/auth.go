package service

import (
	"context"
	"regexp"

	v1 "user/api/auth/v1"
	"user/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"user/internal/pkg/tracing"
	error_reason "user/api/error_reason"
)

// AuthService 实现 AuthService 接口
type AuthService struct {
	v1.UnimplementedAuthServiceServer

	authUsecase *biz.AuthUsecase
	userUsecase *biz.UserUsecase
	logger      *log.Helper
}

// emailRegex 邮箱格式正则表达式
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// validateEmail 验证邮箱格式
//
// 参数:
//   - email: 待验证的邮箱地址
//
// 返回值:
//   - error: 验证失败时返回错误，验证成功时返回 nil
func validateEmail(email string) error {
	if email == "" {
		return error_reason.ErrorUserInvalidEmail("邮箱不能为空")
	}

	// 检查邮箱长度（最大254字符是RFC 5321规定的）
	if len(email) > 254 {
		return error_reason.ErrorUserInvalidEmail("邮箱长度不能超过254个字符")
	}

	// 检查邮箱格式
	if !emailRegex.MatchString(email) {
		return error_reason.ErrorUserInvalidEmail("邮箱格式不正确")
	}

	return nil
}

// NewAuthService 创建 AuthService 实例
func NewAuthService(authUsecase *biz.AuthUsecase, userUsecase *biz.UserUsecase, logger log.Logger) *AuthService {
	return &AuthService{
		authUsecase: authUsecase,
		userUsecase: userUsecase,
		logger:      log.NewHelper(logger),
	}
}

// SendRegisterCode 发送注册验证码
func (s *AuthService) SendRegisterCode(ctx context.Context, req *v1.SendRegisterCodeRequest) (*v1.SendRegisterCodeResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "AuthService.SendRegisterCode")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"operation": "send_register_code",
		"email": req.Email,
	})

	s.logger.WithContext(ctx).Infof("Received SendRegisterCode request for email: %s", req.Email)

	// 验证邮箱格式
	if err := validateEmail(req.Email); err != nil {
		s.logger.WithContext(ctx).Warnf("Invalid email format: %s, error: %v", req.Email, err)
		return nil, err
	}

	err := s.userUsecase.SendRegisterCode(ctx, req.Email)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("SendRegisterCode failed: %v", err)
		return nil, err
	}

	s.logger.WithContext(ctx).Info("SendRegisterCode completed successfully")
	return &v1.SendRegisterCodeResponse{
		Success: true,
		Message: "验证码发送成功",
	}, nil
}

// Register 用户注册
func (s *AuthService) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "AuthService.Register")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"operation": "register",
		"email": req.Email,
		"nickname": req.Nickname,
	})

	s.logger.WithContext(ctx).Infof("Received Register request for email: %s", req.Email)

	user, err := s.userUsecase.Register(ctx, req.Email, req.Password, req.Code, req.Nickname)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("Register failed: %v", err)
		return nil, err
	}

	s.logger.WithContext(ctx).Infof("Register completed successfully for user id: %d", user.ID)
	return &v1.RegisterResponse{
		Id:       user.ID,
		Email:    user.Email,
		Nickname: user.Nickname,
	}, nil
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "AuthService.Login")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"operation": "login",
		"email": req.Email,
	})

	s.logger.WithContext(ctx).Infof("Received Login request for email: %s", req.Email)

	tokenPair, err := s.userUsecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("Login failed: %v", err)
		return nil, err
	}

	s.logger.WithContext(ctx).Info("Login completed successfully")
	return &v1.LoginResponse{
		AccessToken:      tokenPair.AccessToken,
		AccessExpiresIn:  tokenPair.AccessExpiresIn,
		RefreshToken:     tokenPair.RefreshToken,
		RefreshExpiresIn: tokenPair.RefreshExpiresIn,
	}, nil
}

// RefreshToken 刷新Access Token
func (s *AuthService) RefreshToken(ctx context.Context, req *v1.RefreshTokenRequest) (*v1.RefreshTokenResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "AuthService.RefreshToken")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"operation": "refresh_token",
		"token_length": len(req.RefreshToken),
	})

	s.logger.WithContext(ctx).Info("Received RefreshToken request")

	tokenPair, err := s.authUsecase.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("RefreshToken failed: %v", err)
		return nil, err
	}

	s.logger.WithContext(ctx).Info("RefreshToken completed successfully")
	return &v1.RefreshTokenResponse{
		AccessToken:     tokenPair.AccessToken,
		AccessExpiresIn: tokenPair.AccessExpiresIn,
	}, nil
}

// Logout 用户登出
func (s *AuthService) Logout(ctx context.Context, req *v1.LogoutRequest) (*v1.LogoutResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "AuthService.Logout")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"operation": "logout",
		"token_length": len(req.RefreshToken),
	})

	s.logger.WithContext(ctx).Info("Received Logout request")

	err := s.authUsecase.Logout(ctx, req.RefreshToken)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("Logout failed: %v", err)
		return nil, err
	}

	s.logger.WithContext(ctx).Info("Logout completed successfully")
	return &v1.LogoutResponse{
		Success: true,
		Message: "登出成功",
	}, nil
}
