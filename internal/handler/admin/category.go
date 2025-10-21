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

// CategoryAdminHandler 分类后端接口
type CategoryAdminHandler struct {
	categoryService *service.CategoryService
	tokenService    *service.TokenService
}

func NewCategoryAdminHandler(categoryService *service.CategoryService, tsv *service.TokenService) *CategoryAdminHandler {
	return &CategoryAdminHandler{
		categoryService: categoryService,
		tokenService:    tsv,
	}
}

// RegisterAdmin 注册分类后端路由
func (h *CategoryAdminHandler) RegisterAdmin(r *gin.RouterGroup) {

	// 鉴权接口
	privateGroup := r.Group("/category")
	privateGroup.Use(middleware.AuthMiddleware(h.tokenService))
	{
		// 添加分类
		privateGroup.POST("", h.addCategory)
		// 根据分类 ID 数组删除分类
		privateGroup.DELETE("", h.deleteCategoryByIds)
		// 根据分类别名数组删除分类
		privateGroup.DELETE("/slug", h.deleteCategoryBySlugs)
		// 修改分类
		privateGroup.PUT("", h.updateCategory)
		// 根据分类 ID 获取分类
		privateGroup.GET("/:id", h.categoryById)
		// 分页获取分类
		privateGroup.GET("", h.categories)
	}
}

// addCategory 添加分类
func (h *CategoryAdminHandler) addCategory(c *gin.Context) {
	var req struct {
		DisplayName  string  `json:"displayName" binding:"required"`
		Slug         string  `json:"slug" binding:"required"`
		Cover        *string `json:"cover"`
		UnifiedCover *bool   `json:"unifiedCover"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	if util.StringIsBlank(req.DisplayName) || util.StringIsBlank(req.Slug) {
		// 分类名或别名为空
		response.ParamMismatch(c)
		return
	}

	// 添加分类
	category, err := h.categoryService.AddCategory(c, req.DisplayName, req.Slug, req.Cover, req.UnifiedCover)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, category)
}

// deleteCategoryByIds 根据分类 ID 数组删除分类
func (h *CategoryAdminHandler) deleteCategoryByIds(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.ParamMismatch(c)
		return
	}

	if len(ids) == 0 {
		response.OkAndResponse(c, false)
		return
	}

	// 删除分类
	ret, err := h.categoryService.DeleteCategories(c, ids)

	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, ret)
}

// deleteCategoryBySlugs 根据分类别名数组删除分类
func (h *CategoryAdminHandler) deleteCategoryBySlugs(c *gin.Context) {
	var slugs []string

	if err := c.ShouldBindJSON(&slugs); err != nil {
		response.ParamMismatch(c)
		return
	}

	// 删除分类
	ret, err := h.categoryService.DeleteCategoryBySlugs(c, slugs)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, ret)
}

// updateCategory 修改分类
func (h *CategoryAdminHandler) updateCategory(c *gin.Context) {
	var category *models.Category

	if err := c.ShouldBindJSON(&category); err != nil {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.categoryService.UpdateCategory(c, category)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, ret)
}

// categoryById 根据分类 ID 获取分类
func (h *CategoryAdminHandler) categoryById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ParamMismatch(c)
		return
	}

	category, err := h.categoryService.CategoryById(c, uint(id))
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, category)
}

// categories 分页获取分类
func (h *CategoryAdminHandler) categories(c *gin.Context) {
	page, size, err := util.ShouldBindPager(c)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	categories, err := h.categoryService.CategoriesPager(c, page, size)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, categories)
}
