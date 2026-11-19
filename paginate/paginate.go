package paginate

import (
	"errors"
	"fmt"
	"math"

	"gorm.io/gorm"
)

// MaxPageSize 本常量用于定义单次查询返回的最大记录数
const MaxPageSize = 100

// ErrInvalidPage 本错误表示表示分页参数缺失/越界/偏移量不能表示为 int
var ErrInvalidPage = errors.New("分页参数不合法")

// Page 本结构体用于保存校验后的 LIMIT 和 OFFSET, 通过 NewPage 构造
type Page struct {
	size   int // size 本页允许返回的最大条数
	offset int // offset 本页跳过的记录数
}

// NewPage 本函数用于校验分页范围, 并在乘法之前检查偏移是否会溢出
func NewPage(currentPage, itemPerPage int) (Page, error) {
	if currentPage < 1 {
		return Page{}, fmt.Errorf("%w: 页码必须从 1 开始", ErrInvalidPage)
	}

	if itemPerPage < 1 || itemPerPage > MaxPageSize {
		return Page{}, fmt.Errorf("%w: 每页条数须在 1 到 %d 之间", ErrInvalidPage, MaxPageSize)
	}

	if currentPage-1 > math.MaxInt/itemPerPage {
		return Page{}, fmt.Errorf("%w: 页码过大", ErrInvalidPage)
	}

	page := Page{
		size:   itemPerPage,
		offset: (currentPage - 1) * itemPerPage,
	}

	return page, nil
}

// Valid 本方法用于判断 Page 是否由合法参数构造, 可用于拦截零值
func (p Page) Valid() bool {
	return p.size >= 1 && p.size <= MaxPageSize && p.offset >= 0
}

// Offset 本方法用于返回分页起点, 模型可以据此判断是否还有记录
func (p Page) Offset() int {
	return p.offset
}

// Paginate 本方法用于为已排序的 GORM 查询应用分页范围, 可传入 Scopes
func (p Page) Paginate(db *gorm.DB) *gorm.DB {
	return db.Limit(p.size).Offset(p.offset)
}

// TotalPages 本方法用于计算总页数, 只使用整数并避免 total + size - 1 溢出
func (p Page) TotalPages(total int64) int64 {
	if total <= 0 || !p.Valid() {
		return 0
	}
	return 1 + (total-1)/int64(p.size)
}
