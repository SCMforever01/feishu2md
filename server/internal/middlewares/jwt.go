package middlewares

import (
	"feishu2md/server/internal/model"
	"feishu2md/server/pkg/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

// JWTMiddleware 验证 JWT 的中间件
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// 1. 从 header 中查找 "token" 字段
		tokenString = c.GetHeader("token")

		// 2. 如果没有从 header 查找，再从 Authorization header 查找
		if tokenString == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				// Authorization: Bearer <token>
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenString = parts[1]
				}
			}
		}

		// 3. 如果还没有，从 URL 路由参数查找是否拼接了 token
		if tokenString == "" {
			tokenString = c.Param("jwtToken")
		}

		// 如果 Token 仍然为空，返回错误
		if tokenString == "" {
			model.Error(c, 401, "未授权")
			c.Abort()
			return
		}

		// 验证 Token
		claims, err := utils.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid Token", "error": err.Error()})
			c.Abort()
			return
		}

		// 将用户 ID 存储到上下文中，后续处理可以直接使用
		c.Set("userID", claims.UserID)

		// 继续处理请求
		c.Next()
	}
}
