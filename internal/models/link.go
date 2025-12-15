package models

// Link 友情链接
type Link struct {
	// LinkId 友情链接 ID
	LinkId uint `gorm:"column:link_id;primaryKey;autoIncrement" json:"linkId"`
	// DisplayName 链接名称
	DisplayName string `gorm:"column:display_name;size:128;not null" json:"displayName"`
	// Url 链接地址
	Url string `gorm:"column:url;size:512;not null" json:"url"`
	// Logo Logo 地址
	Logo *string `gorm:"column:logo;size:512" json:"logo"`
	// Description 描述
	Description *string `gorm:"column:description;size:512" json:"description"`
	// Priority 优先级（0 默认，1 - 100）
	Priority uint `gorm:"column:priority;default:0;not null" json:"priority"`
	// Remark 备注
	Remark *string `gorm:"column:remark;size:256" json:"remark"`
	// IsLost 是否失联
	IsLost bool `gorm:"column:is_lost;default:false;not null" json:"isLost"`
	// CreateTime 创建时间
	CreateTime int64 `gorm:"column:create_time;autoCreateTime:milli;not null" json:"createTime"`
	// LastModifyTime 最后修改时间
	LastModifyTime *int64 `gorm:"column:last_modify_time" json:"lastModifyTime"`
}

func (Link) TableName() string {
	return "link"
}
