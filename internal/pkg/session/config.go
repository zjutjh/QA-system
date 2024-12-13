package session

import (
	"strings"

	"QA-System/internal/global/config"
	"github.com/zjutjh/WeJH-SDK/redisHelper"
	"github.com/zjutjh/WeJH-SDK/sessionHelper"
)

type driver string

const (
	// Memory 内存
	Memory driver = "memory"
	// Redis redis缓存
	Redis driver = "redis"
)

var defaultName = "qa-session"

func getConfig() sessionHelper.InfoConfig {
	wc := sessionHelper.InfoConfig{}
	wc.Name = defaultName
	if config.Config.IsSet("session.name") {
		wc.Name = strings.ToLower(config.Config.GetString("session.name"))
	}

	wc.SecretKey = strings.ToLower(config.Config.GetString("session.secret"))

	wc.RedisConfig = getRedisConfig()

	return wc
}

func getRedisConfig() *redisHelper.InfoConfig {
	info := redisHelper.InfoConfig{
		Host:     "localhost",
		Port:     "6379",
		DB:       0,
		Password: "",
	}
	if config.Config.IsSet("redis.host") {
		info.Host = config.Config.GetString("redis.host")
	}
	if config.Config.IsSet("redis.port") {
		info.Port = config.Config.GetString("redis.port")
	}
	if config.Config.IsSet("redis.db") {
		info.DB = config.Config.GetInt("redis.db")
	}
	if config.Config.IsSet("redis.pass") {
		info.Password = config.Config.GetString("redis.pass")
	}
	return &info
}
