package router

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"cursor-cmdb-backend/config"
	"cursor-cmdb-backend/controller"
	"cursor-cmdb-backend/middleware"
)

func New(cfg *config.Config, db *gorm.DB, log *zap.Logger) *gin.Engine {
	if cfg.Server.Mode != "" {
		gin.SetMode(cfg.Server.Mode)
	}

	r := gin.New()
	r.Use(cors.New(middleware.CORS()))
	r.Use(middleware.Recovery(log))
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	h := &controller.Handler{Cfg: cfg, DB: db, Log: log}

	v1 := r.Group("/api/v1")
	{
		v1.POST("/login", h.Login)

		authed := v1.Group("")
		authed.Use(middleware.JWTAuth(cfg.JWT.Secret, cfg.JWT.Issuer, cfg.JWT.Audience))
		authed.Use(middleware.CasbinAuth())

		authed.GET("/me", h.Me)

		// 用户管理（admin）
		authed.GET("/users", h.UserList)
		authed.POST("/users", h.UserCreate)
		authed.PUT("/users/:id", h.UserUpdate)
		authed.DELETE("/users/:id", h.UserDelete)
		authed.PUT("/users/:id/roles", h.UserBindRoles)

		// 角色管理（admin）
		authed.GET("/roles", h.RoleList)
		authed.POST("/roles", h.RoleCreate)
		authed.PUT("/roles/:id", h.RoleUpdate)
		authed.DELETE("/roles/:id", h.RoleDelete)
		authed.GET("/roles/:id/menus", h.RoleGetMenus)
		authed.POST("/roles/:id/menus", h.RoleSaveMenus)

		// 菜单管理（admin）
		authed.GET("/menus", h.MenuList)
		authed.POST("/menus", h.MenuCreate)
		authed.PUT("/menus/:id", h.MenuUpdate)
		authed.DELETE("/menus/:id", h.MenuDelete)

		// CMDB资产
		authed.GET("/cmdb/assets", h.AssetList)
		authed.GET("/cmdb/assets/:id", h.AssetGet)
		authed.POST("/cmdb/assets", h.AssetCreate)
		authed.PUT("/cmdb/assets/:id", h.AssetUpdate)
		authed.DELETE("/cmdb/assets/:id", h.AssetDelete)
		authed.POST("/cmdb/assets/batch-delete", h.AssetBatchDelete)
		authed.GET("/cmdb/assets/export", h.AssetExportExcel)

		// 公共接口
		authed.GET("/cmdb/cloud-providers", h.CloudProviders)
	}

	return r
}

