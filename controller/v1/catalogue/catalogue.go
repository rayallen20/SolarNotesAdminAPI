package catalogue

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"SolarNotesAdminAPI/biz"
	reqCatalogue "SolarNotesAdminAPI/param/req/v1/catalogue"
	respCatalogue "SolarNotesAdminAPI/param/resp/v1/catalogue"
	"SolarNotesAdminAPI/resp_body"
	"SolarNotesAdminAPI/sys_error"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// Tree 返回选中书籍的摘要和完整目录.
func Tree(c *gin.Context) {
	req := &reqCatalogue.TreeReq{}

	body := &resp_body.RespBody{}
	if err := bindCatalogueRequest(c, req); err != nil {
		body.ValidateError(err)
		c.JSON(http.StatusOK, body)
		return
	}

	catalogue := &biz.Catalogue{}
	book, roots, err := catalogue.Tree(c.Request.Context(), req)
	if err != nil {
		var dbErr *sys_error.DatabaseError
		switch {
		case errors.Is(err, sys_error.BookNotFoundErr):
			body.BookNotFoundError()
		case errors.As(err, &dbErr):
			body.DBError(dbErr)
		default:
			body.InternalError(err)
		}
		c.JSON(http.StatusOK, body)
		return
	}

	bookResp := &respCatalogue.TreeBookResp{
		ID:            strconv.FormatInt(book.ID, 10),
		Title:         book.Title,
		Sort:          book.Sort,
		DisplayStatus: book.DisplayStatus,
		Catalogues:    treeNodeResponses(roots),
	}
	data := map[string]any{"book": bookResp}
	body.Success(data)

	c.JSON(http.StatusOK, body)
}

// bindCatalogueRequest 读取一个完整 JSON 对象, 然后校验目标请求的 binding 标签.
func bindCatalogueRequest(c *gin.Context, req any) error {
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(req); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("request body must contain one JSON object")
	}
	return binding.Validator.ValidateStruct(req)
}

// treeNodeResponses 保留业务层的顺序, 并把每层 ID 转为字符串.
func treeNodeResponses(nodes []*biz.Catalogue) []*respCatalogue.TreeNodeResp {
	result := make([]*respCatalogue.TreeNodeResp, 0, len(nodes))
	for _, node := range nodes {
		parentID := "0"

		if node.Parent != nil {
			parentID = strconv.FormatInt(node.Parent.ID, 10)
		}

		result = append(result, &respCatalogue.TreeNodeResp{
			ID:            strconv.FormatInt(node.ID, 10),
			ParentID:      parentID,
			Name:          node.Name,
			Type:          node.Type,
			Sort:          node.Sort,
			DisplayStatus: node.DisplayStatus,
			Children:      treeNodeResponses(node.Children),
		})
	}

	return result
}

// Reorder 保存一个文件夹直接子项的阅读顺序和相应的显隐状态.
func Reorder(c *gin.Context) {
	req := &reqCatalogue.ReorderReq{}
	body := &resp_body.RespBody{}
	if err := bindCatalogueRequest(c, req); err != nil {
		body.ValidateError(err)
		c.JSON(http.StatusOK, body)
		return
	}

	catalogue := &biz.Catalogue{}
	if err := catalogue.Reorder(c.Request.Context(), req); err != nil {
		var dbErr *sys_error.DatabaseError
		switch {
		case errors.Is(err, sys_error.CatalogueOrderInvalidErr):
			body.ValidateError(err)
		case errors.Is(err, sys_error.CatalogueNotFoundErr):
			body.CatalogueNotFoundError()
		case errors.Is(err, sys_error.CatalogueOrderChangedErr):
			body.CatalogueOrderChangedError()
		case errors.Is(err, sys_error.BookNotFoundErr):
			body.BookNotFoundError()
		case errors.As(err, &dbErr):
			body.DBError(dbErr)
		default:
			body.InternalError(err)
		}
		c.JSON(http.StatusOK, body)
		return
	}

	body.Success(map[string]any{})
	c.JSON(http.StatusOK, body)
}

// Display 校验请求, 设置单个节点的显示状态并返回统一响应.
func Display(c *gin.Context) {
	req := &reqCatalogue.DisplayReq{}
	body := &resp_body.RespBody{}
	err := bindCatalogueRequest(c, req)
	if err != nil {
		body.ValidateError(err)
		c.JSON(http.StatusOK, body)
		return
	}

	catalogue := &biz.Catalogue{}
	err = catalogue.Display(c.Request.Context(), req)
	if err != nil {
		var dbErr *sys_error.DatabaseError
		switch {
		case errors.Is(err, sys_error.CatalogueNotFoundErr):
			body.CatalogueNotFoundError()
		case errors.Is(err, sys_error.BookNotFoundErr):
			body.BookNotFoundError()
		case errors.Is(err, sys_error.CatalogueSortRequiredErr):
			body.CatalogueSortRequiredError()
		case errors.As(err, &dbErr):
			body.DBError(dbErr)
		default:
			body.InternalError(err)
		}
		c.JSON(http.StatusOK, body)
		return
	}

	body.Success(map[string]any{})
	c.JSON(http.StatusOK, body)
}
