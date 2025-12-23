package repository

import (
	"context"
	"nola-go/internal/db"
	"nola-go/internal/models"
	"nola-go/internal/models/enum"
	"nola-go/internal/models/request"
	"nola-go/internal/util"
	"time"

	"gorm.io/gorm"
)

// LinkRepository 友情链接 Repo 接口
type LinkRepository interface {
	// AddLink 添加友情链接
	AddLink(ctx context.Context, link *models.Link) (*models.Link, error)
	// DeleteLinks 删除友情链接
	DeleteLinks(ctx context.Context, ids []uint) (bool, error)
	// UpdateLink 修改友情链接
	UpdateLink(ctx context.Context, link *request.LinkRequest) (bool, error)
	// Links 获取所有友情链接
	Links(ctx context.Context, sort *enum.LinkSort) ([]*models.Link, error)
	// LinksPager 分页获取友情链接
	LinksPager(ctx context.Context, page, size int, sort *enum.LinkSort) (*models.Pager[models.Link], error)
	// LinkCount 友情链接数量
	LinkCount(ctx context.Context) (int64, error)
}

type linkRepo struct {
	db *gorm.DB
}

func NewLinkRepository(db *gorm.DB) LinkRepository {
	return &linkRepo{
		db: db,
	}
}

// AddLink 添加友情链接
func (r *linkRepo) AddLink(ctx context.Context, link *models.Link) (*models.Link, error) {
	if util.StringIsNilOrBlank(link.Logo) {
		link.Logo = nil
	}

	if util.StringIsNilOrBlank(link.Description) {
		link.Description = nil
	}
	if util.StringIsNilOrBlank(link.Remark) {
		link.Remark = nil
	}

	err := r.db.WithContext(ctx).Create(link).Error
	if err != nil {
		return nil, err
	}
	return link, nil
}

// DeleteLinks 删除友情链接
func (r *linkRepo) DeleteLinks(ctx context.Context, ids []uint) (bool, error) {
	if len(ids) == 0 {
		return false, nil
	}

	ret := r.db.WithContext(ctx).Delete(&models.Link{}, ids)
	if ret.Error != nil {
		return false, ret.Error
	}
	return ret.RowsAffected > 0, nil
}

// UpdateLink 修改友情链接
func (r *linkRepo) UpdateLink(ctx context.Context, link *request.LinkRequest) (bool, error) {
	updates := map[string]any{
		"display_name":     link.DisplayName,
		"url":              link.Url,
		"priority":         link.Priority,
		"is_lost":          link.IsLost,
		"last_modify_time": time.Now().UnixMilli(),
	}
	if util.StringIsNilOrBlank(link.Logo) {
		updates["logo"] = nil
	} else {
		updates["logo"] = link.Logo
	}

	if util.StringIsNilOrBlank(link.Description) {
		updates["description"] = nil
	} else {
		updates["description"] = link.Description
	}

	if util.StringIsNilOrBlank(link.Remark) {
		updates["remark"] = nil
	} else {
		updates["remark"] = link.Remark
	}
	ret := r.db.WithContext(ctx).
		Where("link_id = ?", link.LinkId).
		Model(&models.Link{}).
		Updates(updates)
	if ret.Error != nil {
		return false, ret.Error
	}

	return ret.RowsAffected > 0, nil
}

// Links 获取所有友情链接
func (r *linkRepo) Links(ctx context.Context, sort *enum.LinkSort) ([]*models.Link, error) {
	var links []*models.Link
	err := r.sqlQueryLinks(ctx, sort).Find(&links).Error
	if err != nil {
		return nil, err
	}
	return links, nil
}

// LinksPager 分页获取友情链接
func (r *linkRepo) LinksPager(ctx context.Context, page, size int, sort *enum.LinkSort) (*models.Pager[models.Link], error) {
	if page == 0 {
		// 获取所有链接
		data, err := r.Links(ctx, sort)
		if err != nil {
			return nil, err
		}
		return &models.Pager[models.Link]{
			Page:       0,
			Size:       0,
			Data:       data,
			TotalData:  int64(len(data)),
			TotalPages: 1,
		}, nil
	}

	pager, err := db.PagerBuilder[models.Link](ctx, r.db, page, size, func(query *gorm.DB) *gorm.DB {
		return r.sqlQueryLinks(ctx, sort)
	})

	if err != nil {
		return nil, err
	}

	return pager, nil
}

// LinkCount 友情链接数量
func (r *linkRepo) LinkCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Link{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// sqlQueryLinks 构建连接查询 SQL
func (r *linkRepo) sqlQueryLinks(ctx context.Context, sort *enum.LinkSort) *gorm.DB {
	query := r.db.WithContext(ctx).Table("link l")
	if sort != nil {
		switch *sort {
		case enum.LinkSortPriorityDesc:
			query = query.Order("l.priority DESC")
		case enum.LinkSortPriorityAsc:
			query = query.Order("l.priority ASC")
		case enum.LinkSortCreateTimeDesc:
			query = query.Order("l.create_time DESC")
		case enum.LinkSortCreateTimeAsc:
			query = query.Order("l.create_time ASC")
		case enum.LinkSortModifyTimeDesc:
			query = query.Order("l.last_modify_time DESC")
		case enum.LinkSortModifyTimeAsc:
			query = query.Order("l.last_modify_time ASC")
		}
	} else {
		// 默认优先级降序
		query = query.Order("l.priority DESC")
	}

	return query
}
