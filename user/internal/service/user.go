package service

import (
	"context"
	"strconv"

	v1 "user/api/user/v1"
	"user/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ExtractUserID 从 HTTP 请求上下文中提取用户ID（由Nginx JWT校验后设置）
func ExtractUserID(ctx context.Context, logger *log.Helper) (int64, error) {
	// 从 HTTP 请求中获取用户ID（由Nginx JWT校验后设置）
	req, ok := http.RequestFromServerContext(ctx)
	if !ok {
		logger.Log(log.LevelWarn, "Failed to get request from context")
		return 0, biz.ErrInvalidToken
	}

	// 从 X-User-ID 头获取用户ID（由Nginx设置）
	userIDStr := req.Header.Get("X-User-ID")
	if userIDStr == "" {
		logger.Log(log.LevelWarn, "No X-User-ID header provided")
		return 0, biz.ErrInvalidToken
	}

	// 解析用户ID
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		logger.Log(log.LevelWarn, "Invalid X-User-ID format: ", userIDStr)
		return 0, biz.ErrInvalidToken
	}

	logger.Log(log.LevelInfo, "User extracted from header, userID: ", userID)
	return userID, nil
}

// UserService 实现 UserService 接口
type UserService struct {
	v1.UnimplementedUserServiceServer

	userUsecase *biz.UserUsecase
	logger      *log.Helper
}

// NewUserService 创建 UserService 实例
func NewUserService(userUsecase *biz.UserUsecase, logger log.Logger) *UserService {
	return &UserService{
		userUsecase: userUsecase,
		logger:      log.NewHelper(logger),
	}
}

// GetCurrentUser 获取当前用户资料
func (s *UserService) GetCurrentUser(ctx context.Context, req *v1.GetCurrentUserRequest) (*v1.GetCurrentUserResponse, error) {
	s.logger.Log(log.LevelInfo, "Received GetCurrentUser request")

	// 提取并验证用户ID
	userID, err := ExtractUserID(ctx, s.logger)
	if err != nil {
		s.logger.Log(log.LevelError, "GetCurrentUser authentication failed: ", err)
		return &v1.GetCurrentUserResponse{}, nil
	}

	// 获取用户信息
	user, err := s.userUsecase.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Log(log.LevelError, "GetCurrentUser failed: ", err)
		return &v1.GetCurrentUserResponse{}, nil
	}

	return &v1.GetCurrentUserResponse{
		Id:        user.ID,
		Email:     user.Email,
		Nickname:  user.Nickname,
		AvatarUrl: user.AvatarURL,
		IsPremium: user.IsPremium == 1,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}, nil
}

// UpdateCurrentUser 更新当前用户资料
func (s *UserService) UpdateCurrentUser(ctx context.Context, req *v1.UpdateCurrentUserRequest) (*v1.UpdateCurrentUserResponse, error) {
	s.logger.Log(log.LevelInfo, "Received UpdateCurrentUser request")

	// 提取并验证用户ID
	userID, err := ExtractUserID(ctx, s.logger)
	if err != nil {
		s.logger.Log(log.LevelError, "UpdateCurrentUser authentication failed: ", err)
		return &v1.UpdateCurrentUserResponse{}, nil
	}

	// 构建更新请求
	updateReq := &biz.UpdateUserRequest{
		Nickname:  &req.Nickname,
		AvatarURL: &req.AvatarUrl,
	}

	// 调用业务逻辑更新用户信息
	err = s.userUsecase.UpdateUser(ctx, userID, updateReq)
	if err != nil {
		s.logger.Log(log.LevelError, "UpdateCurrentUser failed: ", err)
		return &v1.UpdateCurrentUserResponse{}, nil
	}

	// 获取更新后的用户信息
	user, err := s.userUsecase.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Log(log.LevelError, "Failed to get updated user info: ", err)
		return &v1.UpdateCurrentUserResponse{}, nil
	}

	return &v1.UpdateCurrentUserResponse{
		Id:        user.ID,
		Email:     user.Email,
		Nickname:  user.Nickname,
		AvatarUrl: user.AvatarURL,
		IsPremium: user.IsPremium == 1,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}, nil
}
