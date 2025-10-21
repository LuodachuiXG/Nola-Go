package router

import (
	"nola-go/internal/handler/admin"
	"nola-go/internal/handler/api"
	"nola-go/internal/service"

	"github.com/gin-gonic/gin"
)

type Deps struct {
	TokenService    *service.TokenService
	UserService     *service.UserService
	PostService     *service.PostService
	ConfigService   *service.ConfigService
	TagService      *service.TagService
	CategoryService *service.CategoryService
}

// SetupRouters 初始化 Gin 路由
func SetupRouters(r *gin.Engine, deps *Deps) *gin.Engine {

	// 静态资源

	// 后台接口（需要登录，登录拦截中间件在 Handler 内部细化设置）
	adminHandler := r.Group("/admin")
	{
		// 用户接口
		userHandler := admin.NewUserAdminHandler(deps.UserService, deps.TokenService)
		userHandler.RegisterAdmin(adminHandler)

		// 配置接口
		configHandler := admin.NewConfigAdminHandler(deps.ConfigService, deps.UserService, deps.TokenService)
		configHandler.RegisterAdmin(adminHandler)

		// 标签接口
		tagHandler := admin.NewTagAdminHandler(deps.TagService, deps.TokenService)
		tagHandler.RegisterAdmin(adminHandler)

		// 分类接口
		categoryHandler := admin.NewCategoryAdminHandler(deps.CategoryService, deps.TokenService)
		categoryHandler.RegisterAdmin(adminHandler)

	}

	// 博客接口（无需登录）
	apiHandler := r.Group("/api")
	{
		// 用户接口
		userHandler := api.NewUserApiHandler(deps.UserService)
		userHandler.RegisterApi(apiHandler)

		// 配置接口
		configHandler := api.NewConfigApiHandler(deps.ConfigService, deps.UserService)
		configHandler.RegisterApi(apiHandler)

		// 标签接口
		tagHandler := api.NewTagApiHandler(deps.TagService)
		tagHandler.RegisterApi(apiHandler)

		// 分类接口
		categoryHandler := api.NewCategoryApiHandler(deps.CategoryService)
		categoryHandler.RegisterApi(apiHandler)
	}

	return r
}
