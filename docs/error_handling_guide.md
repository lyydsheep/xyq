# User微服务错误处理改进指南

## 概述

本指南详细说明了user微服务错误处理的改进方案，基于Kratos框架的最佳实践，为用户提供更加友好、标准、可维护的错误响应机制。

## 改进前后对比

### 改进前问题
1. **错误定义不统一**：直接在Go代码中定义错误，缺乏标准化
2. **错误码不统一**：部分使用业务前缀，部分使用系统前缀
3. **错误信息不友好**：英文错误信息对中文用户不友好
4. **缺乏错误元数据**：无法提供额外的错误上下文信息
5. **错误断言困难**：没有生成错误检查函数
6. **不符合最佳实践**：未遵循Kratos错误处理规范

### 改进后优势
1. **类型安全**：通过proto定义错误，自动生成代码
2. **标准统一**：与gRPC状态码一致，遵循Kratos规范
3. **用户友好**：中文错误提示，清晰易懂
4. **可扩展**：支持错误元数据和详细信息
5. **开发友好**：提供IsXXXError()断言函数
6. **易于维护**：集中的错误定义和消息映射

## 核心功能

### 1. 错误定义（Proto）

#### 文件位置
- `user/api/user/v1/error_reason.proto`

#### 错误类型
- **UserService错误**：18种用户相关错误
- **AuthService错误**：12种认证相关错误
- **System错误**：7种系统级错误

#### 错误码格式
`{模块}_{状态码}_{序号}`

示例：
- `USER_401_01`: 无效Token
- `USER_409_01`: 邮箱已存在
- `USER_404_01`: 用户不存在

### 2. 自动生成的错误处理函数

#### 错误断言函数
```go
// 检查是否为无效Token错误
if v1.IsUserInvalidToken(err) {
    // 处理逻辑
}

// 检查是否为用户不存在错误
if v1.IsUserNotFound(err) {
    // 处理逻辑
}
```

#### 错误创建函数
```go
// 创建无效Token错误
return nil, v1.ErrorUserInvalidToken("用户未登录或Token无效")

// 创建用户不存在错误
return nil, v1.ErrorUserNotFound("用户不存在")
```

### 3. 友好的错误消息

#### 错误消息映射
文件：`user/internal/service/error_messages.go`

特点：
- 支持中文友好错误提示
- 集中管理，易于维护
- 支持多语言扩展

示例消息：
```
"USER_INVALID_TOKEN": "访问令牌无效，请重新登录"
"USER_EMAIL_ALREADY_EXISTS": "该邮箱已被注册"
"USER_DATABASE_ERROR": "数据库操作失败"
```

### 4. 标准错误响应格式

#### 结构定义
```go
type StandardErrorResponse struct {
    Code    int                    `json:"code"`              // HTTP状态码
    Reason  string                 `json:"reason"`            // 错误原因
    Message string                 `json:"message"`           // 用户友好的错误信息
    Details map[string]interface{} `json:"details,omitempty"` // 错误详情
    Meta    map[string]string      `json:"meta,omitempty"`    // 错误元数据
}
```

#### 示例响应
```json
{
  "code": 401,
  "reason": "USER_INVALID_TOKEN",
  "message": "访问令牌无效，请重新登录",
  "meta": {
    "request_id": "req_xxxxxx",
    "timestamp": "2023-01-01T00:00:00Z"
  }
}
```

## 使用指南

### 1. 在服务中使用

```go
// UserService 示例
func (s *UserService) GetCurrentUser(ctx context.Context, req *v1.GetCurrentUserRequest) (*v1.GetCurrentUserResponse, error) {
    // 检查用户认证
    if !isAuthenticated(ctx) {
        return nil, v1.ErrorUserInvalidToken("用户未登录或Token无效")
    }
    
    // 查找用户
    user, err := s.repo.FindUserByID(userID)
    if err != nil {
        return nil, v1.ErrorUserDatabaseError("数据库查询失败")
    }
    
    if user == nil {
        return nil, v1.ErrorUserNotFound("用户不存在")
    }
    
    return &v1.GetCurrentUserResponse{
        Id:       user.ID,
        Email:    user.Email,
        Nickname: user.Nickname,
    }, nil
}
```

### 2. 错误断言

```go
// 在业务逻辑中进行错误分类处理
func handleUserError(err error) {
    switch {
    case v1.IsUserInvalidToken(err):
        // 重定向到登录页面
        http.Redirect(w, r, "/login", http.StatusUnauthorized)
        
    case v1.IsUserNotFound(err):
        // 显示404页面
        http.NotFound(w, r)
        
    case v1.IsUserEmailAlreadyExists(err):
        // 返回错误信息给前端
        http.Error(w, "该邮箱已被注册", http.StatusConflict)
        
    default:
        // 处理其他错误
        http.Error(w, "服务器内部错误", http.StatusInternalServerError)
    }
}
```

### 3. HTTP中间件

```go
// 错误处理中间件
func ErrorMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if rec := recover(); rec != nil {
                // 处理panic
                errorResponse := service.NewStandardErrorResponse(
                    v1.ErrorUserInternalError("服务内部错误")
                )
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(errorResponse.Code)
                json.NewEncoder(w).Encode(errorResponse)
            }
        }()
        
        next.ServeHTTP(w, r)
    })
}
```

## 生成和维护

### 1. 生成错误代码

```bash
# 生成错误处理代码
make errors

# 或生成完整的API代码
make api
```

### 2. 添加新错误

1. 在`error_reason.proto`中添加新的错误枚举
2. 运行`make errors`重新生成代码
3. 在`error_messages.go`中添加对应的友好消息

### 3. 错误消息本地化

当前只支持中文，可以扩展支持多语言：

```go
var ErrorMessageMap = map[string]map[string]string{
    "zh": {
        "USER_INVALID_TOKEN": "访问令牌无效，请重新登录",
        "USER_EMAIL_ALREADY_EXISTS": "该邮箱已被注册",
    },
    "en": {
        "USER_INVALID_TOKEN": "Access token is invalid, please login again",
        "USER_EMAIL_ALREADY_EXISTS": "Email already exists",
    },
}
```

## 最佳实践

### 1. 错误分类
- **4xx**：客户端错误，提供用户友好的提示
- **5xx**：服务端错误，隐藏技术细节，提供通用提示
- **429**：频率限制错误，提供重试时间建议

### 2. 错误消息原则
- 使用中文，避免技术术语
- 给出明确的解决建议
- 保持一致性

### 3. 错误记录
- 记录详细的错误信息和上下文
- 使用结构化日志
- 保护用户隐私信息

## 迁移指南

### 从旧系统迁移

1. **识别现有错误**：列出所有业务错误
2. **映射到新错误**：将旧错误映射到新的错误码
3. **更新代码**：替换错误创建和断言逻辑
4. **测试验证**：确保所有错误场景正确处理

### 兼容性考虑

- 保持HTTP状态码不变
- 新增错误码不影响现有业务
- 可以同时支持新旧错误格式

## 监控和告警

### 1. 错误统计
- 错误类型分布
- 错误频率监控
- 用户体验指标

### 2. 告警设置
- 高频错误告警
- 系统错误告警
- 用户行为异常告警

## 总结

通过这次改进，user微服务的错误处理能力得到了显著提升：

1. **标准化**：符合Kratos最佳实践
2. **友好性**：中文错误提示，提升用户体验
3. **可维护性**：集中的错误定义和消息管理
4. **开发效率**：自动生成错误处理函数
5. **可扩展性**：支持错误元数据和多语言

这套错误处理机制将为user微服务提供更加健壮、用户友好的错误响应能力。
