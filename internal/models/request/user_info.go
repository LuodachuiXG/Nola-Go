package request

// UserInfoRequest 用户信息请求结构体
type UserInfoRequest struct {
	Username    string  `json:"username" binding:"required"`
	Email       string  `json:"email" binding:"required"`
	DisplayName string  `json:"displayName" binding:"required"`
	Description *string `json:"description"`
	Avatar      *string `json:"avatar"`
}
