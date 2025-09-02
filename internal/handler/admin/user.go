package admin

import (
	"nola-go/internal/middleware"
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
	value, exists := c.Get("uid")
	if !exists {
		response.FailAndResponse(c, "未知用户")
		return
	}

	userId, ret := value.(uint)
	if !ret {
		response.FailAndResponse(c, "未知用户")
		return
	}

	user, err := h.userService.UserById(c, userId)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, user)
}

// loginUser 用户登录
func (h *UserAdminHandler) loginUser(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
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
