package middleware

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"cursor-cmdb-backend/model"
	"cursor-cmdb-backend/utils"
)

func APIKeyAuth(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-Api-Key")
		if key == "" {
			c.Next()
			return
		}

		hash := sha256.Sum256([]byte(key))
		keyHash := hex.EncodeToString(hash[:])

		var apiKey model.APIKey
		if err := model.DB.Where("key_hash = ? AND status = 1", keyHash).First(&apiKey).Error; err != nil {
			log.Warn("auth", zap.String("path", c.Request.URL.Path), zap.String("error", "无效的 API Key"))
			utils.Fail(c, 401, "无效的 API Key")
			c.Abort()
			return
		}

		c.Set(CtxUserID, uint(0))
		c.Set(CtxUsername, "api_key:"+apiKey.Name)
		c.Set(CtxRoles, []string{})
		c.Set(CtxScope, apiKey.Scope)
		c.Next()
	}
}
