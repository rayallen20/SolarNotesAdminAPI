package biz

import (
	"SolarNotesAdminAPI/auth"
	"SolarNotesAdminAPI/conf"
	"SolarNotesAdminAPI/model"
	reqAdmin "SolarNotesAdminAPI/param/req/v1/admin"
	"SolarNotesAdminAPI/sys_error"
	"context"
	"errors"
)

// Admin 本结构体用于表示管理员业务实体
type Admin struct {
	ID          int64  // ID 管理员ID
	Username    string // Username 管理员用户名
	Password    string // Password 管理员密码
	AccessToken string // AccessToken 管理员登录token
}

// Login 本方法用于处理管理员登录业务逻辑,包括:
// - 验证用户名和密码
// - 生成访问令牌
// - 更新数据库中的令牌哈希值
func (a *Admin) Login(ctx context.Context, req *reqAdmin.LoginReq) error {
	if a == nil {
		return &sys_error.InternalError{
			Cause: errors.New("nil biz entity"),
		}
	}

	a.flush()

	if ctx == nil || req == nil || req.Admin == nil {
		return sys_error.InvalidLoginRequestErr
	}

	record := &model.Admin{}
	err := record.FindByName(ctx, req.Admin.Username)
	if errors.Is(err, model.AdminNotFoundErr) {
		return sys_error.InvalidCredentialsErr
	}

	if err != nil {
		return &sys_error.DatabaseError{
			Cause: err,
		}
	}

	matched, err := auth.MatchSHA256(req.Admin.Password, record.PasswordHash)
	if err != nil {
		return &sys_error.InternalError{
			Cause: err,
		}
	}
	if !matched {
		return sys_error.InvalidCredentialsErr
	}

	config := conf.Get()
	if config == nil {
		return &sys_error.InternalError{
			Cause: errors.New("configuration is not loaded"),
		}
	}
	options := &auth.JWTOptions{
		Issuer:     config.JWT.Issuer,
		Audience:   config.JWT.Audience,
		Lifetime:   config.JWT.Lifetime(),
		SigningKey: config.JWT.SigningKey(),
	}
	token, err := auth.IssueAdminToken(record.ID, options)
	if err != nil {
		return &sys_error.InternalError{Cause: err}
	}

	updated, err := record.UpdateTokenHash(ctx, auth.SHA256Hex(token))
	if err != nil {
		return &sys_error.DatabaseError{Cause: err}
	}
	if !updated {
		return sys_error.LoginStateChangedErr
	}

	a.fillByORM(ctx, record)
	a.AccessToken = token

	return nil
}

// flush 本方法用于清空管理员业务实体的所有字段,以防止实例复用导致的旧数据残留
func (a *Admin) flush() {
	a.ID = 0
	a.Username = ""
	a.Password = ""
	a.AccessToken = ""
}

// fillByORM 本方法用于根据给定的管理员ORM对象填充管理员业务结构体
func (a *Admin) fillByORM(ctx context.Context, record *model.Admin) {
	if a == nil || record == nil {
		return
	}

	a.ID = record.ID
	a.Username = record.Name
	a.Password = record.PasswordHash
}
