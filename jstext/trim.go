package jstext

import (
	"strings"
	"unicode"
)

// TrimSpace 本函数用于按 JavaScript String.prototype.trim 的规则移除字符串两端的空白和换行
func TrimSpace(s string) string {
	return strings.TrimFunc(s, isTrimSpace)
}

// isTrimSpace 本函数用于判断字符是否属于 JavaScript trim 使用的空白或换行集合
func isTrimSpace(r rune) bool {
	switch r {
	case '\t', '\n', '\v', '\f', '\r', '\u2028', '\u2029', '\uFEFF':
		return true
	default:
		return unicode.Is(unicode.Zs, r)
	}
}
