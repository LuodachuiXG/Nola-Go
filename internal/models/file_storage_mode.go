package models

import "nola-go/internal/models/enum"

// FileStorageModes 文件存储方式表结构
type FileStorageModes struct {
	// FileStorageModeId 文件存储方式 ID
	FileStorageModeId uint `gorm:"column:file_storage_mode_id;primaryKey;autoIncrement" json:"fileStorageModeId"`
	// StorageMode 文件存储方式
	StorageMode *enum.FileStorageMode `gorm:"column:storage_mode;type:varchar(48);not null" json:"storageMode"`
	// Config 配置
	Config string `gorm:"column:config;type:text;not null" json:"config"`
}

func (FileStorageModes) TableName() string {
	return "file_storage_mode"
}
