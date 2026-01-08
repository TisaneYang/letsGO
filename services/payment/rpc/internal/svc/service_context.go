package svc

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"

	"letsgo/services/order/rpc/order_client"
	"letsgo/services/payment/model"
	"letsgo/services/payment/rpc/internal/config"
)

type ServiceContext struct {
	Config config.Config

	// Database models
	PaymentModel model.PaymentModel

	// Redis cache
	Redis redis.Redis

	// Kafka brokers and topics
	KafkaBrokers []string
	KafkaTopics  struct {
		PaymentSuccess string
		PaymentFailed  string
	}

	// RPC Clients
	OrderRpc order_client.Order
}

func NewServiceContext(c config.Config) *ServiceContext {
	// Initialize PostgreSQL connection
	db, err := sql.Open("postgres", c.DB.DataSource)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}

	// Configure connection pool
	db.SetMaxOpenConns(c.DBPool.MaxOpenConns)
	db.SetMaxIdleConns(c.DBPool.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(c.DBPool.ConnMaxLifetime) * time.Second)

	// Test connection
	if err := db.Ping(); err != nil {
		panic(fmt.Sprintf("failed to ping database: %v", err))
	}

	sqlConn := sqlx.NewSqlConnFromDB(db)

	// Initialize Redis
	rds, err := redis.NewRedis(redis.RedisConf{
		Host: c.RedisConf[0].Host,
		Type: c.RedisConf[0].Type,
		Pass: c.RedisConf[0].Pass,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to connect to redis: %v", err))
	}

	ctx := &ServiceContext{
		Config: c,

		// Models
		PaymentModel: model.NewPaymentModel(sqlConn),

		// Redis
		Redis: *rds,

		// Kafka
		KafkaBrokers: c.Kafka.Brokers,

		// RPC Clients
		OrderRpc: order_client.NewOrder(zrpc.MustNewClient(c.OrderRpc)),
	}

	// Set Kafka topic names
	ctx.KafkaTopics.PaymentSuccess = c.Kafka.Topics.PaymentSuccess
	ctx.KafkaTopics.PaymentFailed = c.Kafka.Topics.PaymentFailed

	return ctx
}
