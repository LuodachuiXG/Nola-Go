package repository

import (
	"context"
	"nola-go/internal/models"

	"gorm.io/gorm"
)

// ConfigRepository 配置 Repo 接口
type ConfigRepository interface {
	// AddConfig 添加配置
	AddConfig(ctx context.Context, config *models.Config) (*models.Config, error)
	// DeleteConfig 删除配置
	DeleteConfig(ctx context.Context, key models.ConfigKey) (bool, error)
	// UpdateConfig 更新配置
	UpdateConfig(ctx context.Context, config *models.Config) (bool, error)
	// Config 获取配置
	Config(ctx context.Context, key models.ConfigKey) (*string, error)
}

type configRepo struct {
	db *gorm.DB
}

// NewConfigRepository 创建配置 Repo
func NewConfigRepository(db *gorm.DB) ConfigRepository {
	return &configRepo{
		db: db,
	}
}

// AddConfig 添加配置
func (r *configRepo) AddConfig(ctx context.Context, config *models.Config) (*models.Config, error) {
	err := r.db.WithContext(ctx).Create(config).Error
	if err != nil {
		return nil, err
	}
	return config, nil
}

// DeleteConfig 删除配置
func (r *configRepo) DeleteConfig(ctx context.Context, key models.ConfigKey) (bool, error) {
	err := r.db.WithContext(ctx).Where("`key` = ?", key).Delete(&models.Config{}).Error
	if err != nil {
		return false, err
	}
	return true, nil
}

// UpdateConfig 更新配置
func (r *configRepo) UpdateConfig(ctx context.Context, config *models.Config) (bool, error) {
	err := r.db.WithContext(ctx).Where("`key` = ?", config.Key).Updates(config).Error
	if err != nil {
		return false, err
	}
	return true, nil
}

// Config 获取配置
func (r *configRepo) Config(ctx context.Context, key models.ConfigKey) (*string, error) {
	var config models.Config
	err := r.db.WithContext(ctx).Where("`key` = ?", key).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config.Value, nil
}
