package models

import "nola-go/internal/models/enum"

// File 文件结构体
type File struct {
	// FileId 文件 ID
	FileId uint `gorm:"column:file_id;primaryKey;autoIncrement" json:"fileId"`
	// FileGroupId 文件组 ID
	FileGroupId *uint `gorm:"column:file_group_id" json:"fileGroupId"`
	// DisplayName 文件名
	DisplayName string `gorm:"column:display_name;type:varchar(512);not null" json:"displayName"`
	// Size 文件大小
	Size int64 `gorm:"column:size;not null" json:"size"`
	// StorageMode 文件存储方式
	StorageMode enum.FileStorageMode `gorm:"column:storage_mode;type:varchar(48);not null" json:"storageMode"`
	// CreateTime 创建时间戳毫秒
	CreateTime int64 `gorm:"column:create_time;autoCreateTime:milli;not null" json:"createTime"`
}

func (File) TableName() string {
	return "file"
}
