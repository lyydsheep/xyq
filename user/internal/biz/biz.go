package biz

import (
	biz2 "edge/internal/biz"
	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewGreeterUsecase,
	NewUserUsecase,
	biz2.NewAuthUsecase,
)
