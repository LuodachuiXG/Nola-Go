package enum

import (
	"encoding/json"
	"fmt"
)

// MenuTarget 菜单项打开方式
type MenuTarget string

const (
	// MenuTargetBlank 新窗口打开（默认）
	MenuTargetBlank MenuTarget = "BLANK"
	// MenuTargetSelf 当前窗口打开
	MenuTargetSelf MenuTarget = "SELF"
	// MenuTargetParent 父窗口打开
	MenuTargetParent MenuTarget = "PARENT"
	// MenuTargetTop 顶级窗口打开
	MenuTargetTop MenuTarget = "TOP"
)

func MenuTargetPtr(s MenuTarget) *MenuTarget {
	return &s
}

// MenuTargetValueOf 尝试将字符串转为菜单打开方式枚举
func MenuTargetValueOf(s string) *MenuTarget {
	switch s {
	case "BLANK":
		return MenuTargetPtr(MenuTargetBlank)
	case "SELF":
		return MenuTargetPtr(MenuTargetSelf)
	case "PARENT":
		return MenuTargetPtr(MenuTargetParent)
	case "TOP":
		return MenuTargetPtr(MenuTargetTop)
	default:
		return nil
	}
}

// UnmarshalJSON 自定义反序列化，验证枚举值
func (ls *MenuTarget) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	// 验证是否为有效枚举值
	if v := MenuTargetValueOf(s); v == nil {
		return fmt.Errorf("invalid MenuTarget: %s", s)
	}
	*ls = MenuTarget(s)
	return nil
}
