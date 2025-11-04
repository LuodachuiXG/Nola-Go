package enum

import (
	"encoding/json"
	"fmt"
)

// PostVisible 文章可见性
type PostVisible string

const (
	// PostVisibleVisible 可见
	PostVisibleVisible PostVisible = "VISIBLE"

	// PostVisibleHidden 隐藏
	PostVisibleHidden PostVisible = "HIDDEN"
)

func PostVisiblePtr(v PostVisible) *PostVisible {
	return &v
}

// PostVisibleValueOf 尝试将字符串转为文章可见性枚举
func PostVisibleValueOf(s string) *PostVisible {
	switch s {
	case string(PostVisibleVisible):
		return PostVisiblePtr(PostVisibleVisible)
	case string(PostVisibleHidden):
		return PostVisiblePtr(PostVisibleHidden)
	default:
		return nil
	}
}

// UnmarshalJSON 自定义反序列化，验证枚举值
func (pv *PostVisible) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	// 验证是否为有效枚举值
	if enum := PostVisibleValueOf(s); enum == nil {
		return fmt.Errorf("invalid PostVisible: %s", s)
	}
	*pv = PostVisible(s)
	return nil
}
