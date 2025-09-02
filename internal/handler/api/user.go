package api

import (
	"net/http"
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
	r.GET("/user", func(ctx *gin.Context) {

		ctx.JSON(http.StatusConflict, response.OK("Hello"))
	})
}
