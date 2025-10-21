package api

import (
	"nola-go/internal/models/response"
	"nola-go/internal/service"
	"nola-go/internal/util"

	"github.com/gin-gonic/gin"
)

// CategoryApiHandler 分类博客接口
type CategoryApiHandler struct {
	categoryService *service.CategoryService
}

func NewCategoryApiHandler(categoryService *service.CategoryService) *CategoryApiHandler {
	return &CategoryApiHandler{
		categoryService: categoryService,
	}
}

// RegisterApi 注册分类博客路由
func (h *CategoryApiHandler) RegisterApi(r *gin.RouterGroup) {

	publicGroup := r.Group("/category")
	{
		publicGroup.GET("", h.getCategory)
	}
}

// getCategory 获取分类
func (h *CategoryApiHandler) getCategory(c *gin.Context) {
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
