package dao

import (
	"context"

	"QA-System/internal/model"
)

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
