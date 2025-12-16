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

	// Database connection pool settings
	DBPool struct {
		MaxOpenConns    int
		MaxIdleConns    int
		ConnMaxLifetime int
	}

	// Redis cache configuration
	RedisConf cache.CacheConf

	// JWT authentication configuration
	AuthConf struct {
		AccessSecret string
		AccessExpire int64
	}
}
