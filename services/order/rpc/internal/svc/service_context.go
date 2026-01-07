package svc

import (
	"time"

	_ "github.com/lib/pq"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"

	"letsgo/services/cart/rpc/cart_client"
	"letsgo/services/order/model"
	"letsgo/services/order/rpc/internal/config"
	"letsgo/services/product/rpc/product_client"
)

type ServiceContext struct {
	Config config.Config

	// Database models
	OrderModel     model.OrderModel
	OrderItemModel model.OrderItemModel

	// Redis cache
	Redis redis.Redis

	// Kafka brokers and topics (we'll create Writer in logic when needed)
	KafkaBrokers []string
	KafkaTopics  struct {
		OrderCreated       string
		OrderPaid          string
		OrderShipped       string
		OrderCompleted     string
		OrderCancelled     string
		OrderStatusChanged string
	}

	// RPC clients
	ProductRpc product_client.Product
	CartRpc    cart_client.Cart
}

func NewServiceContext(c config.Config) *ServiceContext {
	// Initialize PostgreSQL connection
	conn := sqlx.NewSqlConn("postgres", c.DB.DataSource)

	// Set connection pool parameters
	sqlDB, err := conn.RawDB()
	if err == nil && sqlDB != nil {
		sqlDB.SetMaxOpenConns(c.DBPool.MaxOpenConns)
		sqlDB.SetMaxIdleConns(c.DBPool.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(time.Duration(c.DBPool.ConnMaxLifetime) * time.Second)
	}

	// Initialize Redis
	rds := redis.MustNewRedis(c.RedisConf[0].RedisConf)

	ctx := &ServiceContext{
		Config: c,

		// Models
		OrderModel:     model.NewOrderModel(conn),
		OrderItemModel: model.NewOrderItemModel(conn),

		// Redis
		Redis: *rds,

		// Kafka
		KafkaBrokers: c.Kafka.Brokers,

		// RPC Clients
		ProductRpc: product_client.NewProduct(zrpc.MustNewClient(c.ProductRpc)),
		CartRpc:    cart_client.NewCart(zrpc.MustNewClient(c.CartRpc)),
	}

	// Set Kafka topic names
	ctx.KafkaTopics.OrderCreated = c.Kafka.Topics.OrderCreated
	ctx.KafkaTopics.OrderPaid = c.Kafka.Topics.OrderPaid
	ctx.KafkaTopics.OrderShipped = c.Kafka.Topics.OrderShipped
	ctx.KafkaTopics.OrderCompleted = c.Kafka.Topics.OrderCompleted
	ctx.KafkaTopics.OrderCancelled = c.Kafka.Topics.OrderCancelled
	ctx.KafkaTopics.OrderStatusChanged = c.Kafka.Topics.OrderStatusChanged

	return ctx
}
