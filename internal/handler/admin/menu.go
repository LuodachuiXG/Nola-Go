package admin

import (
	"fmt"
	"nola-go/internal/middleware"
	"nola-go/internal/models/request"
	"nola-go/internal/models/response"
	"nola-go/internal/service"
	"nola-go/internal/util"

	"github.com/gin-gonic/gin"
)

// MenuAdminHandler 菜单后端接口
type MenuAdminHandler struct {
	menuService  *service.MenuService
	tokenService *service.TokenService
}

func NewMenuAdminHandler(msv *service.MenuService, tsv *service.TokenService) *MenuAdminHandler {
	return &MenuAdminHandler{
		menuService:  msv,
		tokenService: tsv,
	}
}

func (h *MenuAdminHandler) RegisterAdmin(r *gin.RouterGroup) {
	privateGroup := r.Group("/menu")
	privateGroup.Use(middleware.AuthMiddleware(h.tokenService))
	{
		// 添加菜单
		privateGroup.POST("", h.addMenu)
		// 删除菜单
		privateGroup.DELETE("", h.deleteMenus)
		// 修改菜单
		privateGroup.PUT("", h.updateMenu)
		// 获取菜单
		privateGroup.GET("", h.getMenus)
		// 添加菜单项
		privateGroup.POST("/item", h.addMenuItem)
		// 删除菜单项
		privateGroup.DELETE("/item", h.deleteMenuItems)
		// 修改菜单项
		privateGroup.PUT("/item", h.updateMenuItem)
		// 获取菜单项 - 菜单 ID
		privateGroup.GET("/item/:menuId", h.getMenuItems)
	}
}

// addMenu 添加菜单
func (h *MenuAdminHandler) addMenu(c *gin.Context) {
	var req *request.MenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("%v", req)
		response.ParamMismatch(c)
		return
	}

	ret, err := h.menuService.AddMenu(c, req)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// deleteMenus 删除菜单
func (h *MenuAdminHandler) deleteMenus(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.menuService.DeleteMenu(c, ids)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// updateMenu 修改菜单
func (h *MenuAdminHandler) updateMenu(c *gin.Context) {
	var req *request.MenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}
	ret, err := h.menuService.UpdateMenu(c, req)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// getMenus 获取菜单
func (h *MenuAdminHandler) getMenus(c *gin.Context) {
	page, size, err := util.ShouldBindPager(c)
	if err != nil {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.menuService.MenusPager(c, page, size)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// addMenuItem 添加菜单项
func (h *MenuAdminHandler) addMenuItem(c *gin.Context) {
	var req *request.MenuItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	// 菜单项参数校验
	if req.ParentMenuId <= 0 || req.Index < 0 {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.menuService.AddMenuItem(c, req)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)

}

// deleteMenuItems 删除菜单项
func (h *MenuAdminHandler) deleteMenuItems(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.menuService.DeleteMenuItems(c, ids)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// updateMenuItem 修改菜单项
func (h *MenuAdminHandler) updateMenuItem(c *gin.Context) {
	var req *request.MenuItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	// 参数校验
	if req.MenuItemId == nil || req.ParentMenuId <= 0 || req.Index < 0 {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.menuService.UpdateMenuItem(c, req)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// getMenuItems 获取菜单项
func (h *MenuAdminHandler) getMenuItems(c *gin.Context) {
	var req struct {
		// MenuId 菜单 ID
		MenuId uint `uri:"menuId" bind:"required"`
		// Tree 是否构建树型
		Tree *bool `form:"tree"`
	}

	if err := c.ShouldBindUri(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	if req.Tree == nil {
		// Tree 默认为 true
		req.Tree = util.BoolPtr(true)
	}

	ret, err := h.menuService.MenuItems(c, req.MenuId, *req.Tree)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}
