package request

// LinkRequest 友联请求结构体
type LinkRequest struct {
	// LinkId 友联 ID
	LinkId *uint `json:"linkId"`
	// DisplayName 链接名称
	DisplayName string `json:"displayName" binding:"required"`
	// Url 链接地址
	Url string `json:"url" binding:"required"`
	// Logo Logo 地址
	Logo *string `json:"logo"`
	// Description 描述
	Description *string `json:"description"`
	// Priority 优先级（0 默认，1 - 100）
	Priority uint `json:"priority"`
	// IsLost 是否已失联
	IsLost bool `json:"isLost"`
	// Remark 备注
	Remark *string `json:"remark"`
}
