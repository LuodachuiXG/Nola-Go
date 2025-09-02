package router

import (
	"nola-go/internal/handler/admin"
	"nola-go/internal/handler/api"
	"nola-go/internal/service"

	"github.com/gin-gonic/gin"
)

type Deps struct {
	TokenService *service.TokenService
	UserService  *service.UserService
	PostService  *service.PostService
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
	}

	// 博客接口（无需登录）
	apiHandler := r.Group("/api")
	{
		// 用户接口
		userHandler := api.NewUserApiHandler(deps.UserService)
		userHandler.RegisterApi(apiHandler)
	}

	return r
}
