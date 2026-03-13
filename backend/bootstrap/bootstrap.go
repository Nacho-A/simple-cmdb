package bootstrap

import (
	"fmt"

	casbinx "cursor-cmdb-backend/casbin"
	"cursor-cmdb-backend/model"
	"cursor-cmdb-backend/utils"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&model.User{}, &model.Role{}, &model.Menu{}, &model.CMDBAsset{}, &model.APIKey{}); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}
	return nil
}

func SeedDefaults(db *gorm.DB) error {
	// 1) roles
	roles := []model.Role{
		{Name: "admin", Description: "系统管理员"},
		{Name: "operator", Description: "运维/操作员"},
		{Name: "viewer", Description: "只读访客"},
	}
	for _, r := range roles {
		var cnt int64
		if err := db.Model(&model.Role{}).Where("name = ?", r.Name).Count(&cnt).Error; err != nil {
			return err
		}
		if cnt == 0 {
			if err := db.Create(&r).Error; err != nil {
				return err
			}
		}
	}

	var adminRole model.Role
	if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		return err
	}
	var operatorRole model.Role
	if err := db.Where("name = ?", "operator").First(&operatorRole).Error; err != nil {
		return err
	}
	var viewerRole model.Role
	if err := db.Where("name = ?", "viewer").First(&viewerRole).Error; err != nil {
		return err
	}

	// 2) admin user
	var adminUser model.User
	err := db.Preload("Roles").Where("username = ?", "admin").First(&adminUser).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if err == gorm.ErrRecordNotFound {
		hash, _ := utils.HashPassword("admin123")
		adminUser = model.User{
			Username: "admin",
			Password: hash,
			Nickname: "管理员",
			Email:    "admin@example.com",
			Status:   1,
		}
		if err := db.Create(&adminUser).Error; err != nil {
			return err
		}
	}

	if err := db.Model(&adminUser).Association("Roles").Replace(&adminRole); err != nil {
		return err
	}

	// 3) menus (two-level)
	// 一级：仪表盘、CMDB、系统管理
	var menuDashboard model.Menu
	_ = db.Where("path = ?", "/dashboard").First(&menuDashboard).Error
	if menuDashboard.ID == 0 {
		menuDashboard = model.Menu{
			Name:      "仪表盘",
			Path:      "/dashboard",
			Component: "views/dashboard/index.vue",
			Icon:      "Odometer",
			ParentID:  nil,
			Order:     1,
			Hidden:    false,
		}
		if err := db.Create(&menuDashboard).Error; err != nil {
			return err
		}
	}

	var menuCMDB model.Menu
	_ = db.Where("path = ?", "/cmdb").First(&menuCMDB).Error
	if menuCMDB.ID == 0 {
		menuCMDB = model.Menu{
			Name:      "资产管理",
			Path:      "/cmdb",
			Component: "",
			Icon:      "Box",
			ParentID:  nil,
			Order:     2,
			Hidden:    false,
		}
		if err := db.Create(&menuCMDB).Error; err != nil {
			return err
		}
	}

	var menuAssets model.Menu
	_ = db.Where("path = ?", "/cmdb/assets").First(&menuAssets).Error
	if menuAssets.ID == 0 {
		menuAssets = model.Menu{
			Name:      "CMDB资产",
			Path:      "/cmdb/assets",
			Component: "views/cmdb/assets/index.vue",
			Icon:      "Monitor",
			ParentID:  &menuCMDB.ID,
			Order:     1,
			Hidden:    false,
		}
		if err := db.Create(&menuAssets).Error; err != nil {
			return err
		}
	}

	var menuSystem model.Menu
	_ = db.Where("path = ?", "/system").First(&menuSystem).Error
	if menuSystem.ID == 0 {
		menuSystem = model.Menu{
			Name:      "系统管理",
			Path:      "/system",
			Component: "",
			Icon:      "Setting",
			ParentID:  nil,
			Order:     9,
			Hidden:    false,
		}
		if err := db.Create(&menuSystem).Error; err != nil {
			return err
		}
	}

	ensureChild := func(path, name, component, icon string, order int) (*model.Menu, error) {
		var m model.Menu
		_ = db.Where("path = ?", path).First(&m).Error
		if m.ID == 0 {
			m = model.Menu{
				Name:      name,
				Path:      path,
				Component: component,
				Icon:      icon,
				ParentID:  &menuSystem.ID,
				Order:     order,
				Hidden:    false,
			}
			if err := db.Create(&m).Error; err != nil {
				return nil, err
			}
		}
		return &m, nil
	}

	menuUser, err := ensureChild("/system/user", "用户管理", "views/system/user/index.vue", "User", 1)
	if err != nil {
		return err
	}
	menuRole, err := ensureChild("/system/role", "角色管理", "views/system/role/index.vue", "Avatar", 2)
	if err != nil {
		return err
	}
	menuMenu, err := ensureChild("/system/menu", "菜单管理", "views/system/menu/index.vue", "Menu", 3)
	if err != nil {
		return err
	}

	// 4) role_menus
	// admin: all
	if err := db.Model(&adminRole).Association("Menus").Replace(&menuDashboard, &menuCMDB, &menuAssets, &menuSystem, menuUser, menuRole, menuMenu); err != nil {
		return err
	}
	// operator/viewer: dashboard + assets
	if err := db.Model(&operatorRole).Association("Menus").Replace(&menuDashboard, &menuCMDB, &menuAssets); err != nil {
		return err
	}
	if err := db.Model(&viewerRole).Association("Menus").Replace(&menuDashboard, &menuCMDB, &menuAssets); err != nil {
		return err
	}

	return nil
}

func SeedCasbinPolicies() error {
	if casbinx.Enforcer == nil {
		return fmt.Errorf("casbin enforcer nil")
	}

	// 幂等：如果已有 policy 则不重复写入
	existing := casbinx.Enforcer.GetPolicy()
	if len(existing) == 0 {
		_, _ = casbinx.Enforcer.AddPolicy("admin", "/api/v1/*", "*")

		_, _ = casbinx.Enforcer.AddPolicy("operator", "/api/v1/me", "GET")
		_, _ = casbinx.Enforcer.AddPolicy("viewer", "/api/v1/me", "GET")

		_, _ = casbinx.Enforcer.AddPolicy("operator", "/api/v1/cmdb/cloud-providers", "GET")
		_, _ = casbinx.Enforcer.AddPolicy("viewer", "/api/v1/cmdb/cloud-providers", "GET")

		// 资产：覆盖列表/详情/导出/批量等
		_, _ = casbinx.Enforcer.AddPolicy("operator", "/api/v1/cmdb/assets*", "GET")
		_, _ = casbinx.Enforcer.AddPolicy("operator", "/api/v1/cmdb/assets*", "POST")
		_, _ = casbinx.Enforcer.AddPolicy("operator", "/api/v1/cmdb/assets*", "PUT")
		_, _ = casbinx.Enforcer.AddPolicy("operator", "/api/v1/cmdb/assets*", "DELETE")

		_, _ = casbinx.Enforcer.AddPolicy("viewer", "/api/v1/cmdb/assets*", "GET")

		if err := casbinx.Enforcer.SavePolicy(); err != nil {
			return err
		}
	}
	return nil
}
