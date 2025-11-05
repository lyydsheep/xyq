package service

import (
	"context"
	"time"

	v1 "user/api/helloworld/v1"
	"user/internal/biz"
	"user/internal/pkg/tracing"

	"go.opentelemetry.io/otel/attribute"
)

// GreeterService is a greeter service.
type GreeterService struct {
	v1.UnimplementedGreeterServer

	uc *biz.GreeterUsecase
}

// NewGreeterService new a greeter service.
func NewGreeterService(uc *biz.GreeterUsecase) *GreeterService {
	return &GreeterService{uc: uc}
}

// SayHello implements helloworld.GreeterServer.
func (s *GreeterService) SayHello(ctx context.Context, in *v1.HelloRequest) (*v1.HelloReply, error) {
	// 开始一个新的 span 用于追踪整个 SayHello 操作
	ctx, span := tracing.StartSpan(ctx, "greeter-service.say-hello")
	defer span.End()

	// 添加标签信息
	span.SetAttributes(
		attribute.String("service.name", "greeter"),
		attribute.String("operation", "say_hello"),
		attribute.String("input.name", in.Name),
	)

	// 添加开始事件
	tracing.AddSpanEvent(ctx, "operation.started", map[string]interface{}{
		"timestamp":  time.Now().Unix(),
		"input_name": in.Name,
	})

	// 模拟一些业务处理逻辑
	time.Sleep(50 * time.Millisecond)

	// 调用业务逻辑层
	g, err := s.uc.CreateGreeter(ctx, &biz.Greeter{Hello: in.Name})
	if err != nil {
		// 记录错误事件
		tracing.AddSpanEvent(ctx, "business.error", map[string]interface{}{
			"error":     err.Error(),
			"operation": "create_greeter",
		})
		span.SetAttributes(attribute.String("error", err.Error()))
		return nil, err
	}

	// 添加成功事件
	tracing.AddSpanEvent(ctx, "business.success", map[string]interface{}{
		"result_message":     "Hello " + g.Hello,
		"processing_time_ms": 50,
	})

	// 模拟响应处理
	time.Sleep(20 * time.Millisecond)

	return &v1.HelloReply{Message: "Hello " + g.Hello}, nil
}
