package admin

import (
	"nola-go/internal/middleware"
	"nola-go/internal/models/enum"
	"nola-go/internal/models/request"
	"nola-go/internal/models/response"
	"nola-go/internal/service"
	"nola-go/internal/util"

	"github.com/gin-gonic/gin"
)

// LinkAdminHandler 友情链接后端接口
type LinkAdminHandler struct {
	linkService  *service.LinkService
	tokenService *service.TokenService
}

func NewLinkAdminHandler(linkService *service.LinkService, tsv *service.TokenService) *LinkAdminHandler {
	return &LinkAdminHandler{
		linkService:  linkService,
		tokenService: tsv,
	}
}

// RegisterAdmin 注册友情链接后端路由
func (h *LinkAdminHandler) RegisterAdmin(r *gin.RouterGroup) {
	privateGroup := r.Group("/link")
	privateGroup.Use(middleware.AuthMiddleware(h.tokenService))
	{
		// 添加友联
		privateGroup.POST("", h.addLink)
		// 删除友联
		privateGroup.DELETE("", h.deleteLink)
		// 修改友情链接
		privateGroup.PUT("", h.updateLink)
		// 获取友情链接
		privateGroup.GET("", h.getLinks)
	}
}

// addLink 添加友联
func (h *LinkAdminHandler) addLink(c *gin.Context) {
	var req *request.LinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	// 添加友联
	link, err := h.linkService.AddLink(c, req)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, link)
}

// deleteLink 删除友联
func (h *LinkAdminHandler) deleteLink(c *gin.Context) {
	var linkIds []uint
	if err := c.ShouldBindJSON(&linkIds); err != nil {
		response.ParamMismatch(c)
		return
	}

	if len(linkIds) == 0 {
		response.OkAndResponse(c, false)
		return
	}

	// 删除友联
	ret, err := h.linkService.DeleteLinks(c, linkIds)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, ret)
}

// updateLink 修改友情链接
func (h *LinkAdminHandler) updateLink(c *gin.Context) {
	var req *request.LinkRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	// 修改时，ID 不能为空
	if req.LinkId == nil {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.linkService.UpdateLink(c, req)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, ret)
}

// getLinks 获取友情链接
func (h *LinkAdminHandler) getLinks(c *gin.Context) {
	page, size, err := util.ShouldBindPager(c)
	if err != nil {
		response.ParamMismatch(c)
		return
	}

	var req struct {
		Sort *enum.LinkSort `form:"sort"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	links, err := h.linkService.LinksPager(c, page, size, req.Sort)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, links)
}
