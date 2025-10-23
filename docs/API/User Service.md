好的，我已根据您的要求，对 **User & Auth API** 文档进行了最终的润色和优化，使其更加规范、专业和易读。

以下是最终的 **User & Auth API 文档**。

---

## 1. 用户与认证 API 文档 (User Service)
**基础路径:** `/v1/auth`, `/v1/users`  
**鉴权机制:** 基于 JWT 的长/短 Token 机制 (Access Token 用于业务请求，Refresh Token 用于刷新 Access Token)。  
**认证约定:** 所有需要认证的接口，均需在 HTTP Header 中携带 `Authorization: Bearer <Access Token>`。

| ID | 功能模块 | 接口路径 | 方法 | 描述 | 鉴权要求 | 依赖 PRD |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **U1** | 注册流程 | `/v1/auth/register-code` | `POST` | 发送注册邮箱验证码。 | 无 | F1.1 |
| **U2** | 注册流程 | `/v1/auth/register` | `POST` | **用户账户注册**，创建用户并分配初始点数。 | 无 | F1.1 |
| **U3** | 登录鉴权 | `/v1/auth/login` | `POST` | **用户登录**，获取 Access Token 和 Refresh Token。 | 无 | F1.2, F1.5 |
| **U4** | 登录鉴权 | `/v1/auth/refresh` | `POST` | 使用 Refresh Token 换取新的 Access Token。 | Refresh Token | F1.5 |
| **U5** | 登录鉴权 | `/v1/auth/logout` | `POST` | **用户登出**，使 Refresh Token 失效。 | Access Token | F1.5 |
| **U6** | 资料管理 | `/v1/users/me` | `GET` | **获取当前用户资料**（基础信息）。 | Access Token | 基础 |
| **U7** | 资料管理 | `/v1/users/me` | `PUT` | 更新当前用户资料（昵称、头像等）。 | Access Token | 基础 |
| **U8** | 点数管理 | `/v1/users/me/points` | `GET` | **查询当前可用点数**。 | Access Token | F1.4 |
| **U9** | 点数管理 | `/v1/users/me/point-logs` | `GET` | 查询点数消耗和交易流水记录（分页）。 | Access Token | F1.4 |


---

### 详细接口规范 (核心接口)
#### U2: 用户账户注册 (POST /v1/auth/register)
+ **功能描述:** 通过邮箱验证码完成新用户注册。
+ **请求 Body:**

```json
{
  "email": "string",       // 邮箱地址，唯一且格式正确
  "password": "string",    // 用户密码
  "code": "string",        // 邮箱收到的验证码
  "nickname": "string"     // 用户昵称
}
```

+ **成功响应 (201 Created):** 返回新创建用户的基本信息。

```json
{
  "id": 12345,
  "email": "user@example.com",
  "nickname": "故事创造者"
}
```

#### U3: 用户登录 (POST /v1/auth/login)
+ **功能描述:** 使用邮箱和密码登录，返回长短 Token 进行鉴权。
+ **请求 Body:**

```json
{
  "email": "string",
  "password": "string"
}
```

+ **成功响应 (200 OK):**

```json
{
  "access_token": "string",      // 短期访问令牌 (例如 1 小时有效期)
  "access_expires_in": 3600,     // Access Token 有效期 (秒)
  "refresh_token": "string",     // 长期刷新令牌 (例如 7 天有效期)
  "refresh_expires_in": 604800   // Refresh Token 有效期 (秒)
}
```

#### U4: 刷新 Access Token (POST /v1/auth/refresh)
+ **功能描述:** 在 Access Token 过期后，使用 Refresh Token 无感获取新的 Access Token。
+ **请求 Body:**

```json
{
  "refresh_token": "string"
}
```

+ **成功响应 (200 OK):**

```json
{
  "access_token": "string",      // 新 Access Token
  "access_expires_in": 3600      // 新 Access Token 有效期 (秒)
}
```

+ **常见错误:** `401 Unauthorized` (Refresh Token 无效或过期，需要用户重新登录)。

#### U8: 查询当前可用点数 (GET /v1/users/me/points)
+ **功能描述:** 获取当前用户可用于 AI 生成的剩余点数。
+ **成功响应 (200 OK):**

```json
{
  "user_id": 12345,
  "current_points": 50,    // 当前可用点数
  "total_consumed": 150    // 历史总消耗点数
}
```

---

**User & Auth API** 设计已完成。我们可以继续设计下一个模块：**Creation & Management API**。

