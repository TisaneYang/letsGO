package svc

import (
	"time"

	"letsgo/services/product/model"
	"letsgo/services/product/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config       config.Config
	ProductModel model.ProductModel
	Redis        redis.Redis
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

	rds := redis.MustNewRedis(c.RedisConf[0].RedisConf)

	return &ServiceContext{
		Config:       c,
		ProductModel: model.NewProductModel(conn),
		Redis:        *rds,
	}
}
