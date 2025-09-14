package database

import (
	"context"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/zhanghuachuan/water-reminder/types"
)

var (
	DB    *gorm.DB
	Redis *RedisClient
)

type RedisClient struct {
	*redis.Client
}

// InitFromEnv 从环境变量初始化数据库和Redis
func InitFromEnv() error {
	// 初始化MySQL
	dsn := os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") +
		"@tcp(" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + ")/" +
		os.Getenv("DB_NAME") + "?charset=utf8mb4&parseTime=True&loc=Local"

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	// 自动迁移表结构
	err = DB.AutoMigrate(
		&types.User{},
		&types.ReminderConfig{},
		&types.WaterRecord{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto migrate tables: %w", err)
	}

	// 初始化Redis
	redisOpt := &redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	}
	Redis = &RedisClient{redis.NewClient(redisOpt)}
	if _, err := Redis.Ping(context.Background()).Result(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return nil
}

func GetDB() *gorm.DB {
	return DB
}

func GetRedis() *RedisClient {
	return Redis
}
