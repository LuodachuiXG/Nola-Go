package api

import (
	"nola-go/internal/models/response"
	"nola-go/internal/service"
	"nola-go/internal/util"

	"github.com/gin-gonic/gin"
)

// MenuApiHandler 菜单博客接口
type MenuApiHandler struct {
	menuService *service.MenuService
}

func NewMenuApiHandler(msv *service.MenuService) *MenuApiHandler {
	return &MenuApiHandler{
		menuService: msv,
	}
}

func (h *MenuApiHandler) RegisterApi(r *gin.RouterGroup) {
	publicGroup := r.Group("/menu")
	{
		// 获取主菜单
		publicGroup.GET("", h.getMainMenu)
	}
}

// getMainMenu 获取主菜单
func (h *MenuApiHandler) getMainMenu(c *gin.Context) {
	var req struct {
		Tree *bool `form:"tree"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	if req.Tree == nil {
		// 默认为 true
		req.Tree = util.BoolPtr(true)
	}

	ret, err := h.menuService.MainMenu(c, *req.Tree)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}
