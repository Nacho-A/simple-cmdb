# 服务列表页面实现计划

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 新增"服务列表"页面，按 service_name 从 cmdb_assets 表聚合展示 IP 信息。

**Architecture:** 后端新增聚合查询接口，前端新增独立页面，bootstrap 新增菜单项和 Casbin 策略。

**Tech Stack:** Go 1.22+, Gin, GORM v2, Vue3, Element Plus

---

## File Structure

### 新增文件
| 文件 | 职责 |
|------|------|
| `frontend/src/views/cmdb/services/index.vue` | 服务列表页面 |
| `frontend/src/types/service.ts` | 服务数据类型定义 |

### 修改文件
| 文件 | 改动 |
|------|------|
| `backend/controller/cmdb_asset.go` | 新增 `ServiceList` 方法 |
| `backend/router/router.go` | 注册服务列表路由 |
| `backend/bootstrap/bootstrap.go` | 新增菜单项和 Casbin 策略 |

---

## Chunk 1: 后端实现

### Task 1: 新增服务列表接口

**Files:**
- Modify: `backend/controller/cmdb_asset.go`

- [ ] **Step 1: 添加 ServiceList 方法**

在 `backend/controller/cmdb_asset.go` 文件末尾添加：

```go
type ServiceItem struct {
	ServiceName string   `json:"service_name"`
	PrivateIPs  []string `json:"private_ips"`
	PublicIPs   []string `json:"public_ips"`
	AssetCount  int      `json:"asset_count"`
}

func (h *Handler) ServiceList(c *gin.Context) {
	page, pageSize, offset, limit := utils.GetPage(c)
	q := strings.TrimSpace(c.Query("q"))

	dbq := h.DB.Model(&model.CMDBAsset{})
	if q != "" {
		dbq = dbq.Where("service_name LIKE ?", "%"+q+"%")
	}

	type serviceRow struct {
		ServiceName string
		PrivateIP   string
		PublicIP    string
	}

	var rows []serviceRow
	if err := dbq.Select("service_name, private_ip, public_ip").Find(&rows).Error; err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}

	serviceMap := make(map[string]*ServiceItem)
	for _, row := range rows {
		if row.ServiceName == "" {
			continue
		}
		item, ok := serviceMap[row.ServiceName]
		if !ok {
			item = &ServiceItem{
				ServiceName: row.ServiceName,
				PrivateIPs:  []string{},
				PublicIPs:   []string{},
			}
			serviceMap[row.ServiceName] = item
		}
		item.AssetCount++
		if row.PrivateIP != "" && !contains(item.PrivateIPs, row.PrivateIP) {
			item.PrivateIPs = append(item.PrivateIPs, row.PrivateIP)
		}
		if row.PublicIP != "" && !contains(item.PublicIPs, row.PublicIP) {
			item.PublicIPs = append(item.PublicIPs, row.PublicIP)
		}
	}

	items := make([]ServiceItem, 0, len(serviceMap))
	for _, v := range serviceMap {
		items = append(items, *v)
	}

	total := len(items)
	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	utils.OK(c, gin.H{
		"items":     items[start:end],
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: 提交**

```bash
git add backend/controller/cmdb_asset.go
git commit -m "feat: add ServiceList API for aggregating services"
```

---

### Task 2: 注册路由

**Files:**
- Modify: `backend/router/router.go`

- [ ] **Step 1: 添加服务列表路由**

在 `backend/router/router.go` 中，在 CMDB 资产路由区域添加：

```go
// 服务列表
authed.GET("/cmdb/services", h.ServiceList)

// CMDB资产
authed.GET("/cmdb/assets", h.AssetList)
```

- [ ] **Step 2: 提交**

```bash
git add backend/router/router.go
git commit -m "feat: register ServiceList route"
```

---

### Task 3: 添加 Casbin 策略

**Files:**
- Modify: `backend/bootstrap/bootstrap.go`

- [ ] **Step 1: 添加服务列表 Casbin 策略**

在 `backend/bootstrap/bootstrap.go` 的 `SeedCasbinPolicies` 函数中，找到 `if len(existing) == 0` 块内的资产策略后面（约第 216 行附近），添加：

```go
_, _ = casbinx.Enforcer.AddPolicy("viewer", "/api/v1/cmdb/assets*", "GET")

// 服务列表
_, _ = casbinx.Enforcer.AddPolicy("operator", "/api/v1/cmdb/services", "GET")
_, _ = casbinx.Enforcer.AddPolicy("viewer", "/api/v1/cmdb/services", "GET")

if err := casbinx.Enforcer.SavePolicy(); err != nil {
```

- [ ] **Step 2: 提交**

```bash
git add backend/bootstrap/bootstrap.go
git commit -m "feat: add Casbin policy for services API"
```

---

### Task 4: 编译验证后端

- [ ] **Step 1: 编译后端**

```bash
cd backend && go build -o cmdb.exe .
```

预期：编译成功无错误

---

## Chunk 2: 前端实现

### Task 5: 新增服务类型定义

**Files:**
- Create: `frontend/src/types/service.ts`

- [ ] **Step 1: 创建类型文件**

```typescript
export interface ServiceItem {
  service_name: string
  private_ips: string[]
  public_ips: string[]
  asset_count: number
}
```

- [ ] **Step 2: 提交**

```bash
git add frontend/src/types/service.ts
git commit -m "feat: add ServiceItem type definition"
```

---

### Task 6: 新增服务列表页面

**Files:**
- Create: `frontend/src/views/cmdb/services/index.vue`

- [ ] **Step 1: 创建页面文件**

```vue
<template>
  <div class="page page--table">
    <SearchForm @search="onSearch" @reset="onReset">
      <div class="filters">
        <el-input v-model="query.q" placeholder="服务名称" clearable />
      </div>
    </SearchForm>

    <div class="table-block">
      <ElTablePro :total="total" v-model:page="page" v-model:pageSize="pageSize" @change="fetchList">
        <template #toolbar>
          <div class="title">服务列表</div>
          <div class="meta">共 {{ total }} 个服务</div>
        </template>

        <div class="table-inner">
          <el-table :data="items" border height="100%">
            <el-table-column prop="service_name" label="服务名称" min-width="180" />
            <el-table-column label="私网IP" min-width="280">
              <template #default="{ row }">
                <el-tag
                  v-for="ip in row.private_ips"
                  :key="ip"
                  type="info"
                  effect="dark"
                  size="small"
                  style="margin-right: 4px; margin-bottom: 4px"
                >
                  {{ ip }}
                </el-tag>
                <span v-if="!row.private_ips?.length" class="empty-text">-</span>
              </template>
            </el-table-column>
            <el-table-column label="公网IP" min-width="200">
              <template #default="{ row }">
                <el-tag
                  v-for="ip in row.public_ips"
                  :key="ip"
                  type="warning"
                  effect="dark"
                  size="small"
                  style="margin-right: 4px; margin-bottom: 4px"
                >
                  {{ ip }}
                </el-tag>
                <span v-if="!row.public_ips?.length" class="empty-text">-</span>
              </template>
            </el-table-column>
            <el-table-column prop="asset_count" label="资产数量" width="100" align="center" />
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button size="small" type="primary" text @click="viewAssets(row)">
                  查看资产
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </ElTablePro>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import SearchForm from '@/components/SearchForm.vue'
import ElTablePro from '@/components/ElTablePro.vue'
import { request } from '@/utils/request'
import type { ApiResp } from '@/types/api'
import type { ServiceItem } from '@/types/service'

const router = useRouter()
const items = ref<ServiceItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)

const query = reactive({
  q: '',
})

async function fetchList() {
  const resp = await request.get<
    ApiResp<{ items: ServiceItem[]; total: number; page: number; page_size: number }>
  >('/cmdb/services', {
    params: {
      page: page.value,
      page_size: pageSize.value,
      ...query,
    },
  })
  items.value = resp.data.data.items || []
  total.value = resp.data.data.total || 0
}

function onSearch() {
  page.value = 1
  fetchList()
}

function onReset() {
  query.q = ''
  page.value = 1
  fetchList()
}

function viewAssets(row: ServiceItem) {
  router.push({ path: '/cmdb/assets', query: { service_name: row.service_name } })
}

onMounted(() => {
  fetchList()
})
</script>

<style scoped>
.page--table {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 16px;
}
.page--table .table-block {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  margin-top: 12px;
}
.table-inner {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}
.filters {
  display: grid;
  grid-template-columns: 1fr;
  gap: 10px;
}
.title {
  font-weight: 800;
}
.meta {
  opacity: 0.75;
  font-size: 12px;
  align-self: center;
}
.empty-text {
  color: #999;
}
</style>
```

- [ ] **Step 2: 提交**

```bash
git add frontend/src/views/cmdb/services/index.vue
git commit -m "feat: add service list page"
```

---

### Task 7: 新增菜单项

**Files:**
- Modify: `backend/bootstrap/bootstrap.go`

- [ ] **Step 1: 添加服务列表菜单**

在 `backend/bootstrap/bootstrap.go` 中，找到 `menuAssets` 创建代码后面（约第 126 行），添加：

```go
var menuServices model.Menu
_ = db.Where("path = ?", "/cmdb/services").First(&menuServices).Error
if menuServices.ID == 0 {
	menuServices = model.Menu{
		Name:      "服务列表",
		Path:      "/cmdb/services",
		Component: "views/cmdb/services/index.vue",
		Icon:      "List",
		ParentID:  &menuCMDB.ID,
		Order:     2,
		Hidden:    false,
	}
	if err := db.Create(&menuServices).Error; err != nil {
		return err
	}
}
```

然后修改角色菜单关联（约第 180 行），添加 `&menuServices`：

```go
// admin: all
if err := db.Model(&adminRole).Association("Menus").Replace(&menuDashboard, &menuCMDB, &menuAssets, &menuServices, &menuSystem, menuUser, menuRole, menuMenu, menuAPIKey); err != nil {
	return err
}
// operator/viewer: dashboard + assets + services
if err := db.Model(&operatorRole).Association("Menus").Replace(&menuDashboard, &menuCMDB, &menuAssets, &menuServices); err != nil {
	return err
}
if err := db.Model(&viewerRole).Association("Menus").Replace(&menuDashboard, &menuCMDB, &menuAssets, &menuServices); err != nil {
	return err
}
```

- [ ] **Step 2: 提交**

```bash
git add backend/bootstrap/bootstrap.go
git commit -m "feat: add service list menu item"
```

---

## Chunk 3: 验证

### Task 8: 编译验证

- [ ] **Step 1: 编译后端**

```bash
cd backend && go build -o cmdb.exe .
```

预期：编译成功

- [ ] **Step 2: 启动后端服务**

```bash
cd backend && ./cmdb.exe
```

预期：服务启动成功，数据库表和菜单自动创建

---

### Task 9: API 测试

- [ ] **Step 1: 测试服务列表 API**

```bash
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/cmdb/services
```

预期响应：
```json
{
  "code": 200,
  "data": {
    "items": [...],
    "total": N,
    "page": 1,
    "page_size": 10
  }
}
```

---

### Task 10: 前端验证

- [ ] **Step 1: 启动前端开发服务**

```bash
cd frontend && npm run dev
```

- [ ] **Step 2: 浏览器验证**

打开浏览器访问前端，登录后在"资产管理"菜单下应该看到"服务列表"菜单项，点击进入应能正常显示服务列表。

- [ ] **Step 3: 查看提交历史**

```bash
git log --oneline -8
```

---

## Summary

| 任务 | 文件 | 说明 |
|------|------|------|
| Task 1 | `controller/cmdb_asset.go` | 服务聚合接口 |
| Task 2 | `router/router.go` | 路由注册 |
| Task 3 | `bootstrap/bootstrap.go` | Casbin 策略 |
| Task 5 | `types/service.ts` | 类型定义 |
| Task 6 | `views/cmdb/services/index.vue` | 前端页面 |
| Task 7 | `bootstrap/bootstrap.go` | 菜单项 |