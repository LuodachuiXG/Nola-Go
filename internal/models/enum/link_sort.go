package enum

import (
	"encoding/json"
	"fmt"
)

// LinkSort 友情链接排序
type LinkSort string

const (
	// LinkSortPriorityDesc 优先级降序排序
	LinkSortPriorityDesc LinkSort = "PRIORITY_DESC"
	// LinkSortPriorityAsc 优先级升序排序
	LinkSortPriorityAsc LinkSort = "PRIORITY_ASC"
	// LinkSortCreateTimeDesc 创建时间降序排序
	LinkSortCreateTimeDesc LinkSort = "CREATE_TIME_DESC"
	// LinkSortCreateTimeAsc 创建时间升序排序
	LinkSortCreateTimeAsc LinkSort = "CREATE_TIME_ASC"
	// LinkSortModifyTimeDesc 修改时间降序排序
	LinkSortModifyTimeDesc LinkSort = "MODIFY_TIME_DESC"
	// LinkSortModifyTimeAsc 修改时间升序排序
	LinkSortModifyTimeAsc LinkSort = "MODIFY_TIME_ASC"
)

func LinkSortPtr(s LinkSort) *LinkSort {
	return &s
}

// LinkSortValueOf 尝试将字符串转为友情链接排序枚举
func LinkSortValueOf(s string) *LinkSort {
	switch s {
	case "PRIORITY_DESC":
		return LinkSortPtr(LinkSortPriorityDesc)
	case "PRIORITY_ASC":
		return LinkSortPtr(LinkSortPriorityAsc)
	case "CREATE_TIME_DESC":
		return LinkSortPtr(LinkSortCreateTimeDesc)
	case "CREATE_TIME_ASC":
		return LinkSortPtr(LinkSortCreateTimeAsc)
	case "MODIFY_TIME_DESC":
		return LinkSortPtr(LinkSortModifyTimeDesc)
	case "MODIFY_TIME_ASC":
		return LinkSortPtr(LinkSortModifyTimeAsc)
	default:
		return nil
	}
}

// UnmarshalJSON 自定义反序列化，验证枚举值
func (ls *LinkSort) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	// 验证是否为有效枚举值
	if v := LinkSortValueOf(s); v == nil {
		return fmt.Errorf("invalid LinkSort: %s", s)
	}
	*ls = LinkSort(s)
	return nil
}
