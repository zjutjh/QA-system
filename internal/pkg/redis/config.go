package redis

import (
	"github.com/go-redis/redis/v8"
	"github.com/zjutjh/WeJH-SDK/redisHelper"
)

// RedisClient Redis客户端
var RedisClient *redis.Client

func init() {
	info := getConfig()

	RedisClient = redisHelper.Init(&info)
}
