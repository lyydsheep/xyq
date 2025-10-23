### 如何规范微服务状态码？
规范通常分为两个层面：**HTTP 状态码（Status Code）** 和 **业务错误码（Business Error Code）**。

#### 1. HTTP 状态码（Status Code）规范
**原则：** 严格遵循 HTTP/1.1 规范，将状态码用于表达请求的**结果类型**，而不是具体的业务原因。

| 状态码范围 | 含义 | 规范示例 | 常见场景 |
| :--- | :--- | :--- | :--- |
| **2xx** | **成功 (Success)** | `200 OK` (通用成功/GET), `201 Created` (资源创建成功/POST), `202 Accepted` (异步任务已接受) | U3 登录成功，U6 获取资料成功 |
| **4xx** | **客户端错误 (Client Error)** | `400 Bad Request` (请求体格式错误、参数校验失败) | U2 注册时，昵称或密码格式不正确 |
|  |  | `401 Unauthorized` (缺少认证信息或 Token 过期/无效) | 访问 U6 缺少 Access Token 或 Token 过期 |
|  |  | `403 Forbidden` (认证通过但无权限访问资源) | 普通用户尝试访问付费用户特有功能 |
|  |  | `404 Not Found` (请求资源不存在) | 请求一个不存在的绘本或用户 ID |
|  |  | `409 Conflict` (资源冲突，如已存在) | U2 注册时，邮箱已被占用 (`email` 唯一键冲突) |
|  |  | `429 Too Many Requests` (被限流) | 用户请求 QPS 超过限制 |
| **5xx** | **服务器错误 (Server Error)** | `500 Internal Server Error` (通用内部错误) | DB 连接失败、微服务间调用超时 |
|  |  | `503 Service Unavailable` (服务暂时不可用) | 服务熔断、系统维护 |


#### 2. 业务错误码（Business Error Code）规范
仅使用 HTTP 状态码不足以表达所有的业务逻辑错误（例如：点数不足、验证码错误）。因此，需要在响应体中增加一个自定义的**业务错误码** (`code`) 和**错误信息** (`message`)。

**规范格式：**

对于所有非 `2xx` 的响应，返回统一的 JSON 格式：

```json
{
  "code": "string | integer", // 业务错误码 (用于程序判断)
  "message": "string",        // 错误信息 (用于用户展示或日志记录)
  "details": "object | null"  // 详细信息（可选，例如哪个字段校验失败）
}
```

**业务错误码编码规则（推荐）：**

采用分段或前缀规则，将错误码与对应的微服务和错误类型关联起来，例如：`服务标识 + 错误类型 + 序号`。

以本项目的 **User Service (用户服务)** 为例：

| 错误码 | HTTP 状态码 | 描述 | 对应场景 |
| :--- | :--- | :--- | :--- |
| `USER_40001` | `400 Bad Request` | **验证码错误或已过期** | U2 注册时，`code` 字段校验失败 |
| `USER_40002` | `400 Bad Request` | 邮箱格式不正确 | U1/U2 邮箱格式校验失败 |
| `USER_40101` | `401 Unauthorized` | Access Token 无效或缺失 | U6 访问时 Token 无效 |
| `USER_40102` | `401 Unauthorized` | Refresh Token 无效或缺失 | U4 刷新失败 |
| `USER_40901` | `409 Conflict` | 邮箱已被注册 | U2 注册时邮箱已存在 |
| `USER_40301` | `403 Forbidden` | **点数不足** | 尝试生成绘本（F1.4）时点数不足 |
| `SYS_50001` | `500 Internal Server Error` | 数据库操作失败 | 任何服务内部数据库错误 |
| `SYS_50401` | `504 Gateway Timeout` | 后端服务超时 | 服务间调用超时 |


#### 总结：User & Auth 失败状态码规范示例
| 接口 | 失败场景 | HTTP Status Code | 业务错误码 (示例) |
| :--- | :--- | :--- | :--- |
| **U2 注册** | 验证码错误 | `400 Bad Request` | `USER_40001` |
| **U2 注册** | 邮箱已注册 | `409 Conflict` | `USER_40901` |
| **U3 登录** | 密码错误 | `401 Unauthorized` | `USER_40103` (密码错误) |
| **U6 获取资料** | Token 过期 | `401 Unauthorized` | `USER_40101` (Token 无效) |
| **U8 查点数** | 内部 DB 错误 | `500 Internal Server Error` | `SYS_50001` (DB 错误) |


