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
		// 获取博客信息
		privateGroup.GET("/blog", h.getBlogInfo)
		// 修改博客信息
		privateGroup.PUT("/blog", h.updateBlogInfo)

		// 修改备案信息
		privateGroup.PUT("/icp", h.updateIcp)
		// 获取备案信息
		privateGroup.GET("/icp", h.getIcp)
	}

	// 无鉴权接口
	publicGroup := r.Group("/config")
	{
		// 初始化博客信息
		publicGroup.POST("/blog", h.initBlogInfo)
		// 初始化管理员（博主）
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

// getBlogInfo 获取博客信息
func (h *ConfigAdminHandler) getBlogInfo(c *gin.Context) {
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

// updateBlogInfo 修改博客信息
func (h *ConfigAdminHandler) updateBlogInfo(c *gin.Context) {

	var req struct {
		Title    string  `json:"title" binding:"required"`
		Subtitle *string `json:"subtitle"`
		Logo     *string `json:"logo"`
		Favicon  *string `json:"favicon"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	blogInfo, err := h.configService.BlogInfo(c)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	if blogInfo == nil {
		response.FailAndResponse(c, "请先初始化博客")
		return
	}

	blogInfo.Title = util.StringPtr(req.Title)
	blogInfo.Subtitle = req.Subtitle
	blogInfo.Logo = req.Logo
	blogInfo.Favicon = req.Favicon

	ret, err := h.configService.SetBlogInfo(c, blogInfo)

	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, ret)
}

// updateIcp 修改备案信息
func (h *ConfigAdminHandler) updateIcp(c *gin.Context) {
	var req struct {
		Icp    *string `json:"icp"`
		Police *string `json:"police"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.configService.SetICPFiling(c, &models.ICPFiling{
		ICP:    req.Icp,
		Police: req.Police,
	})

	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, ret)
}

// getIcp 获取备案信息
func (h *ConfigAdminHandler) getIcp(c *gin.Context) {
	icp, err := h.configService.ICPFiling(c)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, icp)
}
