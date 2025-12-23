package request

// MenuRequest 菜单请求
type MenuRequest struct {
	// MenuId 菜单 ID
	MenuId *uint `json:"menuId"`
	// DisplayName 菜单名称
	DisplayName string `json:"displayName" binding:"required"`
	// IsMain 是否是主菜单
	IsMain bool `json:"isMain"`
}
