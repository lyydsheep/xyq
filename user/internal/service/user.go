package service

import (
	"context"
	"strconv"

	v1 "user/api/user/v1"
	"user/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
	"google.golang.org/protobuf/types/known/timestamppb"
	"user/internal/pkg/tracing"
	error_reason "user/api/error_reason"
)

// ExtractUserID 从 HTTP 请求上下文中提取用户ID（由Nginx JWT校验后设置）
func ExtractUserID(ctx context.Context, logger *log.Helper) (int64, error) {
	ctx, span := tracing.StartSpan(ctx, "Service.ExtractUserID")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"operation": "extract_user_id",
	})

	// 从 HTTP 请求中获取用户ID（由Nginx JWT校验后设置）
	req, ok := http.RequestFromServerContext(ctx)
	if !ok {
		logger.WithContext(ctx).Warn("Failed to get request from context")
		return 0, error_reason.ErrorUserInvalidToken("用户认证信息无效")
	}

	// 从 X-User-ID 头获取用户ID（由Nginx设置）
	userIDStr := req.Header.Get("X-User-ID")
	if userIDStr == "" {
		logger.WithContext(ctx).Warn("No X-User-ID header provided")
		return 0, error_reason.ErrorUserInvalidToken("用户认证信息缺失")
	}

	// 解析用户ID
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		logger.WithContext(ctx).Warnf("Invalid X-User-ID format: %s", userIDStr)
		return 0, error_reason.ErrorUserInvalidToken("用户ID格式无效")
	}

	logger.WithContext(ctx).Infof("User extracted from header, userID: %d", userID)
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
	ctx, span := tracing.StartSpan(ctx, "UserService.GetCurrentUser")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"operation": "get_current_user",
	})

	s.logger.WithContext(ctx).Info("Received GetCurrentUser request")

	userID, err := ExtractUserID(ctx, s.logger)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("GetCurrentUser authentication failed: %v", err)
		return nil, err
	}

	user, err := s.userUsecase.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("GetCurrentUser failed: %v", err)
		return nil, err
	}

	s.logger.WithContext(ctx).Infof("Successfully retrieved current user with id: %d", user.ID)
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
	ctx, span := tracing.StartSpan(ctx, "UserService.UpdateCurrentUser")
	defer span.End()

	tracing.AddSpanTags(ctx, map[string]interface{}{
		"operation": "update_current_user",
	})

	s.logger.WithContext(ctx).Info("Received UpdateCurrentUser request")

	userID, err := ExtractUserID(ctx, s.logger)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("UpdateCurrentUser authentication failed: %v", err)
		return &v1.UpdateCurrentUserResponse{}, nil
	}

	updateReq := &biz.UpdateUserRequest{
		Nickname:  &req.Nickname,
		AvatarURL: &req.AvatarUrl,
	}

	err = s.userUsecase.UpdateUser(ctx, userID, updateReq)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("UpdateCurrentUser failed: %v", err)
		return &v1.UpdateCurrentUserResponse{}, nil
	}

	user, err := s.userUsecase.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("Failed to get updated user info: %v", err)
		return &v1.UpdateCurrentUserResponse{}, nil
	}

	s.logger.WithContext(ctx).Infof("Successfully updated current user with id: %d", user.ID)
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
