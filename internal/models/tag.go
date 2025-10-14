package models

// Tag 标签
type Tag struct {
	// TagId 标签 ID
	TagId uint `gorm:"column:tag_id;primaryKey;autoIncrement" json:"tagId"`
	// DisplayName 标签名
	DisplayName string `gorm:"column:display_name;size:256;not null" json:"displayName"`
	// Slug 标签别名
	Slug string `gorm:"column:slug;size:128;uniqueIndex;not null" json:"slug"`
	// Color 标签颜色
	Color *string `gorm:"column:color;size:24" json:"color"`
	// PostCount 文章数量（无数据库字段）
	PostCount *uint `gorm:"-" json:"postCount"`
}

func (Tag) TableName() string {
	return "tag"
}
