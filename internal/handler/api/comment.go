package api

import (
	"nola-go/internal/models"
	"nola-go/internal/models/request"
	"nola-go/internal/models/response"
	"nola-go/internal/service"
	"nola-go/internal/util"

	"github.com/gin-gonic/gin"
)

// CommentApiHandler 评论博客接口
type CommentApiHandler struct {
	commentService *service.CommentService
}

func NewCommentApiHandler(csv *service.CommentService) *CommentApiHandler {
	return &CommentApiHandler{
		commentService: csv,
	}
}

// RegisterApi 注册评论博客路由
func (h *CommentApiHandler) RegisterApi(r *gin.RouterGroup) {

	publicGroup := r.Group("/comment")
	{
		// 添加评论
		publicGroup.POST("", h.addComment)
		// 根据文章 ID 或别名获取评论
		publicGroup.GET("", h.getComments)
	}
}

// addComment 添加评论
func (h *CommentApiHandler) addComment(c *gin.Context) {
	var req *request.CommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.commentService.AddComment(c, models.Comment{
		PostId:          req.PostId,
		ParentCommentId: req.ParentCommentId,
		ReplyCommentId:  req.ReplyCommentId,
		Content:         req.Content,
		Site:            req.Site,
		DisplayName:     req.DisplayName,
		Email:           req.Email,
		IsPass:          false,
	}, true)

	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, ret)

}

// getComments 根据文章 ID 或别名获取评论
func (h *CommentApiHandler) getComments(c *gin.Context) {
	page, size, err := util.ShouldBindPager(c)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	var req struct {
		// PostId 可空文章 ID
		PostId *string `form:"id"`
		// Slug 可空文章别名
		Slug *string `form:"slug"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	if req.PostId != nil && !util.StringIsNumber(*req.PostId) {
		response.ParamMismatch(c)
		return
	}

	var postId *uint
	if req.PostId != nil {
		if pi, err := util.StringToUint(*req.PostId); err == nil {
			postId = &pi
		} else {
			response.ParamMismatch(c)
			return
		}
	}

	if postId == nil && util.StringIsNilOrBlank(req.Slug) {
		response.FailAndResponse(c, "文章 ID 和文章别名至少提供一个")
		return
	}

	ret, err := h.commentService.Comments(c, page, size, postId, req.Slug, nil, nil, util.BoolPtr(true), nil, nil, true)

	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}
