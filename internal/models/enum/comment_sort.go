package enum

import (
	"encoding/json"
	"fmt"
)

// CommentSort 评论排序
type CommentSort string

const (
	// CommentSortCreateDesc 创建时间降序
	CommentSortCreateDesc CommentSort = "CREATE_DESC"
	// CommentSortCreateAsc 创建时间升序
	CommentSortCreateAsc CommentSort = "CREATE_ASC"
)

func CommentSortPtr(s CommentSort) *CommentSort {
	return &s
}

// CommentSortValueOf 尝试将字符串转为评论排序枚举
func CommentSortValueOf(s string) *CommentSort {
	switch s {
	case "CREATE_DESC":
		return CommentSortPtr(CommentSortCreateDesc)
	case "CREATE_ASC":
		return CommentSortPtr(CommentSortCreateAsc)
	default:
		return nil
	}
}

// UnmarshalJSON 自定义反序列化，验证枚举值
func (ps *CommentSort) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	// 验证是否为有效枚举值
	if enum := CommentSortValueOf(s); enum == nil {
		return fmt.Errorf("invalid CommentSort: %s", s)
	}
	*ps = CommentSort(s)
	return nil
}
