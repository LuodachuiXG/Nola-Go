package models

import "nola-go/internal/models/enum"

// FileIndex 文件索引结构体
type FileIndex struct {
	// FileId 文件 ID
	FileId *uint `json:"fileId"`
	// Name 文件名（文件名是分组名加上文件名，如：/img/1.jpg）
	Name string `json:"name" binding:"required"`
	// StorageMode 文件存储方式
	StorageMode enum.FileStorageMode `json:"storageMode" binding:"required"`
}
