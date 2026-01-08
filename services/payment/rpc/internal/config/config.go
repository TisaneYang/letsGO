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

	// Database pool configuration
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
			PaymentSuccess string
			PaymentFailed  string
		}
	}

	// Third-party payment gateway configuration
	Alipay struct {
		AppId      string
		PrivateKey string
		PublicKey  string
		NotifyUrl  string
		ReturnUrl  string
		SignType   string
	}

	WechatPay struct {
		AppId     string
		MchId     string
		ApiKey    string
		NotifyUrl string
	}

	// Mock payment configuration (for development/testing)
	MockPayment struct {
		Enabled      bool
		AutoSuccess  bool
		DelaySeconds int
	}

	// Order service RPC
	OrderRpc zrpc.RpcClientConf

	// Payment settings
	Payment struct {
		Timeout  int // Payment timeout in seconds
		MaxRetry int // Max retry times
	}
}
