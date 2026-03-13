package middleware

import (
	"net/url"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	casbinx "cursor-cmdb-backend/casbin"
	"cursor-cmdb-backend/utils"
)

func CasbinAuth(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exists := c.Get(CtxScope); exists {
			c.Next()
			return
		}

		if casbinx.Enforcer == nil {
			log.Error("permission", zap.String("error", "权限引擎未初始化"))
			utils.Fail(c, 500, "权限引擎未初始化")
			c.Abort()
			return
		}

		v, ok := c.Get(CtxRoles)
		if !ok {
			log.Warn("permission", zap.String("path", c.Request.URL.Path), zap.String("error", "无角色信息"))
			utils.Fail(c, 403, "无权限")
			c.Abort()
			return
		}

		roles, _ := v.([]string)
		if len(roles) == 0 {
			log.Warn("permission", zap.String("path", c.Request.URL.Path), zap.String("error", "角色为空"))
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

		log.Warn("permission",
			zap.String("path", path),
			zap.String("method", method),
			zap.Strings("roles", roles),
			zap.String("error", "Casbin 拒绝访问"),
		)
		utils.Fail(c, 403, "无权限")
		c.Abort()
	}
}
