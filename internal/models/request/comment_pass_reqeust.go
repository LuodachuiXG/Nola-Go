package request

// CommentPassRequest 评论通过审核结构体
type CommentPassRequest struct {
	// Ids 评论 ID 数组
	Ids []uint `json:"ids" binding:"required"`
	// IsPass 是否通过审核
	IsPass *bool `json:"isPass" binding:"required"`
}
