package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"SolarNotesAdminAPI/db"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PlanetNotFoundErr 本错误用于将数据库记录不存在转换为业务层可识别的错误
var PlanetNotFoundErr = errors.New("planet not found")

// PlanetDuplicateKeyErr 表示插入触发唯一键冲突(唯一键可能来自 Name 或 ID,不能直接解释为 Name 已存在)
var PlanetDuplicateKeyErr = errors.New("planet duplicate key")

// PlanetEditChangedErr 表示被锁定记录中的标题或简介与原文不一致.
var PlanetEditChangedErr = errors.New("planet edit snapshot has changed")

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

// Update 锁定天体, 核对原文, 成功提交后将保存的记录回填到接收者.
func (p *Planet) Update(ctx context.Context, expectedTitle, expectedIntro string) error {
	if p == nil {
		return errors.New("model.Planet.Update: nil receiver")
	}
	if ctx == nil {
		return errors.New("model.Planet.Update: nil context")
	}
	if p.ID <= 0 {
		return errors.New("model.Planet.Update: planet ID must be positive")
	}
	if db.DB == nil {
		return errors.New("database is not initialized")
	}

	tx := db.DB.WithContext(ctx).Begin(&sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if tx.Error != nil {
		return tx.Error
	}
	defer tx.Rollback()

	var current Planet
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id", "name", "title", "intro").
		Where("id = ?", p.ID).Take(&current).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return PlanetNotFoundErr
	}
	if err != nil {
		return err
	}
	if current.Title != expectedTitle || current.Intro != expectedIntro {
		return PlanetEditChangedErr
	}

	if current.Title != p.Title || current.Intro != p.Intro {
		result := tx.Model(&current).Updates(map[string]any{
			"title": p.Title,
			"intro": p.Intro,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return fmt.Errorf("planet %d edit affected %d rows", p.ID, result.RowsAffected)
		}

		// 在仍持有行锁的事务中读取实际保存的值.
		err = tx.Select("id", "name", "title", "intro").Take(&current, p.ID).Error
		if err != nil {
			return err
		}
	}

	err = tx.Commit().Error
	if err != nil {
		return err
	}
	*p = current
	return nil
}

// FindAll 本方法用于查询已创建的全部天体,按 ID 升序提供稳定的读取顺序
// 使用独立切片承接结果,不将接收者中已有的 ID 或名称作为查询条件
func (p *Planet) FindAll(ctx context.Context) ([]*Planet, error) {
	if p == nil {
		return nil, errors.New("nil planet model")
	}

	if ctx == nil {
		return nil, errors.New("nil planet query context")
	}

	if db.DB == nil {
		return nil, errors.New("database is not initialized")
	}

	planets := make([]*Planet, 0)
	err := db.DB.WithContext(ctx).
		Select("id", "name", "title", "intro").
		Order("id ASC").
		Find(&planets).Error
	if err != nil {
		return nil, err
	}

	return planets, nil
}
