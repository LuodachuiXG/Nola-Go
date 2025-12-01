package api

import (
	"nola-go/internal/models/enum"
	"nola-go/internal/models/response"
	"nola-go/internal/service"
	"nola-go/internal/util"
	"strconv"

	"github.com/gin-gonic/gin"
)

// PostApiHandler 文章博客接口
type PostApiHandler struct {
	postService *service.PostService
}

func NewPostApiHandler(psv *service.PostService) *PostApiHandler {
	return &PostApiHandler{
		postService: psv,
	}
}

func (h *PostApiHandler) RegisterApi(r *gin.RouterGroup) {
	group := r.Group("/post")
	{
		// 分页获取文章
		group.GET("", h.getPost)
		// 获取文章 - 根据文章 ID
		group.GET("/:id", h.getPostById)
		// 获取文章 - 根据别名
		group.GET("/slug/:slug", h.getPostBySlug)
		// 获取文章内容
		group.GET("/content", h.getPostContent)
	}
}

// getPost 分页获取文章
func (h *PostApiHandler) getPost(c *gin.Context) {
	page, size, err := util.ShouldBindPager(c)

	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	var req struct {
		Key        *string `form:"key"`
		TagId      *string `form:"tagId"`
		CategoryId *string `form:"categoryId"`
		Tag        *string `form:"tag"`
		Category   *string `form:"category"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	// 文章标签、分类是否是数字
	if (req.TagId != nil && !util.StringIsNumber(*req.TagId)) ||
		(req.CategoryId != nil && !util.StringIsNumber(*req.CategoryId)) {
		response.ParamMismatch(c)
		return
	}

	var tagId, categoryId *uint
	if req.TagId != nil {
		if tagUint, err := strconv.ParseUint(*req.TagId, 10, 32); err == nil {
			tagId = new(uint)
			*tagId = uint(tagUint)
		} else {
			response.ParamMismatch(c)
			return
		}
	}

	if req.CategoryId != nil {
		if categoryUint, err := strconv.ParseUint(*req.CategoryId, 10, 32); err == nil {
			categoryId = new(uint)
			*categoryId = uint(categoryUint)
		} else {
			response.ParamMismatch(c)
			return
		}
	}

	ret, err := h.postService.ApiPosts(c, page, size, req.Key, tagId, categoryId, req.Tag, req.Category)

	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// getPostById 获取文章 - 根据文章 ID
func (h *PostApiHandler) getPostById(c *gin.Context) {
	var req struct {
		Id uint `uri:"id"`
	}

	if err := c.ShouldBindUri(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.postService.PostById(c, req.Id, true)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	if ret == nil || ret.Status != enum.PostStatusPublished {
		// 文章不存在，或者文章未发布，则返回 404
		response.NotFoundAndResponse(c)
		return
	}

	response.OkAndResponse(c, ret)
}

// getPostBySlug 获取文章 - 根据别名
func (h *PostApiHandler) getPostBySlug(c *gin.Context) {
	var req struct {
		Slug string `uri:"slug"`
	}

	if err := c.ShouldBindUri(&req); err != nil {
		response.ParamMismatch(c)
	}

	ret, err := h.postService.PostBySlug(c, req.Slug, true)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	if ret == nil || ret.Status != enum.PostStatusPublished {
		// 文章不存在，或者文章未发布，返回 404
		response.NotFoundAndResponse(c)
		return
	}
	response.OkAndResponse(c, ret)
}

// getPostContent 获取文章内容
func (h *PostApiHandler) getPostContent(c *gin.Context) {
	var req struct {
		ID       *uint   `form:"id" json:"id,omitempty"`
		Slug     *string `form:"slug" json:"slug,omitempty"`
		Password *string `form:"password" json:"password,omitempty"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	if req.ID == nil && req.Slug == nil {
		// 文章 ID 和别名都为空
		response.FailAndResponse(c, "文章不存在或不可见")
		return
	}

	ret, err := h.postService.ApiPostContent(c, req.ID, req.Slug, req.Password)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	if ret == nil {
		response.FailAndResponse(c, "文章不存在或不可见")
		return
	}

	response.OkAndResponse(c, ret)
}
