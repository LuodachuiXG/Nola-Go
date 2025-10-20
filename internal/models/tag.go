package models

// Tag 标签
type Tag struct {
	// TagId 标签 ID
	TagId uint `gorm:"column:tag_id;primaryKey;autoIncrement" json:"tagId" binding:"required"`
	// DisplayName 标签名
	DisplayName string `gorm:"column:display_name;size:256;not null" json:"displayName" binding:"required"`
	// Slug 标签别名
	Slug string `gorm:"column:slug;size:128;uniqueIndex;not null" json:"slug" binding:"required"`
	// Color 标签颜色
	Color *string `gorm:"column:color;size:24" json:"color"`
	// PostCount 文章数量（无数据库字段，只读字段）
	PostCount int64 `gorm:"->;column:post_count" json:"postCount"`
}

func (Tag) TableName() string {
	return "tag"
}
