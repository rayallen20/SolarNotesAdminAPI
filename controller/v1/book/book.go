package book

import (
	"errors"
	"net/http"

	"SolarNotesAdminAPI/biz"
	reqBook "SolarNotesAdminAPI/param/req/v1/book"
	"SolarNotesAdminAPI/resp_body"
	"SolarNotesAdminAPI/sys_error"

	"github.com/gin-gonic/gin"
)

// Create 本函数用于创建书籍
func Create(c *gin.Context) {
	req := &reqBook.CreateReq{}

	respBody := &resp_body.RespBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		respBody.ValidateError(err)
		c.JSON(http.StatusOK, respBody)
		return
	}

	bookBiz := &biz.Book{}

	err := bookBiz.Create(c.Request.Context(), req)
	if err != nil {
		var dbErr *sys_error.DatabaseError
		switch {
		case errors.Is(err, sys_error.BookPlanetNotFoundErr):
			respBody.BookPlanetNotFoundError()
		case errors.Is(err, sys_error.BookSortOccupiedErr):
			respBody.BookSortOccupiedError()
		case errors.As(err, &dbErr):
			respBody.DBError(dbErr)
		default:
			respBody.InternalError(err)
		}
		c.JSON(http.StatusOK, respBody)
		return
	}

	respBody.Success(map[string]any{})
	c.JSON(http.StatusOK, respBody)
}
