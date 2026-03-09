# Cursor CMDB

企业级配置管理数据库（CMDB），核心功能为管理「服务/主机」资产，提供完整 RBAC 权限体系（Casbin）、JWT 登录与页面级菜单权限控制。

## 技术栈

| 端 | 技术 |
|----|------|
| 前端 | Vue3 (Composition API + `<script setup>`)、Vite 5+、Pinia、Vue Router 4、Element Plus（暗黑主题）、Axios、TypeScript |
| 后端 | Go 1.22+、Gin、GORM v2、MySQL 8.0、Casbin v2、JWT、Viper、Zap |

## 目录结构

```
cursor-cmdb/
├── backend/          # Gin 后端
│   ├── casbin/       # Casbin 模型与初始化
│   ├── config/       # 配置与 Viper
│   ├── controller/   # 接口处理
│   ├── middleware/   # JWT、Casbin、CORS、Recovery
│   ├── model/        # GORM 模型
│   ├── router/       # 路由注册
│   ├── bootstrap/    # 建表与默认数据、Casbin 策略
│   ├── logger/       # Zap 日志
│   ├── utils/        # 响应、分页、密码
│   ├── config/config.yaml
│   └── main.go
├── frontend/         # Vite + Vue3 前端
│   ├── src/
│   │   ├── assets/
│   │   ├── components/
│   │   ├── layouts/
│   │   ├── router/
│   │   ├── stores/
│   │   ├── views/
│   │   ├── utils/
│   │   └── types/
│   └── package.json
├── sql/
│   └── init.sql      # MySQL 初始化（表结构 + 角色 + 菜单 + 角色菜单）
└── README.md
```

## 快速开始

### 1. 数据库

- 安装 MySQL 8.0，创建库：`CREATE DATABASE cursor_cmdb CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;`
- 执行初始化脚本：`mysql -u root -p cursor_cmdb < sql/init.sql`

### 2. 后端

```bash
cd backend
cp config/config.yaml.example config/config.yaml   # 按需修改 DSN、JWT secret 等
go mod tidy
go run main.go
```

- 默认监听 `:8080`
- 配置项 `app.bootstrap: true` 时，首次启动会：自动建表（若用 GORM 迁移）、写入默认角色/菜单/角色菜单、创建管理员 `admin` / `admin123`、写入 Casbin 策略

### 3. 前端

```bash
cd frontend
npm install
npm run dev
```

- 默认 `http://localhost:5173`，代理 `/api` 到后端 `http://localhost:8080`

### 4. 登录与权限

- 打开浏览器访问 `http://localhost:5173`，跳转登录页。
- 使用默认管理员：**用户名** `admin`，**密码** `admin123`。
- 登录成功后进入仪表盘，侧栏根据当前用户角色显示菜单（admin 可见全部，operator/viewer 仅仪表盘与资产管理）。

## 配置说明

### 后端 `config/config.yaml`

| 配置 | 说明 |
|------|------|
| `server.addr` | 服务监听地址，默认 `:8080` |
| `server.mode` | `debug` / `release` |
| `mysql.dsn` | MySQL 连接串，需指定数据库名 |
| `jwt.secret` | JWT 签名密钥，生产环境务必修改 |
| `jwt.expire_h` | Token 有效期（小时） |
| `app.bootstrap` | 是否执行建表与默认数据、Casbin 策略 |

### 前端代理

开发环境在 `vite.config.ts` 中已将 `/api` 代理到后端，生产部署时需在 Nginx 等反向代理中配置 `/api` 指向后端服务。

## API 一览

- **认证**：`POST /api/v1/login`、`GET /api/v1/me`
- **用户**：`GET/POST/PUT/DELETE /api/v1/users`、`PUT /api/v1/users/:id/roles`
- **角色**：`GET/POST/PUT/DELETE /api/v1/roles`、`GET/POST /api/v1/roles/:id/menus`
- **菜单**：`GET/POST/PUT/DELETE /api/v1/menus`
- **CMDB 资产**：`GET/POST/PUT/DELETE /api/v1/cmdb/assets`、`POST /api/v1/cmdb/assets/batch-delete`、`GET /api/v1/cmdb/assets/export`
- **公共**：`GET /api/v1/cmdb/cloud-providers`

统一响应格式：`{ "code": 200, "message": "success", "data": {} }`；错误码 401 未登录、403 无权限、500 服务器错误。

## 部署

- **前端**：`npm run build` 后，将 `dist` 目录部署到 Nginx 等静态服务器，并配置 `/api` 反向代理到后端。
- **后端**：可编译为二进制在服务器运行，或使用 Docker 构建镜像运行；需保证能连 MySQL 且 `config.yaml` 或环境变量正确。

## 截图占位

<!-- 此处可放置登录页、仪表盘、资产管理列表、角色菜单权限等截图 -->

| 登录页 | 仪表盘 | 资产管理 |
|--------|--------|----------|
| （截图占位） | （截图占位） | （截图占位） |

## 许可证

MIT
