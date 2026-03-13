package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"cursor-cmdb-backend/utils"
)

func ScopeAuth(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		scopeVal, exists := c.Get(CtxScope)
		if !exists {
			c.Next()
			return
		}

		scope, ok := scopeVal.(string)
		if !ok {
			c.Next()
			return
		}

		method := c.Request.Method
		allowed := false

		switch scope {
		case "write":
			allowed = true
		case "read":
			allowed = method == "GET"
		}

		if !allowed {
			log.Warn("permission",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", method),
				zap.String("scope", scope),
				zap.String("error", "权限不足"),
			)
			utils.Fail(c, 403, "权限不足")
			c.Abort()
			return
		}

		c.Next()
	}
}
