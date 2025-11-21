# User 微服务拆分总结

## 📋 工作概览

本次工作将 user 微服务从主项目中拆分出来，使其成为一个独立可复用的微服务。

## ✅ 已完成工作

### 1. 清理未使用的功能模块

删除了以下未使用的模块：
- `greeter` - 模板示例代码
- `point` - 积分系统（仅数据模型，未实现业务逻辑）
- `transaction` - 交易系统（仅数据模型，未实现业务逻辑）

**涉及文件：**
- `internal/biz/greeter.go`, `internal/biz/point.go`, `internal/biz/transaction.go`
- `internal/service/greeter.go`
- `internal/data/greeter.go`, `internal/data/point.go`, `internal/data/transaction.go`
- `internal/data/point_test.go`, `internal/data/transaction_test.go`
- `internal/server/http.go` - 移除 greeter 注册
- `internal/server/grpc.go` - 移除 greeter 注册

### 2. 配置参数化

将硬编码的邮件配置抽取为可配置参数：

**新增配置项：**
```yaml
email:
  sender_name: "用户系统"        # 邮件发件人显示名称
  sender_email: "noreply@example.com"  # 发件人邮箱地址
  support_email: "support@example.com" # 客服支持邮箱
  company_name: "您的公司名称"   # 公司名称
  app_name: "您的应用名称"       # 应用名称
```

**代码修改：**
- `internal/conf/conf.proto` - 添加 Email 消息定义
- `internal/biz/user.go` - 添加 EmailConfig 结构体，使用配置参数
- `internal/biz/biz.go` - 添加 NewEmailConfig 和 EmailProvider 函数
- `cmd/auth/main.go` - 传递 Email 配置给 wireApp

### 3. 更新依赖注入

- `internal/biz/biz.go` - 更新 ProviderSet，移除 greeter、point、transaction 相关依赖
- `internal/service/service.go` - 更新 ProviderSet，移除 greeter 相关依赖
- `internal/data/data.go` - 更新 ProviderSet，移除相关仓库依赖
- `cmd/auth/wire.go` - 更新 wireApp 签名，添加 Email 参数
- 重新生成 `wire_gen.go`

### 4. 代码清理

- 移除所有未使用的 import
- 删除示例文件：`internal/service/error_tracing_example.go`
- 更新 Makefile 中的服务名称
- 更新配置文件中的服务名称

### 5. 文档更新

- 重写 `README.md` - 提供完整的使用说明
  - 功能特性介绍
  - 架构设计图
  - 快速开始指南
  - 详细的 API 文档
  - 开发和部署指南
- 更新 `config.yaml` - 添加模板配置和注释

## 🎯 当前服务功能

### AuthService（认证服务）
- ✅ 发送注册验证码
- ✅ 用户注册
- ✅ 用户登录
- ✅ 刷新令牌

### UserService（用户服务）
- ✅ 获取当前用户信息
- ✅ 更新用户资料（昵称、头像）

## 📦 依赖项

**主要依赖：**
- Kratos v2.8.0 - 微服务框架
- GORM v1.31.0 - ORM
- MySQL 驱动
- Redis 客户端
- SendGrid 邮件发送
- JWT 令牌管理
- OpenTelemetry - 链路追踪
- Snowflake - ID 生成

## 🔧 使用方式

### 1. 配置数据库和 Redis

```bash
# MySQL
CREATE DATABASE user_service;

# Redis
# 启动 Redis 服务
```

### 2. 配置服务

```bash
# 复制配置文件
cp configs/config.yaml configs/config.local.yaml

# 编辑配置，修改：
# - 数据库连接信息
# - Redis 连接信息
# - 邮件配置（sender_email, support_email, company_name, app_name）
```

### 3. 设置环境变量

```bash
export DB_PASSWORD=your-mysql-password
export REDIS_PASSWORD=your-redis-password  # 可选
export SENDGRID_API_KEY=your-sendgrid-key
```

### 4. 运行服务

```bash
# 构建
make build

# 运行
./bin/auth -conf ./configs/config.local.yaml
```

## 📡 API 端点

### HTTP 服务
- 端口：8000
- 健康检查：`GET /v1/health`

### gRPC 服务
- 端口：9000

### 主要 API
- `POST /v1/auth/send-register-code` - 发送注册验证码
- `POST /v1/auth/register` - 用户注册
- `POST /v1/auth/login` - 用户登录
- `POST /v1/auth/refresh` - 刷新令牌
- `GET /v1/user/profile` - 获取用户信息（需要 X-User-ID 头）
- `PUT /v1/user/profile` - 更新用户信息（需要 X-User-ID 头）

## 🐳 Docker 部署

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

## ✨ 特性

- 🔐 JWT 认证（访问令牌 + 刷新令牌）
- 📧 SendGrid 邮件发送（支持测试模式）
- 📊 链路追踪（集成 Jaeger）
- 📝 结构化日志
- 🛡️ 错误增强（带 trace ID）
- ⚙️ 配置驱动

## 🔄 下一步工作建议

1. **添加数据库迁移工具** - 自动化表结构创建
2. **完善单元测试** - 提高代码覆盖率
3. **添加性能监控** - 集成 Prometheus
4. **添加限流熔断** - 提高系统稳定性
5. **完善部署文档** - 添加 Kubernetes 部署示例

## 📞 支持

如有问题，请查看 `README.md` 或提交 Issue。
