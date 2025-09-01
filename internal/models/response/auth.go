package response

// AuthResponse 用户登录成功响应结构体
type AuthResponse struct {
	// Username 用户名
	Username string `json:"username"`

	// Email 邮箱
	Email string `json:"email"`

	// DisplayName 昵称
	DisplayName string `json:"displayName"`

	// Description 描述
	Description *string `json:"description,omitempty"`

	// CreateDate 注册时间戳毫秒
	CreateDate int64 `json:"createDate"`

	// LastLoginDate 最后登录日期
	LastLoginDate *int64 `json:"lastLoginDate,omitempty"`

	// Avatar 头像地址
	Avatar *string `json:"avatar,omitempty"`

	// Token 令牌
	Token string `json:"token"`
}
