package response

import "nola-go/internal/models/enum"

// FileResponse 文件响应结构体
type FileResponse struct {
	// FileId 文件 Id
	FileId uint `json:"fileId"`
	// FileGroupId 文件组 Id
	FileGroupId *uint `json:"fileGroupId"`
	// FileGroupName 文件组名
	FileGroupName *string `json:"fileGroupName"`
	// DisplayName 文件名
	DisplayName string `json:"displayName"`
	// Url 文件地址
	Url string `json:"url"`
	// Size 文件大小
	Size int64 `json:"size"`
	// StorageMode 文件存储方式
	StorageMode enum.FileStorageMode `json:"storageMode"`
	// CreateTime 文件创建时间戳
	CreateTime int64 `json:"createTime"`
}
