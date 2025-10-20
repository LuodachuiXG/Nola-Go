package api

import (
	"nola-go/internal/models/response"
	"nola-go/internal/service"
	"nola-go/internal/util"

	"github.com/gin-gonic/gin"
)

// TagApiHandler 标签博客接口 Handler
type TagApiHandler struct {
	tagService *service.TagService
}

// NewTagApiHandler 新建标签博客 Handler
func NewTagApiHandler(tagService *service.TagService) *TagApiHandler {
	return &TagApiHandler{
		tagService: tagService,
	}
}

func (h *TagApiHandler) RegisterApi(r *gin.RouterGroup) {
	publicGroup := r.Group("/tag")
	{
		publicGroup.GET("", h.getTag)
	}
}

// getTag 获取标签
func (h *TagApiHandler) getTag(c *gin.Context) {
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
