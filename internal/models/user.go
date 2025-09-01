package models

// User 用户
type User struct {
	// UserId 用户ID
	UserId uint `gorm:"column:user_id;primaryKey;autoIncrement" json:"userId"`

	// Username 用户名
	Username string `gorm:"column:username;size:64;uniqueIndex;not null" json:"username"`

	// Email 电子邮箱
	Email string `gorm:"column:email;size:128;not null" json:"email"`

	// DisplayName 显示名称
	DisplayName string `gorm:"column:display_name;size:128;not null" json:"displayName"`

	// Password 密码
	Password string `gorm:"column:password;size:128;not null" json:"-"`

	// Salt 盐值
	Salt string `gorm:"column:salt;size:128;not null" json:"-"`

	// Description 描述
	Description *string `gorm:"column:description;size:1024" json:"description,omitempty"`

	// CreateDate 注册日期
	CreateDate int64 `gorm:"column:create_date;not null" json:"createDate"`

	// LastLoginDate 最后登录日期
	LastLoginDate *int64 `gorm:"column:last_login_time;not null" json:"lastLoginDate,omitempty"`

	// Avatar 头像地址
	Avatar *string `gorm:"column:avatar;size:512" json:"avatar,omitempty"`
}

func (User) TableName() string {
	return "user"
}
