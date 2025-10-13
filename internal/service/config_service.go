package service

import (
	"context"
	"nola-go/internal/logger"
	"nola-go/internal/models"
	"nola-go/internal/models/response"
	"nola-go/internal/repository"
	"nola-go/internal/util"

	"go.uber.org/zap"
)

// ConfigService 配置服务
type ConfigService struct {
	configRepo repository.ConfigRepository
}

// NewConfigService 创建配置服务
func NewConfigService(configRepo repository.ConfigRepository) *ConfigService {
	return &ConfigService{
		configRepo: configRepo,
	}
}

// SetConfig 设置配置信息
func (s *ConfigService) SetConfig(ctx context.Context, config *models.Config) (*models.Config, error) {
	oldConfig, err := s.Config(ctx, config.Key)
	if err != nil {
		return nil, err
	}

	if oldConfig != nil {
		// 配置信息已存在，修改配置
		_, err := s.UpdateConfig(ctx, config)
		if err != nil {
			return nil, err
		}

		return config, nil
	} else {
		// 配置信息不存在，添加配置
		ret, err := s.configRepo.AddConfig(ctx, config)
		if err != nil {
			logger.Log.Error("添加配置失败", zap.Error(err))
			return nil, response.ServerError
		}
		return ret, nil
	}
}

// DeleteConfig 删除配置信息
func (s *ConfigService) DeleteConfig(ctx context.Context, key models.ConfigKey) (bool, error) {
	ret, err := s.configRepo.DeleteConfig(ctx, key)
	if err != nil {
		logger.Log.Error("删除配置失败", zap.Error(err))
		return false, response.ServerError
	}
	return ret, nil
}

// UpdateConfig 修改配置信息
func (s *ConfigService) UpdateConfig(ctx context.Context, config *models.Config) (bool, error) {
	ret, err := s.configRepo.UpdateConfig(ctx, config)
	if err != nil {
		logger.Log.Error("修改配置失败", zap.Error(err))
		return false, response.ServerError
	}
	return ret, nil
}

// Config 获取配置信息
func (s *ConfigService) Config(ctx context.Context, key models.ConfigKey) (*string, error) {
	config, err := s.configRepo.Config(ctx, key)
	if err != nil {
		logger.Log.Error("获取配置失败", zap.Error(err))
		return nil, response.ServerError
	}
	return config, nil
}

// SetBlogInfo 设置博客信息
func (s *ConfigService) SetBlogInfo(ctx context.Context, blogInfo *models.BlogInfo) (bool, error) {
	_, err := s.SetConfig(ctx, &models.Config{
		Key:   models.ConfigKeyBlogInfo,
		Value: util.StringDefault(util.ToJsonString(blogInfo), ""),
	})

	if err != nil {
		return false, err
	}

	return true, nil
}

// BlogInfo 获取博客信息
func (s *ConfigService) BlogInfo(ctx context.Context) (*models.BlogInfo, error) {
	blogInfo := &models.BlogInfo{}
	config, err := s.Config(ctx, models.ConfigKeyBlogInfo)
	if err != nil {
		return nil, err
	}

	if err := util.FromJsonString(config, blogInfo); err != nil {
		logger.Log.Error("解析博客信息失败", zap.Error(err))
		return nil, response.ServerError
	}

	return blogInfo, nil
}

// SetICPFiling 设置备案信息
func (s *ConfigService) SetICPFiling(ctx context.Context, filing *models.ICPFiling) (bool, error) {
	_, err := s.SetConfig(ctx, &models.Config{
		Key:   models.ConfigKeyICPFiling,
		Value: util.StringDefault(util.ToJsonString(filing), ""),
	})

	if err != nil {
		return false, err
	}

	return true, nil
}

// ICPFiling 获取备案信息
func (s *ConfigService) ICPFiling(ctx context.Context) (*models.ICPFiling, error) {
	icp := &models.ICPFiling{}
	config, err := s.Config(ctx, models.ConfigKeyICPFiling)
	if err != nil {
		return nil, err
	}

	if err := util.FromJsonString(config, icp); err != nil {
		logger.Log.Error("解析备案信息失败", zap.Error(err))
		return nil, response.ServerError
	}

	return icp, nil
}
