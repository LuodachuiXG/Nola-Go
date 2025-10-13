package admin

import (
	"nola-go/internal/middleware"
	"nola-go/internal/models"
	"nola-go/internal/models/response"
	"nola-go/internal/service"
	"nola-go/internal/util"
	"time"

	"github.com/gin-gonic/gin"
)

// ConfigAdminHandler 配置后端接口
type ConfigAdminHandler struct {
	configService *service.ConfigService
	userService   *service.UserService
	tokenService  *service.TokenService
}

// NewConfigAdminHandler 新建配置后端 Handler
func NewConfigAdminHandler(s *service.ConfigService, usv *service.UserService, tsv *service.TokenService) *ConfigAdminHandler {
	return &ConfigAdminHandler{
		configService: s,
		userService:   usv,
		tokenService:  tsv,
	}
}

// RegisterAdmin 注册配置后端路由
func (h *ConfigAdminHandler) RegisterAdmin(r *gin.RouterGroup) {
	// 鉴权接口
	privateGroup := r.Group("/config")
	privateGroup.Use(middleware.AuthMiddleware(h.tokenService))
	{

	}

	// 无鉴权接口
	publicGroup := r.Group("/config")
	{
		publicGroup.POST("/blog", h.initBlogInfo)
		publicGroup.POST("/blog/admin", h.initBlogger)
	}
}

// 初始化播客信息
func (h *ConfigAdminHandler) initBlogInfo(c *gin.Context) {
	var req struct {
		Title    string `json:"title" binding:"required"`
		Subtitle string `json:"subtitle" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	// 判断博客是否已经完成初始化
	blogInfo, err := h.configService.BlogInfo(c)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	if blogInfo != nil {
		response.FailAndResponse(c, "博客已经创建")
		return
	}

	// 初始化博客
	ret, err := h.configService.SetBlogInfo(c, &models.BlogInfo{
		Title:      &req.Title,
		Subtitle:   &req.Subtitle,
		CreateDate: util.Int64Ptr(time.Now().UnixMilli()),
	})

	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, ret)

}

// initBlogger 初始化管理员
func (h *ConfigAdminHandler) initBlogger(c *gin.Context) {
	var req struct {
		Username    string `json:"username" binding:"required"`
		DisplayName string `json:"displayName" binding:"required"`
		Email       string `json:"email" binding:"required"`
		Password    string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	// 判断博客是否已经完成初始化
	blogInfo, err := h.configService.BlogInfo(c)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	if blogInfo == nil {
		response.FailAndResponse(c, "请先初始化博客")
		return
	}

	// 初始化管理员
	ret, err := h.userService.InitAdmin(c, &models.User{
		Username:    req.Username,
		Email:       req.Email,
		DisplayName: req.DisplayName,
		Password:    req.Password,
		Salt:        "",
	})

	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, ret)
}
