package repository

import (
	"context"
	"errors"
	"nola-go/internal/db"
	"nola-go/internal/models"
	"nola-go/internal/models/enum"
	"time"

	"gorm.io/gorm"
)

// CommentRepository 评论 Repo 接口
type CommentRepository interface {
	// AddComment 添加评论
	AddComment(ctx context.Context, comment *models.Comment) (*models.Comment, error)
	// DeleteCommentById 根据评论 ID 删除评论
	DeleteCommentById(ctx context.Context, id uint) (bool, error)
	// DeleteCommentByIds 根据评论 ID 数组删除评论
	DeleteCommentByIds(ctx context.Context, ids []uint) (bool, error)
	// DeleteCommentByPostId 根据文章 ID 删除评论
	DeleteCommentByPostId(ctx context.Context, postId uint) (bool, error)
	// DeleteCommentByParentIds 根据父评论 ID 数组删除评论
	DeleteCommentByParentIds(ctx context.Context, parentIds []uint) (bool, error)
	// UpdateComment 修改评论
	UpdateComment(ctx context.Context, comment models.Comment) (bool, error)
	// SetCommentPass 批量设置评论是否通过审核
	SetCommentPass(ctx context.Context, ids []uint, isPass bool) (bool, error)
	// CommentByIds 根据评论 ID 数组获取所有评论
	CommentByIds(ctx context.Context, ids []uint) ([]*models.Comment, error)
	// Comments 获取所有评论
	Comments(
		ctx context.Context,
		postId, commentId, parentId *uint,
		isPass *bool, key *string,
		sort *enum.CommentSort,
	) ([]*models.Comment, error)
	// CommentsPager 分页获取所有评论
	CommentsPager(
		ctx context.Context,
		page, size int,
		postId, commentId, parentId *uint,
		isPass *bool, key *string,
		sort *enum.CommentSort,
	) (*models.Pager[models.Comment], error)
	// CommentById 根据评论 ID 获取评论
	CommentById(ctx context.Context, id uint) (*models.Comment, error)
	// CommentByPostId 根据文章 ID 获取所有评论
	CommentByPostId(ctx context.Context, postId uint, isPass bool) ([]*models.Comment, error)
	// CommentCount 获取评论数量
	CommentCount(ctx context.Context) (int64, error)
}

type commentRepo struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepo{
		db: db,
	}
}

// AddComment 添加评论
func (r *commentRepo) AddComment(ctx context.Context, comment *models.Comment) (*models.Comment, error) {
	comment.CreateTime = time.Now().UnixMilli()
	err := r.db.WithContext(ctx).Create(comment).Error
	if err != nil {
		return nil, err
	}
	return comment, nil
}

// DeleteCommentById 根据评论 ID 删除评论
func (r *commentRepo) DeleteCommentById(ctx context.Context, id uint) (bool, error) {
	ret := r.db.WithContext(ctx).Delete(&models.Comment{}, id)
	return ret.RowsAffected > 0, ret.Error
}

// DeleteCommentByIds 根据评论 ID 数组删除评论
func (r *commentRepo) DeleteCommentByIds(ctx context.Context, ids []uint) (bool, error) {
	ret := r.db.WithContext(ctx).Delete(&models.Comment{}, ids)
	return ret.RowsAffected > 0, ret.Error
}

// DeleteCommentByPostId 根据文章 ID 删除评论
func (r *commentRepo) DeleteCommentByPostId(ctx context.Context, postId uint) (bool, error) {
	ret := r.db.WithContext(ctx).
		Delete(&models.Comment{}).
		Where("post_id = ?", postId)
	return ret.RowsAffected > 0, ret.Error
}

// DeleteCommentByParentIds 根据父评论 ID 数组删除评论
func (r *commentRepo) DeleteCommentByParentIds(ctx context.Context, parentIds []uint) (bool, error) {
	ret := r.db.WithContext(ctx).
		Delete(&models.Comment{}).
		Where("parent_comment_id IN ?", parentIds)
	return ret.RowsAffected > 0, ret.Error
}

// UpdateComment 修改评论
func (r *commentRepo) UpdateComment(ctx context.Context, comment models.Comment) (bool, error) {
	updates := map[string]any{
		"content":      comment.Content,
		"site":         comment.Site,
		"display_name": comment.DisplayName,
		"email":        comment.Email,
		"is_pass":      comment.IsPass,
	}

	ret := r.db.WithContext(ctx).
		Model(&models.Comment{}).
		Where("comment_id = ?", comment.CommentId).
		Updates(updates)
	return ret.RowsAffected > 0, ret.Error
}

// SetCommentPass 批量设置评论是否通过审核
func (r *commentRepo) SetCommentPass(ctx context.Context, ids []uint, isPass bool) (bool, error) {
	ret := r.db.WithContext(ctx).
		Model(&models.Comment{}).
		Where("comment_id IN ?", ids).
		Update("is_pass", isPass)
	return ret.RowsAffected > 0, ret.Error
}

// CommentByIds 根据评论 ID 数组获取所有评论
func (r *commentRepo) CommentByIds(ctx context.Context, ids []uint) ([]*models.Comment, error) {
	var comments []*models.Comment
	err := r.db.WithContext(ctx).
		Model(&models.Comment{}).
		Where("comment_id IN ?", ids).
		Find(&comments).Error
	if err != nil {
		return []*models.Comment{}, err
	}
	return comments, nil
}

// Comments 获取所有评论
//   - postId: 文章 ID
//   - commentId: 评论 ID
//   - parentId: 父评论 ID
//   - isPass: 是否通过审核
//   - key: 关键字
//   - sort: 排序方式（默认时间降序）
func (r *commentRepo) Comments(
	ctx context.Context,
	postId, commentId, parentId *uint,
	isPass *bool, key *string,
	sort *enum.CommentSort,
) ([]*models.Comment, error) {
	query := r.commentSQL(ctx, postId, commentId, parentId, isPass, key, sort)

	var comments []*models.Comment

	err := query.Find(&comments).Error
	if err != nil {
		return []*models.Comment{}, err
	}

	return comments, nil
}

// CommentsPager 分页获取所有评论
//   - page: 当前页数
//   - size: 每页条数
//   - postId: 文章 ID
//   - commentId: 评论 ID
//   - parentId: 父评论 ID
//   - isPass: 是否通过审核
//   - key: 关键字
//   - sort: 排序方式（默认时间降序）
func (r *commentRepo) CommentsPager(
	ctx context.Context,
	page,
	size int,
	postId,
	commentId,
	parentId *uint,
	isPass *bool,
	key *string,
	sort *enum.CommentSort,
) (*models.Pager[models.Comment], error) {

	query := r.commentSQL(ctx, postId, commentId, parentId, isPass, key, sort)

	if page == 0 {
		// 获取所有评论
		var comments []*models.Comment
		err := query.Find(&comments).Error

		if err != nil {
			return nil, err
		}
		return &models.Pager[models.Comment]{
			Data:       comments,
			Page:       0,
			Size:       0,
			TotalData:  int64(len(comments)),
			TotalPages: 1,
		}, nil
	}

	pager, err := db.PagerBuilder[models.Comment](ctx, r.db, page, size, func(g *gorm.DB) *gorm.DB {
		return query
	})

	if err != nil {
		return nil, err
	}
	return pager, nil
}

// CommentById 根据评论 ID 获取评论
func (r *commentRepo) CommentById(ctx context.Context, id uint) (*models.Comment, error) {
	var comment *models.Comment
	err := r.db.WithContext(ctx).
		Model(&models.Comment{}).
		Where("comment_id = ?", id).
		First(&comment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return comment, nil
}

// CommentByPostId 根据文章 ID 获取所有评论
func (r *commentRepo) CommentByPostId(ctx context.Context, postId uint, isPass bool) ([]*models.Comment, error) {
	var comments []*models.Comment

	err := r.db.WithContext(ctx).
		Model(&models.Comment{}).
		Where("post_id = ?", postId).
		Where("is_pass = ?", isPass).
		Find(&comments).Error

	if err != nil {
		return []*models.Comment{}, err
	}
	return comments, nil
}

// CommentCount 获取评论数量
func (r *commentRepo) CommentCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Comment{}).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// 构建评论查询 SQL
func (r *commentRepo) commentSQL(
	ctx context.Context,
	postId, commentId, parentId *uint,
	isPass *bool, key *string,
	sort *enum.CommentSort,
) *gorm.DB {
	query := r.db.WithContext(ctx).
		Table("comment c").
		Joins("LEFT JOIN post p ON c.post_id = p.post_id")

	if postId != nil {
		query = query.Where("c.post_id = ?", postId)
	}

	if commentId != nil {
		query = query.Where("c.comment_id = ?", commentId)
	}

	if parentId != nil {
		query = query.Where("c.parent_comment_id = ?", parentId)
	}

	if isPass != nil {
		query = query.Where("c.is_pass = ?", isPass)
	}

	if key != nil {
		query = query.Where("c.content LIKE ? OR c.email LIKE ? OR c.display_name LIKE ?", "%"+*key+"%", "%"+*key+"%", "%"+*key+"%")
	}

	if sort == nil {
		// 默认时间降序
		query = query.Order("c.create_time DESC")
	} else {
		switch *sort {
		case enum.CommentSortCreateDesc:
			query = query.Order("c.create_time DESC")
		case enum.CommentSortCreateAsc:
			query = query.Order("c.create_time ASC")
		}
	}

	return query
}
