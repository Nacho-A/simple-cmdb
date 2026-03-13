package middleware

import (
	"net/url"

	"github.com/gin-gonic/gin"

	casbinx "cursor-cmdb-backend/casbin"
	"cursor-cmdb-backend/utils"
)

func CasbinAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exists := c.Get(CtxScope); exists {
			c.Next()
			return
		}

		if casbinx.Enforcer == nil {
			utils.Fail(c, 500, "权限引擎未初始化")
			c.Abort()
			return
		}

		v, ok := c.Get(CtxRoles)
		if !ok {
			utils.Fail(c, 403, "无权限")
			c.Abort()
			return
		}

		roles, _ := v.([]string)
		if len(roles) == 0 {
			utils.Fail(c, 403, "无权限")
			c.Abort()
			return
		}

		path := c.Request.URL.Path
		if p, err := url.PathUnescape(path); err == nil {
			path = p
		}
		method := c.Request.Method

		for _, role := range roles {
			allowed, err := casbinx.Enforcer.Enforce(role, path, method)
			if err == nil && allowed {
				c.Next()
				return
			}
		}

		utils.Fail(c, 403, "无权限")
		c.Abort()
	}
}
