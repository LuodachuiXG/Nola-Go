package enum

import (
	"encoding/json"
	"fmt"
)

// FileSort 文件排序
type FileSort string

const (
	// FileSortCreateTimeDesc 创建时间降序排序
	FileSortCreateTimeDesc FileSort = "CREATE_TIME_DESC"
	// FileSortCreateTimeAsc 创建时间升序排序
	FileSortCreateTimeAsc FileSort = "CREATE_TIME_ASC"
	// FileSortSizeDesc 文件大小降序
	FileSortSizeDesc FileSort = "SIZE_DESC"
	// FileSortSizeAsc 文件大小升序
	FileSortSizeAsc FileSort = "SIZE_ASC"
)

func FileSortPtr(s FileSort) *FileSort {
	return &s
}

// FileSortValueOf 尝试将字符串转为文件排序枚举
func FileSortValueOf(s string) *FileSort {
	switch s {
	case "CREATE_TIME_DESC":
		return FileSortPtr(FileSortCreateTimeDesc)
	case "CREATE_TIME_ASC":
		return FileSortPtr(FileSortCreateTimeAsc)
	case "SIZE_DESC":
		return FileSortPtr(FileSortSizeDesc)
	case "SIZE_ASC":
		return FileSortPtr(FileSortSizeAsc)
	default:
		return nil
	}
}

// UnmarshalJSON 自定义反序列化，验证枚举值
func (ls *FileSort) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	// 验证是否为有效枚举值
	if v := FileSortValueOf(s); v == nil {
		return fmt.Errorf("invalid FileSort: %s", s)
	}
	*ls = FileSort(s)
	return nil
}
