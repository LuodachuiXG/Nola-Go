package config

// TencentCosConfig 腾讯云对象存储配置
type TencentCosConfig struct {
	// SecretId 密钥 ID
	SecretId string `json:"secretId" binding:"required"`
	// SecretKey 密钥 KEY
	SecretKey string `json:"secretKey" binding:"required"`
	// Region 存储区域
	Region string `json:"region" binding:"required"`
	// Bucket 存储桶
	Bucket string `json:"bucket" binding:"required"`
	// Path 存储路径
	Path *string `json:"path"`
	// Https 是否使用 HTTPS
	Https bool `json:"https"`
}
