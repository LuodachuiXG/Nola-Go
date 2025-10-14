package api

import (
	"nola-go/internal/models/response"
	"nola-go/internal/service"

	"github.com/gin-gonic/gin"
)

// ConfigApiHandler 配置博客接口 Handler
type ConfigApiHandler struct {
	configService *service.ConfigService
	userService   *service.UserService
}

// NewConfigApiHandler 新建配置博客 Handler
func NewConfigApiHandler(configService *service.ConfigService, userService *service.UserService) *ConfigApiHandler {
	return &ConfigApiHandler{
		configService: configService,
		userService:   userService,
	}
}

// RegisterApi 注册配置博客路由
func (h *ConfigApiHandler) RegisterApi(r *gin.RouterGroup) {
	group := r.Group("/config")
	{
		// 获取博客信息
		group.GET("/blog", h.getBlogInfo)
	}
}

// getBlogInfo 获取博客信息
func (h *ConfigApiHandler) getBlogInfo(c *gin.Context) {
	// 获取博客信息
	blogInfo, err := h.configService.BlogInfo(c)

	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	// 获取博主信息
	users, err := h.userService.AllUsers(c)

	if len(users) != 0 && blogInfo != nil {
		// 博主数量不为空，填充博主名称
		blogInfo.Blogger = &users[0].DisplayName
	}

	response.OkAndResponse(c, blogInfo)
}
