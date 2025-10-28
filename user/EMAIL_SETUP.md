# 邮件发送功能配置指南

本文档说明如何配置和使用SendGrid邮件发送功能。

## 概述

我们已经实现了基于SendGrid的邮件发送功能，用于发送用户注册验证码邮件。

## 功能特点

✅ **双格式邮件**：同时发送纯文本和HTML格式
✅ **中文界面**：完整的中文邮件模板
✅ **美观设计**：渐变色设计，响应式布局
✅ **安全提醒**：包含验证码安全使用说明
✅ **环境变量配置**：通过环境变量管理API密钥

## 配置步骤

### 1. 获取SendGrid API Key

1. 注册或登录SendGrid账户：https://app.sendgrid.com
2. 进入 Settings → API Keys
3. 创建新的API Key（选择"Restricted Access"）
4. 复制API Key并妥善保存

### 2. 配置环境变量

编辑项目根目录下的 `.env` 文件：

```bash
# SendGrid API密钥，用于发送验证邮件
SENDGRID_API_KEY=SG.xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

或者设置系统环境变量：

```bash
export SENDGRID_API_KEY=SG.xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

### 3. 配置发送者邮箱

在 `internal/biz/user.go` 文件中，更新发送者邮箱信息（第319行）：

```go
fromEmail := mail.NewEmail("您的公司名称", "noreply@yourdomain.com")
```

**重要**：发送者邮箱域名必须在SendGrid中完成域名验证。

### 4. 配置支持邮箱

在邮件模板的页脚中，更新支持邮箱地址（第403行）：

```go
<p>如有问题请联系 <a href="mailto:support@yourdomain.com">support@yourdomain.com</a></p>
```

## 邮件模板设计

### HTML邮件特性

- **渐变色头部**：紫色渐变设计
- **验证码高亮**：粉色渐变背景，大字号显示
- **安全提醒**：黄色警告框强调安全性
- **响应式布局**：适配移动设备
- **中文字体优化**：适配不同平台的字体

### 邮件内容

- 友好的问候语
- 6位数字验证码
- 10分钟过期提醒
- 安全使用建议
- 支持联系方式

## 测试邮件发送

在开发环境中，可以通过以下方式测试：

```go
// 测试发送验证码
err := userUsecase.SendRegisterCode(context.Background(), "test@example.com")
if err != nil {
    log.Fatal("发送失败:", err)
}
```

## 生产环境部署

### Docker部署

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main cmd/auth/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

ENV SENDGRID_API_KEY=SG.xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

CMD ["./main"]
```

### Kubernetes部署

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: app
spec:
  containers:
  - name: app
    image: your-registry/app:latest
    env:
    - name: SENDGRID_API_KEY
      valueFrom:
        secretKeyRef:
          name: app-secrets
          key: sendgrid-api-key
```

## 故障排除

### 常见问题

1. **API Key未设置**
   - 错误信息：`SENDGRID_API_KEY environment variable is required`
   - 解决方案：检查环境变量是否正确设置

2. **发送失败**
   - 错误信息：`failed to send verification email`
   - 解决方案：检查API Key是否有效，检查发送者邮箱域名是否验证

3. **域名未验证**
   - 错误信息：SendGrid返回4xx错误
   - 解决方案：在SendGrid控制台完成域名验证

### 日志监控

应用程序会记录以下日志：

- `Sending verification email to: <email>` - 开始发送
- `Verification email sent successfully` - 发送成功
- `Failed to send email` - 发送失败

## 安全建议

1. **API Key保护**
   - 不要将API Key硬编码到代码中
   - 使用环境变量或密钥管理服务
   - 定期轮换API Key

2. **域名验证**
   - 必须完成域名所有权验证
   - 配置SPF、DKIM记录

3. **发送限制**
   - SendGrid免费计划：每天100封
   - 监控发送量，避免超额

## 相关文档

- SendGrid官方文档：https://docs.sendgrid.com/
- SendGrid API参考：https://docs.sendgrid.com/api-reference
- 邮件模板设计指南：https://docs.sendgrid.com/ui/sending-email/email-templates

## 技术支持

如有问题，请联系：support@lyydsheep.xyz

---

*最后更新时间：2025年1月*
