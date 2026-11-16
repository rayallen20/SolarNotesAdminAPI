package sys_error

import "errors"

var (
	PlanetExistErr = errors.New("请求创建的天体已存在,无法创建")
)