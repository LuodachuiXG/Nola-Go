package repository

import (
	"context"
	"errors"
	"nola-go/internal/db"
	"nola-go/internal/models"
	"nola-go/internal/util"

	"gorm.io/gorm"
)

// CategoryRepository 分类 Repo 接口
type CategoryRepository interface {
	// AddCategory 添加分类
	AddCategory(ctx context.Context, category *models.Category) (*models.Category, error)
	// DeleteCategories 删除分类 - ID 数组
	DeleteCategories(ctx context.Context, ids []uint) (bool, error)
	// DeleteCategoryBySlugs 删除分类 - Slug 数组
	DeleteCategoryBySlugs(ctx context.Context, slugs []string) (bool, error)
	// UpdateCategory 修改分类
	UpdateCategory(ctx context.Context, category *models.Category) (bool, error)
	// Categories 获取所有分类
	Categories(ctx context.Context) ([]*models.Category, error)
	// CategoryByPostId 获取分类 - 文章 ID
	CategoryByPostId(ctx context.Context, postId uint) (*models.Category, error)
	// CategoriesPager 分页获取所有分类
	CategoriesPager(ctx context.Context, page, size int) (*models.Pager[models.Category], error)
	// TopCategories 获取文章数量最多的 6 个分类
	TopCategories(ctx context.Context) ([]*models.Category, error)
	// CategoryById 获取分类 - ID
	CategoryById(ctx context.Context, id uint) (*models.Category, error)
	// CategoryByDisplayName 获取分类 - 分类名
	CategoryByDisplayName(ctx context.Context, displayName string) (*models.Category, error)
	// CategoryBySlug 获取分类 - 别名
	CategoryBySlug(ctx context.Context, slug string) (*models.Category, error)
	// CategoryBySlugs 获取分类 - 别名数组
	CategoryBySlugs(ctx context.Context, slugs []string) ([]*models.Category, error)
	// CategoryCount 分类数量
	CategoryCount(ctx context.Context) (int64, error)
}

type categoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepo{
		db: db,
	}
}

// AddCategory 添加分类
func (r *categoryRepo) AddCategory(ctx context.Context, category *models.Category) (*models.Category, error) {
	err := r.db.WithContext(ctx).Create(category).Error
	if err != nil {
		return nil, err
	}
	return category, nil
}

// DeleteCategories 删除分类 - ID 数组
func (r *categoryRepo) DeleteCategories(ctx context.Context, ids []uint) (bool, error) {
	if len(ids) == 0 {
		return false, nil
	}

	// 开启事务
	tx := *r.db.WithContext(ctx).Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return false, tx.Error
	}

	// 先尝试删除分类文章关联信息
	ret := tx.Where("`category_id` IN ?", ids).Delete(&models.PostCategory{})
	if err := ret.Error; err != nil {
		tx.Rollback()
		return false, err
	}

	// 删除分类
	ret = tx.Delete(&models.Category{}, ids)
	if err := ret.Error; err != nil {
		tx.Rollback()
		return false, err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return false, err
	}

	return ret.RowsAffected > 0, nil
}

// DeleteCategoryBySlugs 删除分类 - Slug 数组
func (r *categoryRepo) DeleteCategoryBySlugs(ctx context.Context, slugs []string) (bool, error) {
	if len(slugs) == 0 {
		return false, nil
	}

	// 先根据别名获取到对应的分类
	categories, err := r.CategoryBySlugs(ctx, slugs)
	if err != nil {
		return false, err
	}

	// 获取所有分类 ID 数组
	ids := util.Map(categories, func(category *models.Category) uint {
		return category.CategoryId
	})

	// 删除分类
	ret, err := r.DeleteCategories(ctx, ids)
	if err != nil {
		return false, err
	}
	return ret, nil
}

// UpdateCategory 修改分类
func (r *categoryRepo) UpdateCategory(ctx context.Context, category *models.Category) (bool, error) {
	updates := map[string]any{
		"display_name":  category.DisplayName,
		"slug":          category.Slug,
		"cover":         *category.Cover,
		"unified_cover": category.UnifiedCover,
	}
	err := r.db.WithContext(ctx).Where("`category_id` = ?", category.CategoryId).Model(&models.Category{}).Updates(updates).Error
	if err != nil {
		return false, err
	}
	return true, nil
}

// Categories 获取所有分类
func (r *categoryRepo) Categories(ctx context.Context) ([]*models.Category, error) {
	var categories []*models.Category
	if err := r.sqlSelectCategory().WithContext(ctx).Scan(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// CategoriesPager 分页获取所有分类
func (r *categoryRepo) CategoriesPager(ctx context.Context, page, size int) (*models.Pager[models.Category], error) {
	pager, err := db.PagerBuilder[models.Category](ctx, r.db, page, size, func(query *gorm.DB) *gorm.DB {
		return r.sqlSelectCategory()
	})

	if err != nil {
		return nil, err
	}

	return pager, nil
}

// TopCategories 获取文章数量最多的 6 个分类
func (r *categoryRepo) TopCategories(ctx context.Context) ([]*models.Category, error) {
	var categories []*models.Category
	err := r.sqlSelectCategory().WithContext(ctx).Order("post_count DESC").Limit(6).Scan(&categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}

// CategoryByPostId 获取分类 - 文章 ID
func (r *categoryRepo) CategoryByPostId(ctx context.Context, postId uint) (*models.Category, error) {
	var category *models.Category
	err := r.sqlSelectCategory().WithContext(ctx).Where("pc.post_id = ?", postId).First(&category).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return category, nil
}

// CategoryById 获取分类 - ID
func (r *categoryRepo) CategoryById(ctx context.Context, id uint) (*models.Category, error) {
	var category models.Category
	err := r.sqlSelectCategory().WithContext(ctx).Where("c.category_id = ?", id).First(&category).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

// CategoryByDisplayName 获取分类 - 分类名
func (r *categoryRepo) CategoryByDisplayName(ctx context.Context, displayName string) (*models.Category, error) {
	var category models.Category
	err := r.sqlSelectCategory().WithContext(ctx).Where("c.display_name = ?", displayName).First(&category).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

// CategoryBySlug 获取分类 - 别名
func (r *categoryRepo) CategoryBySlug(ctx context.Context, slug string) (*models.Category, error) {
	var category models.Category
	err := r.sqlSelectCategory().WithContext(ctx).Where("c.slug = ?", slug).First(&category).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

// CategoryBySlugs 获取分类 - 别名数组
func (r *categoryRepo) CategoryBySlugs(ctx context.Context, slugs []string) ([]*models.Category, error) {
	var categories []*models.Category
	err := r.sqlSelectCategory().WithContext(ctx).Where("c.slug IN ?", slugs).Scan(&categories).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return categories, nil
}

// CategoryCount 分类数量
func (r *categoryRepo) CategoryCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Category{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// sqlSelectCategory 获取分类和对应的文章数量
func (r *categoryRepo) sqlSelectCategory() *gorm.DB {
	return r.db.Table("category c").
		Joins("LEFT JOIN post_category pc ON c.category_id = pc.category_id").
		Select("c.category_id, c.display_name, c.slug, c.cover, c.unified_cover, COUNT(pc.post_category_id) as post_count").
		Group("c.category_id").
		Order("c.category_id DESC")
}
