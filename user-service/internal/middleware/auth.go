package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ecommerce/user-service/internal/pkg/jwt"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix       = "Bearer "
	UserIDKey          = "user_id"
)

type AuthMiddleware struct {
	jwtMgr *jwt.Manager
}

func NewAuthMiddleware(jwtMgr *jwt.Manager) *AuthMiddleware {
	return &AuthMiddleware{jwtMgr: jwtMgr}
}

// RequireAuth 需要认证的请求拦截
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" || !strings.HasPrefix(authHeader, BearerPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    1001,
				"message": "missing or invalid authorization header",
			})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, BearerPrefix)
		claims, err := m.jwtMgr.ValidateToken(tokenStr)
		if err != nil {
			code := 1002
			msg := "invalid token"
			if err == jwt.ErrTokenExpired {
				code = 1003
				msg = "token expired"
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    code,
				"message": msg,
			})
			return
		}

		if claims.TokenType != "access" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    1004,
				"message": "invalid token type",
			})
			return
		}

		// 将 user_id 注入上下文
		c.Set(UserIDKey, claims.UserID)
		c.Next()
	}
}

// OptionalAuth 可选认证（不强制）
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" || !strings.HasPrefix(authHeader, BearerPrefix) {
			c.Next()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, BearerPrefix)
		claims, err := m.jwtMgr.ValidateToken(tokenStr)
		if err == nil && claims.TokenType == "access" {
			c.Set(UserIDKey, claims.UserID)
		}
		c.Next()
	}
}

// GetUserID 获取当前登录用户ID
func GetUserID(c *gin.Context) (uint64, bool) {
	val, exists := c.Get(UserIDKey)
	if !exists {
		return 0, false
	}
	userID, ok := val.(uint64)
	return userID, ok
}
