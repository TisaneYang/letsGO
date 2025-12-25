package svc

import (
	"letsgo/services/cart/rpc/internal/config"
	"letsgo/services/product/rpc/product"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	Redis      *redis.Redis
	ProductRpc product.ProductClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	// Initialize Redis connection
	rds := redis.MustNewRedis(c.RedisConf)

	// Initialize Product RPC client
	productRpc := product.NewProductClient(zrpc.MustNewClient(c.ProductRpc).Conn())

	return &ServiceContext{
		Config:     c,
		Redis:      rds,
		ProductRpc: productRpc,
	}
}
