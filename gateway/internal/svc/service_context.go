// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"letsgo/gateway/internal/config"
	"letsgo/gateway/internal/middleware"
	"letsgo/services/product/rpc/product_client"
	"letsgo/services/user/rpc/user_client"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	Auth       rest.Middleware
	AdminAuth  rest.Middleware
	UserRpc    user_client.User
	ProductRpc product_client.Product
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:     c,
		Auth:       middleware.NewAuthMiddleware(c.Auth.AccessSecret).Handle,
		AdminAuth:  middleware.NewAdminAuthMiddleware(c.Auth.AccessSecret).Handle,
		UserRpc:    user_client.NewUser(zrpc.MustNewClient(c.UserRpc)),
		ProductRpc: product_client.NewProduct(zrpc.MustNewClient(c.ProductRpc)),
	}
}
