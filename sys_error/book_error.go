package sys_error

import "errors"

var (
	// BookPlanetNotFoundErr 本错误表示表示所选天体已不存在
	BookPlanetNotFoundErr = errors.New("所选天体不存在,请刷新天体选项")

	// BookSortOccupiedErr 本错误表示在给定天体下已有书籍使用了请求的排序值
	BookSortOccupiedErr = errors.New("该天体下的排序值已被占用")

	// BookNotFoundErr 本错误表示表示给定的的书籍不存在
	BookNotFoundErr = errors.New("所选书籍不存在,请刷新书籍选项")
)
