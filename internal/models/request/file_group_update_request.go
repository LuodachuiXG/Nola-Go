package request

// FileGroupUpdateRequest 文件组修改请求结构体
// 文件组的存储方式和文件组地址不能修改
type FileGroupUpdateRequest struct {
	// FileGroupId 文件组 ID
	FileGroupId uint `json:"fileGroupId" binding:"required"`
	// DisplayName 文件组名
	DisplayName string `json:"displayName" binding:"required"`
}
