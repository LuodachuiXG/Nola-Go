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

// CategoryService 分类 Service
type CategoryService struct {
	categoryRepo repository.CategoryRepository
}

// NewCategoryService 创建分类 Service
func NewCategoryService(categoryRepo repository.CategoryRepository) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
	}
}

// AddCategory 添加分类
func (s *CategoryService) AddCategory(c *gin.Context, displayName string, slug string, cover *string, unifiedCover *bool) (*models.Category, error) {
	// 先判断分类别名是否已经存在
	exist, err := s.isSlugExist(c, slug, nil)
	if err != nil {
		return nil, err
	}
	if exist {
		// 分类别名已存在
		return nil, errors.New("分类别名 [" + slug + "] 已经存在")
	}

	// 是否统一封面
	unified := false
	if unifiedCover != nil && *unifiedCover {
		unified = true
	}

	// 添加分类
	ret, err := s.categoryRepo.AddCategory(c, &models.Category{
		DisplayName:  displayName,
		Slug:         slug,
		Cover:        cover,
		UnifiedCover: unified,
	})

	if err != nil {
		return nil, response.ServerError
	}

	// 默认文章数量 0
	ret.PostCount = 0

	return ret, nil
}

// DeleteCategories 根据分类 ID 数组删除分类
func (s *CategoryService) DeleteCategories(c *gin.Context, categoryIds []uint) (bool, error) {
	ret, err := s.categoryRepo.DeleteCategories(c, categoryIds)
	if err != nil {
		logger.Log.Error("删除分类失败 - Ids", zap.Error(err))
		return false, response.ServerError
	}

	return ret, nil
}

// DeleteCategoryBySlugs 根据别名数组删除分类
func (s *CategoryService) DeleteCategoryBySlugs(c *gin.Context, slugs []string) (bool, error) {
	ret, err := s.categoryRepo.DeleteCategoryBySlugs(c, slugs)
	if err != nil {
		logger.Log.Error("删除分类失败 - Slugs", zap.Error(err))
		return false, response.ServerError
	}

	return ret, nil
}

// UpdateCategory 修改分类
func (s *CategoryService) UpdateCategory(c *gin.Context, category *models.Category) (bool, error) {
	// 先判断分类别名是否已存在
	exist, err := s.isSlugExist(c, category.Slug, &category.CategoryId)
	if err != nil {
		return false, err
	}
	if exist {
		// 分类别名已经存在
		return false, errors.New("分类别名 [" + category.Slug + "] 已经存在")
	}

	// 更新分类
	ret, err := s.categoryRepo.UpdateCategory(c, category)

	if err != nil {
		logger.Log.Error("修改分类失败", zap.Error(err))
		return false, response.ServerError
	}

	return ret, nil
}

// Categories 获取所有分类
func (s *CategoryService) Categories(c *gin.Context) ([]*models.Category, error) {
	ret, err := s.categoryRepo.Categories(c)

	if err != nil {
		logger.Log.Error("获取分类失败", zap.Error(err))
		return nil, response.ServerError
	}

	return ret, nil
}

// CategoriesPager 分页获取所有分类
func (s *CategoryService) CategoriesPager(c *gin.Context, page, size int) (*models.Pager[models.Category], error) {
	if page == 0 {
		// 获取所有分类
		categories, err := s.Categories(c)
		if err != nil {
			return nil, err
		}
		return &models.Pager[models.Category]{
			Page:       0,
			Size:       0,
			Data:       categories,
			TotalData:  int64(len(categories)),
			TotalPages: 1,
		}, nil
	}

	ret, err := s.categoryRepo.CategoriesPager(c, page, size)
	if err != nil {
		logger.Log.Error("获取分类失败", zap.Error(err))
		return nil, response.ServerError
	}

	return ret, nil
}

// TopCategories 获取文章数量最多的 6 个分类
func (s *CategoryService) TopCategories(c *gin.Context) ([]*models.Category, error) {
	ret, err := s.categoryRepo.TopCategories(c)
	if err != nil {
		logger.Log.Error("获取分类失败", zap.Error(err))
		return nil, response.ServerError
	}

	return ret, nil
}

// CategoryById 根据分类 ID 获取分类
func (s *CategoryService) CategoryById(c *gin.Context, id uint) (*models.Category, error) {
	ret, err := s.categoryRepo.CategoryById(c, id)
	if err != nil {
		logger.Log.Error("获取分类失败", zap.Error(err))
		return nil, response.ServerError
	}
	return ret, nil
}

// CategoryByDisplayName 根据分类名获取分类
func (s *CategoryService) CategoryByDisplayName(c *gin.Context, displayName string) (*models.Category, error) {
	ret, err := s.categoryRepo.CategoryByDisplayName(c, displayName)
	if err != nil {
		logger.Log.Error("获取分类失败", zap.Error(err))
		return nil, response.ServerError
	}
	return ret, nil
}

// CategoryBySlug 根据分类别名获取分类
func (s *CategoryService) CategoryBySlug(c *gin.Context, slug string) (*models.Category, error) {
	ret, err := s.categoryRepo.CategoryBySlug(c, slug)
	if err != nil {
		logger.Log.Error("获取分类失败", zap.Error(err))
		return nil, response.ServerError
	}
	return ret, nil
}

// CategoryCount 分类数量
func (s *CategoryService) CategoryCount(c *gin.Context) (int64, error) {
	count, err := s.categoryRepo.CategoryCount(c)

	if err != nil {
		logger.Log.Error("获取分类数量失败", zap.Error(err))
		return 0, response.ServerError
	}

	return count, nil
}

// isSlugExist 判断分类别名是否已经存在，并且不是当前分类自己
//   - c: 上下文
//   - slug: 别名
//   - categoryId: 分类 ID，用于排除自己（添加新分类时可以传 nil）
func (s *CategoryService) isSlugExist(c *gin.Context, slug string, categoryId *uint) (bool, error) {
	category, err := s.categoryRepo.CategoryBySlug(c, slug)
	if err != nil {
		return false, response.ServerError
	}

	return category != nil && (categoryId == nil || category.CategoryId != *categoryId), nil
}
