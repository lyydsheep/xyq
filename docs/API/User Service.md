# User & Auth API 文档

## 基础信息
**基础路径:** `/v1/auth`, `/v1/user`  
**鉴权机制:** 基于 JWT 的长/短 Token 机制 (Access Token 用于业务请求，Refresh Token 用于刷新 Access Token)  
**认证约定:** 
- AuthService接口：不需要认证
- UserService接口：从HTTP Header的`X-User-ID`获取用户ID（由Nginx JWT校验后设置）
**用户身份识别:** UserService通过`X-User-ID`Header获取用户ID，AuthService通过用户名密码或token进行认证

---

## AuthService 接口

### AuthService_SendRegisterCode
● **POST**  
● `/v1/auth/send-code`  
● **功能描述:** 发送注册邮箱验证码，支持60秒频率限制

● **请求 Body:**
```json
{
    "email": "string"
}
```

● **成功响应 (200 OK):**
```json
{
    "success": true,
    "message": "验证码发送成功"
}
```

● **邮箱格式错误（HTTP 状态码 400）**
```json
{
    "code": 400,
    "reason": "USER_INVALID_EMAIL", 
    "message": "邮箱不能为空",
    "metadata": {}
}
```

● **邮箱已注册（HTTP 状态码 409）**
```json
{
    "code": 409,
    "reason": "USER_EMAIL_ALREADY_EXISTS",
    "message": "该邮箱已被注册",
    "metadata": {}
}
```

● **请求过快（HTTP 状态码 429）**
```json
{
    "code": 429,
    "reason": "USER_TOO_MANY_REQUESTS",
    "message": "请求过于频繁，请稍后再试",
    "metadata": {}
}
```

● **其他错误响应**
- HTTP 500: `USER_DATABASE_ERROR` - 频率限制检查失败
- HTTP 500: `USER_INTERNAL_ERROR` - 邮件发送失败或验证码存储失败

---

### AuthService_Register
● **POST**  
● `/v1/auth/register`  
● **功能描述:** 用户注册，需要提供正确的邮箱验证码

● **请求 Body:**
```json
{
    "email": "string",
    "password": "string", 
    "code": "string",
    "nickname": "string"
}
```

● **成功响应 (200 OK):**
```json
{
    "id": 12345,
    "email": "user@example.com", 
    "nickname": "故事创造者"
}
```

● **参数缺失（HTTP 状态码 400）**
```json
{
    "code": 400,
    "reason": "USER_INVALID_REQUEST",
    "message": "邮箱、密码和验证码为必填项",
    "metadata": {}
}
```

● **密码过短（HTTP 状态码 400）**
```json
{
    "code": 400,
    "reason": "USER_INVALID_REQUEST", 
    "message": "密码长度至少为6位",
    "metadata": {}
}
```

● **验证码错误（HTTP 状态码 400）**
```json
{
    "code": 400,
    "reason": "USER_INVALID_VERIFICATION_CODE",
    "message": "验证码错误",
    "metadata": {}
}
```

● **验证码过期（HTTP 状态码 400）**
```json
{
    "code": 400,
    "reason": "USER_VERIFICATION_CODE_EXPIRED",
    "message": "验证码已过期",
    "metadata": {}
}
```

● **邮箱已注册（HTTP 状态码 409）**
```json
{
    "code": 409,
    "reason": "USER_EMAIL_ALREADY_EXISTS",
    "message": "该邮箱已被注册",
    "metadata": {}
}
```

● **其他错误响应**
- HTTP 500: `USER_DATABASE_ERROR` - 用户创建失败
- HTTP 500: `USER_INTERNAL_ERROR` - 密码加密失败

---

### AuthService_Login
● **POST**  
● `/v1/auth/login`  
● **功能描述:** 用户登录，返回JWT Access Token和Refresh Token

● **请求 Body:**
```json
{
    "email": "string",
    "password": "string"
}
```

● **成功响应 (200 OK):**
```json
{
    "access_token": "string",
    "access_expires_in": 3600,
    "refresh_token": "string", 
    "refresh_expires_in": 604800
}
```

● **参数缺失（HTTP 状态码 400）**
```json
{
    "code": 400,
    "reason": "USER_INVALID_REQUEST",
    "message": "邮箱和密码为必填项",
    "metadata": {}
}
```

● **认证失败（HTTP 状态码 401）**
```json
{
    "code": 401,
    "reason": "USER_INVALID_CREDENTIALS",
    "message": "用户名或密码错误",
    "metadata": {}
}
```

● **请求过多（HTTP 状态码 429）**
```json
{
    "code": 429,
    "reason": "USER_LOGIN_TOO_MANY",
    "message": "登录尝试过于频繁，请稍后再试",
    "metadata": {}
}
```

● **其他错误响应**
- HTTP 500: `USER_DATABASE_ERROR` - 用户查询失败或令牌存储失败
- HTTP 500: `USER_INTERNAL_ERROR` - 访问令牌或刷新令牌生成失败

---

### AuthService_RefreshToken
● **POST**  
● `/v1/auth/refresh`  
● **功能描述:** 刷新Access Token，原子性操作确保安全

● **请求 Body:**
```json
{
    "refresh_token": "string"
}
```

● **成功响应 (200 OK):**
```json
{
    "access_token": "string",
    "access_expires_in": 3600
}
```

● **Token无效（HTTP 状态码 401）**
```json
{
    "code": 401,
    "reason": "USER_REFRESH_TOKEN_INVALID",
    "message": "刷新令牌无效",
    "metadata": {}
}
```

● **其他错误响应**
- HTTP 500: `USER_DATABASE_ERROR` - 令牌刷新失败
- HTTP 500: `USER_INTERNAL_ERROR` - 访问令牌或刷新令牌生成失败

---

### AuthService_Logout
● **POST**  
● `/v1/auth/logout`  
● **功能描述:** 用户登出，使Refresh Token失效

● **请求 Body:**
```json
{
    "refresh_token": "string"
}
```

● **成功响应 (200 OK):**
```json
{
    "success": true,
    "message": "登出成功"
}
```

● **Token无效（HTTP 状态码 401）**
```json
{
    "code": 401,
    "reason": "USER_REFRESH_TOKEN_INVALID",
    "message": "刷新令牌无效",
    "metadata": {}
}
```

● **其他错误响应**
- HTTP 500: `USER_DATABASE_ERROR` - 令牌删除失败

---

## UserService 接口

### UserService_GetCurrentUser
● **GET**  
● `/v1/user/profile`  
● **功能描述:** 获取当前用户资料

● **请求 Headers:**
```
X-User-ID: 12345
```

● **成功响应 (200 OK):**
```json
{
    "id": 12345,
    "email": "user@example.com",
    "nickname": "故事创造者",
    "avatar_url": "https://example.com/avatar.jpg",
    "is_premium": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
}
```

● **认证失败（HTTP 状态码 401）**
```json
{
    "code": 401,
    "reason": "USER_INVALID_TOKEN",
    "message": "用户认证信息无效",
    "metadata": {}
}
```

● **用户不存在（HTTP 状态码 404）**
```json
{
    "code": 404,
    "reason": "USER_NOT_FOUND",
    "message": "用户不存在",
    "metadata": {}
}
```

● **其他错误响应**
- HTTP 500: `USER_DATABASE_ERROR` - 用户查询失败

---

### UserService_UpdateCurrentUser
● **PUT**  
● `/v1/user/profile`  
● **功能描述:** 更新当前用户资料（昵称、头像）

● **请求 Headers:**
```
X-User-ID: 12345
```

● **请求 Body:**
```json
{
    "nickname": "string",
    "avatar_url": "string"
}
```

● **成功响应 (200 OK):**
```json
{
    "id": 12345,
    "email": "user@example.com",
    "nickname": "故事创造者",
    "avatar_url": "https://example.com/avatar.jpg", 
    "is_premium": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
}
```

● **认证失败（HTTP 状态码 401）**
```json
{
    "code": 401,
    "reason": "USER_INVALID_TOKEN",
    "message": "用户认证信息无效",
    "metadata": {}
}
```

● **昵称冲突（HTTP 状态码 409）**
```json
{
    "code": 409,
    "reason": "USER_NICKNAME_ALREADY_EXISTS",
    "message": "该昵称已被使用",
    "metadata": {}
}
```

● **其他错误响应**
- HTTP 500: `USER_DATABASE_ERROR` - 用户更新失败

---

## 错误响应格式

所有错误响应都遵循Kratos框架的标准格式：

```json
{
    "code": 400,
    "reason": "ERROR_REASON_CODE",
    "message": "用户友好的错误信息",
    "metadata": {}
}
```

## 详细错误码说明

### 认证相关错误 (401)
- `USER_INVALID_TOKEN`: 访问令牌无效
- `USER_TOKEN_EXPIRED`: 访问令牌已过期  
- `USER_INVALID_CREDENTIALS`: 用户名或密码错误
- `USER_REFRESH_TOKEN_INVALID`: 刷新令牌无效

### 请求参数错误 (400)
- `USER_INVALID_EMAIL`: 邮箱格式错误
- `USER_INVALID_VERIFICATION_CODE`: 验证码错误
- `USER_VERIFICATION_CODE_EXPIRED`: 验证码已过期
- `USER_INVALID_REQUEST`: 请求参数无效
- `USER_INVALID_NICKNAME`: 昵称格式错误

### 资源冲突 (409)
- `USER_EMAIL_ALREADY_EXISTS`: 邮箱已被注册
- `USER_NICKNAME_ALREADY_EXISTS`: 昵称已被使用

### 资源不存在 (404)
- `USER_NOT_FOUND`: 用户不存在
- `USER_PROFILE_NOT_FOUND`: 用户资料不存在

### 请求过于频繁 (429)
- `USER_TOO_MANY_REQUESTS`: 请求过于频繁
- `USER_LOGIN_TOO_MANY`: 登录尝试过于频繁

### 系统错误 (500/503)
- `USER_DATABASE_ERROR`: 数据库操作失败
- `USER_INTERNAL_ERROR`: 服务内部错误
- `USER_SERVICE_UNAVAILABLE`: 用户服务暂时不可用

---

## 接口总览

| 接口名 | 方法 | 路径 | 描述 | 认证方式 |
|--------|------|------|------|----------|
| AuthService_SendRegisterCode | POST | `/v1/auth/send-code` | 发送注册验证码 | 无需认证 |
| AuthService_Register | POST | `/v1/auth/register` | 用户注册 | 无需认证 |
| AuthService_Login | POST | `/v1/auth/login` | 用户登录 | 无需认证 |
| AuthService_RefreshToken | POST | `/v1/auth/refresh` | 刷新Token | Refresh Token |
| AuthService_Logout | POST | `/v1/auth/logout` | 用户登出 | Refresh Token |
| UserService_GetCurrentUser | GET | `/v1/user/profile` | 获取用户资料 | X-User-ID Header |
| UserService_UpdateCurrentUser | PUT | `/v1/user/profile` | 更新用户资料 | X-User-ID Header |

---

## 重要说明

1. **验证码机制**: 生成的验证码为6位数字，有效期10分钟
2. **频率限制**: 发送验证码接口有60秒频率限制
3. **密码要求**: 密码长度至少6位
4. **Token有效期**: Access Token 1小时，Refresh Token 7天
5. **认证方式**: UserService使用X-User-ID Header而非JWT Token
6. **原子性操作**: Token刷新使用事务确保原子性
