package data

import (
	"context"
	"errors"
	"fmt"
	"os"
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
	NewDB,
	NewRedis,
	NewUserRepository,
	NewCodeRepository,
	NewAuthRepository,
)

// Data .
type Data struct {
	rds *redis.Client
	db  *gorm.DB
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	// 从环境变量获取Redis密码，优先级最高
	redisPassword := os.Getenv("REDIS_PASSWORD")
	if redisPassword == "" && c.Redis.Password != "" {
		redisPassword = c.Redis.Password
	}

	// 初始化Redis客户端
	rds := redis.NewClient(&redis.Options{
		Addr:     c.Redis.Addr,
		Password: redisPassword,
	})

	// 测试Redis连接
	_, err := rds.Ping(context.Background()).Result()
	if err != nil {
		log.NewHelper(logger).Errorf("Failed to connect to Redis: %v", err)
		return nil, nil, err
	}

	// 从环境变量获取数据库密码，优先级最高
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" && c.Database.Password != "" {
		dbPassword = c.Database.Password
	}

	// 检查密码是否配置
	if dbPassword == "" {
		return nil, nil, errors.New("database password is required, set DB_PASSWORD environment variable or database.password config")
	}

	// 使用拆分的配置组件构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=True&loc=Local",
		c.Database.Username,
		dbPassword,
		c.Database.Host,
		c.Database.Port,
		c.Database.Database,
	)

	// 初始化MySQL数据库连接
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
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

// NewDB 返回MySQL数据库连接
func NewDB(data *Data) *gorm.DB {
	return data.db
}

// NewRedis 返回Redis客户端
func NewRedis(data *Data) *redis.Client {
	return data.rds
}
