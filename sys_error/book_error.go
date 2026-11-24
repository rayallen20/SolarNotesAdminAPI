package sys_error

import "errors"

var (
	// BookPlanetNotFoundErr 本错误表示表示所选天体已不存在
	BookPlanetNotFoundErr = errors.New("所选天体不存在,请刷新天体选项")

	// BookNotFoundErr 本错误表示表示给定的的书籍不存在
	BookNotFoundErr = errors.New("所选书籍不存在,请刷新书籍选项")

	// BookEditChangedErr 提示管理员刷新书单并重新核对编辑内容.
	BookEditChangedErr = errors.New("书籍标题或显示状态已变化,请刷新书单后重新确认")
)
