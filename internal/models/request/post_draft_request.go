package request

// PostDraftRequest 文章草稿请求体
type PostDraftRequest struct {
	// PostId 文章 ID
	PostId uint `json:"postId" binding:"required"`
	// DraftName 草稿名称
	DraftName string `json:"draftName" binding:"required"`
	// Content 草稿内容
	Content string `json:"content" binding:"required"`
}
