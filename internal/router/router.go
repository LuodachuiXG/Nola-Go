package router

import (
	"nola-go/internal/handler"
	"nola-go/internal/middleware"
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

	// 后台接口（需要登录）
	admin := r.Group("/admin")
	admin.Use(middleware.AuthMiddleware(deps.TokenService))
	{
		// 用户接口
		userHandler := handler.NewUserAdminHandler(deps.UserService)
		userHandler.RegisterAdmin(admin)
	}

	// 博客接口（无需登录）
	api := r.Group("/api")
	{
		// 用户接口
		userHandler := handler.NewUserApiHandler(deps.UserService)
		userHandler.RegisterApi(api)
	}

	return r
}
