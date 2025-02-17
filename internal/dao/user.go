package dao

import (
	"context"
	"sync"
	"time"

	"QA-System/internal/model"

	"go.uber.org/zap"
)

// emailCache 用户邮箱缓存
var (
	emailCache sync.Map
	cacheTTL   = 30 * time.Minute
	cacheMux   sync.RWMutex
)

// emailCacheItem 缓存项结构
type emailCacheItem struct {
	email     string
	timestamp time.Time
}

// InitializeCache 初始化用户邮箱缓存
func (d *Dao) InitializeCache() {
	users := []model.User{}
	result := d.orm.Model(&model.User{}).Find(&users)

	if result.Error != nil {
		zap.L().Error("failed to cache user email", zap.Error(result.Error))
		return
	}

	for _, user := range users {
		emailCache.Store(user.ID, emailCacheItem{
			email:     user.NotifyEmail,
			timestamp: time.Now(),
		})
	}
}

// startCacheCleanup 启动清理过期缓存的 goroutine
func (d *Dao) StartCacheCleaner() {
	ticker := time.NewTicker(cacheTTL / 2)
	go func() {
		for range ticker.C {
			now := time.Now()
			emailCache.Range(func(key, value interface{}) bool {
				cacheItem := value.(emailCacheItem)
				if now.Sub(cacheItem.timestamp) > cacheTTL {
					emailCache.Delete(key)
					zap.L().Info("delete expired email cache", zap.Int("uid", key.(int)))
				}
				return true
			})
		}
	}()
}

// GetUserByUsername 根据用户名获取用户
func (d *Dao) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	result := d.orm.WithContext(ctx).Model(&model.User{}).Where("username = ?", username).First(&user)
	return &user, result.Error
}

// GetUserByID 根据用户ID获取用户
func (d *Dao) GetUserByID(ctx context.Context, id int) (*model.User, error) {
	var user model.User
	result := d.orm.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).First(&user)
	return &user, result.Error
}

// CreateUser 创建新用户
func (d *Dao) CreateUser(ctx context.Context, user *model.User) error {
	result := d.orm.WithContext(ctx).Model(&model.User{}).Create(user)
	return result.Error
}

// UpdateUserPassword 更新用户密码
func (d *Dao) UpdateUserPassword(ctx context.Context, uid int, password string) error {
	result := d.orm.WithContext(ctx).Model(&model.User{}).Where("id = ?", uid).Update("password", password)
	return result.Error
}

// UpdateUserEmail 更新用户邮箱
func (d *Dao) UpdateUserEmail(ctx context.Context, uid int, email string) error {
	result := d.orm.WithContext(ctx).Model(&model.User{}).Where("id = ?", uid).Update("notify_email", email)
	if result.Error != nil {
		return result.Error
	}
	// 同步更新缓存
	cacheMux.Lock()
	emailCache.Store(uid, emailCacheItem{
		email:     email,
		timestamp: time.Now(),
	})
	defer cacheMux.Unlock()
	return result.Error
}

// GetUserEmailByID 根据用户ID获取用户邮箱
func (d *Dao) GetUserEmailByID(ctx context.Context, uid int) (string, error) {
	// 尝试从缓存获取
	cacheMux.RLock()
	if item, ok := emailCache.Load(uid); ok {
		cacheItem := item.(emailCacheItem)
		if time.Since(cacheItem.timestamp) < cacheTTL {
			cacheMux.RUnlock()
			return cacheItem.email, nil
		}
		emailCache.Delete(uid)
	}
	defer cacheMux.RUnlock()

	// 缓存未命中，查询数据库
	user, err := d.GetUserByID(ctx, uid)
	if err != nil {
		return "", err
	}
	return user.NotifyEmail, nil
}
