package request

import (
	"nola-go/internal/models/enum"
)

// PostStatusRequest 文章状态请求结构体
type PostStatusRequest struct {
	// PostId 文章 ID
	PostId uint `json:"postId" binding:"required"`
	// Status 状态
	Status *enum.PostStatus `json:"status"`
	// Visible 可见性
	Visible *enum.PostVisible `json:"visible"`
	// Pinned 置顶
	Pinned *bool `json:"pinned"`
}
