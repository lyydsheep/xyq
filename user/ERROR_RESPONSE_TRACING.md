# 错误响应追踪增强

## 概述

用户微服务现在在返回错误响应时会在 `metadata` 中自动包含 `traceid` 和 `spanid`，这将帮助快速定位问题和追踪分布式调用链。

## 功能特性

### ✅ 自动追踪信息注入
- **HTTP 错误响应**：自动为所有 HTTP 错误响应添加追踪信息
- **gRPC 错误响应**：自动为所有 gRPC 错误响应添加追踪信息
- **无侵入设计**：通过中间件实现，对业务代码完全透明

### ✅ 多层级追踪支持
- 从 OpenTelemetry 追踪上下文中提取
- 从 HTTP 请求头中提取（如果存在）
- 自动降级到合适的追踪源

## 错误响应格式

### 增强前
```json
{
  "code": 401,
  "reason": "USER_INVALID_TOKEN",
  "message": "用户认证信息无效",
  "metadata": {}
}
```

### 增强后
```json
{
  "code": 401,
  "reason": "USER_INVALID_TOKEN", 
  "message": "用户认证信息无效",
  "metadata": {
    "traceid": "4bf92f3577b34da6a3ce929d0e0e4736",
    "spanid": "00f0a4777a90e3c10a7c3e5b3e2c3c9d"
  }
}
```

## 实现原理

### 1. 中间件架构

- **HTTP 错误增强中间件** (`HTTPErrorResponseEnhancer`)：处理 HTTP 错误响应
- **gRPC 错误增强中间件** (`GRPCErrorResponseEnhancer`)：处理 gRPC 错误响应
- **通用错误增强中间件** (`ErrorResponseEnhancer`)：通用错误处理

### 2. 集成位置

已在服务器配置中集成：

**HTTP 服务器** (`internal/server/http.go`)：
```go
var opts = []http.ServerOption{
    http.Middleware(
        recovery.Recovery(),
        tracing.Server(),
        tracingpkg.HTTPErrorResponseEnhancer(), // 错误追踪增强
    ),
}
```

**gRPC 服务器** (`internal/server/grpc.go`)：
```go
var opts = []grpc.ServerOption{
    grpc.Middleware(
        recovery.Recovery(),
        tracing.Server(),
        tracingpkg.GRPCErrorResponseEnhancer(), // 错误追踪增强
    ),
}
```

### 3. 追踪信息提取流程

1. **业务执行**：正常执行业务逻辑
2. **错误检测**：如果返回错误，进入错误处理流程
3. **追踪信息提取**：
   - 尝试从 OpenTelemetry 上下文中获取 span 信息
   - 如果是 HTTP 请求，额外检查请求头中的追踪信息
4. **元数据增强**：将 traceid 和 spanid 添加到错误响应的 metadata 中
5. **响应返回**：返回包含追踪信息的增强错误响应

## 工具函数

### 提取追踪信息
```go
// 从错误中提取 traceid 和 spanid
traceID, spanID, hasTrace := tracing.ExtractTraceInfoFromError(err)
if hasTrace {
    fmt.Printf("追踪ID: traceid=%s, spanid=%s\n", traceID, spanID)
}
```

### 格式化错误信息
```go
// 格式化包含追踪信息的错误日志
formattedError := tracing.FormatErrorWithTrace(err)
// 输出: "error: 用户认证信息无效, traceid: 4bf92f3577b34da6a3ce929d0e0e4736, spanid: 00f0a4777a90e3c10a7c3e5b3e2c3c9d"
```

## 最佳实践

### 1. 日志记录
在业务代码中可以使用提供的工具函数来记录包含追踪信息的错误日志：

```go
// 在业务逻辑中
if err != nil {
    logger.WithContext(ctx).Error(tracing.FormatErrorWithTrace(err))
    return nil, err
}
```

### 2. 监控和告警
- 监控错误响应中 metadata 字段的完整性
- 设置告警规则，当错误响应缺少追踪信息时触发告警

### 3. 调试和问题定位
- 客户端可以提供 traceid 和 spanid 给运维团队
- 在分布式追踪系统中快速定位完整的调用链
- 识别问题发生的具体服务和操作

## 兼容性说明

### ✅ 完全向后兼容
- 对现有业务代码零侵入
- 现有的错误处理逻辑无需修改
- 对于没有追踪上下文的请求，metadata 字段将保持为空对象 `{}`

### ✅ 多协议支持
- HTTP RESTful API
- gRPC API
- 统一的追踪信息处理

## 测试验证

### 本地测试
```bash
# 编译测试
cd user && go build ./...

# 启动服务
go run cmd/auth/main.go
```

### 功能验证
1. 发起一个会返回错误的请求
2. 检查响应中的 metadata 字段
3. 确认包含 `traceid` 和 `spanid`

## 扩展建议

### 1. 添加更多追踪信息
可以扩展 metadata 字段以包含更多有用的追踪信息：
- `parent_spanid`: 父span ID
- `request_id`: 请求ID
- `user_id`: 用户ID（如果适用）

### 2. 追踪采样
对于高并发场景，可以考虑实现追踪采样策略，只为部分请求添加详细的追踪信息。

### 3. 性能优化
- 缓存追踪信息解析结果
- 异步写入追踪日志

## 总结

通过这个增强，现在用户微服务的所有错误响应都会自动包含分布式追踪信息，这将大大提升问题定位和系统可观测性的能力。实现方案对业务代码完全透明，是一个优雅且高效的解决方案。
