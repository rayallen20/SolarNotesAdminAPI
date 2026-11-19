package validation

import (
	"errors"
	"reflect"
	"strings"

	reqBook "SolarNotesAdminAPI/param/req/v1/book"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Register 本函数用于注册自定义验证器
func Register() error {
	if binding.Validator == nil {
		return errors.New("gin 参数校验器未启用")
	}

	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok || v == nil {
		return errors.New("无法获取 Gin 默认参数校验器")
	}

	if err := v.RegisterValidation("notblank", notBlank); err != nil {
		return err
	}

	v.RegisterStructValidation(validateCreateBookSort, reqBook.CreateBookReq{})

	return nil
}

// notBlank 本函数用于检查给定字符串是否至少包含一个非空白字符,非字符串字段不通过校验
func notBlank(field validator.FieldLevel) bool {
	value := field.Field()
	if value.Kind() != reflect.String {
		return false
	}
	return strings.TrimSpace(value.String()) != ""
}

// validateCreateBookSort 本函数用于检查可见书籍的排序值(缺失和负数由 Sort 的字段规则报告)
func validateCreateBookSort(sl validator.StructLevel) {
	book := sl.Current().Interface().(reqBook.CreateBookReq)
	if book.Sort != nil && book.DisplayStatus == "visible" && *book.Sort == 0 {
		sl.ReportError(*book.Sort, "Sort", "Sort", "visible_sort_gt_zero", "")
	}
}
