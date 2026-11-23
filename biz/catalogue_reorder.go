package biz

import (
	"fmt"
	"math"

	"SolarNotesAdminAPI/enum"
	"SolarNotesAdminAPI/model"
	reqCatalogue "SolarNotesAdminAPI/param/req/v1/catalogue"
	"SolarNotesAdminAPI/sys_error"
)

// siblingOrder 以 ID 查找请求中的位置. 0 仅在此内存索引中表示待排, 写库时转为 NULL.
type siblingOrder map[int64]uint32

// reorderPlan 保存已通过请求校验的原分组与目标分组.
type reorderPlan struct {
	expected       siblingOrder            // expected 保存打开抽屉时可编排节点的相对位置.
	sortedChildren []reqCatalogue.ChildRef // sortedChildren 按目标阅读顺序保存已排项, 其余项待排.
}

// newReorderPlan 在获取数据库锁之前验证两份分组使用相同的完整成员集合.
func newReorderPlan(input *reqCatalogue.ReorderCatalogueReq) (*reorderPlan, error) {
	if input == nil {
		return nil, fmt.Errorf("%w: catalogue is required", sys_error.CatalogueOrderInvalidErr)
	}
	if input.ParentID <= 0 {
		return nil, fmt.Errorf("%w: parent_id must be positive", sys_error.CatalogueOrderInvalidErr)
	}
	expected, err := indexOrderGroups(input.ParentID, input.Expected, "expected")
	if err != nil {
		return nil, err
	}
	target, err := indexOrderGroups(input.ParentID, input.Target, "target")
	if err != nil {
		return nil, err
	}
	if len(expected) != len(target) {
		return nil, fmt.Errorf("%w: expected and target have different child counts", sys_error.CatalogueOrderInvalidErr)
	}
	for id := range expected {
		if _, exists := target[id]; !exists {
			return nil, fmt.Errorf("%w: target is missing child %d", sys_error.CatalogueOrderInvalidErr, id)
		}
	}
	return &reorderPlan{expected: expected, sortedChildren: input.Target.SortedChildren}, nil
}

// indexOrderGroups 将两组 ID 合并为同一个索引, 同组和跨组重复均可被识别.
func indexOrderGroups(parentID int64, groups *reqCatalogue.OrderGroups, field string) (siblingOrder, error) {
	if groups == nil {
		return nil, fmt.Errorf("%w: %s is required", sys_error.CatalogueOrderInvalidErr, field)
	}
	if groups.SortedChildren == nil || groups.UnsortedChildren == nil {
		return nil, fmt.Errorf("%w: %s must contain both arrays", sys_error.CatalogueOrderInvalidErr, field)
	}
	if uint64(len(groups.SortedChildren)) > math.MaxUint32 {
		return nil, fmt.Errorf("%w: %s exceeds the maximum sort value", sys_error.CatalogueOrderInvalidErr, field)
	}
	order := make(siblingOrder, len(groups.SortedChildren)+len(groups.UnsortedChildren))
	for i, child := range groups.SortedChildren {
		if err := order.add(parentID, child.ID, uint32(i)+1); err != nil {
			return nil, fmt.Errorf("%s.sorted_children: %w", field, err)
		}
	}
	for _, child := range groups.UnsortedChildren {
		if err := order.add(parentID, child.ID, 0); err != nil {
			return nil, fmt.Errorf("%s.unsorted_children: %w", field, err)
		}
	}
	return order, nil
}

// add 只接受同一快照中首次出现的正数子节点 ID.
func (o siblingOrder) add(parentID, id int64, position uint32) error {
	if id <= 0 || id == parentID {
		return fmt.Errorf("%w: child ID %d must be positive and different from parent_id", sys_error.CatalogueOrderInvalidErr, id)
	}
	if _, exists := o[id]; exists {
		return fmt.Errorf("%w: duplicate child ID %d", sys_error.CatalogueOrderInvalidErr, id)
	}
	o[id] = position
	return nil
}

// matches 比对已排序列和待排集合. 输入行中非 NULL 项必须已经按 sort 升序排列.
func (o siblingOrder) matches(rows []model.Catalogue) bool {
	if len(o) != len(rows) {
		return false
	}
	var rankedCount uint64
	for _, row := range rows {
		var position uint64
		if row.Sort != nil {
			rankedCount++
			position = rankedCount
		}
		expectedPosition, exists := o[row.ID]
		if !exists || uint64(expectedPosition) != position {
			return false
		}
	}
	return true
}

// sortPositions 给目标已排项分配正整数序号, 跳过 img 文件夹占用的序号.
func (p *reorderPlan) sortPositions(reserved map[uint32]bool) (siblingOrder, error) {
	if uint64(len(p.sortedChildren))+uint64(len(reserved)) > math.MaxUint32 {
		return nil, fmt.Errorf("%w: not enough available sort positions", sys_error.CatalogueOrderInvalidErr)
	}

	positions := make(siblingOrder, len(p.sortedChildren))
	var rank uint32
	for _, child := range p.sortedChildren {
		rank++
		for reserved[rank] {
			rank++
		}
		positions[child.ID] = rank
	}
	return positions, nil
}

// changes 按目录树返回的同级成员校验快照, 并保留 img 文件夹的序号和显隐.
func (p *reorderPlan) changes(rows []model.Catalogue) (map[int64]model.Catalogue, error) {
	editable := make([]model.Catalogue, 0, len(rows))
	reserved := make(map[uint32]bool)
	for _, row := range rows {
		if isCatalogueImageFolder(row.Type, row.Name) {
			if row.Sort != nil {
				reserved[*row.Sort] = true
			}
			continue
		}
		editable = append(editable, row)
	}
	if !p.expected.matches(editable) {
		return nil, sys_error.CatalogueOrderChangedErr
	}
	positions, err := p.sortPositions(reserved)
	if err != nil {
		return nil, err
	}
	changes := make(map[int64]model.Catalogue)
	for _, row := range editable {
		target := row
		if rank := positions[row.ID]; rank > 0 {
			target.Sort = &rank
			if row.Sort == nil {
				target.DisplayStatus = enum.DisplayVisible
			}
		} else {
			target.Sort = nil
			target.DisplayStatus = enum.DisplayHidden
		}
		if !equalCatalogueSort(row.Sort, target.Sort) || row.DisplayStatus != target.DisplayStatus {
			changes[row.ID] = target
		}
	}
	return changes, nil
}

// equalCatalogueSort 按数值比较排序位置, 两项均为 nil 表示均待排.
func equalCatalogueSort(a, b *uint32) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
