# API Key 鉴权实现计划

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现 API Key 鉴权机制，支持 Jenkins 等外部系统通过 `X-Api-Key` header 调用 CMDB API。

**Architecture:** 扩展现有 JWT 中间件，新增 API Key 认证中间件和 Scope 权限检查。认证流程：请求 → 检查 X-Api-Key → 有则走 API Key 认证 + Scope 检查，无则走 JWT + Casbin 检查。

**Tech Stack:** Go 1.22+, Gin, GORM v2, crypto/sha256

---

## File Structure

### 新增文件
| 文件 | 职责 |
|------|------|
| `backend/model/api_key.go` | API Key 数据模型，包含 GORM 结构体定义 |
| `backend/controller/api_key.go` | API Key CRUD 接口实现 |
| `backend/middleware/api_key.go` | X-Api-Key header 解析与验证 |
| `backend/middleware/scope.go` | 基于 scope 的 HTTP 方法权限检查 |

### 修改文件
| 文件 | 改动 |
|------|------|
| `backend/middleware/context_keys.go` | 新增 `CtxScope` 常量 |
| `backend/middleware/casbin.go` | 检测到 scope 时跳过 Casbin 检查 |
| `backend/router/router.go` | 注册 API Key 管理路由和服务 IP 查询路由 |
| `backend/bootstrap/bootstrap.go` | AutoMigrate 添加 APIKey 模型 |

---

## Chunk 1: 数据模型与基础设施

### Task 1: 新增 API Key 数据模型

**Files:**
- Create: `backend/model/api_key.go`

- [ ] **Step 1: 创建 API Key 模型文件**

```go
package model

import "time"

type APIKey struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"column:name;type:varchar(64);uniqueIndex;not null" json:"name"`
	KeyHash   string    `gorm:"column:key_hash;type:varchar(128);uniqueIndex;not null" json:"-"`
	Scope     string    `gorm:"column:scope;type:varchar(10);not null;default:'read'" json:"scope"`
	Status    int       `gorm:"column:status;type:tinyint;default:1;not null" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (APIKey) TableName() string { return "api_keys" }
```

- [ ] **Step 2: 更新 bootstrap.go 添加 AutoMigrate**

修改 `backend/bootstrap/bootstrap.go` 第 13 行：

```go
if err := db.AutoMigrate(&model.User{}, &model.Role{}, &model.Menu{}, &model.CMDBAsset{}, &model.APIKey{}); err != nil {
```

- [ ] **Step 3: 提交**

```bash
git add backend/model/api_key.go backend/bootstrap/bootstrap.go
git commit -m "feat: add APIKey model and auto migration"
```

---

### Task 2: 新增 context key 常量

**Files:**
- Modify: `backend/middleware/context_keys.go`

- [ ] **Step 1: 添加 CtxScope 常量**

修改 `backend/middleware/context_keys.go`：

```go
package middleware

const (
	CtxUserID   = "userID"
	CtxUsername = "username"
	CtxRoles    = "roles"
	CtxScope    = "scope"
)
```

- [ ] **Step 2: 提交**

```bash
git add backend/middleware/context_keys.go
git commit -m "feat: add CtxScope context key constant"
```

---

## Chunk 2: 认证中间件

### Task 3: 实现 API Key 认证中间件

**Files:**
- Create: `backend/middleware/api_key.go`

- [ ] **Step 1: 创建 API Key 认证中间件**

```go
package middleware

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/gin-gonic/gin"

	"cursor-cmdb-backend/model"
	"cursor-cmdb-backend/utils"
)

func APIKeyAuth() gin.HandlerFunc {
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
```

- [ ] **Step 2: 提交**

```bash
git add backend/middleware/api_key.go
git commit -m "feat: add API Key authentication middleware"
```

---

### Task 4: 实现 Scope 权限检查中间件

**Files:**
- Create: `backend/middleware/scope.go`

- [ ] **Step 1: 创建 Scope 权限检查中间件**

```go
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
```

- [ ] **Step 2: 提交**

```bash
git add backend/middleware/scope.go
git commit -m "feat: add Scope authorization middleware"
```

---

### Task 5: 修改 Casbin 中间件跳过 API Key 请求

**Files:**
- Modify: `backend/middleware/casbin.go`

- [ ] **Step 1: 在 CasbinAuth 开头添加 scope 检测**

修改 `backend/middleware/casbin.go`，在第 13 行 `if casbinx.Enforcer == nil` 之前添加：

```go
func CasbinAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// API Key 请求跳过 Casbin 检查
		if _, exists := c.Get(CtxScope); exists {
			c.Next()
			return
		}

		if casbinx.Enforcer == nil {
```

- [ ] **Step 2: 提交**

```bash
git add backend/middleware/casbin.go
git commit -m "feat: skip Casbin check for API Key requests"
```

---

## Chunk 3: 控制器与路由

### Task 6: 实现 API Key 管理控制器

**Files:**
- Create: `backend/controller/api_key.go`

- [ ] **Step 1: 创建 API Key 控制器**

```go
package controller

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cursor-cmdb-backend/middleware"
	"cursor-cmdb-backend/model"
	"cursor-cmdb-backend/utils"
)

type CreateAPIKeyReq struct {
	Name  string `json:"name" binding:"required"`
	Scope string `json:"scope" binding:"required,oneof=read write"`
}

type APIKeyResp struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Key       string `json:"key,omitempty"`
	Scope     string `json:"scope"`
	Status    int    `json:"status"`
	CreatedAt string `json:"created_at"`
}

func (h *Handler) APIKeyList(c *gin.Context) {
	var keys []model.APIKey
	if err := h.DB.Order("id desc").Find(&keys).Error; err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}

	list := make([]APIKeyResp, 0, len(keys))
	for _, k := range keys {
		list = append(list, APIKeyResp{
			ID:        k.ID,
			Name:      k.Name,
			Scope:     k.Scope,
			Status:    k.Status,
			CreatedAt: k.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	utils.OK(c, gin.H{"list": list, "total": len(list)})
}

func (h *Handler) APIKeyCreate(c *gin.Context) {
	var req CreateAPIKeyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}

	var exist model.APIKey
	if err := h.DB.Where("name = ?", req.Name).First(&exist).Error; err == nil {
		utils.Fail(c, 400, "名称已存在")
		return
	}

	rawKey := generateAPIKey()
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	apiKey := model.APIKey{
		Name:    req.Name,
		KeyHash: keyHash,
		Scope:   req.Scope,
		Status:  1,
	}

	if err := h.DB.Create(&apiKey).Error; err != nil {
		utils.Fail(c, 500, "创建失败")
		return
	}

	utils.OK(c, APIKeyResp{
		ID:        apiKey.ID,
		Name:      apiKey.Name,
		Key:       rawKey,
		Scope:     apiKey.Scope,
		Status:    apiKey.Status,
		CreatedAt: apiKey.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *Handler) APIKeyDelete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.Fail(c, 400, "参数错误")
		return
	}

	result := h.DB.Delete(&model.APIKey{}, id)
	if result.Error != nil {
		utils.Fail(c, 500, "删除失败")
		return
	}
	if result.RowsAffected == 0 {
		utils.Fail(c, 404, "不存在")
		return
	}

	utils.OK(c, nil)
}

func generateAPIKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("sk_live_%s", hex.EncodeToString(b))
}
```

- [ ] **Step 2: 提交**

```bash
git add backend/controller/api_key.go
git commit -m "feat: add API Key CRUD controller"
```

---

### Task 7: 实现服务 IP 查询接口

**Files:**
- Modify: `backend/controller/api_key.go` (添加新函数)

- [ ] **Step 1: 在 api_key.go 添加服务 IP 查询函数**

在 `backend/controller/api_key.go` 末尾添加：

```go
func (h *Handler) ServiceIPs(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		utils.Fail(c, 400, "参数错误")
		return
	}

	var assets []model.CMDBAsset
	if err := h.DB.Where("service_name = ?", serviceName).Find(&assets).Error; err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}

	ips := make([]string, 0)
	for _, a := range assets {
		if a.PrivateIP != "" {
			ips = append(ips, a.PrivateIP)
		}
	}

	utils.OK(c, gin.H{
		"service": serviceName,
		"ips":     ips,
	})
}
```

需要添加 import `"cursor-cmdb-backend/model"` (已存在)

- [ ] **Step 2: 提交**

```bash
git add backend/controller/api_key.go
git commit -m "feat: add service IPs query endpoint"
```

---

### Task 8: 注册路由

**Files:**
- Modify: `backend/router/router.go`

- [ ] **Step 1: 添加中间件到路由组**

修改 `backend/router/router.go`，在第 32-34 行之间：

```go
	authed := v1.Group("")
	authed.Use(middleware.APIKeyAuth())
	authed.Use(middleware.JWTAuth(cfg.JWT.Secret, cfg.JWT.Issuer, cfg.JWT.Audience))
	authed.Use(middleware.ScopeAuth())
	authed.Use(middleware.CasbinAuth())
```

- [ ] **Step 2: 添加 API Key 管理路由**

在 `authed.GET("/me", h.Me)` 之后添加：

```go
		authed.GET("/me", h.Me)

		// API Key 管理（admin）
		authed.GET("/api-keys", h.APIKeyList)
		authed.POST("/api-keys", h.APIKeyCreate)
		authed.DELETE("/api-keys/:id", h.APIKeyDelete)

		// 用户管理（admin）
```

- [ ] **Step 3: 添加服务 IP 查询路由**

在 CMDB 资产路由区域添加：

```go
		// CMDB资产
		authed.GET("/cmdb/assets", h.AssetList)
		authed.GET("/cmdb/assets/:id", h.AssetGet)
		authed.POST("/cmdb/assets", h.AssetCreate)
		authed.PUT("/cmdb/assets/:id", h.AssetUpdate)
		authed.DELETE("/cmdb/assets/:id", h.AssetDelete)
		authed.POST("/cmdb/assets/batch-delete", h.AssetBatchDelete)
		authed.GET("/cmdb/assets/export", h.AssetExportExcel)

		// 服务 IP 查询
		authed.GET("/cmdb/services/:name/ips", h.ServiceIPs)

		// 公共接口
```

- [ ] **Step 4: 提交**

```bash
git add backend/router/router.go
git commit -m "feat: register API Key and service IPs routes"
```

---

### Task 9: 添加 Casbin 策略

**Files:**
- Modify: `backend/bootstrap/bootstrap.go`

- [ ] **Step 1: 在 SeedCasbinPolicies 添加 API Key 管理策略**

修改 `backend/bootstrap/bootstrap.go`，在 `SeedCasbinPolicies` 函数中，第 200-221 行之间添加：

```go
	existing := casbinx.Enforcer.GetPolicy()
	if len(existing) == 0 {
		_, _ = casbinx.Enforcer.AddPolicy("admin", "/api/v1/*", "*")

		// API Key 管理
		_, _ = casbinx.Enforcer.AddPolicy("admin", "/api/v1/api-keys", "GET")
		_, _ = casbinx.Enforcer.AddPolicy("admin", "/api/v1/api-keys", "POST")
		_, _ = casbinx.Enforcer.AddPolicy("admin", "/api/v1/api-keys/*", "DELETE")

		_, _ = casbinx.Enforcer.AddPolicy("operator", "/api/v1/me", "GET")
```

- [ ] **Step 2: 提交**

```bash
git add backend/bootstrap/bootstrap.go
git commit -m "feat: add Casbin policies for API Key management"
```

---

## Chunk 4: 测试与验证

### Task 10: 编译验证

- [ ] **Step 1: 编译后端**

```bash
cd backend && go build -o cmdb.exe .
```

预期：编译成功无错误

- [ ] **Step 2: 启动服务验证**

```bash
cd backend && ./cmdb.exe
```

预期：服务启动成功，数据库表自动创建

---

### Task 11: 功能测试

- [ ] **Step 1: 登录获取 JWT Token**

```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

预期：返回 token

- [ ] **Step 2: 创建 API Key**

```bash
curl -X POST http://localhost:8080/api/v1/api-keys \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"test-key","scope":"read"}'
```

预期：返回 API Key 原文

- [ ] **Step 3: 使用 API Key 查询服务 IP**

```bash
curl -H "X-Api-Key: <api_key>" \
  http://localhost:8080/api/v1/cmdb/services/test-service/ips
```

预期：返回服务 IP 列表

- [ ] **Step 4: 验证权限控制**

使用 read scope 的 API Key 尝试 POST 请求：

```bash
curl -X POST -H "X-Api-Key: <read_scope_key>" \
  http://localhost:8080/api/v1/api-keys \
  -H "Content-Type: application/json" \
  -d '{"name":"fail","scope":"read"}'
```

预期：返回 403 权限不足

---

### Task 12: 最终提交

- [ ] **Step 1: 确认所有更改已提交**

```bash
git status
```

预期：working tree clean

- [ ] **Step 2: 查看提交历史**

```bash
git log --oneline -10
```

预期：看到所有功能提交

---

## Summary

| 任务 | 文件 | 状态 |
|------|------|------|
| Task 1 | `model/api_key.go`, `bootstrap/bootstrap.go` | 数据模型 |
| Task 2 | `middleware/context_keys.go` | Context 常量 |
| Task 3 | `middleware/api_key.go` | API Key 认证 |
| Task 4 | `middleware/scope.go` | Scope 权限检查 |
| Task 5 | `middleware/casbin.go` | Casbin 跳过逻辑 |
| Task 6 | `controller/api_key.go` | CRUD 接口 |
| Task 7 | `controller/api_key.go` | 服务 IP 查询 |
| Task 8 | `router/router.go` | 路由注册 |
| Task 9 | `bootstrap/bootstrap.go` | Casbin 策略 |
| Task 10 | - | 编译验证 |
| Task 11 | - | 功能测试 |
| Task 12 | - | 最终确认 |