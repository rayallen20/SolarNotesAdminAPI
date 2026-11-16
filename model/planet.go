package model

import (
	"SolarNotesAdminAPI/db"
	"context"
	"errors"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// PlanetNotFoundErr 本错误用于将数据库记录不存在转换为业务层可识别的错误
var PlanetNotFoundErr = errors.New("planet not found")

// PlanetDuplicateKeyErr 表示插入触发唯一键冲突(唯一键可能来自 Name 或 ID,不能直接解释为 Name 已存在)
var PlanetDuplicateKeyErr = errors.New("planet duplicate key")

// Planet 本结构体用于和数据库中的planet表进行映射
type Planet struct {
	ID    int64  `gorm:"column:id;primaryKey"` // ID 行星的唯一标识符
	Name  string `gorm:"column:name"`          // Name 行星名称 (用于前台前端匹配3D资源)
	Title string `gorm:"column:title"`         // Title 行星标题
	Intro string `gorm:"column:intro"`         // Intro 行星简介
}

// FindByNames 本方法用于根据给定的行星名称列表查询行星信息
func (p *Planet) FindByNames(ctx context.Context, names []string) ([]*Planet, error) {
	if p == nil {
		return nil, errors.New("nil planet model")
	}
	if ctx == nil {
		return nil, errors.New("nil planet query context")
	}
	if db.DB == nil {
		return nil, errors.New("database is not initialized")
	}

	var planets []*Planet
	err := db.DB.WithContext(ctx).Where("name IN ?", names).Find(&planets).Error
	if err != nil {
		return nil, err
	}

	return planets, nil
}

// FindByName 本方法用于根据给定的行星名称查询行星信息
func (p *Planet) FindByName(ctx context.Context, name string) (*Planet, error) {
	if p == nil {
		return nil, errors.New("nil planet model")
	}
	if ctx == nil {
		return nil, errors.New("nil planet query context")
	}
	if db.DB == nil {
		return nil, errors.New("database is not initialized")
	}

	var planet Planet
	err := db.DB.WithContext(ctx).Where("name = ?", name).First(&planet).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, PlanetNotFoundErr
	}

	if err != nil {
		return nil, err
	}

	return &planet, nil
}

// Insert 本方法用于插入接收者表示的一条天体记录,不覆盖已有记录
// (唯一键冲突通过错误链保留类别和底层原因,供业务层进一步判断)
func (p *Planet) Insert(ctx context.Context) error {
	if p == nil {
		return errors.New("nil planet model")
	}
	if ctx == nil {
		return errors.New("nil planet insert context")
	}
	if db.DB == nil {
		return errors.New("database is not initialized")
	}
	if p.ID <= 0 {
		return errors.New("planet ID must be positive")
	}

	result := db.DB.WithContext(ctx).Create(p)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return errors.Join(PlanetDuplicateKeyErr, err)
		}

		// 1062 表示唯一键冲突
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return errors.Join(PlanetDuplicateKeyErr, err)
		}

		return err
	}

	if result.RowsAffected != 1 {
		return errors.New("planet insert did not affect exactly one row")
	}

	return nil
}
