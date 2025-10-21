package models

// Category 分类
type Category struct {
	// CategoryId 分类 ID
	CategoryId uint `gorm:"column:category_id;primaryKey;autoIncrement" json:"categoryId" binding:"required"`
	// DisplayName 分类名
	DisplayName string `gorm:"column:display_name;size:256;not null" json:"displayName" binding:"required"`
	// Slug 分类别名
	Slug string `gorm:"column:slug;size:128;uniqueIndex;not null" json:"slug" binding:"required"`
	// Cover 封面
	Cover *string `gorm:"column:cover;size:128" json:"cover"`
	// UnifiedCover 是否统一封面（未单独设置封面的文章，使用分类的封面）
	UnifiedCover bool `gorm:"column:unified_cover;default:false;not null" json:"unifiedCover"`
	// PostCount 文章数量（无数据库字段，只读字段）
	PostCount int64 `gorm:"->;column:post_count" json:"postCount"`
}

func (Category) TableName() string {
	return "category"
}
