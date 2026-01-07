package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	// Database configuration
	DB struct {
		DataSource string
	}

	DBPool struct {
		MaxOpenConns    int
		MaxIdleConns    int
		ConnMaxLifetime int
	}

	// Redis configuration
	RedisConf []cache.NodeConf

	// Kafka configuration
	Kafka struct {
		Brokers []string
		Topics  struct {
			OrderCreated       string
			OrderPaid          string
			OrderShipped       string
			OrderCompleted     string
			OrderCancelled     string
			OrderStatusChanged string
		}
	}

	// Product RPC client - to check stock and deduct inventory
	ProductRpc zrpc.RpcClientConf

	// Cart RPC client - to clear cart after order creation
	CartRpc zrpc.RpcClientConf

	// Business configuration
	Order struct {
		CancelTimeout   int64 // seconds
		CompleteTimeout int64 // seconds
	}
}
