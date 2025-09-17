package admin

import (
	"nola-go/internal/middleware"
	"nola-go/internal/models/request"
	"nola-go/internal/models/response"
	"nola-go/internal/service"

	"github.com/gin-gonic/gin"
)

// UserAdminHandler 用户后端接口 Handler
type UserAdminHandler struct {
	userService  *service.UserService
	tokenService *service.TokenService
}

// NewUserAdminHandler 新建用户后端 Handler
func NewUserAdminHandler(s *service.UserService, tsv *service.TokenService) *UserAdminHandler {
	return &UserAdminHandler{userService: s, tokenService: tsv}
}

// RegisterAdmin 注册后端路由
func (h *UserAdminHandler) RegisterAdmin(r *gin.RouterGroup) {
	// 需要鉴权接口
	privateGroup := r.Group("/user")
	privateGroup.Use(middleware.AuthMiddleware(h.tokenService))
	{
		// 验证登录是否过期
		privateGroup.GET("/validate", func(c *gin.Context) {
			response.OkAndResponse(c, true)
		})
		// 获取登录用户信息
		privateGroup.GET("", h.getLoginUser)
		// 修改登录用户信息
		privateGroup.PUT("", h.updateUser)
		// 修改密码
		privateGroup.PUT("/password", h.updatePassword)

	}

	// 无需鉴权接口
	publicGroup := r.Group("/user")
	{
		// 用户登录
		publicGroup.POST("/login", h.loginUser)
	}
}

// getLoginUser 获取登录用户的信息
func (h *UserAdminHandler) getLoginUser(c *gin.Context) {
	userId := c.GetUint("uid")

	if userId == 0 {
		response.UnauthorizedAndResponse(c)
		return
	}

	user, err := h.userService.UserById(c, userId)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, user)
}

// updateUser 修改登录用户信息
func (h *UserAdminHandler) updateUser(c *gin.Context) {
	userId := c.GetUint("uid")

	if userId == 0 {
		response.UnauthorizedAndResponse(c)
		return
	}

	var req request.UserInfoRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	res, err := h.userService.UpdateUser(c, userId, &req)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, res)
}

// updatePassword 修改密码
func (h *UserAdminHandler) updatePassword(c *gin.Context) {
	userId := c.GetUint("uid")

	if userId == 0 {
		response.UnauthorizedAndResponse(c)
		return
	}

	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	res, err := h.userService.UpdatePassword(c, userId, req.Password)

	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, res)
}

// loginUser 用户登录
func (h *UserAdminHandler) loginUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	res, err := h.userService.Login(c, req.Username, req.Password)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, res)
}
