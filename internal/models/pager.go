package models

// Pager 分页结构体
type Pager[T any] struct {
	// Page 当前页
	Page int `json:"page"`
	// Size 每页条数
	Size int `json:"size"`
	// Data 数据数组
	Data []*T `json:"data"`
	// TotalData 总条数
	TotalData int64 `json:"totalData"`
	// TotalPages 总页数
	TotalPages int64 `json:"totalPages"`
}
