package middleware

import (
	"log"
	"net/http"
	"nola-go/internal/models/response"
	"nola-go/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 管理员身份验证中间件
func AuthMiddleware(tokenSvc *service.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Unauthorized())
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")

		// 解析并验证
		claims, err := tokenSvc.ParseAndValidate(tokenStr)
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Unauthorized())
			return
		}

		// 从 claims 中获取 userId
		var userId uint
		switch v := claims["user_id"].(type) {
		case float64:
			userId = uint(v)
		case int64:
			userId = uint(v)
		case uint64:
			userId = uint(v)
		case uint:
			userId = v
		default:
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Unauthorized())
			return
		}

		// 检查 token 是否和 userId 对应
		if !tokenSvc.Verify(userId, tokenStr) {
			// 验证失败
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Unauthorized())
			return
		}

		// 将用户信息放到上下文，供后续 handler 使用
		c.Set("uid", userId)
		if uname, ok := claims["username"].(string); ok {
			c.Set("username", uname)
		}

		// 继续处理请求
		c.Next()
	}
}
