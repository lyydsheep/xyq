# User 微服务 (Auth & User Management)

基于 Kratos 框架开发的用户认证与用户管理微服务，提供完整的用户注册、登录和用户信息管理功能。

## ✨ 功能特性

- 🔐 **用户认证**: 邮箱注册、登录、验证码发送
- 👤 **用户管理**: 获取/更新用户资料、头像、昵称
- 📧 **邮件服务**: 支持 SendGrid 邮件发送（支持测试模式）
- 🔑 **令牌管理**: JWT 访问令牌 + 刷新令牌
- 📊 **链路追踪**: 集成 Jaeger 分布式追踪
- 📝 **日志记录**: 结构化日志输出
- 🛡️ **错误增强**: 带 trace ID 的错误响应

## 🏗️ 架构设计

```
┌─────────────────┐
│   HTTP/gRPC     │  ← API 层
│    服务端       │
└────────┬────────┘
         │
┌────────▼────────┐
│  Service 层     │  ← 业务逻辑层
└────────┬────────┘
         │
┌────────▼────────┐
│   Biz 层        │  ← 用例层
└────────┬────────┘
         │
┌────────▼────────┐
│   Data 层       │  ← 数据访问层
└────────┬────────┘
         │
┌────────▼────────┐
│ MySQL / Redis   │  ← 数据存储
└─────────────────┘
```

## 🚀 快速开始

### 环境要求

- Go 1.24+
- MySQL 5.7+
- Redis 6.0+
- Jaeger (可选，用于链路追踪)

### 1. 安装依赖

```bash
# 下载依赖
go mod download

# 生成相关代码（如果修改了 proto 文件）
make api
make config
make generate
```

### 2. 配置数据库

创建 MySQL 数据库：

```sql
CREATE DATABASE user_service;
```

创建 Redis（无密码或设置密码）。

### 3. 配置服务

复制并修改配置文件：

```bash
cp configs/config.yaml configs/config.local.yaml
```

编辑 `configs/config.local.yaml`:

```yaml
server:
  http:
    addr: 0.0.0.0:8000
  grpc:
    addr: 0.0.0.0:9000

data:
  database:
    driver: mysql
    host: your-mysql-host
    port: 3306
    database: user_service
    username: your-username
    # 建议通过环境变量设置密码
    # password: your-password
  redis:
    addr: your-redis-host:6379
    # password: your-redis-password  # 可选

email:
  sender_name: "用户系统"
  sender_email: "noreply@yourdomain.com"      # 替换为你的邮箱
  support_email: "support@yourdomain.com"     # 替换为你的客服邮箱
  company_name: "你的公司名称"
  app_name: "你的应用名称"

trace:
  endpoint: http://localhost:14268/api/traces  # 可选：Jaeger 地址
  service_name: user-service
```

设置环境变量：

```bash
export DB_PASSWORD=your-mysql-password
export REDIS_PASSWORD=your-redis-password  # 如果 Redis 有密码
export SENDGRID_API_KEY=your-sendgrid-api-key  # 发送邮件用
```

### 4. 运行服务

```bash
# 构建
make build

# 运行
./bin/auth -conf ./configs/config.local.yaml
```

服务启动后：
- HTTP 服务：`http://localhost:8000`
- gRPC 服务：`localhost:9000`
- 健康检查：`http://localhost:8000/v1/health`

## 📚 API 文档

### 认证服务 (AuthService)

#### 1. 发送注册验证码

```http
POST /v1/auth/send-register-code
Content-Type: application/json

{
  "email": "user@example.com"
}
```

#### 2. 用户注册

```http
POST /v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "code": "123456",
  "nickname": "新用户"
}
```

#### 3. 用户登录

```http
POST /v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

#### 4. 刷新令牌

```http
POST /v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "your-refresh-token"
}
```

### 用户服务 (UserService)

#### 1. 获取当前用户信息

需要通过请求头传递用户ID（通常由网关设置）：

```http
GET /v1/user/profile
X-User-ID: 123456
```

#### 2. 更新用户信息

```http
PUT /v1/user/profile
Content-Type: application/json
X-User-ID: 123456

{
  "nickname": "新昵称",
  "avatar_url": "https://example.com/avatar.jpg"
}
```

## 🔧 开发指南

### 项目结构

```
├── api/                  # 生成的 API 代码
│   ├── auth/v1/         # 认证相关 API
│   └── user/v1/         # 用户相关 API
├── cmd/auth/            # 应用入口
├── internal/            # 内部代码
│   ├── biz/             # 业务逻辑层
│   ├── conf/            # 配置定义
│   ├── data/            # 数据访问层
│   ├── pkg/             # 工具包
│   └── service/         # 业务服务层
└── configs/             # 配置文件
```

### 添加新的 API

1. 在 `api/` 目录创建 proto 文件
2. 运行 `make api` 生成代码
3. 运行 `make config` 生成配置
4. 运行 `make generate` 生成 wire 代码

### 测试模式

开发测试时，可以在邮件发送部分使用测试模式：

```bash
# 使用测试 API key（以 test- 开头）
export SENDGRID_API_KEY=test-your-api-key
```

测试模式下不会实际发送邮件，但会记录日志。

## 📦 部署

### Docker 部署

```bash
# 构建镜像
docker build -t user-service:latest .

# 运行容器
docker run -d \
  --name user-service \
  -p 8000:8000 \
  -p 9000:9000 \
  -v /path/to/configs:/data/conf \
  -e DB_PASSWORD=your-password \
  -e REDIS_PASSWORD=your-redis-password \
  -e SENDGRID_API_KEY=your-sendgrid-key \
  user-service:latest
```

### Kubernetes 部署

参考 `k8s/` 目录下的 YAML 文件（如果有）。

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

[根据实际情况填写许可证信息]

## 📞 支持

如有问题请联系：
- 邮箱: support@yourdomain.com
- 文档: [链接到详细文档]

