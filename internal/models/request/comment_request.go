package request

// CommentRequest 评论请求结构体
type CommentRequest struct {
	// CommentId 评论 ID
	CommentId *uint `json:"commentId"`
	// PostId 文章 ID
	PostId uint `json:"postId" binding:"required"`
	// ParentCommentId 父评论 ID
	ParentCommentId *uint `json:"parentCommentId"`
	// ReplyCommentId 回复评论 ID
	ReplyCommentId *uint `json:"replyCommentId"`
	// Content 评论内容
	Content string `json:"content" binding:"required"`
	// Site 站点地址
	Site *string `json:"site"`
	// DisplayName 评论人名称
	DisplayName string `json:"displayName" binding:"required"`
	// Email 评论人邮箱
	Email string `json:"email" binding:"required"`
	// IsPass 是否通过审核
	IsPass bool `json:"isPass"`
}
