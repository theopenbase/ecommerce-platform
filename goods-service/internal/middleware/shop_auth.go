package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	ShopIDKey = "shop_id"
)

// ShopAuth 商家身份认证（从 JWT 解析 shop_id）
// 生产环境中 shop_id 从 JWT Claims 中获取，此处做简化处理
func ShopAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		shopIDStr := c.GetHeader("X-Shop-ID")
		if shopIDStr == "" {
			// 从 Authorization header 解析（此处简化，生产用 JWT）
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    1001,
				"message": "missing shop authentication",
			})
			return
		}
		shopID, err := strconv.ParseUint(shopIDStr, 10, 64)
		if err != nil || shopID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    1002,
				"message": "invalid shop id",
			})
			return
		}
		c.Set(ShopIDKey, shopID)
		c.Next()
	}
}

// GetShopID 获取当前商家 ID
func GetShopID(c *gin.Context) (uint64, bool) {
	val, exists := c.Get(ShopIDKey)
	if !exists {
		return 0, false
	}
	id, ok := val.(uint64)
	return id, ok
}
