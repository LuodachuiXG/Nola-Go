package service

import (
	"context"
	"nola-go/internal/logger"
	"nola-go/internal/models"
	"nola-go/internal/models/enum"
	"nola-go/internal/models/request"
	"nola-go/internal/models/response"
	"nola-go/internal/repository"

	"go.uber.org/zap"
)

type LinkService struct {
	linkRepo repository.LinkRepository
}

// NewLinkService 创建友情链接 Service
func NewLinkService(linkRepo repository.LinkRepository) *LinkService {
	return &LinkService{
		linkRepo: linkRepo,
	}
}

// AddLink 添加友情链接
func (s *LinkService) AddLink(c context.Context, link *request.LinkRequest) (*models.Link, error) {
	if link == nil {
		return nil, nil
	}
	ret, err := s.linkRepo.AddLink(c, &models.Link{
		DisplayName: link.DisplayName,
		Url:         link.Url,
		Logo:        link.Logo,
		Description: link.Description,
		Priority:    link.Priority,
		Remark:      link.Remark,
		IsLost:      link.IsLost,
	})
	if err != nil {
		logger.Log.Error("添加友链失败", zap.Error(err))
		return nil, response.ServerError
	}
	return ret, nil
}

// DeleteLinks 删除友情链接
func (s *LinkService) DeleteLinks(c context.Context, ids []uint) (bool, error) {
	if len(ids) == 0 {
		return false, nil
	}
	ret, err := s.linkRepo.DeleteLinks(c, ids)
	if err != nil {
		logger.Log.Error("删除友联失败", zap.Error(err))
		return false, response.ServerError
	}

	return ret, nil
}

// UpdateLink 修改友联
func (s *LinkService) UpdateLink(c context.Context, link *request.LinkRequest) (bool, error) {
	if link == nil {
		return false, nil
	}

	ret, err := s.linkRepo.UpdateLink(c, link)
	if err != nil {
		logger.Log.Error("修改友联失败", zap.Error(err))
		return false, response.ServerError
	}
	return ret, nil
}

// Links 获取所有友联
func (s *LinkService) Links(c context.Context, sort *enum.LinkSort) ([]*models.Link, error) {
	ret, err := s.linkRepo.Links(c, sort)
	if err != nil {
		logger.Log.Error("获取友联失败", zap.Error(err))
		return nil, response.ServerError
	}
	return ret, nil
}

// LinksPager 分页获取友联
func (s *LinkService) LinksPager(c context.Context, page, size int, sort *enum.LinkSort) (*models.Pager[models.Link], error) {
	if page == 0 {
		// 获取所有友联
		links, err := s.Links(c, sort)
		if err != nil {
			return nil, err
		}
		return &models.Pager[models.Link]{
			Page:       0,
			Size:       0,
			Data:       links,
			TotalData:  int64(len(links)),
			TotalPages: 1,
		}, nil
	}

	ret, err := s.linkRepo.LinksPager(c, page, size, sort)
	if err != nil {
		logger.Log.Error("获取友联失败", zap.Error(err))
		return nil, response.ServerError
	}

	return ret, nil

}

// LinkCount 友联数量
func (s *LinkService) LinkCount(c context.Context) (int64, error) {
	count, err := s.linkRepo.LinkCount(c)
	if err != nil {
		logger.Log.Error("获取友联数量失败", zap.Error(err))
		return 0, response.ServerError
	}
	return count, nil
}
