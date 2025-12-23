package repository

import (
	"context"
	errors "errors"
	"nola-go/internal/db"
	"nola-go/internal/models"
	"nola-go/internal/util"

	"gorm.io/gorm"
)

// TagRepository 标签 Repo 接口
type TagRepository interface {
	// AddTag 添加标签
	AddTag(ctx context.Context, tag *models.Tag) (*models.Tag, error)
	// DeleteTags 删除标签 - ID 数组
	DeleteTags(ctx context.Context, ids []uint) (bool, error)
	// DeleteTagBySlugs 删除标签 - Slug 数组
	DeleteTagBySlugs(ctx context.Context, slugs []string) (bool, error)
	// UpdateTag 修改标签
	UpdateTag(ctx context.Context, tag *models.Tag) (bool, error)
	// Tags 获取所有标签
	Tags(ctx context.Context) ([]*models.Tag, error)
	// TagsPager 分页获取所有标签
	TagsPager(ctx context.Context, page, size int) (*models.Pager[models.Tag], error)
	// TagById 获取标签 - ID
	TagById(ctx context.Context, id uint) (*models.Tag, error)
	// TagByPostId 获取标签 - 文章 ID
	TagByPostId(ctx context.Context, postId uint) ([]*models.Tag, error)
	// TagByIds 获取标签 - ID 数组
	TagByIds(ctx context.Context, ids []uint) ([]*models.Tag, error)
	// TopTags 获取文章数量最多的 6 个标签
	TopTags(ctx context.Context) ([]*models.Tag, error)
	// TagByDisplayName 获取标签 - 标签名
	TagByDisplayName(ctx context.Context, displayName string) (*models.Tag, error)
	// TagBySlug 获取标签 - 别名
	TagBySlug(ctx context.Context, slug string) (*models.Tag, error)
	// TagBySlugs 获取标签 - 别名数组
	TagBySlugs(ctx context.Context, slugs []string) ([]*models.Tag, error)
	// TagCount 标签数量
	TagCount(ctx context.Context) (int64, error)
}

type tagRepo struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepo{
		db: db,
	}
}

// AddTag 添加标签
func (r *tagRepo) AddTag(ctx context.Context, tag *models.Tag) (*models.Tag, error) {
	err := r.db.WithContext(ctx).Create(tag).Error
	if err != nil {
		return nil, err
	}
	return tag, nil
}

// DeleteTags 删除标签 - ID 数组
func (r *tagRepo) DeleteTags(ctx context.Context, tagIds []uint) (bool, error) {
	if len(tagIds) == 0 {
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

	// 先尝试删除标签文章关联信息
	ret := tx.Where("`tag_id` IN ?", tagIds).Delete(&models.PostTag{})
	if err := ret.Error; err != nil {
		tx.Rollback()
		return false, err
	}

	// 删除标签
	ret = tx.Delete(&models.Tag{}, tagIds)
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

// DeleteTagBySlugs 删除标签 - Slug 数组
func (r *tagRepo) DeleteTagBySlugs(ctx context.Context, slugs []string) (bool, error) {
	if len(slugs) == 0 {
		return false, nil
	}

	// 先根据别名获取到对应的标签
	tags, err := r.TagBySlugs(ctx, slugs)
	if err != nil {
		return false, err
	}

	// 获取所有标签 ID 数组
	ids := util.Map(tags, func(tag *models.Tag) uint {
		return tag.TagId
	})

	// 删除标签
	ret, err := r.DeleteTags(ctx, ids)
	if err != nil {
		return false, err
	}
	return ret, nil
}

// UpdateTag 修改标签
func (r *tagRepo) UpdateTag(ctx context.Context, tag *models.Tag) (bool, error) {
	if util.StringIsNilOrBlank(tag.Color) {
		tag.Color = nil
	}

	updates := map[string]any{
		"display_name": tag.DisplayName,
		"slug":         tag.Slug,
		"color":        *tag.Color,
	}

	err := r.db.WithContext(ctx).Where("`tag_id` = ?", tag.TagId).Model(&models.Tag{}).Updates(updates).Error
	if err != nil {
		return false, err
	}
	return true, nil
}

// Tags 获取所有标签
func (r *tagRepo) Tags(ctx context.Context) ([]*models.Tag, error) {
	var tags []*models.Tag
	if err := r.sqlSelectTag().WithContext(ctx).Scan(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

// TagsPager 分页获取所有标签
func (r *tagRepo) TagsPager(ctx context.Context, page, size int) (*models.Pager[models.Tag], error) {
	pager, err := db.PagerBuilder[models.Tag](ctx, r.db, page, size, func(query *gorm.DB) *gorm.DB {
		return r.sqlSelectTag()
	})

	if err != nil {
		return nil, err
	}

	return pager, nil
}

// TagById 获取标签 - ID
func (r *tagRepo) TagById(ctx context.Context, id uint) (*models.Tag, error) {
	var tag *models.Tag
	err := r.sqlSelectTag().WithContext(ctx).Where("t.tag_id = ?", id).First(&tag).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return tag, nil
}

// TagByPostId 获取标签 - 文章 ID
func (r *tagRepo) TagByPostId(ctx context.Context, postId uint) ([]*models.Tag, error) {
	var tags []*models.Tag
	err := r.sqlSelectTag().WithContext(ctx).Where("pt.post_id = ?", postId).Scan(&tags).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []*models.Tag{}, nil
		}
		return nil, err
	}

	if tags == nil {
		return []*models.Tag{}, nil
	}
	return tags, nil
}

// TagByIds 获取标签 - ID 数组
func (r *tagRepo) TagByIds(ctx context.Context, ids []uint) ([]*models.Tag, error) {
	var tags []*models.Tag
	err := r.sqlSelectTag().WithContext(ctx).Where("t.tag_id IN ?", ids).Scan(&tags).Error
	if err != nil {
		return nil, err
	}
	return tags, nil
}

// TopTags 获取文章数量最多的 6 个标签
func (r *tagRepo) TopTags(ctx context.Context) ([]*models.Tag, error) {
	var tags []*models.Tag
	err := r.sqlSelectTag().WithContext(ctx).Order("post_count DESC").Limit(6).Scan(&tags).Error
	if err != nil {
		return nil, err
	}
	return tags, nil
}

// TagByDisplayName 获取标签 - 标签名
func (r *tagRepo) TagByDisplayName(ctx context.Context, displayName string) (*models.Tag, error) {
	var tag models.Tag
	err := r.sqlSelectTag().WithContext(ctx).Where("t.display_name = ?", displayName).First(&tag).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tag, nil
}

// TagBySlug 获取标签 - 别名
func (r *tagRepo) TagBySlug(ctx context.Context, slug string) (*models.Tag, error) {
	var tag models.Tag
	err := r.sqlSelectTag().WithContext(ctx).Where("t.slug = ?", slug).First(&tag).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tag, nil
}

// TagBySlugs 获取标签 - 别名数组
func (r *tagRepo) TagBySlugs(ctx context.Context, slugs []string) ([]*models.Tag, error) {
	var tags []*models.Tag
	err := r.sqlSelectTag().WithContext(ctx).Where("t.slug IN ?", slugs).Scan(&tags).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return tags, nil
}

// TagCount 标签数量
func (r *tagRepo) TagCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Tag{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// sqlSelectTag 获取标签和对应的文章数量
func (r *tagRepo) sqlSelectTag() *gorm.DB {
	return r.db.Table("tag t").
		Joins("LEFT JOIN post_tag pt ON t.tag_id = pt.tag_id").
		Select("t.tag_id, t.display_name, t.slug, t.color, COUNT(pt.post_tag_id) as post_count").
		Group("t.tag_id").
		Order("t.tag_id DESC")
}
