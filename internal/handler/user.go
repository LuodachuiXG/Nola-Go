package handler

import (
	"net/http"
	"nola-go/internal/models/response"
	"nola-go/internal/service"

	"github.com/gin-gonic/gin"
)

// UserAdminHandler 用户后端接口 Handler
type UserAdminHandler struct {
	userService *service.UserService
}

// NewUserAdminHandler 新建用户后端 Handler
func NewUserAdminHandler(s *service.UserService) *UserAdminHandler {
	return &UserAdminHandler{userService: s}
}

// RegisterAdmin 注册后端路由
func (h *UserAdminHandler) RegisterAdmin(r *gin.RouterGroup) {
	r.POST("/user", h.createUser)
	r.GET("/users/:id", h.getUser)
}

func (h *UserAdminHandler) createUser(c *gin.Context) {}

func (h *UserAdminHandler) getUser(c *gin.Context) {}

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
