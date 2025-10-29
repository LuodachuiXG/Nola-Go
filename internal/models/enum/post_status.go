package enum

import (
	"encoding/json"
	"fmt"
)

// PostStatus 文章状态
type PostStatus string

const (
	// PostStatusPublished 已发布
	PostStatusPublished PostStatus = "PUBLISHED"

	// PostStatusDraft 草稿
	PostStatusDraft PostStatus = "DRAFT"

	// PostStatusDeleted 已删除（回收站）
	PostStatusDeleted PostStatus = "DELETED"
)

// PostStatusPtr 获取文章状态指针
func PostStatusPtr(s PostStatus) *PostStatus {
	return &s
}

// PostStatusValueOf 尝试将字符串转为文章状态枚举
func PostStatusValueOf(s string) *PostStatus {
	switch s {
	case "PUBLISHED":
		return PostStatusPtr(PostStatusPublished)
	case "DRAFT":
		return PostStatusPtr(PostStatusDraft)
	case "DELETED":
		return PostStatusPtr(PostStatusDeleted)
	default:
		return nil
	}
}

// UnmarshalJSON 自定义反序列化，验证枚举值
func (ps *PostStatus) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	// 验证是否为有效枚举值
	if enum := PostStatusValueOf(s); enum == nil {
		return fmt.Errorf("invalid PostStatus: %s", s)
	}
	*ps = PostStatus(s)
	return nil
}
