package svc

import (
	"time"

	"letsgo/services/user/model"
	"letsgo/services/user/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config    config.Config
	UserModel model.UserModel
	Redis     *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	// Initialize PostgreSQL connection
	conn := sqlx.NewSqlConn("postgres", c.DB.DataSource)

	// Set connection pool parameters
	db, err := conn.RawDB()
	if err != nil {
		logx.Errorf("Failed to get raw database connection: %v", err)
	} else {
		db.SetMaxOpenConns(c.DBPool.MaxOpenConns)
		db.SetMaxIdleConns(c.DBPool.MaxIdleConns)
		db.SetConnMaxLifetime(time.Duration(c.DBPool.ConnMaxLifetime) * time.Second)
	}

	// Initialize Redis connection
	rds := redis.MustNewRedis(c.RedisConf[0].RedisConf)

	return &ServiceContext{
		Config:    c,
		UserModel: model.NewUserModel(conn),
		Redis:     rds,
	}
}
