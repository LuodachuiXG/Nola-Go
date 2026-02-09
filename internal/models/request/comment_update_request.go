package request

// CommentUpdateRequest 修改评论请求结构体
type CommentUpdateRequest struct {
	// CommentId 评论 ID
	CommentId uint `json:"commentId" binding:"required"`
	// Content 评论内容
	Content string `json:"content" binding:"required"`
	// Site 站点地址
	Site *string `json:"site"`
	// DisplayName 评论者名称
	DisplayName string `json:"displayName" binding:"required"`
	// Email 评论者邮箱
	Email string `json:"email" binding:"required"`
	// IsPass 是否通过审核
	IsPass bool `json:"isPass"`
}
