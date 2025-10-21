package service

import (
	"errors"
	"nola-go/internal/logger"
	"nola-go/internal/models"
	"nola-go/internal/models/response"
	"nola-go/internal/repository"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TagService 标签 Service
type TagService struct {
	tagRepo repository.TagRepository
}

// NewTagService 创建标签 Service
func NewTagService(tagRepo repository.TagRepository) *TagService {
	return &TagService{
		tagRepo: tagRepo,
	}
}

// AddTag 添加标签
func (s *TagService) AddTag(c *gin.Context, displayName string, slug string, color *string) (*models.Tag, error) {
	// 先判断标签别名是否已经存在
	exist, err := s.isSlugExist(c, slug, nil)
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
	ret.PostCount = 0

	return ret, nil
}

// DeleteTags 根据标签 ID 数组删除标签
func (s *TagService) DeleteTags(c *gin.Context, tagIds []uint) (bool, error) {
	ret, err := s.tagRepo.DeleteTags(c, tagIds)
	if err != nil {
		logger.Log.Error("删除标签失败 - Ids", zap.Error(err))
		return false, response.ServerError
	}

	return ret, nil
}

// DeleteTagBySlugs 根据别名数组删除标签
func (s *TagService) DeleteTagBySlugs(c *gin.Context, slugs []string) (bool, error) {
	ret, err := s.tagRepo.DeleteTagBySlugs(c, slugs)
	if err != nil {
		logger.Log.Error("删除标签失败 - Slugs", zap.Error(err))
		return false, response.ServerError
	}

	return ret, nil
}

// UpdateTag 修改标签
func (s *TagService) UpdateTag(c *gin.Context, tag *models.Tag) (bool, error) {

	// 先判断标签别名是否已存在
	exist, err := s.isSlugExist(c, tag.Slug, &tag.TagId)
	if err != nil {
		return false, err
	}
	if exist {
		// 标签别名已经存在
		return false, errors.New("标签别名 [" + tag.Slug + "] 已经存在")
	}

	// 更新标签
	ret, err := s.tagRepo.UpdateTag(c, tag)

	if err != nil {
		logger.Log.Error("修改标签失败", zap.Error(err))
		return false, response.ServerError
	}

	return ret, nil
}

// Tags 获取所有标签
func (s *TagService) Tags(c *gin.Context) ([]*models.Tag, error) {
	ret, err := s.tagRepo.Tags(c)

	if err != nil {
		logger.Log.Error("获取标签失败", zap.Error(err))
		return nil, response.ServerError
	}

	return ret, nil
}

// TopTags 获取文章数量最多的 6 个标签
func (s *TagService) TopTags(c *gin.Context) ([]*models.Tag, error) {
	ret, err := s.tagRepo.TopTags(c)
	if err != nil {
		logger.Log.Error("获取标签失败", zap.Error(err))
		return nil, response.ServerError
	}

	return ret, nil
}

// TagsPager 分页获取所有标签
func (s *TagService) TagsPager(c *gin.Context, page, size int) (*models.Pager[models.Tag], error) {
	if page == 0 {
		// 获取所有标签
		tags, err := s.Tags(c)
		if err != nil {
			return nil, err
		}
		return &models.Pager[models.Tag]{
			Page:       0,
			Size:       0,
			Data:       tags,
			TotalData:  int64(len(tags)),
			TotalPages: 1,
		}, nil
	}

	ret, err := s.tagRepo.TagsPager(c, page, size)
	if err != nil {
		logger.Log.Error("获取标签失败", zap.Error(err))
		return nil, response.ServerError
	}

	return ret, nil
}

// TagById 根据标签 ID 获取标签
func (s *TagService) TagById(c *gin.Context, id uint) (*models.Tag, error) {
	ret, err := s.tagRepo.TagById(c, id)
	if err != nil {
		logger.Log.Error("获取标签失败", zap.Error(err))
		return nil, response.ServerError
	}
	return ret, nil
}

// TagByDisplayName 根据标签名获取标签
func (s *TagService) TagByDisplayName(c *gin.Context, displayName string) (*models.Tag, error) {
	ret, err := s.tagRepo.TagByDisplayName(c, displayName)
	if err != nil {
		logger.Log.Error("获取标签失败", zap.Error(err))
		return nil, response.ServerError
	}
	return ret, nil
}

// TagBySlug 根据标签别名获取标签
func (s *TagService) TagBySlug(c *gin.Context, slug string) (*models.Tag, error) {
	ret, err := s.tagRepo.TagBySlug(c, slug)
	if err != nil {
		logger.Log.Error("获取标签失败", zap.Error(err))
		return nil, response.ServerError
	}
	return ret, nil
}

// TagCount 标签数量
func (s *TagService) TagCount(c *gin.Context) (int64, error) {
	count, err := s.tagRepo.TagCount(c)

	if err != nil {
		logger.Log.Error("获取标签数量失败", zap.Error(err))
		return 0, response.ServerError
	}

	return count, nil
}

// isSlugExist 判断标签别名是否已经存在，并且不是当前标签自己
//   - c: 上下文
//   - slug: 别名
//   - tagId: 标签 ID，用于排除自己（添加新标签时可以传 nil）
func (s *TagService) isSlugExist(c *gin.Context, slug string, tagId *uint) (bool, error) {
	tag, err := s.tagRepo.TagBySlug(c, slug)
	if err != nil {
		return false, response.ServerError
	}

	return tag != nil && (tagId == nil || tag.TagId != *tagId), nil
}

// isIdsExist 根据标签 ID 数组，判断标签是否都存在。
//
// Returns:
//   - []*models.Tag: 如果标签都存在返回空数组，否则返回不存在的 ID 数组
func (s *TagService) isIdsExist(c *gin.Context, tagIds []uint) ([]uint, error) {
	tags, err := s.tagRepo.TagByIds(c, tagIds)
	if err != nil {
		logger.Log.Error("获取标签失败", zap.Error(err))
		return nil, response.ServerError
	}

	if len(tags) == len(tagIds) {
		// 标签都存在，返回空数组
		return []uint{}, nil
	}

	// 检查哪些标签不存在
	var notExistIds []uint

	// 集合存储 ID
	existIds := make(map[uint]any)
	for _, tag := range tags {
		existIds[tag.TagId] = nil
	}

	for _, tagId := range tagIds {
		if _, ok := existIds[tagId]; !ok {
			notExistIds = append(notExistIds, tagId)
		}
	}

	// 返回不存在的标签 ID 数组
	return notExistIds, nil
}
