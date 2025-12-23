package response

import "nola-go/internal/models/enum"

// MenuItemResponse 菜单项响应体
type MenuItemResponse struct {
	// MenuItemId 菜单项 ID
	MenuItemId uint `json:"menuItemId"`
	// DisplayName 菜单项名称
	DisplayName string `json:"displayName"`
	// Href 菜单地址
	Href string `json:"href"`
	// Target 打开方式
	Target enum.MenuTarget `json:"target"`
	// ParentMenuId 父菜单 ID
	ParentMenuId uint `json:"parentMenuId"`
	// ParentMenuItemId 父菜单项 ID
	ParentMenuItemId *uint `json:"parentMenuItemId"`
	// Children 子菜单项
	Children []*MenuItemResponse `json:"children"`
	// Index 菜单项排序索引
	Index uint `json:"index"`
	// CreateTime 创建时间
	CreateTime int64 `json:"createTime"`
	// LastModifyTime 最后修改时间
	LastModifyTime *int64 `json:"lastModifyTime"`
}
