package redis

import (
	"QA-System/internal/global/config"
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/zjutjh/WeJH-SDK/redisHelper"
	"go.uber.org/zap"
)

// Init 初始化 Redis 连接和 Stream
// func Init() error {
func init() {
	// 初始化 Redis 客户端
	info := getConfig()
	RedisClient = redisHelper.Init(&info)

	// 测试连接
	ctx := context.Background()
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		// return err
	}

	// 初始化 Stream 配置
	StreamName = config.Config.GetString("redis.stream.name")
	GroupName = config.Config.GetString("redis.stream.group")

	// 创建消费者组
	err := RedisClient.XGroupCreate(ctx, StreamName, GroupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		// return err
	}

	zap.L().Info("Redis initialized successfully",
		zap.String("stream", StreamName),
		zap.String("group", GroupName))
	// return nil
}

// createStream 创建 Stream（如果不存在）
func createStream(ctx context.Context) error {
	// 检查 Stream 是否存在
	exists, err := RedisClient.Exists(ctx, StreamName).Result()
	if err != nil {
		zap.L().Error("Failed to check if stream exists", zap.Error(err))
		return err
	}

	// 如果 Stream 不存在，则创建一个空消息初始化 Stream
	if exists == 0 {
		if err := RedisClient.XAdd(ctx, &redis.XAddArgs{
			Stream: StreamName,
			Values: map[string]interface{}{"init": "stream_initialized"},
		}).Err(); err != nil {
			zap.L().Error("Failed to initialize stream", zap.Error(err))
			return err
		}
		zap.L().Info("Stream created successfully", zap.String("stream", StreamName))
	}
	return nil
}

// createConsumerGroup 创建消费者组（如果不存在）
func createConsumerGroup(ctx context.Context) error {
	// 创建消费者组，忽略已存在的错误
	err := RedisClient.XGroupCreate(ctx, StreamName, GroupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		zap.L().Error("Failed to create consumer group", zap.Error(err))
		return err
	}
	zap.L().Info("Consumer group created successfully", zap.String("group", GroupName))
	return nil
}

// PublishToStream 发布消息到 Stream
func PublishToStream(ctx context.Context, data map[string]interface{}) error {
	return RedisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: StreamName,
		Values: data,
	}).Err()
}

// ConsumeFromStream 从 Stream 消费消息
func ConsumeFromStream(ctx context.Context, consumerName string) ([]redis.XStream, error) {
	return RedisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    GroupName,
		Consumer: consumerName,
		Streams:  []string{StreamName, ">"},
		Count:    1,
		Block:    0,
	}).Result()
}

// AckMessage 确认消息处理完成
func AckMessage(ctx context.Context, messageID string) error {
	return RedisClient.XAck(ctx, StreamName, GroupName, messageID).Err()
}
