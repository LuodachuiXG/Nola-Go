package enum

import (
	"encoding/json"
	"fmt"
)

// FileStorageMode 文件存储方式枚举类
type FileStorageMode string

const (
	// FileStorageModeLocal 本地存储
	FileStorageModeLocal FileStorageMode = "LOCAL"
	// FileStorageModeTencentCOS 腾讯云对象存储
	FileStorageModeTencentCOS FileStorageMode = "TENCENT_COS"
)

func FileStorageModePtr(s FileStorageMode) *FileStorageMode {
	return &s
}

// FileStorageModeValueOf 尝试将字符串转为文件存储方式枚举
func FileStorageModeValueOf(s string) *FileStorageMode {
	switch s {
	case "LOCAL":
		return FileStorageModePtr(FileStorageModeLocal)
	case "TENCENT_COS":
		return FileStorageModePtr(FileStorageModeTencentCOS)
	default:
		return nil
	}
}

// UnmarshalJSON 自定义反序列化，验证枚举值
func (ls *FileStorageMode) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	// 验证是否为有效枚举值
	if v := FileStorageModeValueOf(s); v == nil {
		return fmt.Errorf("invalid FileStorageMode: %s", s)
	}
	*ls = FileStorageMode(s)
	return nil
}
