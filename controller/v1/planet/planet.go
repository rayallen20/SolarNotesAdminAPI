package planet

import (
	"SolarNotesAdminAPI/biz"
	respPlanet "SolarNotesAdminAPI/param/resp/v1/planet"
	"SolarNotesAdminAPI/resp_body"
	"SolarNotesAdminAPI/sys_error"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ShowStatus 本函数用于显示行星的可用状态信息
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

	for _, planet := range planets {
		planetResp := &respPlanet.ShowStatusPlanetResp{
			Name:  planet.Name,
			Exist: planet.Exist,
		}

		resp.Planets = append(resp.Planets, planetResp)
	}

	data := map[string]interface{}{
		"planets": resp.Planets,
	}
	respBody.Success(data)

	c.JSON(http.StatusOK, respBody)
}
