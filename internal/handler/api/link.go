package api

import (
	"nola-go/internal/models"
	"nola-go/internal/models/response"
	"nola-go/internal/service"
	"nola-go/internal/util"

	"github.com/gin-gonic/gin"
)

// LinkApiHandler 友情链接博客接口
type LinkApiHandler struct {
	linkService *service.LinkService
}

func NewLinkApiHandler(linkService *service.LinkService) *LinkApiHandler {
	return &LinkApiHandler{
		linkService: linkService,
	}
}

// RegisterApi 注册友情链接博客路由
func (h *LinkApiHandler) RegisterApi(r *gin.RouterGroup) {
	publicGroup := r.Group("/link")
	{
		publicGroup.GET("", h.getLinks)
	}
}

// getLinks 获取友情链接
func (h *LinkApiHandler) getLinks(c *gin.Context) {
	page, size, err := util.ShouldBindPager(c)
	if err != nil {
		response.FailAndResponse(c, err.Error())
	}

	links, err := h.linkService.LinksPager(c, page, size, nil)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	// 将友联转为响应体
	linksTransform := util.Map(links.Data, func(link *models.Link) *response.LinkResponse {
		return response.LinkResponseValueOf(link)
	})

	response.OkAndResponse(c, models.Pager[response.LinkResponse]{
		Page:       links.Page,
		Size:       links.Size,
		Data:       linksTransform,
		TotalData:  links.TotalData,
		TotalPages: links.TotalPages,
	})
}
