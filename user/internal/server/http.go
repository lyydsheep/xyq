package server

import (
	authv1 "user/api/auth/v1"
	userv1 "user/api/user/v1"
	"user/internal/conf"
	tracingpkg "user/internal/pkg/tracing"
	"user/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, authService *service.AuthService, userService *service.UserService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			tracingpkg.HTTPErrorResponseEnhancer(), // 添加错误响应增强中间件
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	authv1.RegisterAuthServiceHTTPServer(srv, authService)
	userv1.RegisterUserServiceHTTPServer(srv, userService)
	return srv
}
