package request

import "nola-go/internal/models/enum"

// MenuItemRequest 菜单项请求
type MenuItemRequest struct {
	// MenuItemId 菜单项 ID
	MenuItemId *uint `json:"menuItemId"`
	// DisplayName 菜单项名称
	DisplayName string `json:"displayName" binding:"required"`
	// Href 菜单地址
	Href string `json:"href" binding:"required"`
	// Target 打开方式
	Target *enum.MenuTarget `json:"target"`
	// ParentMenuId 父菜单 ID
	ParentMenuId uint `json:"parentMenuId" binding:"required"`
	// ParentMenuItemId 父菜单项 ID
	ParentMenuItemId *uint `json:"parentMenuItemId"`
	// Index 菜单项排序索引
	Index uint `json:"index"`
}
