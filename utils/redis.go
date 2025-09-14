package utils

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func InitRedis(addr, password string, db int) error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := RedisClient.Ping(ctx).Result()
	return err
}

func GetRedis() *redis.Client {
	return RedisClient
}

// StoreTokenInRedis 存储token到Redis
func StoreTokenInRedis(userID, token string, expiration time.Duration) error {
	ctx := context.Background()
	return RedisClient.Set(ctx, "user_token:"+userID, token, expiration).Err()
}

// GetTokenFromRedis 从Redis获取token
func GetTokenFromRedis(userID string) (string, error) {
	ctx := context.Background()
	return RedisClient.Get(ctx, "user_token:"+userID).Result()
}

// RefreshTokenInRedis 刷新token有效期
func RefreshTokenInRedis(userID string, expiration time.Duration) error {
	ctx := context.Background()
	return RedisClient.Expire(ctx, "user_token:"+userID, expiration).Err()
}

// DeleteTokenFromRedis 从Redis删除token
func DeleteTokenFromRedis(userID string) error {
	ctx := context.Background()
	return RedisClient.Del(ctx, "user_token:"+userID).Err()
}

// IsTokenValid 验证token是否有效
func IsTokenValid(userID, token string) (bool, error) {
	storedToken, err := GetTokenFromRedis(userID)
	if err != nil {
		return false, err
	}
	return storedToken == token, nil
}
