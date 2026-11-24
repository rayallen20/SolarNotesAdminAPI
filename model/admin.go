package model

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"

	"SolarNotesAdminAPI/auth"
	"SolarNotesAdminAPI/db"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AdminNotFoundErr 本错误用于将数据库记录不存在转换为业务层可识别的错误
var AdminNotFoundErr = errors.New("administrator not found")

var (
	// AdminSessionChangedErr 表示原会话已被撤销或替换.
	AdminSessionChangedErr = errors.New("administrator session has changed")
	// AdminPasswordMismatchErr 表示会话仍有效, 但原密码核验未通过.
	AdminPasswordMismatchErr = errors.New("current administrator password does not match")
)

// Admin 本结构体用于和数据库中的admin表进行映射
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

// FindByID 本方法用于根据给定 ID 查询管理员数据.不存在的管理员会返回 AdminNotFoundErr,其他错误会原样返回
func (a *Admin) FindByID(ctx context.Context, id int64) error {
	if a == nil {
		return errors.New("administrator model is nil")
	}
	if ctx == nil {
		return errors.New("administrator query context is nil")
	}
	if id <= 0 {
		return errors.New("administrator ID must be positive")
	}
	if db.DB == nil {
		return errors.New("database is not initialized")
	}

	// 使用临时记录,避免接收者原有 ID 参与 GORM 的查询条件
	record := &Admin{}
	err := db.DB.WithContext(ctx).Where("id = ?", id).Take(record).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return AdminNotFoundErr
	}

	if err != nil {
		return err
	}

	*a = *record
	return nil
}

// ChangePassword 锁定接收者 ID 对应的行, 核对会话和原密码后同时写入两个字段.
func (a *Admin) ChangePassword(ctx context.Context, sessionDigest, currentPasswordDigest, newPasswordDigest [sha256.Size]byte) error {
	if a == nil || a.ID <= 0 {
		return errors.New("model.Admin.ChangePassword: missing administrator ID")
	}
	if ctx == nil {
		return errors.New("model.Admin.ChangePassword: nil context")
	}
	if db.DB == nil {
		return errors.New("database is not initialized")
	}

	tx := db.DB.WithContext(ctx).Begin(&sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if tx.Error != nil {
		return tx.Error
	}
	defer tx.Rollback()

	var locked Admin
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id", "password_hash", "token_hash").
		Where("id = ?", a.ID).Take(&locked).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return AdminNotFoundErr
	}
	if err != nil {
		return err
	}

	err = locked.checkSessionDigest(sessionDigest)
	if err != nil {
		return err
	}
	passwordMatches, err := auth.MatchSHA256Digest(currentPasswordDigest, locked.PasswordHash)
	if err != nil {
		return err
	}
	if !passwordMatches {
		return AdminPasswordMismatchErr
	}

	result := tx.Model(&locked).Updates(map[string]any{
		"password_hash": hex.EncodeToString(newPasswordDigest[:]),
		"token_hash":    "",
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return errors.New("administrator password update did not affect exactly one row")
	}
	return tx.Commit().Error
}

// RevokeSession 锁定当前管理员, 仅撤销与本次请求摘要匹配的会话.
func (a *Admin) RevokeSession(ctx context.Context, sessionDigest [sha256.Size]byte) error {
	if a == nil || a.ID <= 0 {
		return errors.New("model.Admin.RevokeSession: missing administrator ID")
	}
	if ctx == nil {
		return errors.New("model.Admin.RevokeSession: nil context")
	}
	if db.DB == nil {
		return errors.New("database is not initialized")
	}

	tx := db.DB.WithContext(ctx).Begin(&sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if tx.Error != nil {
		return tx.Error
	}
	defer tx.Rollback()

	var locked Admin
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id", "token_hash").
		Where("id = ?", a.ID).Take(&locked).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return AdminNotFoundErr
	}
	if err != nil {
		return err
	}
	err = locked.checkSessionDigest(sessionDigest)
	if err != nil {
		return err
	}

	// locked.ID 来自锁定查询, Update 显式写入空字符串.
	result := tx.Model(&locked).Update("token_hash", "")
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return errors.New("administrator logout did not affect exactly one row")
	}
	return tx.Commit().Error
}

// checkSessionDigest 核对已读行保存的令牌摘要, 供持有行锁的写事务使用.
func (a *Admin) checkSessionDigest(sessionDigest [sha256.Size]byte) error {
	if a.TokenHash == "" {
		return AdminSessionChangedErr
	}
	matched, err := auth.MatchSHA256Digest(sessionDigest, a.TokenHash)
	if err != nil {
		return err
	}
	if !matched {
		return AdminSessionChangedErr
	}
	return nil
}
