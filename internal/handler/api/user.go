package api

import (
	"nola-go/internal/models/response"
	"nola-go/internal/service"

	"github.com/gin-gonic/gin"
)

// UserApiHandler 用户博客接口 Handler
type UserApiHandler struct {
	userService *service.UserService
}

// NewUserApiHandler 新建用户博客 Handler
func NewUserApiHandler(s *service.UserService) *UserApiHandler {
	return &UserApiHandler{userService: s}
}

// RegisterApi 注解博客路由
func (h *UserApiHandler) RegisterApi(r *gin.RouterGroup) {
	// 获取博主信息
	r.GET("/blogger", h.getBloggerInfo)
}

// getBloggerInfo 获取博主信息
func (h *UserApiHandler) getBloggerInfo(c *gin.Context) {
	users, err := h.userService.AllUsers(c)

	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	if len(users) == 0 {
		response.OkAndResponse(c, nil)
		return
	}

	response.OkAndResponse(c, map[string]any{
		"email":       users[0].Email,
		"displayName": users[0].DisplayName,
		"description": users[0].Description,
		"avatar":      users[0].Avatar,
	})
}
