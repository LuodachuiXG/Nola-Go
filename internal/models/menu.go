package models

// Menu 菜单
type Menu struct {
	// MenuId 菜单 ID
	MenuId uint `gorm:"column:menu_id;primaryKey;autoIncrement" json:"menuId"`
	// isMain 是否是主菜单
	IsMain bool `gorm:"column:is_main;not null" json:"isMain"`
	// DisplayName 菜单名（唯一键）
	DisplayName string `gorm:"column:display_name;size:128;not null;uniqueIndex" json:"displayName"`
	// CreateTime 创建时间
	CreateTime int64 `gorm:"column:create_time;autoCreateTime:milli;not null" json:"createTime"`
	// LastModifyTime 最后修改时间
	LastModifyTime *int64 `gorm:"column:last_modify_time" json:"lastModifyTime"`
}

func (Menu) TableName() string {
	return "menu"
}
