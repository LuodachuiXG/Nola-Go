package models

// PostCategory 文章分类表，用于记录文章和分类的关联关系
type PostCategory struct {
	// PostCategoryId 文章分类 ID
	PostCategoryId uint `gorm:"column:post_category_id;primaryKey;autoIncrement" json:"postCategoryId"`
	// PostId 文章 ID
	PostId uint `gorm:"column:post_id;not null" json:"postId"`
	// CategoryId 分类 ID
	CategoryId uint `gorm:"column:category_id;not null" json:"categoryId"`
}

func (PostCategory) TableName() string {
	return "post_category"
}
