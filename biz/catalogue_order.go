package biz

import (
	"cmp"
	"fmt"
	"slices"

	"golang.org/x/text/collate"
)

// catalogueOrder 持有本次组树使用的中文排序器
type catalogueOrder struct {
	names *collate.Collator // names 让名称中的连续数字按数值进行比较
}

// sortChildren 排列指定节点的直接子项, 并检查同一父项下是否有重复序号.
func (o catalogueOrder) sortChildren(parent *Catalogue) error {
	children := parent.Children

	slices.SortFunc(children, o.compare)

	for i := 1; i < len(children); i++ {
		previous, current := children[i-1].Sort, children[i].Sort
		if previous != nil && current != nil && *previous == *current {
			err := fmt.Errorf("folder %d has duplicate child sort %d", parent.ID, *current)
			return err
		}
	}
	return nil
}

// compare 将已排序项放在前面, 待排序项使用中文自然序, 最后以数值 ID 兜底.
func (o catalogueOrder) compare(a, b *Catalogue) int {
	switch {
	case a.Sort != nil && b.Sort == nil:
		return -1
	case a.Sort == nil && b.Sort != nil:
		return 1
	case a.Sort != nil && b.Sort != nil:
		if order := cmp.Compare(*a.Sort, *b.Sort); order != 0 {
			return order
		}
	default:
		if order := o.names.CompareString(a.Name, b.Name); order != 0 {
			return order
		}
	}
	return cmp.Compare(a.ID, b.ID)
}
