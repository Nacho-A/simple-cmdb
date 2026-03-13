# 日志系统完善设计文档

## 概述

完善后端日志系统，实现全局请求日志、接口错误日志、认证/权限错误日志。

## 日志级别规则

| 级别 | 场景 | 示例 |
|------|------|------|
| INFO | 正常请求、业务操作成功 | 请求完成、数据查询成功 |
| WARN | 认证失败、权限不足 | JWT 过期、API Key 无效、Casbin 拒绝 |
| ERROR | 内部错误、数据库错误 | 数据库连接失败、查询失败 |

## 日志格式

```
INFO    request        {"method": "GET", "path": "/api/v1/cmdb/services", "status": 200, "latency": "12.5ms"}
WARN    auth           {"method": "POST", "path": "/api/v1/assets", "status": 401, "error": "JWT 已过期"}
WARN    permission     {"method": "DELETE", "path": "/api/v1/users/1", "status": 403, "error": "无权限"}
ERROR   database       {"method": "POST", "path": "/api/v1/assets", "error": "创建失败: connection refused"}
```

## 文件变更

### 新增文件

| 文件 | 说明 |
|------|------|
| `middleware/logger.go` | 全局请求日志中间件 |

### 修改文件

| 文件 | 改动 |
|------|------|
| `middleware/jwt.go` | 添加认证失败 WARN 日志 |
| `middleware/api_key.go` | 添加 API Key 无效 WARN 日志 |
| `middleware/casbin.go` | 添加权限不足 WARN 日志 |
| `middleware/scope.go` | 添加 Scope 权限不足 WARN 日志 |
| `router/router.go` | 注册 Logger 中间件 |
| 各 controller 文件 | 数据库错误添加 ERROR 日志 |

## 实现细节

### 1. 全局请求日志中间件

```go
func Logger(log *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        method := c.Request.Method

        c.Next()

        latency := time.Since(start)
        status := c.Writer.Status()

        if status >= 500 {
            log.Error("request",
                zap.String("method", method),
                zap.String("path", path),
                zap.Int("status", status),
                zap.Duration("latency", latency),
            )
        } else if status >= 400 {
            log.Warn("request",
                zap.String("method", method),
                zap.String("path", path),
                zap.Int("status", status),
                zap.Duration("latency", latency),
            )
        } else {
            log.Info("request",
                zap.String("method", method),
                zap.String("path", path),
                zap.Int("status", status),
                zap.Duration("latency", latency),
            )
        }
    }
}
```

### 2. 认证/权限日志

**JWT 认证失败：**
- 无 Authorization header → WARN "未登录"
- Token 解析失败 → WARN "登录已过期"

**API Key 认证失败：**
- Key 无效/已禁用 → WARN "无效的 API Key"

**权限不足：**
- Casbin 拒绝 → WARN "无权限"
- Scope 限制 → WARN "权限不足: read scope 不允许 POST"

### 3. 接口错误日志

数据库操作失败时打印 ERROR 日志：
```go
if err := h.DB.Create(&asset).Error; err != nil {
    h.Log.Error("database", zap.String("operation", "create asset"), zap.Error(err))
    utils.Fail(c, 500, "创建失败")
    return
}
```