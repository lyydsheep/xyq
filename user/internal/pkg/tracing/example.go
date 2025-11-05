package tracing

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// ExampleUsage 展示如何在业务代码中使用链路追踪
func ExampleUsage(ctx context.Context) {
	// 开始一个新的 span
	ctx, span := StartSpan(ctx, "business-operation")
	defer span.End()

	// 添加标签
	span.SetAttributes(
		attribute.String("user.id", "12345"),
		attribute.String("operation.type", "example"),
	)

	// 添加事件
	AddSpanEvent(ctx, "operation.started", map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"step":      1,
	})

	// 模拟业务操作
	time.Sleep(100 * time.Millisecond)

	// 添加另一个事件
	AddSpanEvent(ctx, "operation.completed", map[string]interface{}{
		"duration_ms": 100,
		"success":     true,
	})

	// 可以在嵌套调用中继续使用链路追踪
	nestedOperation(ctx)
}

func nestedOperation(ctx context.Context) {
	ctx, span := StartSpan(ctx, "nested-operation")
	defer span.End()

	span.SetAttributes(
		attribute.String("nested.type", "database"),
		attribute.String("query", "SELECT * FROM users"),
	)

	// 模拟数据库查询
	time.Sleep(50 * time.Millisecond)

	AddSpanEvent(ctx, "db.query.executed", map[string]interface{}{
		"rows_affected": 10,
	})
}
