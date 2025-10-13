package models

// Config 配置结构体
type Config struct {
	// ConfigId 配置表 ID
	ConfigId uint `gorm:"column:config_id;primaryKey;autoIncrement" json:"configId"`

	// Key 配置键
	Key ConfigKey `gorm:"column:key;type:varchar(64);uniqueIndex;not null" json:"key"`

	// Value 配置值
	Value string `gorm:"column:value;type:text;not null" json:"value"`
}

func (Config) TableName() string {
	return "config"
}

// ConfigKey 配置键枚举
type ConfigKey string

const (

	// ConfigKeyBlogInfo 博客信息
	ConfigKeyBlogInfo ConfigKey = "BLOG_INFO"

	// ConfigKeyICPFiling ICP 备案信息
	ConfigKeyICPFiling ConfigKey = "ICP_FILING"
)
