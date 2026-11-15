package model

import (
	"SolarNotesAdminAPI/db"
	"context"
)

// Planet 本结构体用于和数据库中的planet表进行映射
type Planet struct {
	ID    int64  `gorm:"column:id;primaryKey"` // ID 行星的唯一标识符
	Name  string `gorm:"column:name"`          // Name 行星名称 (用于前台前端匹配3D资源)
	Title string `gorm:"column:title"`         // Title 行星标题
	Intro string `gorm:"column:intro"`         // Intro 行星简介
}

// FindByNames 本方法用于根据给定的行星名称列表查询行星信息
func (p *Planet) FindByNames(ctx context.Context, names []string) ([]*Planet, error) {
	var planets []*Planet
	err := db.DB.WithContext(ctx).Where("name IN ?", names).Find(&planets).Error
	if err != nil {
		return nil, err
	}

	return planets, nil
}
