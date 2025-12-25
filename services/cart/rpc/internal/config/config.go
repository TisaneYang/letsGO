package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	// Redis configuration for cart storage
	RedisConf redis.RedisConf

	// Cart settings
	Cart struct {
		Expire             int // Cart expiration time in seconds
		MaxItems           int // Maximum items per cart
		MaxQuantityPerItem int // Maximum quantity per item
	}

	// Product RPC client
	ProductRpc zrpc.RpcClientConf
}
