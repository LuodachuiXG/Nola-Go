package admin

import (
	"nola-go/internal/middleware"
	"nola-go/internal/models/response"
	"nola-go/internal/service"
	"nola-go/internal/util"

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
		privateGroup.POST("", h.addTag)
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
