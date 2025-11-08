# User & Auth API 文档

## 基础信息
**基础路径:** `/v1/auth`, `/v1/user`
**鉴权机制:** 基于 JWT 的长/短 Token 机制 (Access Token 用于业务请求，Refresh Token 用于刷新 Access Token)
**架构说明:** 本系统采用 **Nginx + 微服务** 的分层鉴权架构

## 认证架构总览

### 🔐 分层鉴权设计
```
┌─────────────┐     ┌──────────┐     ┌─────────────┐
│   客户端     │────▶│  Nginx   │────▶│  微服务     │
│             │     │  (验证)   │     │  (业务)     │
└─────────────┘     └──────────┘     └─────────────┘
```

### 📝 请求格式对照

#### 1. 客户端 → Nginx（入口请求）
| 接口类型 | 请求头 | 说明 |
|----------|--------|------|
| **UserService** | `Authorization: Bearer <access_token>` | 需要提供JWT Access Token |
| **AuthService** | 无特殊要求 | 直接发送请求即可 |

#### 2. Nginx → 微服务（转发请求）
| 接口类型 | 请求头 | 说明 |
|----------|--------|------|
| **UserService** | `X-User-ID: <user_id>` | Nginx提取用户ID后设置 |
| **AuthService** | 透传 | 直接转发原始请求 |

### 🎯 认证方式
- **AuthService接口**：Nginx无认证，微服务根据需要验证Refresh Token
- **UserService接口**：Nginx验证JWT Access Token，微服务从`X-User-ID`获取用户ID（由Nginx JWT校验后设置）

### 💡 完整请求示例

#### 示例1: 获取用户资料（需要JWT鉴权）
**步骤1：客户端发送请求给Nginx**
```bash
curl -X GET "https://api.example.com/v1/user/profile" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json"
```

**步骤2：Nginx验证并转发给微服务**
```
请求头变化:
- 删除: Authorization: Bearer eyJ...
- 添加: X-User-ID: 12345
```

**步骤3：微服务处理并返回**
```json
{
    "id": 12345,
    "email": "user@example.com",
    "nickname": "故事创造者"
}
```

#### 示例2: 用户登录（无需JWT鉴权）
**步骤1：客户端发送请求给Nginx**
```bash
curl -X POST "https://api.example.com/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

**步骤2：Nginx直接转发给微服务**
```
请求头: 无变化，直接透传
```

**步骤3：微服务处理并返回Token**
```json
{
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "access_expires_in": 3600,
    "refresh_expires_in": 604800
}
```

**用户身份识别:** UserService通过`X-User-ID`Header获取用户ID，AuthService通过用户名密码或Refresh Token进行认证

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

● **密码强度要求:**
  - 长度：8-16位字符
  - 必须包含至少一个数字（0-9）
  - 必须包含至少一个字母（a-z或A-Z）
  - 允许包含字母、数字和常见特殊字符

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

● **密码格式错误（HTTP 状态码 400）**
```json
{
    "code": 400,
    "reason": "USER_INVALID_REQUEST",
    "message": "密码长度至少8位",
    "metadata": {}
}
```

● **密码过长（HTTP 状态码 400）**
```json
{
    "code": 400,
    "reason": "USER_INVALID_REQUEST",
    "message": "密码长度不能超过16位",
    "metadata": {}
}
```

● **密码缺少数字（HTTP 状态码 400）**
```json
{
    "code": 400,
    "reason": "USER_INVALID_REQUEST",
    "message": "密码必须包含至少一个数字",
    "metadata": {}
}
```

● **密码缺少字母（HTTP 状态码 400）**
```json
{
    "code": 400,
    "reason": "USER_INVALID_REQUEST",
    "message": "密码必须包含至少一个字母",
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
● **鉴权说明:** 需要有效的Refresh Token（在数据库中验证）

● **请求 Body:**
```json
{
    "refresh_token": "string"
}
```

● **请求示例:**
```bash
curl -X POST "https://api.example.com/v1/auth/refresh" \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "eyJhbGciOiJIUzI1NiIs..."}'
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
● **鉴权说明:** 需要有效的Refresh Token（在数据库中验证）

● **请求 Body:**
```json
{
    "refresh_token": "string"
}
```

● **请求示例:**
```bash
curl -X POST "https://api.example.com/v1/auth/logout" \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "eyJhbGciOiJIUzI1NiIs..."}'
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

#### 客户端请求格式（Nginx接收）

● **请求 Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json
```

● **请求示例:**
```bash
curl -X GET "https://api.example.com/v1/user/profile" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json"
```

#### 微服务接收格式（Nginx转发后）

● **请求 Headers:**
```
X-User-ID: 12345
Content-Type: application/json
```

● **说明:**
- 客户端发送JWT Access Token（Authorization: Bearer）
- Nginx验证Token后，将用户ID提取到X-User-ID头
- 微服务从X-User-ID头获取用户身份信息

#### 成功响应 (200 OK)
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

#### 错误响应

● **认证失败（Nginx返回 401）**
```json
{
    "code": 401,
    "reason": "USER_INVALID_TOKEN",
    "message": "访问令牌无效或已过期",
    "metadata": {}
}
```

● **用户不存在（微服务返回 404）**
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

#### 客户端请求格式（Nginx接收）

● **请求 Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json
```

● **请求 Body:**
```json
{
    "nickname": "string",
    "avatar_url": "string"
}
```

● **请求示例:**
```bash
curl -X PUT "https://api.example.com/v1/user/profile" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "nickname": "新昵称",
    "avatar_url": "https://example.com/new-avatar.jpg"
  }'
```

#### 微服务接收格式（Nginx转发后）

● **请求 Headers:**
```
X-User-ID: 12345
Content-Type: application/json
```

● **请求 Body:**
```json
{
    "nickname": "string",
    "avatar_url": "string"
}
```

● **说明:**
- 客户端发送JWT Access Token（Authorization: Bearer）
- Nginx验证Token后，将用户ID提取到X-User-ID头
- 微服务从X-User-ID头获取用户身份信息，body内容保持不变

#### 成功响应 (200 OK)
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

#### 错误响应

● **认证失败（Nginx返回 401）**
```json
{
    "code": 401,
    "reason": "USER_INVALID_TOKEN",
    "message": "访问令牌无效或已过期",
    "metadata": {}
}
```

● **昵称冲突（微服务返回 409）**
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

| 接口名 | 方法 | 路径 | Nginx认证方式 | 微服务认证方式 | 说明 |
|--------|------|------|---------------|----------------|------|
| AuthService_SendRegisterCode | POST | `/v1/auth/send-code` | 无需认证 | 无需认证 | 公共接口，发送验证码 |
| AuthService_Register | POST | `/v1/auth/register` | 无需认证 | 无需认证 | 公共接口，用户注册 |
| AuthService_Login | POST | `/v1/auth/login` | 无需认证 | 无需认证 | 公共接口，返回Access Token |
| AuthService_RefreshToken | POST | `/v1/auth/refresh` | 无需认证 | Refresh Token | 需有效Refresh Token |
| AuthService_Logout | POST | `/v1/auth/logout` | 无需认证 | Refresh Token | 需有效Refresh Token |
| UserService_GetCurrentUser | GET | `/v1/user/profile` | **JWT Access Token** | X-User-ID Header | Nginx验证JWT，提取UserID |
| UserService_UpdateCurrentUser | PUT | `/v1/user/profile` | **JWT Access Token** | X-User-ID Header | Nginx验证JWT，提取UserID |

### 认证流程说明

#### 1. 客户端 → Nginx（入口）
- **UserService接口**：发送 `Authorization: Bearer <access_token>`
- **AuthService接口**：无需特殊头，直接发送请求

#### 2. Nginx → 微服务（转发）
- **UserService接口**：Nginx验证JWT后，设置 `X-User-ID: <user_id>`
- **AuthService接口**：直接转发请求

---

## 重要说明

1. **验证码机制**: 生成的验证码为6位数字，有效期10分钟
2. **频率限制**: 发送验证码接口有60秒频率限制
3. **密码强度要求**: 密码长度8-16位，必须包含至少一个数字和至少一个字母
4. **Token有效期**: Access Token 1小时，Refresh Token 7天
5. **认证方式**: UserService使用X-User-ID Header而非JWT Token
6. **原子性操作**: Token刷新使用事务确保原子性

---

## Nginx 鉴权配置

为了在微服务架构中实现统一的鉴权，建议在Nginx层进行JWT Access Token的验证，然后通过请求头将用户ID传递给后端微服务。

### 配置方案

#### 1. 安装必要的模块

需要安装`ngx_http_auth_jwt_module`模块（或使用第三方模块如`ngx-restful-jwt`）：

```bash
# 使用OpenResty（推荐）
sudo yum install openresty

# 或编译Nginx时添加JWT模块
--add-module=/path/to/nginx-jwt-module
```

#### 2. Nginx 配置示例

**环境变量配置：**

```bash
# /etc/nginx/conf.d/jwt.conf
# JWT访问令牌密钥
env JWT_ACCESS_SECRET="your-jwt-access-secret-key";

# 用户服务Upstream
upstream user_service {
    server user-service-1:8000;
    server user-service-2:8000;
    # 可根据需要添加更多实例
}
```

**Nginx主配置：**

```nginx
# /etc/nginx/nginx.conf

http {
    # JWT验证配置
    jwt_key_header x-jwt-key;  # 从header中获取JWT密钥
    # 或使用固定密钥
    # jwt_key "your-jwt-access-secret-key";

    # 频次限制
    limit_req_zone $binary_remote_addr zone=auth:10m rate=10r/s;

    server {
        listen 80;
        server_name api.example.com;

        # 解析用户服务请求
        location ~ ^/v1/user/ {
            # 频率限制
            limit_req zone=auth burst=20 nodelay;

            # 启用JWT验证
            auth_jwt "User Service API" token_key="$access_token";
            auth_jwt_key_file /etc/nginx/jwt_keys.json;

            # 验证失败处理
            error_page 401 = @jwt_error;

            # 从JWT声明中提取用户ID并设置请求头
            access_by_lua_block {
                -- 从请求参数或header中获取access_token
                local token = ngx.var.access_token
                if token == "" then
                    token = ngx.req.get_headers()["Authorization"]
                    if token and token:match("Bearer%s+(.+)") then
                        token = token:match("Bearer%s+(.+)")
                    end
                end

                if token then
                    -- 解析JWT token
                    local jwt = require "resty.jwt"
                    local jwt_obj = jwt:verify(ngx.shared.jwt_keys:get("access_secret"), token, {
                        sub = { type = "string" },
                        exp = { type = "number" }
                    })

                    if jwt_obj.valid then
                        local user_id = jwt_obj.payload.sub
                        -- 设置请求头，供后端服务使用
                        ngx.req.set_header("X-User-ID", user_id)
                        ngx.req.set_header("X-User-ID-Str", user_id)
                    else
                        ngx.exit(ngx.HTTP_UNAUTHORIZED)
                    end
                else
                    ngx.log(ngx.ERR, "No access token provided")
                    ngx.exit(ngx.HTTP_UNAUTHORIZED)
                end
            }

            # 转发请求到用户服务
            proxy_pass http://user_service;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;

            # 保留用户相关头信息
            proxy_set_header X-User-ID $http_x_user_id;
        }

        # JWT错误处理
        location @jwt_error {
            return 401 '{"code": 401, "reason": "USER_INVALID_TOKEN", "message": "访问令牌无效或已过期", "metadata": {}}';
        }

        # 健康检查接口
        location /health {
            access_log off;
            return 200 "healthy\n";
            add_header Content-Type text/plain;
        }
    }
}
```

#### 3. 加载JWT密钥的Lua脚本

```lua
-- /etc/nginx/jwt_keys.json
{
  "keys": [
    {
      "alg": "HS256",
      "secret": "your-jwt-access-secret-key"
    }
  ]
}
```

```lua
-- /etc/nginx/init_by_lua.lua
-- 初始化JWT密钥
local jwt_keys = ngx.shared.jwt_keys
jwt_keys:set("access_secret", os.getenv("JWT_ACCESS_SECRET") or "your-jwt-access-secret-key")
```

#### 4. Docker Compose 配置示例

```yaml
# docker-compose.yml
version: '3.8'

services:
  nginx:
    image: openresty/openresty:latest
    container_name: api-gateway
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/conf:/etc/nginx/conf.d
      - ./nginx/lua:/etc/nginx/lua
    environment:
      - JWT_ACCESS_SECRET=${JWT_ACCESS_SECRET}
    depends_on:
      - user-service

  user-service:
    image: user-service:latest
    container_name: user-service
    environment:
      - JWT_ACCESS_SECRET=${JWT_ACCESS_SECRET}
      - JWT_REFRESH_SECRET=${JWT_REFRESH_SECRET}
    # 其他配置...
```

### 鉴权流程

1. **客户端请求**：客户端将JWT Access Token放在请求头或查询参数中
   ```
   Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
   ```

2. **Nginx验证**：Nginx验证Access Token的有效性
   - 验证Token签名
   - 检查Token是否过期
   - 从Token的`sub`声明中提取用户ID

3. **设置请求头**：Nginx将用户ID设置到`X-User-ID`头中
   ```
   X-User-ID: 12345
   ```

4. **转发请求**：Nginx将请求转发到后端微服务
   - 保留原始请求头
   - 添加`X-User-ID`头

5. **微服务处理**：微服务从`X-User-ID`头中获取用户ID
   - 不需要再验证JWT Token
   - 直接使用用户ID进行业务处理

### 优势

1. **性能提升**：避免每个微服务都验证JWT Token，降低延迟
2. **安全统一**：统一在Nginx层进行鉴权，策略一致性更好
3. **微服务解耦**：微服务只关注业务逻辑，不关心鉴权实现
4. **扩展性更好**：后续添加新微服务时不需要重复实现鉴权逻辑
5. **日志统一**：所有鉴权日志集中在Nginx，便于审计和监控

### 注意事项

1. **Token过期处理**：建议在Nginx返回401时，引导客户端使用Refresh Token获取新的Access Token
2. **错误信息**：Nginx错误页面应保持与API文档一致的错误格式
3. **监控告警**：建议监控JWT验证失败率，及时发现异常
4. **高可用**：Nginx作为API网关，应配置多实例并使用负载均衡
5. **安全性**：
   - 使用环境变量管理JWT密钥
   - 定期轮换JWT密钥
   - 确保Nginx与微服务之间的内部网络是安全的
