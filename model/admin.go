package model

import (
	"SolarNotesAdminAPI/db"
	"context"
	"errors"

	"gorm.io/gorm"
)

// AdminNotFoundErr 本错误用于将数据库记录不存在转换为业务层可识别的错误
var AdminNotFoundErr = errors.New("administrator not found")

type Admin struct {
	ID           int64  `gorm:"column:id;primaryKey"` // ID 管理员ID
	Name         string `gorm:"column:name"`          // Name 管理员用户名
	PasswordHash string `gorm:"column:password_hash"` // PasswordHash 管理员密码哈希值
	TokenHash    string `gorm:"column:token_hash"`    // TokenHash 管理员登录令牌哈希值
}

// FindByName 本方法用于根据给定的管理员用户名查询管理员信息,并将查询结果填充到接收者中
func (a *Admin) FindByName(ctx context.Context, username string) error {
	if a == nil {
		return errors.New("administrator model is nil")
	}

	if db.DB == nil {
		return errors.New("database is not initialized")
	}

	// 使用临时记录接收查询结果 避免接收者已有 ID 导致隐式参与查询条件
	record := &Admin{}
	err := db.DB.WithContext(ctx).
		Select("id", "name", "password_hash").
		Where("name = ?", username).
		Take(record).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return AdminNotFoundErr
	}

	if err != nil {
		return err
	}

	*a = *record
	return nil
}

// UpdateTokenHash 本方法用于更新管理员的登录令牌哈希值,并将更新结果保存到数据库中
// 返回一个bool值表示是否成功更新了记录,因为 UPDATE 在没有实际修改数据时不会返回错误,所以需要通过 RowsAffected 来判断是否有记录被更新
func (a *Admin) UpdateTokenHash(ctx context.Context, newTokenHash string) (bool, error) {
	if a == nil {
		return false, errors.New("administrator model is nil")
	}
	if db.DB == nil {
		return false, errors.New("database is not initialized")
	}

	result := db.DB.WithContext(ctx).
		Model(&Admin{}).
		Where("id = ?", a.ID).
		Where("CAST(password_hash AS BINARY) = CAST(? AS BINARY)", a.PasswordHash).
		Update("token_hash", newTokenHash)

	if result.Error != nil {
		return false, result.Error
	}

	if result.RowsAffected != 1 {
		return false, nil
	}

	a.TokenHash = newTokenHash
	return true, nil
}
