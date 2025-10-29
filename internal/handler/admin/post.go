package admin

import (
	"nola-go/internal/middleware"
	"nola-go/internal/models/enum"
	"nola-go/internal/models/request"
	"nola-go/internal/models/response"
	"nola-go/internal/service"
	"nola-go/internal/util"
	"strconv"

	"github.com/gin-gonic/gin"
)

// PostAdminHandler 文章后端接口
type PostAdminHandler struct {
	postService  *service.PostService
	tokenService *service.TokenService
}

func NewPostAdminHandler(psv *service.PostService, tsv *service.TokenService) *PostAdminHandler {
	return &PostAdminHandler{
		postService:  psv,
		tokenService: tsv,
	}
}

func (h *PostAdminHandler) RegisterAdmin(r *gin.RouterGroup) {
	privateGroup := r.Group("/post")
	privateGroup.Use(middleware.AuthMiddleware(h.tokenService))
	{
		// 添加文章
		privateGroup.POST("", h.addPost)
		// 回收文章 - 根据文章 ID 数组
		privateGroup.PUT("/recycle", h.recyclePost)
		// 恢复文章 - 根据文章 ID 数组
		privateGroup.PUT("/restore/:status", h.restorePost)
		// 删除文章 - 根据文章 ID 数组
		privateGroup.DELETE("", h.deletePost)
		// 修改文章
		privateGroup.PUT("", h.updatePost)
		// 修改文章状态（文章状态、可见性、置顶）
		privateGroup.PUT("/status", h.updatePostStatus)
		// 获取文章
		privateGroup.GET("", h.getPost)
		// 获取文章所有内容，包括正文和所有草稿
		privateGroup.GET("/content/:id", h.getAllContent)
		// 获取文章 - 根据文章 ID
		privateGroup.GET("/:id", h.getPostById)
		// 获取文章 - 根据文章别名
		privateGroup.GET("/slug/:slug", h.getPostBySlug)
		// 修改文章正文
		privateGroup.PUT("/publish", h.updatePostContent)
		// 获取文章正文
		privateGroup.GET("/publish/:id", h.getPostContent)
		// 添加文章草稿
		privateGroup.POST("/draft", h.addPostDraft)
		// 删除文章草稿 - 根据草稿名数组
		privateGroup.DELETE("/draft/:id", h.deletePostDraft)
		// 修改文章草稿
		privateGroup.PUT("/draft", h.updatePostDraft)
		// 修改文章草稿名
		privateGroup.PUT("/draft/name", h.updatePostDraftName)
		// 将文章草稿转换为文章正文
		privateGroup.PUT("/draft/publish", h.updatePostDraftToPublish)
		// 获取文章草稿
		privateGroup.PUT("/:id/draft/:draftName", h.getPostDraft)
	}
}

// addPost 添加文章
func (h *PostAdminHandler) addPost(c *gin.Context) {
	var req *request.PostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	if req.Status == enum.PostStatusDeleted {
		// 文章状态不能设置为已删除
		response.ParamMismatch(c)
		return
	}

	ret, err := h.postService.AddPost(c, req)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, ret)
}

// recyclePost 回收文章 - 根据文章 ID 数组
func (h *PostAdminHandler) recyclePost(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.postService.UpdatePostStatusToDeleted(c, ids)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// restorePost 恢复文章 - 根据文章 ID 数组
func (h *PostAdminHandler) restorePost(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.ParamMismatch(c)
		return
	}

	var uri struct {
		Status string `uri:"status"`
	}

	if err := c.ShouldBindUri(&uri); err != nil {
		response.ParamMismatch(c)
		return
	}

	statusEnum := enum.PostStatusValueOf(uri.Status)
	if statusEnum == nil {
		response.ParamMismatch(c)
		return
	}

	// 状态不能为已删除
	if *statusEnum == enum.PostStatusDeleted {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.postService.UpdatePostStatusTo(c, ids, *statusEnum)

	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// deletePost 删除文章 - 根据文章 ID 数组
func (h *PostAdminHandler) deletePost(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.postService.DeletePosts(c, ids)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// updatePost 修改文章
func (h *PostAdminHandler) updatePost(c *gin.Context) {
	var req *request.PostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	if req.PostId == nil {
		// 没有传文章 ID
		response.ParamMismatch(c)
		return
	}

	if req.Encrypted != nil && *req.Encrypted == true && util.StringIsNilOrBlank(req.Password) {
		// 文章设为加密，但是没有提供密码
		response.FailAndResponse(c, "文章设为加密需要提供密码")
		return
	}

	ret, err := h.postService.UpdatePost(c, req)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// updatePostStatus 修改文章状态（文章状态、可见性、置顶）
func (h *PostAdminHandler) updatePostStatus(c *gin.Context) {
	var req *request.PostStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.postService.UpdatePostStatus(c, req)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// getPost 获取文章
func (h *PostAdminHandler) getPost(c *gin.Context) {
	page, size, err := util.ShouldBindPager(c)
	if err != nil {
		response.FailAndResponse(c, err.Error())
	}

	var req struct {
		Status   string  `form:"status"`
		Visible  string  `form:"visible"`
		key      *string `form:"key"`
		tag      *string `form:"tag"`
		category *string `form:"category"`
		sort     *string `form:"sort"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	// 文章状态和可见性
	statusEnum := enum.PostStatusValueOf(req.Status)
	visibleEnum := enum.PostVisibleValueOf(req.Visible)
	if statusEnum == nil || visibleEnum == nil {
		response.ParamMismatch(c)
		return
	}

	// 文章排序
	var sortEnum *enum.PostSort = nil
	if req.sort != nil {
		sortEnum = enum.PostSortValueOf(*req.sort)
		if sortEnum == nil {
			response.ParamMismatch(c)
			return
		}
	}

	// 文章标签、分类是否是数字
	if (req.tag != nil && !util.StringIsNumber(*req.tag)) ||
		(req.category != nil && !util.StringIsNumber(*req.category)) {
		response.ParamMismatch(c)
		return
	}

	var tagId, categoryId *uint
	if req.tag != nil {
		if tagUint, err := strconv.ParseUint(*req.tag, 10, 32); err == nil {
			tagId = new(uint)
			*tagId = uint(tagUint)
		} else {
			response.ParamMismatch(c)
			return
		}
	}

	if req.category != nil {
		if categoryUint, err := strconv.ParseUint(*req.category, 10, 32); err == nil {
			categoryId = new(uint)
			*categoryId = uint(categoryUint)
		} else {
			response.ParamMismatch(c)
			return
		}
	}

	ret, err := h.postService.PostPager(c, page, size, statusEnum, visibleEnum, req.key, tagId, categoryId, sortEnum)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, ret)
}

// getAllContent 获取文章所有内容，包括正文和所有草稿
func (h *PostAdminHandler) getAllContent(c *gin.Context) {
	var req struct {
		Id uint `uri:"id"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		response.ParamMismatch(c)
		return
	}
	ret, err := h.postService.PostContents(c, req.Id)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// getPostByIds 获取文章 - 根据文章 ID
func (h *PostAdminHandler) getPostById(c *gin.Context) {
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
	response.OkAndResponse(c, ret)
}

// getPostBySlug 获取文章 - 根据文章别名
func (h *PostAdminHandler) getPostBySlug(c *gin.Context) {
	var req struct {
		Slug string `uri:"slug"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		response.ParamMismatch(c)
		return
	}
	ret, err := h.postService.PostBySlug(c, req.Slug, true)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// updatePostContent 修改文章正文
func (h *PostAdminHandler) updatePostContent(c *gin.Context) {
	var req *request.PostContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}
	ret, err := h.postService.UpdatePostContent(c, *req, enum.PostContentStatusPublished, nil)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// getPostContent 获取文章正文
func (h *PostAdminHandler) getPostContent(c *gin.Context) {
	var req struct {
		Id uint `uri:"id"`
	}

	if err := c.ShouldBindUri(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.postService.PostContent(c, req.Id, enum.PostContentStatusPublished, nil)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// addPostDraft 添加文章草稿
func (h *PostAdminHandler) addPostDraft(c *gin.Context) {
	var req *request.PostDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.postService.AddPostDraft(c, req)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// deletePostDraft 删除文章草稿 - 根据草稿名数组
func (h *PostAdminHandler) deletePostDraft(c *gin.Context) {
	var uri struct {
		Id uint `uri:"id"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		response.ParamMismatch(c)
		return
	}
	var draftNames []string
	if err := c.ShouldBindJSON(&draftNames); err != nil {
		response.ParamMismatch(c)
		return
	}
	ret, err := h.postService.DeletePostContent(c, uri.Id, enum.PostContentStatusDraft, draftNames)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// updatePostDraft 修改文章草稿
func (h *PostAdminHandler) updatePostDraft(c *gin.Context) {
	var req *request.PostDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}
	ret, err := h.postService.UpdatePostContent(c, request.PostContentRequest{PostId: req.PostId, Content: req.Content}, enum.PostContentStatusDraft, &req.DraftName)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// updatePostDraftName 修改文章草稿名
func (h *PostAdminHandler) updatePostDraftName(c *gin.Context) {
	var req struct {
		PostId uint `json:"postId" binding:"required"`
		// OldName 旧草稿名
		OldName string `json:"oldName" binding:"required"`
		// NewName 新草稿名
		NewName string `json:"newName" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.postService.UpdatePostDraftName(c, req.PostId, req.OldName, req.NewName)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// updatePostDraftToPublish 将文章草稿转换为文章正文
func (h *PostAdminHandler) updatePostDraftToPublish(c *gin.Context) {
	var req struct {
		PostId uint `json:"postId" binding:"required"`
		// DraftName 草稿名
		DraftName string `json:"draftName" binding:"required"`
		// DeleteContent 是否删除原来的正文
		DeleteContent bool `json:"deleteContent"`
		// ContentName 正文草稿名
		ContentName *string `json:"contentName"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.postService.UpdatePostDraftToContent(c, req.PostId, req.DraftName, req.DeleteContent, req.ContentName)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// getPostDraft 获取文章草稿
func (h *PostAdminHandler) getPostDraft(c *gin.Context) {
	var req struct {
		Id        uint   `uri:"id"`
		DraftName string `uri:"draftName"`
	}

	ret, err := h.postService.PostContent(c, req.Id, enum.PostContentStatusDraft, &req.DraftName)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}
