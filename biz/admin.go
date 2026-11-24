package biz

import (
	"SolarNotesAdminAPI/auth"
	"SolarNotesAdminAPI/conf"
	"SolarNotesAdminAPI/model"
	reqAdmin "SolarNotesAdminAPI/param/req/v1/admin"
	"SolarNotesAdminAPI/sys_error"
	"context"
	"crypto/sha256"
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

	a.flush(ctx)

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
func (a *Admin) flush(ctx context.Context) {
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

// Authenticate 本方法用于验证管理员的访问令牌是否有效,并根据验证结果返回相应的错误信息
func (a *Admin) Authenticate(ctx context.Context, rawToken string) error {
	if a == nil {
		return &sys_error.InternalError{
			Cause: errors.New("nil administrator entity"),
		}
	}

	a.flush(ctx)

	if ctx == nil {
		return &sys_error.InternalError{
			Cause: errors.New("nil authentication context"),
		}
	}

	if rawToken == "" {
		return auth.InvalidTokenErr
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

	// 复用已有解析器 先验证签名/算法/标准声明/有效期/管理员 ID
	adminID, err := auth.ParseAdminToken(rawToken, options)
	if err != nil {
		if errors.Is(err, auth.ExpiredTokenErr) {
			return err
		}

		if errors.Is(err, auth.InvalidTokenErr) {
			return err
		}

		return &sys_error.InternalError{
			Cause: err,
		}
	}

	record := &model.Admin{}
	err = record.FindByID(ctx, adminID)
	if errors.Is(err, model.AdminNotFoundErr) {
		return auth.InvalidTokenErr
	}

	if err != nil {
		return &sys_error.DatabaseError{
			Cause: err,
		}
	}

	if record.TokenHash == "" {
		return auth.InvalidTokenErr
	}

	// MatchSHA256 对原始 JWT 计算摘要,并进行恒定时间比较
	matched, err := auth.MatchSHA256(rawToken, record.TokenHash)
	if err != nil {
		// 非空摘要格式损坏属于内部数据错误,不能放行请求
		return &sys_error.InternalError{
			Cause: err,
		}
	}

	if !matched {
		return auth.InvalidTokenErr
	}

	a.fillByORM(ctx, record)
	return nil
}

// ChangePassword 为接收者代表的已鉴权管理员修改密码并撤销原会话.
func (a *Admin) ChangePassword(ctx context.Context, req *reqAdmin.ChangePasswordReq, sessionDigest [sha256.Size]byte) error {
	if a == nil || a.ID <= 0 {
		return &sys_error.InternalError{Cause: errors.New("missing authenticated administrator ID")}
	}
	if ctx == nil {
		return &sys_error.InternalError{Cause: errors.New("nil password change context")}
	}
	if req == nil || req.Admin == nil {
		return &sys_error.InternalError{Cause: errors.New("missing password change request")}
	}

	currentDigest := sha256.Sum256([]byte(req.Admin.CurrentPassword))
	newDigest := sha256.Sum256([]byte(req.Admin.NewPassword))
	record := &model.Admin{ID: a.ID}
	err := record.ChangePassword(ctx, sessionDigest, currentDigest, newDigest)
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, model.AdminNotFoundErr), errors.Is(err, model.AdminSessionChangedErr):
		return auth.InvalidTokenErr
	case errors.Is(err, model.AdminPasswordMismatchErr):
		return sys_error.CurrentPasswordIncorrectErr
	case errors.Is(err, auth.InvalidHashErr):
		return &sys_error.InternalError{Cause: err}
	default:
		return &sys_error.DatabaseError{Cause: err}
	}
}
