package book

// ReorderReq 定义保存某个天体的完整书序所需的请求.
type ReorderReq struct {
	Planet *ReorderPlanetReq `json:"planet" binding:"required"` // Planet 是调序的范围.
}

// ReorderPlanetReq 将选中的天体与编辑前后的书序放在一起.
type ReorderPlanetReq struct {
	ID       int64         `json:"id,string" binding:"gt=0"`    // ID 是十进制字符串形式的天体 ID.
	Expected *BookSequence `json:"expected" binding:"required"` // Expected 是开始调序时的书单.
	Target   *BookSequence `json:"target" binding:"required"`   // Target 是要保存的书单.
}

// BookSequence 按数组位置表达阅读次序. 显式空数组表示空书单.
type BookSequence struct {
	Books []BookOrderRef `json:"books" binding:"required,dive"` // Books 包含显示及隐藏书籍.
}

// BookOrderRef 用 ID 引用一本书.
type BookOrderRef struct {
	ID int64 `json:"id,string" binding:"gt=0"` // ID 全程使用正 int64 精度.
}

// IDs 按原顺序提取主键, 同时保留缺少数组与空数组的区别.
func (s *BookSequence) IDs() []int64 {
	if s == nil || s.Books == nil {
		return nil
	}
	ids := make([]int64, len(s.Books))
	for i, ref := range s.Books {
		ids[i] = ref.ID
	}
	return ids
}
