package admin

import (
	"SolarNotesAdminAPI/biz"
	"SolarNotesAdminAPI/diagnostics"
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
		var sizeErr *http.MaxBytesError
		if errors.As(err, &sizeErr) {
			respBody.ValidateError(errors.New("请求体超过大小限制"))
			c.JSON(http.StatusOK, respBody)
			return
		}

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
