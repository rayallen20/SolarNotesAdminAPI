package article

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"SolarNotesAdminAPI/biz"
	reqArticle "SolarNotesAdminAPI/param/req/v1/article"
	respArticle "SolarNotesAdminAPI/param/resp/v1/article"
	"SolarNotesAdminAPI/resp_body"
	"SolarNotesAdminAPI/sys_error"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// bindPreview 读取一个 JSON 对象, 检查多余输入并执行 binding 校验.
func bindPreview(c *gin.Context, req *reqArticle.PreviewReq) error {
	decoder := json.NewDecoder(c.Request.Body)
	err := decoder.Decode(req)
	if err != nil {
		return err
	}
	err = decoder.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("request body must contain one JSON object")
	}
	return binding.Validator.ValidateStruct(req)
}

// Preview 用工程的统一包装器返回文件预览或公开的业务错误.
func Preview(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	body := &resp_body.RespBody{}
	req := &reqArticle.PreviewReq{}
	err := bindPreview(c, req)
	if err != nil {
		body.ValidateError(err)
		c.JSON(http.StatusOK, body)
		return
	}

	article := &biz.Article{}
	err = article.Preview(c.Request.Context(), req)
	if err != nil {
		var dbErr *sys_error.DatabaseError
		switch {
		case errors.Is(err, sys_error.CatalogueNotFoundErr):
			body.CatalogueNotFoundError()
		case errors.Is(err, sys_error.ArticlePreviewUnavailableErr):
			body.ArticlePreviewUnavailableError()
		case errors.As(err, &dbErr):
			body.DBError(dbErr)
		default:
			body.InternalError(err)
		}
		c.JSON(http.StatusOK, body)
		return
	}

	dto := &respArticle.PreviewResp{
		ID:      strconv.FormatInt(article.ID, 10),
		Title:   article.Title,
		Content: article.Content,
	}
	body.Success(map[string]any{"article": dto})
	c.JSON(http.StatusOK, body)
}
