# 服务列表页面设计文档

## 概述

新增"服务列表"页面，按 `service_name` 从现有 `cmdb_assets` 表聚合展示服务 IP 信息。

## 目标

- 新增独立菜单项"服务列表"
- 按服务名聚合展示私网IP、公网IP列表
- 展示每个服务下的资产数量
- 支持按服务名称搜索

## 数据模型

无需新增数据表，直接从 `cmdb_assets` 表聚合查询。

**聚合逻辑：**
```sql
SELECT 
  service_name,
  GROUP_CONCAT(DISTINCT private_ip) as private_ips,
  GROUP_CONCAT(DISTINCT public_ip) as public_ips,
  COUNT(*) as asset_count
FROM cmdb_assets
GROUP BY service_name
```

## API 接口

### 服务列表查询

```
GET /api/v1/cmdb/services
Authorization: Bearer <jwt_token> 或 X-Api-Key: <api_key>

Query Parameters:
  - q: 服务名称模糊搜索（可选）
  - page: 页码
  - page_size: 每页数量

Response:
{
  "code": 200,
  "data": {
    "items": [
      {
        "service_name": "LOAN-API",
        "private_ips": ["172.16.1.1", "172.16.2.2"],
        "public_ips": ["1.2.3.4"],
        "asset_count": 2
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10
  }
}
```

## 文件变更

### 新增文件

| 文件 | 说明 |
|------|------|
| `frontend/src/views/cmdb/services/index.vue` | 服务列表页面 |
| `frontend/src/types/service.ts` | 服务数据类型定义 |

### 修改文件

| 文件 | 改动 |
|------|------|
| `backend/controller/cmdb.go` | 新增 `ServiceList` 方法 |
| `backend/router/router.go` | 注册服务列表路由 |
| `backend/bootstrap/bootstrap.go` | 新增菜单项 |

## 前端页面

**布局：**
- 搜索栏：服务名称
- 表格列：服务名称、私网IP列表、公网IP列表、资产数量、操作（查看资产）

**交互：**
- 私网IP、公网IP 以标签形式展示
- 点击"查看资产"跳转到资产管理页面并筛选该服务

## 菜单结构

```
资产管理
├── CMDB资产
└── 服务列表  ← 新增
```

## 权限

- 访问权限：与 CMDB 资产相同，admin/operator/viewer 均可查看
- API Key：read scope 可访问