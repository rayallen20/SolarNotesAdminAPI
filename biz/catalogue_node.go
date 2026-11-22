package biz

import (
	"fmt"

	"SolarNotesAdminAPI/enum"
	"SolarNotesAdminAPI/model"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// buildCatalogueTree 将单本书的平面记录挂接为双向树, 空书返回空数组.
func buildCatalogueTree(bookID int64, rows []model.Catalogue) ([]*Catalogue, error) {
	if len(rows) == 0 {
		return []*Catalogue{}, nil
	}

	byID := make(map[int64]*Catalogue, len(rows))

	var root *Catalogue
	for i := range rows {
		row := &rows[i]

		if row.BookID != bookID {
			err := fmt.Errorf("node %d belongs to book %d, expected %d", row.ID, row.BookID, bookID)
			return nil, err
		}
		if _, exists := byID[row.ID]; exists {
			err := fmt.Errorf("duplicate catalogue ID %d", row.ID)
			return nil, err
		}

		if err := validateCatalogueRow(row); err != nil {
			return nil, err
		}

		node := &Catalogue{
			ID:            row.ID,
			Type:          row.Type,
			Name:          row.Name,
			Intro:         row.Intro,
			Sort:          row.Sort,
			DisplayStatus: row.DisplayStatus,
			Children:      []*Catalogue{},
		}

		if row.ParentID == 0 {
			if root != nil {
				err := fmt.Errorf("multiple roots in book %d", bookID)
				return nil, err
			}

			root = node
		}

		byID[node.ID] = node
	}

	if root == nil {
		err := fmt.Errorf("root missing in nonempty book %d", bookID)
		return nil, err
	}

	// 按模型行的 ParentID 挂接两端, 根节点的 Parent 保持 nil
	for _, row := range rows {
		if row.ParentID == 0 {
			continue
		}

		parent := byID[row.ParentID]
		if parent == nil || parent.Type != enum.TypeFolder {
			err := fmt.Errorf("node %d has missing or non-folder parent %d", row.ID, row.ParentID)
			return nil, err
		}

		node := byID[row.ID]
		node.Parent = parent
		parent.Children = append(parent.Children, node)
	}

	// 只沿 Children 从根向下遍历. 每个非根节点恰有一个父节点, 不可达分支必然含环
	order := catalogueOrder{names: collate.New(language.Make("zh-CN"), collate.Numeric)}

	queue := []*Catalogue{root}
	for i := 0; i < len(queue); i++ {
		node := queue[i]
		if err := order.sortChildren(node); err != nil {
			return nil, err
		}

		queue = append(queue, node.Children...)
	}

	if len(queue) != len(rows) {
		err := fmt.Errorf("cycle outside the root of book %d", bookID)
		return nil, err
	}

	return []*Catalogue{root}, nil
}

// validateCatalogueRow 在挂接父节点前, 按数据库 ParentID 区分根和普通节点并校验字段
func validateCatalogueRow(row *model.Catalogue) error {
	if row.ID <= 0 || row.ParentID < 0 {
		err := fmt.Errorf("invalid catalogue identity: id=%d parent_id=%d", row.ID, row.ParentID)
		return err
	}

	if row.Type != enum.TypeFolder && row.Type != enum.TypeFile {
		err := fmt.Errorf("node %d has unknown type %q", row.ID, row.Type)
		return err
	}

	if row.DisplayStatus != enum.DisplayVisible && row.DisplayStatus != enum.DisplayHidden {
		err := fmt.Errorf("node %d has unknown display status %q", row.ID, row.DisplayStatus)
		return err
	}

	if row.ParentID == 0 {
		if row.Type != enum.TypeFolder || row.Sort == nil || *row.Sort != 0 {
			err := fmt.Errorf("root %d must be a folder with sort 0", row.ID)
			return err
		}

		return nil
	}

	if row.Sort == nil && row.DisplayStatus != enum.DisplayHidden {
		err := fmt.Errorf("unsorted node %d must be hidden", row.ID)
		return err
	}

	if row.Sort != nil && *row.Sort == 0 {
		err := fmt.Errorf("non-root node %d has invalid sort %d", row.ID, *row.Sort)
		return err
	}

	return nil
}
