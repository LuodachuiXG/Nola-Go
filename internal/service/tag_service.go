package service

import (
	"errors"
	"nola-go/internal/models"
	"nola-go/internal/models/response"
	"nola-go/internal/repository"
	"nola-go/internal/util"

	"github.com/gin-gonic/gin"
)

// TagService 标签 Service
type TagService struct {
	tagRepo repository.TagRepository
}

// NewTagService 创建标签 Service
func NewTagService(configRepo repository.TagRepository) *TagService {
	return &TagService{
		tagRepo: configRepo,
	}
}

// AddTag 添加标签
func (s *TagService) AddTag(c *gin.Context, displayName string, slug string, color *string) (*models.Tag, error) {
	// 先判断标签别名是否已经存在
	exist, err := s.IsSlugExist(c, slug, nil)
	if err != nil {
		return nil, err
	}
	if exist {
		// 标签别名已存在
		return nil, errors.New("标签别名 [" + slug + "] 已经存在")
	}
	ret, err := s.tagRepo.AddTag(c, &models.Tag{
		DisplayName: displayName,
		Slug:        slug,
		Color:       color,
	})

	if err != nil {
		return nil, response.ServerError
	}

	// 默认文章数量 0
	ret.PostCount = util.UintPrt(0)

	return ret, nil
}

// IsSlugExist 判断标签别名是否已经存在，并且不是当前标签自己
//   - c: 上下文
//   - slug: 别名
//   - tagId: 标签 ID，用于排除自己（添加新标签时可以传 nil）
func (s *TagService) IsSlugExist(c *gin.Context, slug string, tagId *uint) (bool, error) {
	tag, err := s.tagRepo.TagBySlug(c, slug)
	if err != nil {
		return false, response.ServerError
	}

	return tag != nil && (tagId == nil || tag.TagId != *tagId), nil
}
