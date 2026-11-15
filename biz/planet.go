package biz

import (
	"SolarNotesAdminAPI/model"
	"SolarNotesAdminAPI/sys_error"
	"context"
	"errors"
)

// names 行星名称列表,用于前端匹配 3D 资源
var names = [9]string{
	"Sun",
	"Mercury",
	"Venus",
	"Earth",
	"Mars",
	"Jupiter",
	"Saturn",
	"Uranus",
	"Neptune",
}

// Planet 本结构体用于表示星球业务实体
type Planet struct {
	ID    int64  // ID 行星的唯一标识符
	Name  string // Name 行星名称 (用于前台前端匹配3D资源)
	Title string // Title 行星标题
	Intro string // Intro 行星简介
	Exist bool   // Exist 行星是否存在 (用于判定能否创建该天体)
}

// ShowStatus 本方法用于查询所有行星的状态,并返回一个包含行星状态的切片
func (p *Planet) ShowStatus(ctx context.Context) ([]*Planet, error) {
	if p == nil || ctx == nil {
		return nil, &sys_error.InternalError{
			Cause: errors.New("nil planet entity or context"),
		}
	}

	planetNames := make([]string, 0, len(names))
	for _, name := range names {
		planetNames = append(planetNames, name)
	}

	orm := &model.Planet{}
	records, err := orm.FindByNames(ctx, planetNames)
	if err != nil {
		return nil, &sys_error.DatabaseError{Cause: err}
	}

	recordsByName := make(map[string]*model.Planet, len(records))
	for _, record := range records {
		if record == nil {
			return nil, &sys_error.InternalError{
				Cause: errors.New("nil planet record"),
			}
		}

		recordsByName[record.Name] = record
	}

	planets := make([]*Planet, 0, len(recordsByName))
	for _, name := range planetNames {
		planet := &Planet{
			Name:  name,
			Exist: false,
		}

		record, exists := recordsByName[name]
		if exists {
			planet.fillByORM(ctx, record)
			planet.Exist = true
		}

		planets = append(planets, planet)
	}

	return planets, nil
}

// fillByORM 本方法用于将数据库查询结果填充到业务实体中
func (p *Planet) fillByORM(ctx context.Context, record *model.Planet) {
	p.ID = record.ID
	p.Name = record.Name
	p.Title = record.Title
	p.Intro = record.Intro
}
