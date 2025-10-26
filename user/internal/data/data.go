package data

import (
	"context"
	"user/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewGreeterRepo,
	NewUserRepository,
	NewUserPointRepository,
	NewPointTransactionRepository,
	NewCodeRepository,
)

// Data .
type Data struct {
	rds *redis.Client
	db  *gorm.DB
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	// 初始化Redis客户端
	rds := redis.NewClient(&redis.Options{
		Addr: c.Redis.Addr,
	})

	// 测试Redis连接
	_, err := rds.Ping(context.Background()).Result()
	if err != nil {
		log.NewHelper(logger).Errorf("Failed to connect to Redis: %v", err)
		return nil, nil, err
	}

	// 初始化MySQL数据库连接
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{})
	if err != nil {
		log.NewHelper(logger).Errorf("Failed to connect to MySQL: %v", err)
		return nil, nil, err
	}

	// 测试MySQL连接
	sqlDB, err := db.DB()
	if err != nil {
		log.NewHelper(logger).Errorf("Failed to get underlying SQL DB: %v", err)
		return nil, nil, err
	}

	err = sqlDB.Ping()
	if err != nil {
		log.NewHelper(logger).Errorf("Failed to ping MySQL: %v", err)
		return nil, nil, err
	}

	d := &Data{
		rds: rds,
		db:  db,
	}

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
		_ = rds.Close()
		_ = sqlDB.Close()
	}
	return d, cleanup, nil
}

// RedisClient 返回Redis客户端
func (d *Data) RedisClient() *redis.Client {
	return d.rds
}

// DB 返回MySQL数据库客户端
func (d *Data) DB() *gorm.DB {
	return d.db
}
