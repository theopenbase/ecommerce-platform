package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
)

const (
	maxWrongAttempts = 5
	lockoutSeconds  = 30 * 60 // 30分钟
	wrongCountKeyTpl = "paypwd:wrong:%d"
	lockedKeyTpl     = "paypwd:locked:%d"
)

// PayPasswordMiddleware 创建支付密码验证中间件
// 触发条件：退款金额 > 200 元
func PayPasswordMiddleware(cache *redis.Client, getUserByID func(ctx context.Context, userID uint64) (string, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 header 获取退款金额
		amountStr := c.GetHeader("X-Refund-Amount")
		if amountStr == "" {
			c.Next()
			return
		}
		var amount float64
		fmt.Sscanf(amountStr, "%f", &amount)
		if amount <= 200 {
			c.Next()
			return
		}

		// 获取 userID（从 JWT 或 header）
		userIDStr := c.GetHeader("X-User-ID")
		var userID uint64
		fmt.Sscanf(userIDStr, "%d", &userID)
		if userID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 4101, "message": "未登录"})
			return
		}

		pwd := c.GetHeader("X-Pay-Password")
		if pwd == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 4101, "message": "支付密码不能为空（退款金额>200）"})
			return
		}

		// 检查是否被锁定
		lockedKey := fmt.Sprintf(lockedKeyTpl, userID)
		if locked, _ := cache.Exists(c.Request.Context(), lockedKey).Result(); locked > 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 4102, "message": "支付密码已锁定，请30分钟后重试"})
			return
		}

		// 获取用户支付密码哈希
		pwdHash, err := getUserByID(c.Request.Context(), userID)
		if err != nil || pwdHash == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 4103, "message": "未设置支付密码"})
			return
		}

		// bcrypt 比对
		if err := bcrypt.CompareHashAndPassword([]byte(pwdHash), []byte(pwd)); err != nil {
			wrongKey := fmt.Sprintf(wrongCountKeyTpl, userID)
			count, _ := cache.Incr(c.Request.Context(), wrongKey).Result()
			if count == 1 {
				cache.Expire(c.Request.Context(), wrongKey, lockoutSeconds)
			}
			remaining := maxWrongAttempts - int(count)
			if count >= maxWrongAttempts {
				cache.Set(c.Request.Context(), lockedKey, "1", lockoutSeconds)
				cache.Del(c.Request.Context(), wrongKey)
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 4104, "message": "连续5次错误，支付密码已锁定30分钟"})
				return
			}
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    4105,
				"message": fmt.Sprintf("支付密码错误，剩余%d次机会", remaining),
			})
			return
		}

		// 验证成功，清除错误计数
		cache.Del(c.Request.Context(), fmt.Sprintf(wrongCountKeyTpl, userID))
		c.Next()
	}
}
