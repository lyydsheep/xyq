package service

import (
	"context"

	v1 "user/api/auth/v1"
	"user/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// AuthService 实现 AuthService 接口
type AuthService struct {
	v1.UnimplementedAuthServiceServer

	authUsecase *biz.AuthUsecase
	userUsecase *biz.UserUsecase
	logger      *log.Helper
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
	s.logger.Log(log.LevelInfo, "Received SendRegisterCode request for email: ", req.Email)

	// 调用业务逻辑
	err := s.userUsecase.SendRegisterCode(ctx, req.Email)
	if err != nil {
		s.logger.Log(log.LevelError, "SendRegisterCode failed: ", err)
		return &v1.SendRegisterCodeResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.SendRegisterCodeResponse{
		Success: true,
		Message: "验证码发送成功",
	}, nil
}

// Register 用户注册
func (s *AuthService) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	s.logger.Log(log.LevelInfo, "Received Register request for email: ", req.Email)

	// 调用业务逻辑
	user, err := s.userUsecase.Register(ctx, req.Email, req.Password, req.Code, req.Nickname)
	if err != nil {
		s.logger.Log(log.LevelError, "Register failed: ", err)
		return &v1.RegisterResponse{
			Id:       0,
			Email:    "",
			Nickname: "",
		}, nil
	}

	return &v1.RegisterResponse{
		Id:       user.ID,
		Email:    user.Email,
		Nickname: user.Nickname,
	}, nil
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	s.logger.Log(log.LevelInfo, "Received Login request for email: ", req.Email)

	// 调用业务逻辑
	tokenPair, err := s.userUsecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		s.logger.Log(log.LevelError, "Login failed: ", err)
		return &v1.LoginResponse{
			AccessToken:      "",
			AccessExpiresIn:  0,
			RefreshToken:     "",
			RefreshExpiresIn: 0,
		}, nil
	}

	return &v1.LoginResponse{
		AccessToken:      tokenPair.AccessToken,
		AccessExpiresIn:  tokenPair.AccessExpiresIn,
		RefreshToken:     tokenPair.RefreshToken,
		RefreshExpiresIn: tokenPair.RefreshExpiresIn,
	}, nil
}

// RefreshToken 刷新Access Token
func (s *AuthService) RefreshToken(ctx context.Context, req *v1.RefreshTokenRequest) (*v1.RefreshTokenResponse, error) {
	s.logger.Log(log.LevelInfo, "Received RefreshToken request")

	// 调用业务逻辑
	tokenPair, err := s.authUsecase.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		s.logger.Log(log.LevelError, "RefreshToken failed: ", err)
		return &v1.RefreshTokenResponse{
			AccessToken:     "",
			AccessExpiresIn: 0,
		}, nil
	}

	return &v1.RefreshTokenResponse{
		AccessToken:     tokenPair.AccessToken,
		AccessExpiresIn: tokenPair.AccessExpiresIn,
	}, nil
}

// Logout 用户登出
func (s *AuthService) Logout(ctx context.Context, req *v1.LogoutRequest) (*v1.LogoutResponse, error) {
	s.logger.Log(log.LevelInfo, "Received Logout request")

	// 调用业务逻辑
	err := s.authUsecase.Logout(ctx, req.RefreshToken)
	if err != nil {
		s.logger.Log(log.LevelError, "Logout failed: ", err)
		return &v1.LogoutResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.LogoutResponse{
		Success: true,
		Message: "登出成功",
	}, nil
}
