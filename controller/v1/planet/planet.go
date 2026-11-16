package planet

import (
	"SolarNotesAdminAPI/biz"
	"SolarNotesAdminAPI/param/req/v1/planet"
	respPlanet "SolarNotesAdminAPI/param/resp/v1/planet"
	"SolarNotesAdminAPI/resp_body"
	"SolarNotesAdminAPI/sys_error"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ShowStatus 本函数用于显示天体的可用状态信息
func ShowStatus(c *gin.Context) {
	respBody := &resp_body.RespBody{}
	planetBiz := &biz.Planet{}
	planets, err := planetBiz.ShowStatus(c.Request.Context())
	if err != nil {
		var dbErr *sys_error.DatabaseError
		if errors.As(err, &dbErr) {
			respBody.DBError(dbErr)
			c.JSON(http.StatusOK, respBody)
			return
		}

		respBody.InternalError(err)
		c.JSON(http.StatusOK, respBody)
		return
	}

	resp := &respPlanet.ShowStatusResp{
		Planets: make([]*respPlanet.ShowStatusPlanetResp, 0, len(planets)),
	}

	for _, planetBiz = range planets {
		planetResp := &respPlanet.ShowStatusPlanetResp{
			Name:  planetBiz.Name,
			Exist: planetBiz.Exist,
		}

		resp.Planets = append(resp.Planets, planetResp)
	}

	data := map[string]interface{}{
		"planets": resp.Planets,
	}
	respBody.Success(data)

	c.JSON(http.StatusOK, respBody)
}

// Create 本函数用于创建天体
func Create(c *gin.Context) {
	req := &planet.CreateReq{}
	respBody := &resp_body.RespBody{}

	err := c.ShouldBindJSON(req)
	if err != nil {
		respBody.ValidateError(err)
		c.JSON(http.StatusOK, respBody)
		return
	}

	planetBiz := &biz.Planet{}
	err = planetBiz.Create(c.Request.Context(), req)
	if err != nil {
		var dbErr *sys_error.DatabaseError
		switch {
		case errors.Is(err, sys_error.PlanetExistErr):
			respBody.PlanetExistError()
		case errors.As(err, &dbErr):
			respBody.DBError(err)
		default:
			respBody.InternalError(err)
		}

		c.JSON(http.StatusOK, respBody)
		return
	}

	respBody.Success(make(map[string]any, 0))
	c.JSON(http.StatusOK, respBody)
}
