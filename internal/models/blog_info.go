package models

// BlogInfo 博客信息
type BlogInfo struct {
	// Title 博客标题
	Title *string `json:"title"`
	// Subtitle 博客副标题
	Subtitle *string `json:"subtitle"`
	// Blogger 博客作者（博主）
	Blogger *string `json:"blogger"`
	// Logo 博客 Logo
	Logo *string `json:"logo"`
	// Favicon 博客 favicon
	Favicon *string `json:"favicon"`
	// CreateDate 博客创建时间
	CreateDate *int64 `json:"createDate"`
}
