# User微服务错误处理分析与改进方案

## 当前错误处理状况分析

### 1. 现有错误处理机制

**优点：**
- 已建立了基本的错误码体系（如USER_40001、SYS_50001等）
- 有错误到HTTP状态码的映射
- 使用了业务层错误定义

**问题：**
1. **缺乏结构化定义**：错误直接在Go代码中定义，没有使用proto规范
2. **错误码不统一**：部分使用业务前缀，部分使用系统前缀
3. **错误信息不够友好**：英文错误信息对中文用户不友好
4. **缺乏错误元数据**：无法提供额外的错误上下文信息
5. **错误断言困难**：没有生成错误检查函数
6. **标准化程度低**：不符合Kratos错误处理最佳实践

### 2. 现有错误码分析

**业务错误：**
- USER_40001: 验证码错误或已过期
- USER_40002: 邮箱格式不正确
- USER_40101: Access Token 无效或缺失
- USER_40102: Refresh Token 无效或缺失
- USER_40103: 用户名或密码错误
- USER_40401: 用户不存在
- USER_40901: 邮箱已被注册
- USER_42901: 请求过于频繁

**系统错误：**
- SYS_50001: 数据库操作失败

## Kratos错误处理最佳实践

### 标准错误响应结构
```json
{
  "code": 500,           // HTTP状态码
  "reason": "USER_NOT_FOUND",  // 业务错误原因
  "message": "用户未找到",      // 用户友好的错误信息
  "metadata": {           // 错误元数据
    "request_id": "xxx",
    "timestamp": "2023-xx-xx"
  }
}
```

### 核心优势
1. **类型安全**：通过proto定义错误
2. **自动生成**：生成错误断言函数
3. **标准统一**：与gRPC状态码一致
4. **可扩展性**：支持错误元数据
5. **开发友好**：提供IsXXXError()断言函数

## 改进方案

### 1. 错误定义标准化

使用proto定义所有业务错误：

```proto
syntax = "proto3";

package user.v1;

option go_package = "user/api/user/v1;v1";
option java_package = "user.api.user.v1";

import "errors/errors.proto";

// 错误枚举定义
enum ErrorReason {
  option (errors.default_code) = 500;
  
  // 认证相关错误 (401)
  USER_INVALID_TOKEN = 0 [(errors.code) = 401];
  USER_TOKEN_EXPIRED = 1 [(errors.code) = 401];
  USER_INVALID_CREDENTIALS = 2 [(errors.code) = 401];
  
  // 请求错误 (400)
  USER_INVALID_EMAIL = 3 [(errors.code) = 400];
  USER_INVALID_VERIFICATION_CODE = 4 [(errors.code) = 400];
  USER_VERIFICATION_CODE_EXPIRED = 5 [(errors.code) = 400];
  USER_INVALID_REQUEST = 6 [(errors.code) = 400];
  
  // 资源冲突 (409)
  USER_EMAIL_ALREADY_EXISTS = 7 [(errors.code) = 409];
  
  // 资源不存在 (404)
  USER_NOT_FOUND = 8 [(errors.code) = 404];
  
  // 请求过多 (429)
  USER_TOO_MANY_REQUESTS = 9 [(errors.code) = 429];
  
  // 系统错误 (500)
  USER_DATABASE_ERROR = 10 [(errors.code) = 500];
  USER_INTERNAL_ERROR = 11 [(errors.code) = 500];
}
```

### 2. 错误码体系设计

**错误码格式：** `{模块}_{状态码}_{序号}`
- 模块：USER（用户模块）、AUTH（认证模块）、SYS（系统）
- 状态码：标准HTTP状态码
- 序号：2位数字，用于区分同类型的不同错误

**错误码映射：**
- USER_401_01: 无效Token
- USER_401_02: Token过期
- USER_401_03: 凭证无效
- USER_400_01: 邮箱格式错误
- USER_400_02: 验证码错误
- USER_400_03: 验证码过期
- USER_400_04: 请求参数无效
- USER_409_01: 邮箱已存在
- USER_404_01: 用户不存在
- USER_429_01: 请求过于频繁
- USER_500_01: 数据库错误
- USER_500_02: 内部错误

### 3. 错误消息本地化

为每种错误提供中英文友好的错误消息：

```go
// 错误消息映射表
var ErrorMessages = map[string][2]string{
  "USER_INVALID_TOKEN": {
    "zh": "访问令牌无效，请重新登录",
    "en": "Access token is invalid, please login again"
  },
  "USER_EMAIL_ALREADY_EXISTS": {
    "zh": "该邮箱已被注册",
    "en": "Email already exists"
  },
  // ... 更多错误消息
}
```

### 4. 统一错误响应格式

```go
type StandardErrorResponse struct {
  Code    int                    `json:"code"`              // HTTP状态码
  Reason  string                 `json:"reason"`            // 错误原因
  Message string                 `json:"message"`           // 用户友好的错误信息
  Details map[string]interface{} `json:"details,omitempty"` // 错误详情
  Meta    map[string]string      `json:"meta,omitempty"`    // 错误元数据
}
```

### 5. 错误断言函数

生成标准化的错误断言函数：

```go
func IsUserInvalidToken(err error) bool {
  if err == nil {
    return false
  }
  e := errors.FromError(err)
  return e.Reason == ErrorReason_USER_INVALID_TOKEN.String() && e.Code == 401
}

func NewUserInvalidToken(format string, args ...interface{}) *errors.Error {
  return errors.New(401, ErrorReason_USER_INVALID_TOKEN.String(), 
    fmt.Sprintf(format, args...))
}
```

## 实施计划

1. **阶段一**：定义proto错误枚举并生成相关代码
2. **阶段二**：重构现有错误处理逻辑
3. **阶段三**：更新服务层的错误响应
4. **阶段四**：完善错误消息和元数据
5. **阶段五**：更新文档和测试

## 预期收益

1. **开发效率提升**：自动生成错误断言函数
2. **维护性改善**：统一的错误处理规范
3. **用户体验优化**：友好的中文错误提示
4. **监控能力增强**：结构化错误元数据
5. **调试效率提高**：详细的错误上下文信息

## 风险评估

1. **迁移成本**：需要重构现有错误处理代码
2. **兼容性**：确保API向后兼容
3. **学习成本**：团队需要学习新的错误处理模式

通过以上改进，user微服务将具备更加友好、标准、可维护的错误处理能力。
