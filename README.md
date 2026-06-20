# 智能眼镜巡检任务管理系统

面向巡检场景的完整 MVP：后台配置巡检模板和任务计划，系统按计划生成巡检任务，Android 智能眼镜端执行任务并提交照片/文字证据，后台跟踪任务进度、异常缺陷和设备状态。

## 功能范围

- 后台管理端
  - 登录、工作台、巡检模板、任务计划、任务管理、缺陷管理、设备管理
  - 组织管理、用户管理、角色权限、菜单权限
  - likeadmin 风格布局：左侧可收缩菜单、顶部面包屑、页面多 Tab、右上角用户菜单
  - 国网绿主题
  - **数据范围权限控制**：支持全部数据、本组织及下级、仅本组织、仅自己四级权限
- 后端 API
  - Gin + GORM 分层单体
  - JWT 认证，区分 admin 和 glasses scope
  - 巡检模板、计划、任务、节点结果、缺陷、设备、附件证据等核心模型
  - 任务状态机和缺陷状态机
  - Scheduler 按计划生成任务，使用数据库唯一索引保证幂等
  - MinIO/S3 预签名上传和证据元数据
  - **DataScope 数据范围过滤中间件**：自动按角色权限过滤查询结果
- Android 眼镜端 baseline
  - Kotlin 数据模型
  - 简单任务详情页面
  - 节点提交前的文字/照片必填校验模型

## 目录结构

```text
AI Glasses/
  admin/                 Vue 3 + Element Plus 后台管理端
  android/               Android 智能眼镜端 baseline
  server/                Go + Gin 后端服务
  HMS/likeadmin_go-master/ likeadmin 参考/demo 工程
  docker-compose.yml     可选本地基础设施编排
```

## 开发人员手册

新增功能、页面、接口、模型和配置时，请先阅读：

```text
DEVELOPER_GUIDE.md
```

## 快速启动

前提：MySQL、Redis、RabbitMQ、MinIO 等基础服务已通过 `docker compose up -d` 启动。

### 一键启动前后端

```bash
./scripts/dev.sh
```

脚本会同时启动后端（Go API）和前端（Vite），启动完成后显示访问地址。按 `Ctrl+C` 停止所有服务。

也可以单独启动：

```bash
./scripts/dev.sh backend    # 仅后端
./scripts/dev.sh frontend   # 仅前端
```

### 在 VS Code 中启动

项目已配置 `.vscode/launch.json` 和 `.vscode/tasks.json`，打开项目文件夹后即可使用：

**方式一：调试启动（推荐）**

1. 按 `F5` 或点击运行面板，选择 **"启动前后端 (全部)"**
2. 后端和前端会同时启动，支持断点调试
3. 也可以单独选择 **"启动后端 (Go API)"** 或 **"启动前端 (Vite)"**

> 调试 Go 后端需要安装 [Go 扩展](https://marketplace.visualstudio.com/items?itemName=golang.go)（`golang.go`）和 `dlv` 调试器。

**方式二：任务启动**

1. 按 `Cmd+Shift+B`（macOS）或 `Ctrl+Shift+B`（Windows/Linux）执行默认构建任务
2. 会自动并行启动前后端

其他可用任务（`Cmd+Shift+P` → `Tasks: Run Task`）：

| 任务名称 | 说明 |
| --- | --- |
| 启动后端 | 启动 Go API 服务 |
| 启动前端 | 启动 Vite 开发服务 |
| 启动前后端 (全部) | 并行启动前后端 |
| 初始化数据库 | 执行数据库迁移和种子数据 |
| 安装前端依赖 | 执行 npm install |
| 后端测试 | 运行 go test |
| 前端构建 | 运行 npm run build |

## 本地依赖

当前本地开发默认使用：

- MySQL：`127.0.0.1:3306`
- Redis：`127.0.0.1:6379`
- RabbitMQ：`127.0.0.1:5672`
- MinIO：`127.0.0.1:9000`

已验证的本地 Docker 容器名称：

- MySQL：`mysql`
- Redis：`aiglasses-redis`
- RabbitMQ：`aiglasses-rabbitmq`

## 后端配置

后端默认读取：

```text
server/config.yaml
```

也可以用环境变量覆盖：

```bash
CONFIG_FILE=/path/to/config.yaml go run ./cmd/api
```

当前本地配置示例：

```yaml
http_addr: ":8080"
database_dsn: "root:123456@tcp(127.0.0.1:3306)/aiglasses?charset=utf8mb4&parseTime=True&loc=UTC"
jwt_secret: "dev-only-change-me"
access_token_ttl: "30m"
refresh_token_ttl: "720h"
redis_addr: "127.0.0.1:6379"
redis_password: ""
rabbitmq_url: "amqp://aiglasses:aiglasses@127.0.0.1:5672/"
s3_endpoint: "127.0.0.1:9000"
s3_access_key: "minioadmin"
s3_secret_key: "minioadmin"
s3_bucket: "aiglasses-evidence"
s3_use_ssl: false
scheduler_lookback: "24h"
required_photo_max_bytes: 10485760
audio_max_bytes: 31457280
```

## 初始化数据库

```bash
cd server
go run ./cmd/initdb
```

初始化会执行：

- GORM AutoMigrate
- 默认角色、用户、班组、演示设备种子数据

默认账号：

| 入口 | 用户名 | 说明 |
| --- | --- | --- |
| 后台 | `admin` | 系统管理员 |
| 眼镜端 | `inspector` | 巡检员 |

## 数据范围权限

系统实现了基于角色的四级数据范围权限控制：

| 范围级别 | 值 | 说明 |
| --- | --- | --- |
| 全部数据 | `all` | 可查看所有组织和所有用户的数据（超级管理员） |
| 本组织及下级 | `org_and_sub` | 可查看本组织及所有下级组织的数据 |
| 仅本组织 | `org_only` | 只能查看本组织内部的数据（默认） |
| 仅自己 | `self_only` | 只能查看自己创建或分配给自己的数据 |

### 数据库迁移

首次初始化全新数据库时执行：

```bash
cd server
go run ./cmd/initdb
```

如果是已有数据库拉取近期更新后的代码，需要先备份数据库，然后执行近期结构更新脚本。

别人拉取本次提交后的更新步骤：

```bash
git pull
mysql -u <user> -p <database> < server/scripts/update_recent_schema.sql
# 然后重启后端 API
```

如果当前 shell 已在 `server/` 目录，也可以执行：

```bash
mysql -u <user> -p <database> < scripts/update_recent_schema.sql
```

更新内容包括：
- `users` 表补齐个人资料、头像、组织、角色字段
- `organizations` 表补齐 `parent_code` 字段和索引
- `roles` 表补齐角色编码、描述、数据范围、排序、状态字段和索引
- `permissions` 表补齐菜单权限管理所需字段，例如图标、路由、组件、权限标识、显示状态
- 为内置用户回填 `role_id`，为内置角色回填 `data_scope`

执行脚本后重启后端 API，让新字段和新接口一起生效。

### 配置方式

1. 登录后台管理端
2. 进入「系统管理 → 角色管理」
3. 新增/编辑角色时选择「数据范围」
4. 不同角色的用户登录后自动应用对应的数据过滤

### 已接入数据范围的模块

- ✅ 用户管理：按用户所属组织过滤
- ✅ 任务管理：按任务执行者所属组织过滤
- 其他模块可参照 `users.ListWithScope` / `tasks.AdminListWithScope` 模式扩展

演示设备：

```text
GLASS-DEMO-001
```

## 启动后端

```bash
cd server
go run ./cmd/api
```

后端地址：

```text
http://127.0.0.1:8080
```

登录接口示例：

```bash
curl -X POST http://127.0.0.1:8080/api/admin/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin"}'
```

## 启动后台管理端

```bash
cd admin
npm install
npm run dev
```

打开 Vite 输出的地址，例如：

```text
http://localhost:5180/login
```

后台登录：

```text
用户名：admin
```

## 启动 Android baseline

Android 工程位于：

```text
android/
```

可使用 Android Studio 打开，或在具备 Gradle/Android SDK 的环境中执行构建。当前 baseline 重点验证任务节点模型和基础 UI，不包含完整相机、上传队列和硬件适配。

## likeadmin demo 工程

参考/demo 工程位于：

```text
HMS/likeadmin_go-master/
```

本地 demo 已配置：

- 后端：`http://127.0.0.1:8001`
- 前端：`http://localhost:5178`
- 数据库：`likeadmin`
- 账号：`admin`
- 密码：`123456`

本项目后台已参考其交互方式实现：可收缩左侧菜单、多 Tab、右上角用户菜单、Element Plus 管理端视觉风格。

## 验证命令

后端：

```bash
cd server
go test ./...
```

后台：

```bash
cd admin
npm run build
```

## 常见问题

### 登录失败

确认后端正在监听：

```bash
lsof -nP -iTCP:8080 -sTCP:LISTEN
```

确认前端访问的是当前 Vite 输出端口，并且 Vite proxy 指向 `http://127.0.0.1:8080`。

### Redis/RabbitMQ 未连接

确认容器已启动并映射端口：

```bash
docker ps | grep -E 'aiglasses-redis|aiglasses-rabbitmq'
```

默认端口：

- Redis：`6379`
- RabbitMQ：`5672`
- RabbitMQ 管理后台：`15672`

### MySQL 连接失败

确认数据库存在：

```bash
docker exec mysql mysql -uroot -p123456 -e "SHOW DATABASES;"
```

如需重新创建：

```bash
docker exec mysql mysql -uroot -p123456 -e "CREATE DATABASE IF NOT EXISTS aiglasses CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```
