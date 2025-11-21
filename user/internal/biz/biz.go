package biz

import (
	"github.com/google/wire"
	"user/internal/pkg/snowflake"
	"user/internal/conf"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewUserUsecase,
	NewAuthUsecase,
	NewEmailConfig,
	wire.Bind(new(SnowflakeIDGenerator), new(*snowflake.SnowflakeGenerator)),
	snowflake.DefaultSnowflakeConfig,
	snowflake.NewSnowflakeGenerator,
)

// NewEmailConfig 创建邮件配置
func NewEmailConfig(c *conf.Email) EmailConfig {
	return EmailConfig{
		SenderName:   c.SenderName,
		SenderEmail:  c.SenderEmail,
		SupportEmail: c.SupportEmail,
		CompanyName:  c.CompanyName,
		AppName:      c.AppName,
	}
}

// EmailProvider 提供 Email 配置给 wire 使用
func EmailProvider(bootstrap *conf.Bootstrap) *conf.Email {
	return bootstrap.Email
}
