package svc

import (
	"letsgo/gateway/internal/config"
	"letsgo/gateway/internal/middleware"
	"letsgo/services/cart/rpc/cart_client"
	"letsgo/services/order/rpc/order_client"
	"letsgo/services/payment/rpc/payment_client"
	"letsgo/services/product/rpc/product_client"
	"letsgo/services/user/rpc/user_client"
	"time"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	Auth       rest.Middleware
	AdminAuth  rest.Middleware
	Timeout    rest.Middleware
	UserRpc    user_client.User
	ProductRpc product_client.Product
	CartRpc    cart_client.Cart
	OrderRpc   order_client.Order
	PaymentRpc payment_client.Payment
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:     c,
		Auth:       middleware.NewAuthMiddleware(c.Auth.AccessSecret).Handle,
		AdminAuth:  middleware.NewAdminAuthMiddleware(c.Auth.AccessSecret).Handle,
		Timeout:    middleware.NewTimeoutMiddleware(time.Duration(c.RequestTimeout) * time.Millisecond).Handle,
		UserRpc:    user_client.NewUser(zrpc.MustNewClient(c.UserRpc)),
		ProductRpc: product_client.NewProduct(zrpc.MustNewClient(c.ProductRpc)),
		CartRpc:    cart_client.NewCart(zrpc.MustNewClient(c.CartRpc)),
		OrderRpc:   order_client.NewOrder(zrpc.MustNewClient(c.OrderRpc)),
		PaymentRpc: payment_client.NewPayment(zrpc.MustNewClient(c.PaymentRpc)),
	}
}
