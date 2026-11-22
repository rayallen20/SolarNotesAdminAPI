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
	if err := bindTreeRequest(c, req); err != nil {
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

// bindTreeRequest 只接受一个完整 JSON 值, 然后执行 Gin 的字段校验.
func bindTreeRequest(c *gin.Context, req *reqCatalogue.TreeReq) error {
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(req); err != nil {
		return err
	}

	err := decoder.Decode(&struct{}{})
	if err != io.EOF {
		err = errors.New("request body must contain one JSON object")
		return err
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