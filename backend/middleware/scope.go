package middleware

import (
	"github.com/gin-gonic/gin"

	"cursor-cmdb-backend/utils"
)

func ScopeAuth() gin.HandlerFunc {
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
			utils.Fail(c, 403, "权限不足")
			c.Abort()
			return
		}

		c.Next()
	}
}
