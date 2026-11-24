package sys_error

import "errors"

var (
	PlanetExistErr = errors.New("请求创建的天体已存在,无法创建")

	// PlanetNotFoundErr 表示正在编辑的天体已不存在.
	PlanetNotFoundErr = errors.New("所选天体不存在,请刷新天体列表")

	// PlanetEditChangedErr 表示有管理员已修改了快照中的标题或简介.
	PlanetEditChangedErr = errors.New("天体标题或简介已变化,请刷新后重新确认")
)
