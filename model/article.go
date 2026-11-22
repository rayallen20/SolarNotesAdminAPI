package model

import (
	"context"
	"errors"

	"SolarNotesAdminAPI/db"
	"SolarNotesAdminAPI/enum"

	"gorm.io/gorm"
)

// Article 映射 article 表中预览接口使用的列.
type Article struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement:false"` // ID 是文章主键.
	CatalogueID int64  `gorm:"column:catalogue_id"`                      // CatalogueID 是唯一的文件节点引用.
	Path        string `gorm:"column:path"`                              // Path 是 Markdown 的存储路径.
}

// ArticlePreviewUnavailableErr 表示节点类型或文章记录不满足预览条件.
var ArticlePreviewUnavailableErr = errors.New("article preview unavailable")

// LoadForPreview 按节点 ID 加载文章并返回节点名称. 数据库只读一次.
func (a *Article) LoadForPreview(ctx context.Context, catalogueID int64) (string, error) {
	if db.DB == nil {
		return "", errors.New("database is not initialized")
	}

	// 左连接的文章字段可空, 即使 article 表的对应列定义为 NOT NULL.
	var row struct {
		Name      string  // Name 是 catalogue.name.
		NodeType  string  // NodeType 是 catalogue.type.
		ArticleID *int64  // ArticleID 为 nil 表示没有关联文章.
		Path      *string // Path 只用于后端定位源文件.
	}
	err := db.DB.WithContext(ctx).Model(&Catalogue{}).
		Select("catalogue.name, catalogue.type AS node_type, article.id AS article_id, article.path").
		Joins("LEFT JOIN article ON article.catalogue_id = catalogue.id").
		Where("catalogue.id = ?", catalogueID).
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", CatalogueNotFoundErr
	}
	if err != nil {
		return "", err
	}
	if row.NodeType != enum.TypeFile || row.ArticleID == nil || row.Path == nil || *row.Path == "" {
		return "", ArticlePreviewUnavailableErr
	}
	*a = Article{ID: *row.ArticleID, CatalogueID: catalogueID, Path: *row.Path}
	return row.Name, nil
}
