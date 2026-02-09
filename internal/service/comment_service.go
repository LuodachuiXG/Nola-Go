package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"nola-go/internal/logger"
	"nola-go/internal/models"
	"nola-go/internal/models/enum"
	"nola-go/internal/models/response"
	"nola-go/internal/repository"
	"nola-go/internal/util"

	"go.uber.org/zap"
)

// CommentService 评论 Service
type CommentService struct {
	commentRepo repository.CommentRepository
	postRepo    repository.PostRepository
}

// NewCommentService 创建评论 Service
func NewCommentService(commentRepo repository.CommentRepository, postRepo repository.PostRepository) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		postRepo:    postRepo,
	}
}

// AddComment 添加评论
//   - c: 上下文
//   - comment: 评论
//   - isApiRequest: 是否是 API 请求（非管理员请求，即从博客前端提交的请求）
func (s *CommentService) AddComment(
	c context.Context,
	comment models.Comment,
	isApiRequest bool,
) (*models.Comment, error) {
	// 检查评论对应的文章是否存在
	post, err := s.postRepo.PostById(c, comment.PostId, false)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("获取文章 [%d] 失败", comment.PostId), zap.Error(err))
		return nil, response.ServerError
	}

	if post == nil {
		return nil, errors.New(fmt.Sprintf("文章 [%d] 不存在", comment.PostId))
	}

	if isApiRequest && !post.AllowComment {
		// 当前添加的评论是博客前端访客提交，但是当前文章禁止评论
		return nil, errors.New(fmt.Sprintf("文章 [%d] 禁止评论", comment.PostId))
	}

	// 检查父评论是否存在
	if comment.ParentCommentId != nil {
		parentComment, err := s.CommentById(c, *comment.ParentCommentId)
		if err != nil {
			logger.Log.Error(fmt.Sprintf("获取父评论 [%d] 失败", *comment.ParentCommentId), zap.Error(err))
		}
		if parentComment == nil {
			return nil, errors.New(fmt.Sprintf("父评论 [%d] 不存在", *comment.ParentCommentId))
		}
	}

	if comment.ReplyCommentId != nil {
		if comment.ParentCommentId == nil {
			// 有回复的评论，但是父评论为 nil，则当前评论未指定父评论
			return nil, errors.New("未指定父评论")
		}

		// 检查回复的评论是否存在
		replyComment, err := s.CommentById(c, *comment.ReplyCommentId)
		if err != nil {
			logger.Log.Error(fmt.Sprintf("获取回复评论 [%d] 失败", *comment.ReplyCommentId), zap.Error(err))
			return nil, response.ServerError
		}

		if replyComment == nil || *replyComment.ParentCommentId != *comment.ParentCommentId {
			// 回复的评论不存在，或者两个评论的父评论不一致
			return nil, errors.New(fmt.Sprintf("回复的评论 [%d] 不存在", *comment.ReplyCommentId))
		}

		// 设置 replyDisplayName
		comment.ReplyDisplayName = &replyComment.DisplayName
	}

	// 评论内容为空
	if util.StringIsBlank(comment.Content) {
		return nil, errors.New("评论内容不能为空")
	}

	// 名称不能为空
	if util.StringIsBlank(comment.DisplayName) {
		return nil, errors.New("名称不能为空")
	}

	// 站点不合法
	if comment.Site != nil {
		if *comment.Site != "/" && !util.StringIsUrl(*comment.Site) {
			return nil, errors.New("站点格式错误")
		}
	}

	// 邮箱不合法
	if !util.StringIsEmail(comment.Email) {
		return nil, errors.New("邮箱格式错误")
	}

	// 添加评论
	ret, err := s.commentRepo.AddComment(c, &comment)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("添加评论 [%d] 失败", comment.PostId), zap.Error(err))
		return ret, response.ServerError
	}
	return ret, nil
}

// DeleteCommentById 根据评论 ID 删除评论
func (s *CommentService) DeleteCommentById(c context.Context, id uint) (bool, error) {
	// 先获取要删除的评论
	deleteComment, err := s.CommentById(c, id)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("获取评论 [%d] 失败", id), zap.Error(err))
		return false, response.ServerError
	}

	if deleteComment == nil {
		return false, nil
	}

	// 删除评论
	ret, err := s.commentRepo.DeleteCommentById(c, id)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("删除评论 [%d] 失败", id), zap.Error(err))
		return ret, response.ServerError
	}

	// 如果父评论为空，那当前评论是顶层评论，那就有可能有子评论，同步删除子评论
	if ret && deleteComment.ParentCommentId == nil {
		_, err := s.DeleteCommentByParentIds(c, []uint{id})
		if err != nil {
			logger.Log.Error(fmt.Sprintf("同步删除子评论 [%d] 失败", id), zap.Error(err))
		}
	}

	return ret, nil
}

// DeleteCommentByIds 根据评论 ID 数组删除评论
func (s *CommentService) DeleteCommentByIds(c context.Context, ids []uint) (bool, error) {
	// 先尝试获取所有评论
	comments, err := s.commentRepo.CommentByIds(c, ids)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("获取评论 [%v] 失败", ids), zap.Error(err))
		return false, response.ServerError
	}

	if len(comments) == 0 {
		return false, nil
	}

	// 删除所有评论
	commentIds := util.Map(comments, func(comment *models.Comment) uint {
		return comment.CommentId
	})
	ret, err := s.commentRepo.DeleteCommentByIds(c, commentIds)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("删除评论 [%v] 失败", ids), zap.Error(err))
		return ret, response.ServerError
	}

	// 删除上面被删除的评论的可能存在的子评论
	if ret {
		// 父评论 ID 为空的评论是顶层评论，可能有子评论
		toBeDeleted := util.Filter(comments, func(comment *models.Comment) bool {
			return comment.ParentCommentId == nil
		})
		toBeDeletedIds := util.Map(toBeDeleted, func(comment *models.Comment) uint {
			return comment.CommentId
		})

		if len(toBeDeletedIds) > 0 {
			_, err := s.DeleteCommentByParentIds(c, toBeDeletedIds)
			if err != nil {
				logger.Log.Error(fmt.Sprintf("同步删除子评论 [%v] 失败", ids), zap.Error(err))
			}
		}
	}

	return ret, nil
}

// DeleteCommentByPostId 根据文章 ID 删除评论
func (s *CommentService) DeleteCommentByPostId(c context.Context, postId uint) (bool, error) {
	ret, err := s.commentRepo.DeleteCommentByPostId(c, postId)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("删除文章 [%d] 的评论失败", postId), zap.Error(err))
		return ret, response.ServerError
	}
	return ret, nil
}

// DeleteCommentByParentIds 根据父评论 ID 数组删除评论（删除的是这些父评论的子评论，父评论不会删除）
func (s *CommentService) DeleteCommentByParentIds(c context.Context, parentIds []uint) (bool, error) {
	ret, err := s.commentRepo.DeleteCommentByParentIds(c, parentIds)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("删除父评论 [%v] 的子评论失败", parentIds), zap.Error(err))
		return ret, response.ServerError
	}
	return ret, nil
}

// UpdateComment 修改评论
// 仅可修改以下字段：content、site、displayName、email、isPass
func (s *CommentService) UpdateComment(c context.Context, comment models.Comment) (bool, error) {
	if util.StringIsBlank(comment.Content) {
		return false, errors.New("评论内容不能为空")
	}

	if comment.Site != nil {
		if *comment.Site != "/" && !util.StringIsUrl(*comment.Site) {
			return false, errors.New("站点格式错误")
		}
	}

	if util.StringIsBlank(comment.DisplayName) {
		return false, errors.New("名称不能为空")
	}

	if !util.StringIsEmail(comment.Email) {
		return false, errors.New("邮箱格式错误")
	}

	ret, err := s.commentRepo.UpdateComment(c, comment)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("修改评论 [%d] 失败", comment.CommentId), zap.Error(err))
		return ret, response.ServerError
	}
	return ret, nil
}

// SetCommentPass 批量设置评论是否通过审核
func (s *CommentService) SetCommentPass(c context.Context, ids []uint, isPass bool) (bool, error) {
	ret, err := s.commentRepo.SetCommentPass(c, ids, isPass)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("批量设置评论 [%v] 是否通过审核失败", ids), zap.Error(err))
		return ret, response.ServerError
	}
	return ret, nil
}

// Comments 分页获取所有评论
//   - page: 当前页数
//   - size: 每页条数
//   - postId: 文章 ID
//   - slug: 文章别名
//   - commentId: 评论 ID
//   - parentId: 父评论 ID
//   - isPass: 是否通过审核
//   - key: 关键字（内容、名称、邮箱）
//   - sort: 排序方式（默认时间降序）
//   - tree: 是否将子评论放置到父评论的 children 字段中 (默认 false)，
//     此项为 true 时，commentId, parentId, email, displayName, key 参数无效。
func (s *CommentService) Comments(
	c context.Context,
	page, size int,
	postId *uint,
	slug *string,
	commentId, parentId *uint,
	isPass *bool,
	key *string,
	sort *enum.CommentSort,
	tree bool,
) (*models.Pager[models.Comment], error) {
	var mPostId = postId

	if slug != nil && mPostId == nil {
		// 文章别名不为空，文章 ID 为空
		// 先根据文章别名获取到文章，然后再填充文章 ID（此处为了博客前端可以通过文章别名获取评论）
		post, err := s.postRepo.PostBySlug(c, *slug, false)
		if err != nil {
			logger.Log.Error(fmt.Sprintf("根据文章别名 [%s] 获取文章失败", *slug), zap.Error(err))
			return nil, response.ServerError
		}

		if post == nil {
			return nil, errors.New(fmt.Sprintf("文章 [%s] 不存在", *slug))
		}

		mPostId = &post.PostId
	}

	if tree {
		// 需要把子评论放到父评论的 children 字段中
		comments, err := s.commentRepo.Comments(c, mPostId, commentId, parentId, isPass, key, sort)

		if err != nil {
			logger.Log.Error(fmt.Sprintf("获取文章 [%d] 的评论失败", *mPostId), zap.Error(err))
			return nil, response.ServerError
		}

		// 将评论 ID 和评论简历映射
		var idToComment = util.AssociateBy(comments, func(comment *models.Comment) uint {
			return comment.CommentId
		})

		// comments 是所有平铺的评论切片，ParentCommentId 为 nil 的评论默认为是顶层评论（父评论）
		for _, comment := range comments {
			if comment.ParentCommentId != nil {
				// 当前评论的父评论 ID 不为空，获得父评论
				parent, ok := idToComment[*comment.ParentCommentId]
				if !ok || parent == nil {
					continue
				}
				parent.Children = append(parent.Children, *comment)
			}
		}

		comments = util.Filter(comments, func(comment *models.Comment) bool {
			// 此时 comments 中的所有父评论的 children 中都已经填充了子评论
			// 过滤掉子评论
			return comment.ParentCommentId == nil
		})

		// 根据分页拆分
		var group [][]*models.Comment
		if size <= 0 {
			group = [][]*models.Comment{comments}
		} else {
			group = util.Chunk(comments, size)
		}

		if len(group) == 0 {
			return &models.Pager[models.Comment]{
				Data:       []*models.Comment{},
				Page:       page,
				Size:       size,
				TotalData:  0,
				TotalPages: 0,
			}, nil
		}

		if page-1 < 0 || page-1 >= len(group) {
			page = 0
			size = 0
		}

		pager := &models.Pager[models.Comment]{
			Data:       group[util.CoerceAtLeast(page-1, 0)],
			Page:       page,
			Size:       size,
			TotalData:  int64(len(comments)),
			TotalPages: int64(math.Ceil(float64(len(comments)) / float64(util.CoerceAtLeast(size, 1)))),
		}

		return pager, nil
	}

	// 平铺获取所有评论
	ret, err := s.commentRepo.CommentsPager(c, page, size, mPostId, commentId, parentId, isPass, key, sort)
	if err != nil {
		logger.Log.Error("分页获取评论失败", zap.Error(err))
		return nil, response.ServerError
	}
	return ret, nil
}

// CommentById 根据评论 ID 获取评论
func (s *CommentService) CommentById(c context.Context, id uint) (*models.Comment, error) {
	ret, err := s.commentRepo.CommentById(c, id)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("获取评论 [%d] 失败", id), zap.Error(err))
		return ret, response.ServerError
	}
	return ret, nil
}

// CommentCount 获取评论数量
func (s *CommentService) CommentCount(c context.Context) (int64, error) {
	count, err := s.commentRepo.CommentCount(c)
	if err != nil {
		logger.Log.Error("获取评论数量失败", zap.Error(err))
		return count, response.ServerError
	}
	return count, nil
}
