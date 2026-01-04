package enum

import (
	"encoding/json"
	"fmt"
)

// DiarySort 日记排序
type DiarySort string

const (
	// DiarySortCreateTimeDesc 创建时间降序排序
	DiarySortCreateTimeDesc DiarySort = "CREATE_TIME_DESC"
	// DiarySortCreateTimeAsc 创建时间升序排序
	DiarySortCreateTimeAsc DiarySort = "CREATE_TIME_ASC"
	// DiarySortModifyTimeDesc 修改时间降序排序
	DiarySortModifyTimeDesc DiarySort = "MODIFY_TIME_DESC"
	// DiarySortModifyTimeAsc 修改时间升序排序
	DiarySortModifyTimeAsc DiarySort = "MODIFY_TIME_ASC"
)

func DiarySortPtr(s DiarySort) *DiarySort {
	return &s
}

// DiarySortValueOf 尝试将字符串转为日记排序枚举
func DiarySortValueOf(s string) *DiarySort {
	switch s {
	case "CREATE_TIME_DESC":
		return DiarySortPtr(DiarySortCreateTimeDesc)
	case "CREATE_TIME_ASC":
		return DiarySortPtr(DiarySortCreateTimeAsc)
	case "MODIFY_TIME_DESC":
		return DiarySortPtr(DiarySortModifyTimeDesc)
	case "MODIFY_TIME_ASC":
		return DiarySortPtr(DiarySortModifyTimeAsc)
	default:
		return nil
	}
}

// UnmarshalJSON 自定义反序列化，验证枚举值
func (ls *DiarySort) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	// 验证是否为有效枚举值
	if v := DiarySortValueOf(s); v == nil {
		return fmt.Errorf("invalid DiarySort: %s", s)
	}
	*ls = DiarySort(s)
	return nil
}
