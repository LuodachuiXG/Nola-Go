package response

import "nola-go/internal/models"

// LinkResponse 友情链接响应
type LinkResponse struct {
	// DisplayName 链接名称
	DisplayName string `json:"displayName"`
	// Url 链接地址
	Url string `json:"url"`
	// Logo Logo 地址
	Logo *string `json:"logo"`
	// Description 描述
	Description *string `json:"description"`
	// isLost 是否失联
	IsLost bool `json:"isLost"`
}

// LinkResponseValueOf 将链接转为链接响应
func LinkResponseValueOf(link *models.Link) *LinkResponse {
	return &LinkResponse{
		DisplayName: link.DisplayName,
		Url:         link.Url,
		Logo:        link.Logo,
		Description: link.Description,
		IsLost:      link.IsLost,
	}
}
