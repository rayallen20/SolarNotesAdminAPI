package book

import (
	"SolarNotesAdminAPI/paginate"
	respPagination "SolarNotesAdminAPI/param/resp/pagination"
	respBook "SolarNotesAdminAPI/param/resp/v1/book"
	respPlanet "SolarNotesAdminAPI/param/resp/v1/planet"
	"errors"
	"net/http"
	"strconv"

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

// List 本函数用于分页查询书籍列表
func List(c *gin.Context) {
	req := &reqBook.ListReq{}
	respBody := &resp_body.RespBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		respBody.ValidateError(err)
		c.JSON(http.StatusOK, respBody)
		return
	}

	bookBiz := &biz.Book{}
	books, totalItem, totalPage, err := bookBiz.List(c.Request.Context(), req)
	if err != nil {
		var dbErr *sys_error.DatabaseError
		switch {
		case errors.Is(err, paginate.ErrInvalidPage):
			respBody.ValidateError(err)
		case errors.As(err, &dbErr):
			respBody.DBError(dbErr)
		default:
			respBody.InternalError(err)
		}
		c.JSON(http.StatusOK, respBody)
		return
	}

	// 显式初始化非 nil 切片, 确保空页输出 [] 而非 null
	listResp := &respBook.ListResp{
		Books: make([]*respBook.ListBookResp, 0, len(books)),
		Pagination: &respPagination.Pagination{
			CurrentPage: *req.Pagination.CurrentPage,
			ItemPerPage: *req.Pagination.ItemPerPage,
			TotalPage:   totalPage,
			TotalItem:   totalItem,
		},
	}

	for _, item := range books {
		// 两种 ID 均转为字符串, 其余字段通过响应 DTO 显式选择
		bookResp := &respBook.ListBookResp{
			ID:            strconv.FormatInt(item.ID, 10),
			Title:         item.Title,
			Sort:          item.Sort,
			DisplayStatus: item.DisplayStatus,
			Planet: &respPlanet.ListPlanetResp{
				ID:    strconv.FormatInt(item.Planet.ID, 10),
				Name:  item.Planet.Name,
				Title: item.Planet.Title,
				Intro: item.Planet.Intro,
			},
		}

		listResp.Books = append(listResp.Books, bookResp)
	}

	respBody.Success(map[string]any{
		"books":      listResp.Books,
		"pagination": listResp.Pagination,
	})

	c.JSON(http.StatusOK, respBody)
}

// Options 本函数用于展示给定天体下的书籍选项
func Options(c *gin.Context) {
	req := &reqBook.OptionsReq{}
	respBody := &resp_body.RespBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		respBody.ValidateError(err)
		c.JSON(http.StatusOK, respBody)
		return
	}

	bookBiz := &biz.Book{}
	books, err := bookBiz.Options(c.Request.Context(), req)
	if err != nil {
		var dbErr *sys_error.DatabaseError
		switch {
		case errors.Is(err, sys_error.BookPlanetNotFoundErr):
			respBody.BookPlanetNotFoundError()
		case errors.As(err, &dbErr):
			respBody.DBError(dbErr)
		default:
			respBody.InternalError(err)
		}
		c.JSON(http.StatusOK, respBody)
		return
	}

	options := make([]respBook.OptionResp, 0, len(books))
	for _, item := range books {
		options = append(options, respBook.OptionResp{
			ID:            strconv.FormatInt(item.ID, 10),
			Title:         item.Title,
			Sort:          item.Sort,
			DisplayStatus: item.DisplayStatus,
		})
	}

	data := map[string]any{
		"books": options,
	}
	respBody.Success(data)
	c.JSON(http.StatusOK, respBody)
}
