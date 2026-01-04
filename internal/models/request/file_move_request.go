package request

// FileMoveRequest 文件移动请求结构体
type FileMoveRequest struct {
	// FileIds 要移动的文件 ID 数组
	FileIds []uint `json:"fileIds" binding:"required"`
	// NewFileGroupId 新的文件组 ID
	NewFileGroupId *uint `json:"newFileGroupId"`
}
