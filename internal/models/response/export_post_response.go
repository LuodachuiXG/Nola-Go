package response

// ExportPostResponse 导出文章响应结果
type ExportPostResponse struct {
	// Path 导出文件路径
	Path string `json:"path"`
	// Count 总数量
	Count int `json:"count"`
	// SuccessCount 成功数量
	SuccessCount int `json:"successCount"`
	// FailCount 失败数量
	FailCount int `json:"failCount"`
	// FailResult 失败信息
	FailResult []string `json:"failResult"`
}
