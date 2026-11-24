package admin

import (
	"SolarNotesAdminAPI/auth"
	"SolarNotesAdminAPI/biz"
	"SolarNotesAdminAPI/diagnostics"
	"SolarNotesAdminAPI/middleware"
	"SolarNotesAdminAPI/param/req/v1/admin"
	respAdmin "SolarNotesAdminAPI/param/resp/v1/admin"
	"SolarNotesAdminAPI/resp_body"
	"SolarNotesAdminAPI/sys_error"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Login 本函数用于处理管理员登录行为
func Login(c *gin.Context) {
	req := &admin.LoginReq{}
	respBody := &resp_body.RespBody{}

	err := c.ShouldBindJSON(req)
	if err != nil {
		respBody.ValidateError(err)
		c.JSON(http.StatusOK, respBody)
		return
	}

	adminBiz := &biz.Admin{}
	err = adminBiz.Login(c.Request.Context(), req)
	if err != nil {
		var dbErr *sys_error.DatabaseError
		switch {
		case errors.Is(err, sys_error.InvalidLoginRequestErr):
			respBody.ValidateError(err)
		case errors.Is(err, sys_error.InvalidCredentialsErr):
			respBody.LoginCredentialError()
		case errors.Is(err, sys_error.LoginStateChangedErr):
			respBody.LoginStateChangedError()
		case errors.As(err, &dbErr):
			diagnostics.LogAdminLoginFailure(err)
			respBody.DBError(dbErr)
		default:
			diagnostics.LogAdminLoginFailure(err)
			respBody.InternalError(err)
		}

		c.JSON(http.StatusOK, respBody)
		return
	}

	loginResp := &respAdmin.LoginResp{}
	loginResp.Admin = &respAdmin.LoginAdminResp{
		ID:          strconv.FormatInt(adminBiz.ID, 10),
		Username:    adminBiz.Username,
		AccessToken: adminBiz.AccessToken,
	}

	data := map[string]interface{}{
		"admin": loginResp.Admin,
	}
	respBody.Success(data)

	c.JSON(http.StatusOK, respBody)
}

// ChangePassword 读取本次会话, 绑定请求, 提交成功后返回空对象.
func ChangePassword(c *gin.Context) {
	respBody := &resp_body.RespBody{}
	adminID, hasID := middleware.CurrentAdminID(c)
	sessionDigest, hasDigest := middleware.CurrentAdminTokenDigest(c)
	if !hasID || !hasDigest {
		respBody.InvalidToken()
		c.JSON(http.StatusOK, respBody)
		return
	}

	req := &admin.ChangePasswordReq{}
	err := c.ShouldBindJSON(req)
	if err != nil {
		// 固定校验消息, 避免将密码请求的解析细节写入响应.
		respBody.ValidateError(nil)
		c.JSON(http.StatusOK, respBody)
		return
	}

	administrator := &biz.Admin{ID: adminID}
	err = administrator.ChangePassword(c.Request.Context(), req, sessionDigest)
	if err != nil {
		var dbErr *sys_error.DatabaseError
		switch {
		case errors.Is(err, auth.InvalidTokenErr):
			respBody.InvalidToken()
		case errors.Is(err, sys_error.CurrentPasswordIncorrectErr):
			respBody.CurrentPasswordIncorrectError()
		case errors.As(err, &dbErr):
			respBody.DBError(dbErr)
		default:
			respBody.InternalError(&sys_error.InternalError{Cause: err})
		}
		c.JSON(http.StatusOK, respBody)
		return
	}

	respBody.Success(map[string]any{})
	c.JSON(http.StatusOK, respBody)
}
