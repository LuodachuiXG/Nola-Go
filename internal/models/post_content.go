package models

import "nola-go/internal/models/enum"

// PostContent 文章内容
type PostContent struct {
	// PostContentId 文章内容 ID
	PostContentId uint `gorm:"column:post_content_id;primaryKey;autoIncrement" json:"postContentId"`
	// PostId 文章 ID
	PostId uint `gorm:"column:post_id;not null" json:"postId"`
	// Content 内容
	Content string `gorm:"column:content;type:text COLLATE utf8mb4_general_ci;not null" json:"content"`
	// HTML （由 content 解析得来）
	HTML string `gorm:"column:html;type:text COLLATE utf8mb4_general_ci;not null" json:"html"`
	// Status 状态
	Status enum.PostContentStatus `gorm:"column:status;type:varchar(24);not null" json:"status"`
	// DraftName 草稿名
	DraftName *string `gorm:"column:draft_name;size:256" json:"draftName"`
	// LastModifyTime 最后修改时间
	LastModifyTime *int64 `gorm:"column:last_modify_time" json:"lastModifyTime"`
}

func (PostContent) TableName() string {
	return "post_content"
}
