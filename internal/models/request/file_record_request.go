package request

import "nola-go/internal/models/enum"

// FileRecordRequest 文件记录请求结构体
type FileRecordRequest struct {
	// Name 文件名（含类型后缀）
	Name string `json:"name" binding:"required"`
	// Size 文件大小（字节 Bytes）
	Size int64 `json:"size" binding:"required"`
	// StorageMode 文件存储策略（nil 默认本地存储 LOCAL）
	StorageMode *enum.FileStorageMode `json:"storageMode"`
	// FileGroupId 文件组 ID（nil 默认不分组）
	FileGroupId *uint `json:"fileGroupId"`
}
