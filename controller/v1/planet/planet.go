package planet

import (
	"SolarNotesAdminAPI/biz"
	"SolarNotesAdminAPI/param/req/v1/planet"
	respPlanet "SolarNotesAdminAPI/param/resp/v1/planet"
	"SolarNotesAdminAPI/resp_body"
	"SolarNotesAdminAPI/sys_error"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ShowStatus 返回天体创建状态及本次完整性检测摘要.
func ShowStatus(c *gin.Context) {
	respBody := &resp_body.RespBody{}
	planetBiz := &biz.Planet{}
	status, err := planetBiz.ShowStatus(c.Request.Context())
	if err != nil {
		var dbErr *sys_error.DatabaseError
		if errors.As(err, &dbErr) {
			respBody.DBError(dbErr)
		} else {
			respBody.InternalError(err)
		}
		c.JSON(http.StatusOK, respBody)
		return
	}

	resp := respPlanet.ShowStatusResp{
		Planets: make([]*respPlanet.ShowStatusPlanetResp, 0, len(status.Planets)),
		Readiness: respPlanet.ReadinessResp{
			Ready:         status.Ready,
			RequiredCount: status.RequiredCount,
			CreatedCount:  status.CreatedCount,
			MissingNames:  status.MissingNames,
		},
	}
	for _, item := range status.Planets {
		resp.Planets = append(resp.Planets, &respPlanet.ShowStatusPlanetResp{
			Name:  item.Name,
			Exist: item.Exist,
		})
	}

	respBody.Success(map[string]any{
		"planets":   resp.Planets,
		"readiness": resp.Readiness,
	})
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
			respBody.DBError(dbErr)
		default:
			respBody.InternalError(err)
		}

		c.JSON(http.StatusOK, respBody)
		return
	}

	respBody.Success(make(map[string]any, 0))
	c.JSON(http.StatusOK, respBody)
}

// Update 绑定天体编辑请求, 返回与列表单行同形的已保存记录.
func Update(c *gin.Context) {
	req := &planet.UpdateReq{}
	respBody := &resp_body.RespBody{}
	err := c.ShouldBindJSON(req)
	if err != nil {
		respBody.ValidateError(err)
		c.JSON(http.StatusOK, respBody)
		return
	}

	var item biz.Planet
	err = item.Update(c.Request.Context(), req)
	if err != nil {
		var dbErr *sys_error.DatabaseError
		switch {
		case errors.Is(err, sys_error.PlanetNotFoundErr):
			respBody.PlanetNotFoundError()
		case errors.Is(err, sys_error.PlanetEditChangedErr):
			respBody.PlanetEditChangedError()
		case errors.As(err, &dbErr):
			respBody.DBError(dbErr)
		default:
			respBody.InternalError(err)
		}
		c.JSON(http.StatusOK, respBody)
		return
	}

	row := respPlanet.ListPlanetResp{
		ID:    strconv.FormatInt(item.ID, 10),
		Name:  item.Name,
		Title: item.Title,
		Intro: item.Intro,
	}
	respBody.Success(map[string]any{"planet": row})
	c.JSON(http.StatusOK, respBody)
}

// List 本函数展示天体列表
func List(c *gin.Context) {
	respBody := &resp_body.RespBody{}

	planetBiz := &biz.Planet{}
	planets, err := planetBiz.List(c.Request.Context())
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

	// 显式初始化空切片,确保没有记录时输出 [] 而非 null
	listResp := &respPlanet.ListResp{
		Planets: make([]*respPlanet.ListPlanetResp, 0, len(planets)),
	}
	for _, item := range planets {
		planetResp := &respPlanet.ListPlanetResp{
			ID:    strconv.FormatInt(item.ID, 10),
			Name:  item.Name,
			Title: item.Title,
			Intro: item.Intro,
		}
		listResp.Planets = append(listResp.Planets, planetResp)
	}

	respBody.Success(map[string]any{
		"planets": listResp.Planets,
	})
	c.JSON(http.StatusOK, respBody)
}
