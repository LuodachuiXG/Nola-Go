package request

// PostContentRequest 文章内容请求数据类
type PostContentRequest struct {
	// PostId 文章 ID
	PostId uint `json:"postId" binding:"required"`
	// Content 文章内容
	Content string `json:"content" binding:"required"`
}
