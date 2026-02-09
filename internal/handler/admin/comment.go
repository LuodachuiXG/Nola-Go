package admin

import (
	"nola-go/internal/middleware"
	"nola-go/internal/models"
	"nola-go/internal/models/enum"
	"nola-go/internal/models/request"
	"nola-go/internal/models/response"
	"nola-go/internal/service"
	"nola-go/internal/util"

	"github.com/gin-gonic/gin"
)

// CommentAdminHandler 评论后端接口
type CommentAdminHandler struct {
	commentService *service.CommentService
	tokenService   *service.TokenService
}

func NewCommentAdminHandler(csv *service.CommentService, tsv *service.TokenService) *CommentAdminHandler {
	return &CommentAdminHandler{
		commentService: csv,
		tokenService:   tsv,
	}
}

// RegisterAdmin 注册评论后端路由
func (h *CommentAdminHandler) RegisterAdmin(r *gin.RouterGroup) {

	// 鉴权接口
	privateGroup := r.Group("/comment")
	privateGroup.Use(middleware.AuthMiddleware(h.tokenService))
	{
		// 添加评论
		privateGroup.POST("", h.addComment)
		// 删除评论
		privateGroup.DELETE("", h.deleteComment)
		// 修改评论
		privateGroup.PUT("", h.updateComment)
		// 修改评论是否通过审核
		privateGroup.PUT("/pass", h.updateCommentPass)
		// 获取评论
		privateGroup.GET("", h.getComments)
	}
}

// addComment 添加评论
func (h *CommentAdminHandler) addComment(c *gin.Context) {
	var newComment *request.CommentRequest
	if err := c.ShouldBindBodyWithJSON(&newComment); err != nil {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.commentService.AddComment(c, models.Comment{
		PostId:          newComment.PostId,
		ParentCommentId: newComment.ParentCommentId,
		ReplyCommentId:  newComment.ReplyCommentId,
		Content:         newComment.Content,
		Site:            newComment.Site,
		DisplayName:     newComment.DisplayName,
		Email:           newComment.Email,
		IsPass:          newComment.IsPass,
	}, false)

	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, ret)
}

// deleteComment 删除评论
func (h *CommentAdminHandler) deleteComment(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.commentService.DeleteCommentByIds(c, ids)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// updateComment 修改评论
func (h *CommentAdminHandler) updateComment(c *gin.Context) {
	var comment *request.CommentUpdateRequest
	if err := c.ShouldBindJSON(&comment); err != nil {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.commentService.UpdateComment(c, models.Comment{
		PostId:      0,
		CommentId:   comment.CommentId,
		Content:     comment.Content,
		Site:        comment.Site,
		DisplayName: comment.DisplayName,
		Email:       comment.Email,
		IsPass:      comment.IsPass,
	})

	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// updateCommentPass 修改评论是否通过审核
func (h *CommentAdminHandler) updateCommentPass(c *gin.Context) {
	var req *request.CommentPassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.commentService.SetCommentPass(c, req.Ids, *req.IsPass)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// getComments 获取评论
func (h *CommentAdminHandler) getComments(c *gin.Context) {

	page, size, err := util.ShouldBindPager(c)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	var req struct {
		// PostId 可空文章 ID
		PostId *string `form:"postId"`
		// CommentId 可空评论 ID
		CommentId *string `form:"commentId"`
		// ParentId 可空父评论 ID
		ParentId *string `form:"parentCommentId"`
		// IsPass 可空是否通过审核
		IsPass *bool `form:"isPass"`
		// Key 可空关键词
		Key *string `form:"key"`
		// Sort 可空的排序方式
		Sort *string `form:"sort"`
		// Tree 可空的是否树形结构
		Tree *bool `form:"tree"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	// 文章 ID、评论 ID、父评论 ID
	var postId *uint = nil
	var commentId, parentId *uint
	if (req.PostId != nil && !util.StringIsNumber(*req.PostId)) ||
		(req.CommentId != nil && !util.StringIsNumber(*req.CommentId)) ||
		(req.ParentId != nil && !util.StringIsNumber(*req.ParentId)) {
		response.ParamMismatch(c)
		return
	}

	if req.PostId != nil {
		if pi, err := util.StringToUint(*req.PostId); err == nil {
			postId = &pi
		} else {
			response.ParamMismatch(c)
			return
		}
	}

	if req.CommentId != nil {
		if ci, err := util.StringToUint(*req.CommentId); err == nil {
			commentId = &ci
		} else {
			response.ParamMismatch(c)
			return
		}
	}

	if req.ParentId != nil {
		if pi, err := util.StringToUint(*req.ParentId); err == nil {
			parentId = &pi
		} else {
			response.ParamMismatch(c)
			return
		}
	}

	// 排序方式
	var sort *enum.CommentSort
	if req.Sort != nil {
		sort = enum.CommentSortValueOf(*req.Sort)
		if sort == nil {
			response.ParamMismatch(c)
			return
		}
	}

	// 树形默认 false
	if req.Tree == nil {
		req.Tree = util.BoolPtr(false)
	}

	ret, err := h.commentService.Comments(
		c, page, size, postId, nil, commentId, parentId, req.IsPass, req.Key, sort, *req.Tree,
	)

	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, ret)
}
