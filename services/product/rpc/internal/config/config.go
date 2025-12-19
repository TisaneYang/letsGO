package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	// PostgreSQL - Product Core Data
	DB struct {
		DataSource string
	}

	// Database connection pool settings
	DBPool struct {
		MaxOpenConns    int
		MaxIdleConns    int
		ConnMaxLifetime int
	}

	// Redis configuration
	RedisConf []cache.NodeConf

	// Cache settings
	Cache struct {
		ProductExpire int
		ListExpire    int
		SearchExpire  int
	}
}
