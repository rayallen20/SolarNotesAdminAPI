package validation

import (
	"errors"
	"reflect"
	"strings"

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

	return v.RegisterValidation("notblank", notBlank)
}

// notBlank 本函数用于检查给定字符串是否至少包含一个非空白字符,非字符串字段不通过校验
func notBlank(field validator.FieldLevel) bool {
	value := field.Field()
	if value.Kind() != reflect.String {
		return false
	}
	return strings.TrimSpace(value.String()) != ""
}
