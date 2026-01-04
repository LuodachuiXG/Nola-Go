package models

import "nola-go/internal/models/enum"

// FileGroup 文件组结构体
type FileGroup struct {
	// FileGroupId 文件组 ID
	FileGroupId uint `gorm:"column:file_group_id;primaryKey;autoIncrement" json:"fileGroupId"`
	// DisplayName 文件组名
	DisplayName string `gorm:"column:display_name;type:varchar(128);not null" json:"displayName"`
	// Path 文件组路径
	Path string `gorm:"column:path;type:varchar(128);not null" json:"path"`
	// StorageMode 文件存储方式
	StorageMode enum.FileStorageMode `gorm:"column:storage_mode;type:varchar(48);not null" json:"storageMode"`
}

func (FileGroup) TableName() string {
	return "file_group"
}
