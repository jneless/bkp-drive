package middleware

import (
	"net/http"
	"strings"
	
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	
	"bkp-drive/internal/models"
)

var jwtSecret []byte

// InitJWT 初始化JWT密钥
func InitJWT(secret string) {
	jwtSecret = []byte(secret)
}

// Claims JWT载荷
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// AuthMiddleware JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Success: false,
				Error:   "缺少认证令牌",
			})
			c.Abort()
			return
		}
		
		// 检查Bearer前缀
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Success: false,
				Error:   "认证令牌格式错误",
			})
			c.Abort()
			return
		}
		
		// 解析token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Success: false,
				Error:   "认证令牌无效",
			})
			c.Abort()
			return
		}
		
		// 提取用户信息
		if claims, ok := token.Claims.(*Claims); ok {
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
		} else {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Success: false,
				Error:   "认证令牌解析失败",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}