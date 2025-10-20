package models

// PostTag 文章标签表，用于记录文章和标签的关联关系
type PostTag struct {
	// PostTagId 文章标签 ID
	PostTagId uint `gorm:"column:post_tag_id;primaryKey;autoIncrement" json:"postTagId"`
	// PostId 文章 ID
	PostId uint `gorm:"column:post_id;not null" json:"postId"`
	// TagId 标签 ID
	TagId uint `gorm:"column:tag_id;not null" json:"tagId"`
}

func (PostTag) TableName() string {
	return "post_tag"
}
