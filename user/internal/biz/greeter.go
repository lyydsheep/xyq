package biz

import (
	"context"
	"time"

	v1 "user/api/helloworld/v1"
	"user/internal/pkg/tracing"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel/attribute"
)

var (
	// ErrUserNotFound is user not found.
	ErrUserNotFound = errors.NotFound(v1.ErrorReason_USER_NOT_FOUND.String(), "user not found")
)

// Greeter is a Greeter model.
type Greeter struct {
	Hello string
}

// GreeterRepo is a Greater repo.
type GreeterRepo interface {
	Save(context.Context, *Greeter) (*Greeter, error)
	Update(context.Context, *Greeter) (*Greeter, error)
	FindByID(context.Context, int64) (*Greeter, error)
	ListByHello(context.Context, string) ([]*Greeter, error)
	ListAll(context.Context) ([]*Greeter, error)
}

// GreeterUsecase is a Greeter usecase.
type GreeterUsecase struct {
	repo GreeterRepo
	log  *log.Helper
}

// NewGreeterUsecase new a Greeter usecase.
func NewGreeterUsecase(repo GreeterRepo, logger log.Logger) *GreeterUsecase {
	return &GreeterUsecase{repo: repo, log: log.NewHelper(logger)}
}

// CreateGreeter creates a Greeter, and returns the new Greeter.
func (uc *GreeterUsecase) CreateGreeter(ctx context.Context, g *Greeter) (*Greeter, error) {
	// 开始业务逻辑层的 span
	ctx, span := tracing.StartSpan(ctx, "greeter-usecase.create-greeter")
	defer span.End()

	// 添加业务层标签
	span.SetAttributes(
		attribute.String("layer", "business"),
		attribute.String("usecase", "greeter"),
		attribute.String("operation", "create"),
		attribute.String("input.hello", g.Hello),
	)

	uc.log.WithContext(ctx).Infof("CreateGreeter: %v", g.Hello)

	// 添加业务事件
	tracing.AddSpanEvent(ctx, "business.logic.start", map[string]interface{}{
		"method": "CreateGreeter",
		"input":  g.Hello,
	})

	// 模拟业务逻辑处理时间
	time.Sleep(30 * time.Millisecond)

	// 调用数据访问层
	result, err := uc.repo.Save(ctx, g)
	if err != nil {
		tracing.AddSpanEvent(ctx, "business.logic.error_reason", map[string]interface{}{
			"error_reason": err.Error(),
			"method":       "repo.Save",
		})
		span.SetAttributes(attribute.String("error_reason", err.Error()))
		return nil, err
	}

	// 记录成功事件
	tracing.AddSpanEvent(ctx, "business.logic.success", map[string]interface{}{
		"result":             result.Hello,
		"processing_time_ms": 30,
	})

	return result, nil
}
