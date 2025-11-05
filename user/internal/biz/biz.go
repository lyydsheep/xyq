package biz

import (
	"github.com/google/wire"
	"user/internal/pkg/snowflake"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewGreeterUsecase,
	NewUserUsecase,
	NewAuthUsecase,
	wire.Bind(new(SnowflakeIDGenerator), new(*snowflake.SnowflakeGenerator)),
	snowflake.DefaultSnowflakeConfig,
	snowflake.NewSnowflakeGenerator,
)
