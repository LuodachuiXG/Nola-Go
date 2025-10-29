package enum

import (
	"encoding/json"
	"fmt"
)

// PostSort 文章排序
type PostSort string

const (
	// PostSortCreateDesc 创建时间降序
	PostSortCreateDesc PostSort = "CREATE_DESC"
	// PostSortCreateAsc 创建时间升序
	PostSortCreateAsc PostSort = "CREATE_ASC"
	// PostSortModifyDesc 修改时间降序
	PostSortModifyDesc PostSort = "MODIFY_DESC"
	// PostSortModifyAsc 修改时间升序
	PostSortModifyAsc PostSort = "MODIFY_ASC"
	// PostSortVisitDesc 访问量降序
	PostSortVisitDesc PostSort = "VISIT_DESC"
	// PostSortVisitAsc 访问量升序
	PostSortVisitAsc PostSort = "VISIT_ASC"
	// PostSortPinned 置顶排序
	PostSortPinned PostSort = "PINNED"
)

func PostSortPtr(s PostSort) *PostSort {
	return &s
}

// PostSortValueOf 尝试将字符串转为文章排序枚举
func PostSortValueOf(s string) *PostSort {
	switch s {
	case "CREATE_DESC":
		return PostSortPtr(PostSortCreateDesc)
	case "CREATE_ASC":
		return PostSortPtr(PostSortCreateAsc)
	case "MODIFY_DESC":
		return PostSortPtr(PostSortModifyDesc)
	case "MODIFY_ASC":
		return PostSortPtr(PostSortModifyAsc)
	case "VISIT_DESC":
		return PostSortPtr(PostSortVisitDesc)
	case "VISIT_ASC":
		return PostSortPtr(PostSortVisitAsc)
	case "PINNED":
		return PostSortPtr(PostSortPinned)
	default:
		return nil
	}
}

// UnmarshalJSON 自定义反序列化，验证枚举值
func (ps *PostSort) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	// 验证是否为有效枚举值
	if enum := PostSortValueOf(s); enum == nil {
		return fmt.Errorf("invalid PostSort: %s", s)
	}
	*ps = PostSort(s)
	return nil
}
