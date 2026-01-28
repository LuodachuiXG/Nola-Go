package models

import (
	"nola-go/internal/models/enum"
)

// PostMetaData 文章导出元数据
type PostMetaData struct {
	Title               string            `json:"title"`
	AutoGenerateExcerpt bool              `json:"autoGenerateExcerpt"`
	Excerpt             string            `json:"excerpt"`
	Slug                string            `json:"slug"`
	Cover               *string           `json:"cover"`
	AllowComment        bool              `json:"allowComment"`
	Pinned              bool              `json:"pinned"`
	Status              enum.PostStatus   `json:"status"`
	Visible             enum.PostVisible  `json:"visible"`
	Visit               uint              `json:"visit"`
	Category            *CategoryMetaData `json:"category"`
	Tags                []TagMetaData     `json:"tags"`
	CreateTime          int64             `json:"createTime"`
	Comments            []CommentMetaData `json:"comments"`
}

// CategoryMetaData 分类元数据
type CategoryMetaData struct {
	DisplayName string `json:"displayName"`
	Slug        string `json:"slug"`
}

// TagMetaData 标签元数据
type TagMetaData struct {
	DisplayName string `json:"displayName"`
	Slug        string `json:"slug"`
}

// CommentMetaData 评论元数据 TODO("待补充")
type CommentMetaData struct{}
