# User 微服务链路追踪实现指南

## 概述

本微服务已成功集成基于 OpenTelemetry 的链路追踪功能，支持分布式请求追踪、性能监控和错误定位。

## 实现特性

- ✅ **自动传播**: traceID 和 spanID 在服务间自动传播
- ✅ **双协议支持**: HTTP 和 gRPC 协议都支持链路追踪
- ✅ **Jaeger 集成**: 通过 Jaeger 进行链路数据收集和可视化
- ✅ **可配置采样**: 支持配置采样率控制性能影响
- ✅ **标准兼容**: 基于 OpenTelemetry 标准

## 配置说明

### 配置文件 (`configs/config.yaml`)

```yaml
trace:
  endpoint: http://localhost:14268/api/traces  # Jaeger 收集器端点
  service_name: user-service                   # 服务名称
  sampler: 1.0                                 # 采样率 (1.0 = 100%)
  batcher: jaeger                              # 批处理器类型
```

### 配置参数说明

| 参数 | 说明 | 默认值 | 示例 |
|------|------|--------|------|
| `endpoint` | Jaeger 收集器端点地址 | - | `http://localhost:14268/api/traces` |
| `service_name` | 服务名称，用于在 Jaeger 中识别 | - | `user-service` |
| `sampler` | 采样率，0.0-1.0 之间 | 1.0 | `0.1` (10% 采样) |
| `batcher` | 批处理器类型 | jaeger | `jaeger` |

## 使用指南

### 1. 启动 Jaeger 服务

使用 Docker 启动 Jaeger:

```bash
docker run -d \
  --name jaeger \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 16686:16686 \
  -p 14268:14268 \
  jaegertracing/all-in-one:latest
```

访问 Jaeger UI: http://localhost:16686

### 2. 启动服务

```bash
go run ./cmd/auth -conf ./configs
```

### 3. 发送测试请求

HTTP 接口测试:
```bash
curl http://localhost:8000/helloworld/{name}
```

gRPC 接口测试:
```bash
# 使用 grpcurl 或其他 gRPC 客户端工具
```

### 4. 查看链路追踪

在 Jaeger UI 中:
1. 选择服务: `user-service`
2. 点击 "Find Traces" 查看链路
3. 点击具体链路查看详细信息

## 代码中使用链路追踪

### 基本用法

```go
import "user/internal/pkg/tracing"

func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    // 开始一个新的 span
    ctx, span := tracing.StartSpan(ctx, "user-service.get-user")
    defer span.End()

    // 添加标签
    span.SetAttributes(
        attribute.String("user.id", req.Id),
        attribute.String("operation", "get_user"),
    )

    // 添加事件
    tracing.AddSpanEvent(ctx, "database.query.start", map[string]interface{}{
        "query": "SELECT * FROM users WHERE id = ?",
        "user_id": req.Id,
    })

    // 业务逻辑
    user, err := s.repo.GetUser(ctx, req.Id)
    if err != nil {
        // 添加错误事件
        tracing.AddSpanEvent(ctx, "database.query.error_reason", map[string]interface{}{
            "error_reason": err.Error(),
        })
        return nil, err
    }

    tracing.AddSpanEvent(ctx, "database.query.success", map[string]interface{}{
        "user_found": user != nil,
    })

    return &pb.GetUserResponse{User: user}, nil
}
```

### 在数据访问层使用

```go
func (r *userRepo) GetUser(ctx context.Context, id string) (*User, error) {
    ctx, span := tracing.StartSpan(ctx, "user-repo.get-user")
    defer span.End()

    span.SetAttributes(
        attribute.String("repository.method", "GetUser"),
        attribute.String("user.id", id),
    )

    var user User
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
    if err != nil {
        return nil, err
    }

    return &user, nil
}
```

## 架构说明

### 文件结构

```
user/
├── internal/
│   ├── conf/
│   │   ├── conf.proto           # 配置结构定义
│   │   └── conf.pb.go           # 生成的配置代码
│   ├── pkg/
│   │   └── tracing/             # 链路追踪模块
│   │       ├── provider.go      # OpenTelemetry 提供者配置
│   │       ├── tracer.go        # 追踪工具函数
│   │       └── example.go       # 使用示例
│   └── server/
│       ├── http.go              # HTTP 服务器 (已添加 tracing 中间件)
│       └── grpc.go              # gRPC 服务器 (已添加 tracing 中间件)
├── cmd/auth/
│   ├── main.go                  # 应用入口 (已初始化链路追踪)
│   ├── wire.go                  # 依赖注入配置
│   └── wire_gen.go              # 生成的 wire 代码
├── configs/
│   └── config.yaml              # 配置文件 (已添加 trace 配置)
└── go.mod                       # 依赖管理
```

### 中间件集成

- **HTTP 中间件**: `internal/server/http.go:21`
- **gRPC 中间件**: `internal/server/grpc.go:21`
- **自动追踪**: 所有 HTTP 和 gRPC 请求都会自动创建 span

## 监控指标

### 可追踪的信息

- **请求链路**: 完整的请求处理路径
- **性能指标**: 各个操作的耗时统计
- **错误信息**: 请求处理中的错误和异常
- **自定义标签**: 业务相关的上下文信息
- **服务依赖**: 跨服务调用关系

### 常用标签

- `service.name`: 服务名称
- `service.version`: 服务版本
- `http.method`: HTTP 方法
- `http.status_code`: HTTP 状态码
- `grpc.method`: gRPC 方法名
- `grpc.status_code`: gRPC 状态码

## 性能考虑

### 采样率建议

- **开发环境**: `1.0` (100% 采样)
- **测试环境**: `0.1` (10% 采样)
- **生产环境**: `0.01` (1% 采样)

### 性能影响

- **CPU 开销**: 约 5-15%
- **内存开销**: 每个 span 约 1KB
- **网络开销**: 批量发送，可配置

## 故障排查

### 常见问题

1. **链路数据未显示**
   - 检查 Jaeger 服务是否启动
   - 验证配置文件中的 endpoint
   - 检查网络连接

2. **traceID/spanID 为空**
   - 确认中间件已正确配置
   - 检查日志输出是否包含 trace 信息

3. **性能影响过大**
   - 降低采样率
   - 检查 span 数量是否过多

### 调试方法

```bash
# 查看日志中的 trace 信息
grep "trace.id" /var/log/user-service.log

# 检查 Jaeger 连接
curl http://localhost:14268/api/traces
```

## 最佳实践

1. **合理命名**: span 名称应该清晰表达操作内容
2. **适当标签**: 添加有意义的业务标签，避免敏感信息
3. **事件记录**: 在关键操作点添加事件记录
4. **错误处理**: 在异常情况下记录错误信息
5. **性能监控**: 定期检查链路追踪的性能影响

## 后续优化

- [ ] 添加自定义指标 (Metrics)
- [ ] 集成日志关联
- [ ] 添加告警规则
- [ ] 性能优化和调优