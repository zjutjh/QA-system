package service

import (
	"QA-System/internal/pkg/redis"
	"context"
	"strconv"
	"time"
)

func GetUserLimit(c context.Context, stu_id string, sid int) (int, error) {
	// 从 redis 中获取用户的对该问卷的访问次数
	item := "survey_" + strconv.Itoa(sid) + "_stu_" + stu_id
	var limit int
	err := redis.RedisClient.Get(c, item).Scan(&limit)
	return limit, err
}

func SetUserLimit(c context.Context, stu_id string, sid int, limit int) error {
	// 设置用户的对该问卷的访问次数
	item := "survey_" + strconv.Itoa(sid) + "_stu_" + stu_id
	err := redis.RedisClient.Set(c, item, limit, 24*time.Hour).Err()
	return err
}

func InscUserLimit(c context.Context, stu_id string, sid int) error {
	// 更新用户的对该问卷的访问次数
	item := "survey_" + strconv.Itoa(sid) + "_stu_" + stu_id
	err := redis.RedisClient.Decr(c, item).Err()
	return err

}
