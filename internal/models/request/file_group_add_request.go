package request

import "nola-go/internal/models/enum"

// FileGroupAddRequest 文件组请求
type FileGroupAddRequest struct {
	// DisplayName 文件组名
	DisplayName string `json:"displayName" binding:"required"`
	// Path 文件组路径
	Path string `json:"path" binding:"required"`
	// StorageMode 文件存储方式
	StorageMode enum.FileStorageMode `json:"storageMode" binding:"required"`
}
