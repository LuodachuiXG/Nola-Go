package response

// Pager 分页结构体
type Pager struct {
	// Page 当前页
	Page int `json:"page"`
	// Size 每页条数
	Size int `json:"size"`
	// Data 数据集合
	Data any `json:"data"`
	// TotalData 总数据条数
	TotalData int64 `json:"totalData"`
	// TotalPages 总页数
	TotalPages int `json:"totalPages"`
}
