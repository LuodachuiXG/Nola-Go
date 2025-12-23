package models

import "nola-go/internal/models/enum"

// MenuItem 菜单项
type MenuItem struct {
	// MenuItemId 菜单项 ID
	MenuItemId uint `gorm:"column:menu_item_id;primaryKey;autoIncrement" json:"menuItemId"`
	// DisplayName 菜单项名
	DisplayName string `gorm:"column:display_name;size:128;not null" json:"displayName"`
	// Href 菜单项地址
	Href string `gorm:"column:href;size:512;not null" json:"href"`
	// Target 打开方式
	Target enum.MenuTarget `gorm:"column:target;type:varchar(12);size:12;not null" json:"target"`
	// ParentMenuId 父菜单 ID
	ParentMenuId *uint `gorm:"column:parent_menuId" json:"parentMenuId"`
	// ParentMenuItemId 父菜单项 ID
	ParentMenuItemId *uint `gorm:"column:parent_menu_item_id" json:"parentMenuItemId"`
	// Index 菜单项排序索引
	Index uint `gorm:"column:index;default:0;not null" json:"index"`
	// CreateTime 创建时间
	CreateTime int64 `gorm:"column:create_time;autoCreateTime:milli;not null" json:"createTime"`
	// LastModifyTime 最后修改时间
	LastModifyTime *int64 `gorm:"column:last_modify_time" json:"lastModifyTime"`
}

func (MenuItem) TableName() string {
	return "menu_item"
}
