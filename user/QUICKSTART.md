# 快速入门指南

本文档帮助您快速启动项目并配置所有必需的环境变量。

## 🚀 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 配置环境变量

复制环境变量模板：

```bash
cp .env.example .env
```

编辑 `.env` 文件，设置所有必需的变量：

```bash
# 数据库
DB_PASSWORD=your_database_password

# Redis（可选）
REDIS_PASSWORD=

# JWT密钥
JWT_ACCESS_SECRET=your_jwt_access_secret_key
JWT_REFRESH_SECRET=your_jwt_refresh_secret_key

# SendGrid邮件服务
SENDGRID_API_KEY=your_sendgrid_api_key
```

### 3. 生成强密钥

使用以下命令生成安全的密钥：

```bash
# 生成JWT密钥
openssl rand -base64 32

# 或使用其他密码生成工具
```

### 4. 启动数据库和Redis

确保MySQL和Redis服务正在运行：

```bash
# MySQL (示例)
mysql -u root -p

# Redis (示例)
redis-cli ping
# 应该返回 PONG
```

### 5. 运行应用

```bash
go run cmd/auth/main.go
```

## 📧 邮件发送配置

### SendGrid设置

1. **注册SendGrid**
   - 访问 https://app.sendgrid.com
   - 创建免费账户

2. **验证域名**
   - 在SendGrid控制台添加您的域名
   - 配置DNS记录（SPF、DKIM）

3. **创建API Key**
   - Settings → API Keys
   - 创建新的Restricted Access API Key
   - 复制API Key到 `.env` 文件

4. **测试邮件发送**

```bash
# 编译测试程序
go build -o test_email test_email.go

# 运行测试
./test_email
```

## 🔧 关键配置项说明

### 必需的环境变量

| 变量名 | 说明 | 示例 |
|--------|------|------|
| `DB_PASSWORD` | 数据库密码 | `MyP@ssw0rd123` |
| `JWT_ACCESS_SECRET` | JWT访问令牌密钥 | `base64编码的32字节随机字符串` |
| `JWT_REFRESH_SECRET` | JWT刷新令牌密钥 | `base64编码的32字节随机字符串` |
| `SENDGRID_API_KEY` | SendGrid API密钥 | `SG.xxxxxx...` |

### 可选的环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `REDIS_PASSWORD` | Redis密码 | 空（无密码） |

## 🏃‍♂️ 常用命令

```bash
# 构建应用
make build

# 运行测试
make test

# 生成proto文件
make config

# 生成所有文件
make all

# 清理
make clean
```

## 🔍 验证安装

### 1. 检查数据库连接

```bash
# 应用程序启动后应看到：
# "connecting to MySQL... OK"
```

### 2. 检查Redis连接

```bash
# 应用程序启动后应看到：
# "connecting to Redis... OK"
```

### 3. 测试JWT功能

```bash
# 检查应用日志，应看到：
# "Starting server on :8000 (HTTP) and :9000 (gRPC)"
```

### 4. 测试邮件发送

```bash
# 触发验证码发送，查看日志：
# "Verification code sent successfully to: <email>"
```

## 🚨 故障排除

### 数据库连接失败

```
Error: failed to connect to MySQL
```

**解决方案**：
- 检查MySQL服务是否运行：`systemctl status mysql`
- 检查配置文件中的数据库设置
- 验证用户名和密码

### Redis连接失败

```
Error: failed to connect to Redis
```

**解决方案**：
- 检查Redis服务是否运行：`systemctl status redis`
- 检查Redis配置（默认端口6379）
- 验证Redis密码（如果设置了）

### 邮件发送失败

```
Error: failed to send verification email
```

**解决方案**：
- 验证`SENDGRID_API_KEY`是否正确
- 检查发送者邮箱域名是否验证
- 查看SendGrid控制台的Activity Feed

### JWT密钥未设置

```
Error: JWT_ACCESS_SECRET environment variable is required
```

**解决方案**：
```bash
# 确保环境变量已设置
export JWT_ACCESS_SECRET=$(openssl rand -base64 32)
export JWT_REFRESH_SECRET=$(openssl rand -base64 32)
```

## 📚 相关文档

- [邮件发送配置指南](./EMAIL_SETUP.md) - 详细的邮件发送配置
- [API文档](./api/) - REST API接口文档
- [数据库设计](./docs/database.md) - 数据库表结构

## 💡 提示

1. **开发环境**：使用`.env`文件管理环境变量
2. **生产环境**：使用Docker Secrets或Kubernetes Secrets
3. **密钥管理**：定期轮换密钥，使用强随机字符串
4. **监控**：配置日志聚合，监控关键指标

## 🆘 获取帮助

- 查看日志文件：`tail -f logs/app.log`
- 常见问题：[FAQ](./docs/FAQ.md)
- 技术支持：support@lyydsheep.xyz

---

祝您使用愉快！🎉
