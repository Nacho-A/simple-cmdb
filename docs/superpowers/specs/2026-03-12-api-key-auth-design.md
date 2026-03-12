# API Key 鉴权设计文档

## 概述

为支持 Jenkins 脚本等外部系统调用 CMDB API，新增 API Key 鉴权机制，与现有 JWT 鉴权并存。

## 目标

- 支持通过 `X-Api-Key` header 认证
- API Key 支持 read/write 两种权限范围
- 提供 API Key 管理接口（CRUD）
- 新增按服务名查询 IP 的接口

## 数据模型

### api_keys 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint unsigned | 主键，自增 |
| name | varchar(64) | Key 名称，用于识别用途 |
| key_hash | varchar(128) | API Key 的 SHA256 哈希值 |
| scope | enum('read','write') | 权限范围 |
| status | tinyint | 1=启用，0=禁用 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

**索引：**
- `key_hash` 唯一索引（加速认证查询）
- `name` 普通索引（管理查询）

**安全设计：**
- API Key 原文仅在创建时返回一次
- 数据库存储 SHA256 哈希值
- Key 格式：`sk_live_<32位随机字符串>`

## 认证流程

```
请求进入
    │
    ▼
检查 X-Api-Key header
    │
    ├── 有 X-Api-Key ──────────────────────────────┐
    │   1. 计算 key_hash = sha256(key)              │
    │   2. 查询 api_keys 表 WHERE key_hash = ?      │
    │   3. 验证记录存在且 status = 1                │
    │   4. 设置 context:                            │
    │      - CtxUserID = 0                          │
    │      - CtxUsername = "api_key:<name>"         │
    │      - CtxRoles = []                          │
    │      - CtxScope = "read" 或 "write"           │
    │   5. Scope 权限检查                           │
    └───────────────────────────────────────────────┘
    │
    └── 无 X-Api-Key ──────────────────────────────┐
        走原有 JWT 流程                             │
        → Casbin 权限检查                           │
        └───────────────────────────────────────────┘
```

### Scope 权限规则

| Scope | 允许的 HTTP 方法 |
|-------|------------------|
| read | GET |
| write | GET, POST, PUT, DELETE |

## 文件变更

### 新增文件

| 文件 | 说明 |
|------|------|
| `model/api_key.go` | API Key 数据模型 |
| `controller/api_key.go` | API Key 管理接口 |
| `middleware/api_key.go` | API Key 认证中间件 |
| `middleware/scope.go` | Scope 权限检查中间件 |

### 修改文件

| 文件 | 改动 |
|------|------|
| `middleware/context_keys.go` | 新增 `CtxScope` 常量 |
| `middleware/casbin.go` | 检测到 scope 时跳过 Casbin |
| `router/router.go` | 注册新路由 |
| `bootstrap/bootstrap.go` | 初始化默认 API Key 表结构 |

## API 接口

### API Key 管理（需 JWT + admin 角色）

#### 创建 API Key

```
POST /api/v1/api-keys
Authorization: Bearer <jwt_token>

Request:
{
  "name": "jenkins-ci",
  "scope": "read"
}

Response:
{
  "code": 200,
  "data": {
    "id": 1,
    "name": "jenkins-ci",
    "key": "sk_live_abc123def456...",  // 仅创建时返回
    "scope": "read",
    "status": 1,
    "created_at": "2026-03-12T10:00:00Z"
  }
}
```

#### 查询 API Key 列表

```
GET /api/v1/api-keys
Authorization: Bearer <jwt_token>

Response:
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 1,
        "name": "jenkins-ci",
        "scope": "read",
        "status": 1,
        "created_at": "2026-03-12T10:00:00Z"
      }
    ],
    "total": 1
  }
}
```

#### 删除 API Key

```
DELETE /api/v1/api-keys/:id
Authorization: Bearer <jwt_token>

Response:
{
  "code": 200,
  "message": "success"
}
```

### 服务 IP 查询

```
GET /api/v1/cmdb/services/:name/ips
X-Api-Key: <api_key>  或  Authorization: Bearer <jwt_token>

Response:
{
  "code": 200,
  "data": {
    "service": "web-api",
    "ips": ["192.168.1.10", "192.168.1.11"]
  }
}
```

**实现逻辑：**
- 查询 `cmdb_assets` 表中 `service_name = :name` 的记录
- 返回 `private_ip` 字段组成的数组
- 过滤掉空 IP

## 使用示例

### Jenkins 脚本调用

```bash
# 获取服务 IP 列表
curl -H "X-Api-Key: sk_live_abc123def456..." \
  http://cmdb.example.com/api/v1/cmdb/services/web-api/ips
```

### 创建 API Key（管理员操作）

```bash
# 登录获取 JWT
TOKEN=$(curl -s -X POST http://cmdb.example.com/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.data.token')

# 创建 API Key
curl -X POST http://cmdb.example.com/api/v1/api-keys \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"jenkins-ci","scope":"read"}'
```

## 安全考虑

1. **Key 存储**：数据库只存哈希，原文仅创建时返回一次
2. **传输安全**：生产环境必须使用 HTTPS
3. **权限隔离**：API Key 独立于用户系统，不继承用户权限
4. **审计追踪**：可通过 Key name 追踪调用来源
5. **最小权限**：默认只给 read 权限

## 兼容性

- 现有 JWT 认证流程不受影响
- Casbin 权限控制仅对 JWT 用户生效
- API Key 认证走独立的 scope 检查