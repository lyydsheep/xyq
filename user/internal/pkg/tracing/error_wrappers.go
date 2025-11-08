package tracing

import (
	"context"

	error_reason "user/api/error_reason"

	errors "github.com/go-kratos/kratos/v2/errors"
)

// WrapErrorWithTrace 通用错误包装函数，为已有错误添加追踪信息
func WrapErrorWithTrace(ctx context.Context, err *errors.Error) *errors.Error {
	if err == nil {
		return nil
	}

	traceInfo := ExtractTraceInfo(ctx)

	// 如果有追踪信息且错误还没有这些信息，则添加
	if traceInfo.TraceID != "" || traceInfo.SpanID != "" {
		// 检查错误是否已经有追踪信息
		existingMetadata := err.Metadata
		hasTraceID := false
		hasSpanID := false

		if existingMetadata != nil {
			_, hasTraceID = existingMetadata["traceid"]
			_, hasSpanID = existingMetadata["spanid"]
		}

		// 创建新的 metadata
		metadata := make(map[string]string)

		// 添加现有 metadata
		if existingMetadata != nil {
			for k, v := range existingMetadata {
				metadata[k] = v
			}
		}

		// 添加追踪信息（如果还没有）
		if !hasTraceID && traceInfo.TraceID != "" {
			metadata["traceid"] = traceInfo.TraceID
		}
		if !hasSpanID && traceInfo.SpanID != "" {
			metadata["spanid"] = traceInfo.SpanID
		}

		err = err.WithMetadata(metadata)
	}

	return err
}

// UserError 包装用户相关错误，添加追踪信息
func UserError(ctx context.Context, reason, format string, args ...interface{}) *errors.Error {
	var err *errors.Error

	switch reason {
	case "USER_INVALID_TOKEN":
		err = error_reason.ErrorUserInvalidToken(format, args...)
	case "USER_TOKEN_EXPIRED":
		err = error_reason.ErrorUserTokenExpired(format, args...)
	case "USER_INVALID_CREDENTIALS":
		err = error_reason.ErrorUserInvalidCredentials(format, args...)
	case "USER_REFRESH_TOKEN_INVALID":
		err = error_reason.ErrorUserRefreshTokenInvalid(format, args...)
	case "USER_INVALID_EMAIL":
		err = error_reason.ErrorUserInvalidEmail(format, args...)
	case "USER_INVALID_VERIFICATION_CODE":
		err = error_reason.ErrorUserInvalidVerificationCode(format, args...)
	case "USER_VERIFICATION_CODE_EXPIRED":
		err = error_reason.ErrorUserVerificationCodeExpired(format, args...)
	case "USER_INVALID_REQUEST":
		err = error_reason.ErrorUserInvalidRequest(format, args...)
	case "USER_INVALID_NICKNAME":
		err = error_reason.ErrorUserInvalidNickname(format, args...)
	case "USER_EMAIL_ALREADY_EXISTS":
		err = error_reason.ErrorUserEmailAlreadyExists(format, args...)
	case "USER_NICKNAME_ALREADY_EXISTS":
		err = error_reason.ErrorUserNicknameAlreadyExists(format, args...)
	case "USER_NOT_FOUND":
		err = error_reason.ErrorUserNotFound(format, args...)
	case "USER_PROFILE_NOT_FOUND":
		err = error_reason.ErrorUserProfileNotFound(format, args...)
	case "USER_TOO_MANY_REQUESTS":
		err = error_reason.ErrorUserTooManyRequests(format, args...)
	case "USER_LOGIN_TOO_MANY":
		err = error_reason.ErrorUserLoginTooMany(format, args...)
	case "USER_DATABASE_ERROR":
		err = error_reason.ErrorUserDatabaseError(format, args...)
	case "USER_INTERNAL_ERROR":
		err = error_reason.ErrorUserInternalError(format, args...)
	case "USER_SERVICE_UNAVAILABLE":
		err = error_reason.ErrorUserServiceUnavailable(format, args...)
	default:
		err = error_reason.ErrorUserInvalidToken(format, args...)
	}

	return WrapErrorWithTrace(ctx, err)
}

// AuthError 包装认证相关错误，添加追踪信息
func AuthError(ctx context.Context, reason, format string, args ...interface{}) *errors.Error {
	var err *errors.Error

	switch reason {
	case "AUTH_INVALID_CREDENTIALS":
		err = error_reason.ErrorAuthInvalidCredentials(format, args...)
	case "AUTH_TOKEN_INVALID":
		err = error_reason.ErrorAuthTokenInvalid(format, args...)
	case "AUTH_TOKEN_EXPIRED":
		err = error_reason.ErrorAuthTokenExpired(format, args...)
	case "AUTH_REFRESH_TOKEN_INVALID":
		err = error_reason.ErrorAuthRefreshTokenInvalid(format, args...)
	case "AUTH_INVALID_REQUEST":
		err = error_reason.ErrorAuthInvalidRequest(format, args...)
	case "AUTH_INVALID_EMAIL":
		err = error_reason.ErrorAuthInvalidEmail(format, args...)
	case "AUTH_INVALID_CODE":
		err = error_reason.ErrorAuthInvalidCode(format, args...)
	case "AUTH_EMAIL_EXISTS":
		err = error_reason.ErrorAuthEmailExists(format, args...)
	case "AUTH_TOO_MANY_REQUESTS":
		err = error_reason.ErrorAuthTooManyRequests(format, args...)
	case "AUTH_LOGIN_BLOCKED":
		err = error_reason.ErrorAuthLoginBlocked(format, args...)
	case "AUTH_DATABASE_ERROR":
		err = error_reason.ErrorAuthDatabaseError(format, args...)
	case "AUTH_SERVICE_ERROR":
		err = error_reason.ErrorAuthServiceError(format, args...)
	case "AUTH_SERVICE_UNAVAILABLE":
		err = error_reason.ErrorAuthServiceUnavailable(format, args...)
	default:
		err = error_reason.ErrorAuthInvalidCredentials(format, args...)
	}

	return WrapErrorWithTrace(ctx, err)
}

// SystemError 包装系统相关错误，添加追踪信息
func SystemError(ctx context.Context, reason, format string, args ...interface{}) *errors.Error {
	var err *errors.Error

	switch reason {
	case "DATABASE_CONNECTION_ERROR":
		err = error_reason.ErrorDatabaseConnectionError(format, args...)
	case "DATABASE_OPERATION_ERROR":
		err = error_reason.ErrorDatabaseOperationError(format, args...)
	case "DATABASE_TIMEOUT_ERROR":
		err = error_reason.ErrorDatabaseTimeoutError(format, args...)
	case "SERVICE_UNAVAILABLE":
		err = error_reason.ErrorServiceUnavailable(format, args...)
	case "SERVICE_OVERLOADED":
		err = error_reason.ErrorServiceOverloaded(format, args...)
	case "REDIS_CONNECTION_ERROR":
		err = error_reason.ErrorRedisConnectionError(format, args...)
	case "EXTERNAL_SERVICE_ERROR":
		err = error_reason.ErrorExternalServiceError(format, args...)
	default:
		err = error_reason.ErrorServiceUnavailable(format, args...)
	}

	return WrapErrorWithTrace(ctx, err)
}
