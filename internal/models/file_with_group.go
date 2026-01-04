package models

import "nola-go/internal/models/enum"

// FileWithGroup 文件和文件组数据类
type FileWithGroup struct {
	// FileId 文件 ID
	FileId uint `gorm:"column:fileId" json:"fileId"`
	// FileGroupId 文件组 ID
	FileGroupId *uint `gorm:"column:fileGroupId" json:"fileGroupId"`
	// FileName 文件名
	FileName string `gorm:"column:fileName"json:"fileName"`
	// FileGroupName 文件组名
	FileGroupName *string `gorm:"column:fileGroupName" json:"fileGroupName"`
	// FileGroupPath 文件组路径
	FileGroupPath *string `gorm:"column:fileGroupPath" json:"fileGroupPath"`
	// Size 文件大小
	Size int64 `gorm:"column:size" json:"size"`
	// StorageMode 文件存储方式
	StorageMode enum.FileStorageMode `gorm:"column:storageMode" json:"storageMode"`
	// CreateTime 文件创建时间
	CreateTime int64 `gorm:"column:createTime" json:"createTime"`
}
