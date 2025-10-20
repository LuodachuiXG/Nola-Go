package admin

import (
	"nola-go/internal/middleware"
	"nola-go/internal/models"
	"nola-go/internal/models/response"
	"nola-go/internal/service"
	"nola-go/internal/util"
	"strconv"

	"github.com/gin-gonic/gin"
)

// TagAdminHandler 标签后端接口
type TagAdminHandler struct {
	tagService   *service.TagService
	tokenService *service.TokenService
}

func NewTagAdminHandler(tagService *service.TagService, tsv *service.TokenService) *TagAdminHandler {
	return &TagAdminHandler{
		tagService:   tagService,
		tokenService: tsv,
	}
}

// RegisterAdmin 注册标签后端路由
func (h *TagAdminHandler) RegisterAdmin(r *gin.RouterGroup) {

	// 鉴权接口
	privateGroup := r.Group("/tag")
	privateGroup.Use(middleware.AuthMiddleware(h.tokenService))
	{
		// 添加标签
		privateGroup.POST("", h.addTag)
		// 根据标签 ID 数组删除标签
		privateGroup.DELETE("", h.deleteTagByIds)
		// 根据标签别名数组删除标签
		privateGroup.DELETE("/slug", h.deleteTagBySlugs)
		// 修改标签
		privateGroup.PUT("", h.updateTag)
		// 根据标签 ID 获取标签
		privateGroup.GET("/:id", h.tagById)
		// 分页获取标签
		privateGroup.GET("", h.tags)
	}
}

// addTag 添加标签
func (h *TagAdminHandler) addTag(c *gin.Context) {
	var req struct {
		DisplayName string  `json:"displayName" binding:"required"`
		Slug        string  `json:"slug" binding:"required"`
		Color       *string `json:"color"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	if util.StringIsBlank(req.DisplayName) || util.StringIsBlank(req.Slug) {
		// 标签名或别名为空
		response.ParamMismatch(c)
		return
	}

	// 添加标签
	tag, err := h.tagService.AddTag(c, req.DisplayName, req.Slug, req.Color)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, tag)
}

// deleteTagByIds 根据标签 ID 数组删除标签
func (h *TagAdminHandler) deleteTagByIds(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.ParamMismatch(c)
		return
	}

	// 删除标签
	ret, err := h.tagService.DeleteTags(c, ids)

	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, ret)
}

// deleteTagBySlugs 根据标签别名数组删除标签
func (h *TagAdminHandler) deleteTagBySlugs(c *gin.Context) {
	var slugs []string

	if err := c.ShouldBindJSON(&slugs); err != nil {
		response.ParamMismatch(c)
		return
	}

	// 删除标签
	ret, err := h.tagService.DeleteTagBySlugs(c, slugs)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, ret)
}

// updateTag 修改标签
func (h *TagAdminHandler) updateTag(c *gin.Context) {
	var tag *models.Tag

	if err := c.ShouldBindJSON(&tag); err != nil {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.tagService.UpdateTag(c, tag)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, ret)
}

// tagById 根据标签 ID 获取标签
func (h *TagAdminHandler) tagById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamMismatch(c)
		return
	}

	tag, err := h.tagService.TagById(c, uint(id))
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, tag)
}

// tags 分页获取标签
func (h *TagAdminHandler) tags(c *gin.Context) {
	page, size, err := util.ShouldBindPager(c)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	tags, err := h.tagService.TagsPager(c, page, size)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, tags)
}
