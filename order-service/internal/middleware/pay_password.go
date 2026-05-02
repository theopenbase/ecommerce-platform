package middleware

import (
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
	payPwdKeyTpl     = "pay_pwd:%d" // 用户支付密码哈希 Redis key
)

// PayPasswordMiddleware 支付密码验证中间件
// 触发条件：X-Refund-Amount > 200 元
// 1. 从 Redis pay_pwd:{userID} 获取支付密码哈希
// 2. 未设置 → code 4001 "请先设置支付密码"
// 3. 已设置 → bcrypt 比对
//    - 匹配 → next()
//    - 不匹配 → code 4002 "支付密码错误"
// 4. 错误计数：5次错误锁30分钟
func PayPasswordMiddleware(cache *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		// 从 Redis 获取支付密码哈希
		pwdKey := fmt.Sprintf(payPwdKeyTpl, userID)
		pwdHash, err := cache.Get(c.Request.Context(), pwdKey).Result()
		if err == redis.Nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 4001, "message": "请先设置支付密码"})
			return
		} else if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"code": 5000, "message": "服务异常"})
			return
		}

		// bcrypt 比对
		if err := bcrypt.CompareHashAndPassword([]byte(pwdHash), []byte(pwd)); err != nil {
			wrongKey := fmt.Sprintf(wrongCountKeyTpl, userID)
			count, _ := cache.Incr(c.Request.Context(), wrongKey).Result()
			if count == 1 {
				cache.Expire(c.Request.Context(), wrongKey, lockoutSeconds)
			}
			if count >= maxWrongAttempts {
				cache.Set(c.Request.Context(), lockedKey, "1", lockoutSeconds)
				cache.Del(c.Request.Context(), wrongKey)
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 4104, "message": "连续5次错误，支付密码已锁定30分钟"})
				return
			}
			remaining := maxWrongAttempts - int(count)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    4002,
				"message": fmt.Sprintf("支付密码错误，剩余%d次机会", remaining),
			})
			return
		}

		// 验证成功，清除错误计数
		cache.Del(c.Request.Context(), fmt.Sprintf(wrongCountKeyTpl, userID))
		c.Next()
	}
}
