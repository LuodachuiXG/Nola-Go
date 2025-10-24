package response

// PostContentApiResponse 博客 API 文章内容响应体
type PostContentApiResponse struct {
	// Post 文章信息
	Post PostApiResponse `json:"post"`
	// Content 文章正文
	Content string `json:"content"`
}
