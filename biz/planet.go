package biz

import (
	"SolarNotesAdminAPI/model"
	"SolarNotesAdminAPI/param/req/v1/planet"
	"SolarNotesAdminAPI/snowflake"
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

func (p *Planet) Create(ctx context.Context, req *planet.CreateReq) error {
	if p == nil {
		return &sys_error.InternalError{
			Cause: errors.New("nil planet entity"),
		}
	}
	p.flush(ctx)

	if ctx == nil {
		return &sys_error.InternalError{
			Cause: errors.New("nil context"),
		}
	}

	if req == nil {
		return &sys_error.InternalError{
			Cause: errors.New("nil create request"),
		}
	}

	if req.Planet == nil {
		return &sys_error.InternalError{
			Cause: errors.New("nil create planet request"),
		}
	}

	orm := &model.Planet{}
	record, err := orm.FindByName(ctx, req.Planet.Name)
	if err != nil {
		if !errors.Is(err, model.PlanetNotFoundErr) {
			return &sys_error.DatabaseError{Cause: err}
		}
	}

	if record != nil {
		return sys_error.PlanetExistErr
	}

	orm = &model.Planet{
		ID:    snowflake.Next(),
		Name:  req.Planet.Name,
		Title: req.Planet.Title,
		Intro: req.Planet.Intro,
	}
	err = orm.Insert(ctx)
	if err != nil {
		if errors.Is(err, model.PlanetDuplicateKeyErr) {
			existRecord, findErr := orm.FindByName(ctx, req.Planet.Name)
			if existRecord != nil {
				return sys_error.PlanetExistErr
			}

			if findErr != nil {
				if !errors.Is(findErr, model.PlanetNotFoundErr) {
					return &sys_error.DatabaseError{
						Cause: errors.Join(err, findErr),
					}
				}
			}
		}

		return &sys_error.DatabaseError{Cause: err}
	}

	p.fillByORM(ctx, orm)
	p.Exist = true

	return nil
}

// flush 本方法用于清空天体业务实体的所有字段,以防止实例复用导致的旧数据残留
func (p *Planet) flush(ctx context.Context) {
	p.ID = 0
	p.Name = ""
	p.Title = ""
	p.Intro = ""
	p.Exist = false
}

// List 本方法用于查询已创建的天体列表,将模型记录转换为业务实体
// 已知天体按 names 顺序排列;其他名称保留在末尾,沿用模型的 ID 升序
// 空表返回空切片,不补充尚未创建的天体
func (p *Planet) List(ctx context.Context) ([]*Planet, error) {
	if p == nil {
		return nil, &sys_error.InternalError{
			Cause: errors.New("nil planet entity"),
		}
	}

	if ctx == nil {
		return nil, &sys_error.InternalError{
			Cause: errors.New("nil context"),
		}
	}

	orm := &model.Planet{}
	records, err := orm.FindAll(ctx)
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

	// 使用同一份 names 作为展示顺序来源,跳过尚未创建的天体
	orderedRecords := make([]*model.Planet, 0, len(records))
	for _, name := range names {
		record, exists := recordsByName[name]
		if !exists {
			continue
		}
		orderedRecords = append(orderedRecords, record)
		delete(recordsByName, name)
	}

	// 保留其他入口可能写入的未知名称,按原查询顺序放到末尾
	// 遍历 records 而非 map,确保这些记录也有确定顺序
	for _, record := range records {
		if _, exists := recordsByName[record.Name]; exists {
			orderedRecords = append(orderedRecords, record)
		}
	}

	planets := make([]*Planet, 0, len(orderedRecords))
	for _, record := range orderedRecords {
		// 每条记录使用独立实体,避免列表元素共享同一个指针
		planetBiz := &Planet{}
		planetBiz.fillByORM(ctx, record)
		planetBiz.Exist = true
		planets = append(planets, planetBiz)
	}

	return planets, nil
}
