# 开发人员手册

本文档用于说明本项目新增代码时的放置位置、分层职责和基础规范。新增功能前请先确认代码应属于后端、后台管理端还是 Android 眼镜端，并保持现有目录结构一致。

## 一、整体目录职责

```text
AI Glasses/
  admin/    Vue 3 + Element Plus 后台管理端
  android/  Android 智能眼镜端 baseline
  server/   Go + Gin 后端服务
```

- 后端业务能力放在 `server/`。
- 管理后台页面、菜单、路由和接口调用放在 `admin/`。
- 眼镜端任务执行模型和移动端界面放在 `android/`。
- 本地运行、初始化和常见问题写入 `README.md`。
- 开发结构、代码归属和新增规则写入本文档。

## 二、后端开发规范

后端采用 Gin + GORM 分层单体结构。新增代码时优先按业务域拆分，避免把业务逻辑写在路由层或数据库层。

### 1. 数据模型放置位置

数据表模型统一放在：

```text
server/internal/platform/database/models.go
```

新增表或字段时需要同步检查：

```text
server/internal/platform/database/migrate.go
```

规则：

- GORM 模型结构体放在 `models.go`。
- 自动迁移列表放在 `migrate.go` 的 `AutoMigrate()` 中。
- 表之间的关联字段应使用明确的 `ID` 字段，例如 `TaskID`、`NodeID`、`DeviceID`。
- 常用查询字段需要考虑索引，例如状态、截止时间、用户 ID、设备 ID。
- 不要在模型文件中写业务判断，模型只描述数据结构。

### 2. 业务逻辑放置位置

业务逻辑按领域放在：

```text
server/internal/<业务域>/service.go
```

现有示例：

```text
server/internal/templates/service.go     巡检模板业务
server/internal/plans/service.go         任务计划业务
server/internal/tasks/service.go         巡检任务业务
server/internal/defects/service.go       缺陷业务
server/internal/devices/service.go       设备业务
server/internal/attachments/service.go   附件业务
server/internal/auth/auth.go             登录、JWT 和认证中间件
```

规则：

- Handler 只负责参数解析、调用 Service、返回响应。
- Service 负责业务校验、状态流转、事务和跨表写入。
- 数据库读写通过 GORM 在 Service 内完成。
- 涉及任务状态的判断应调用 `server/internal/tasks/state_machine.go`。
- 不要把状态流转规则散落在多个 Handler 中。

### 3. 后端路由和接口放置位置

HTTP 路由和 Handler 统一放在：

```text
server/internal/httpapi/handlers.go
```

规则：

- 新增后台接口放在 `/api/admin/...` 分组下。
- 新增眼镜端接口放在 `/api/glasses/...` 分组下。
- 后台接口使用 admin scope。
- 眼镜端接口使用 glasses scope。
- 路由注册写在 `Register()` 中。
- 具体处理函数按业务含义命名，例如 `createTemplate`、`listTasks`、`submitNodeResult`。
- Handler 中不要直接写复杂业务规则，应调用对应 Service。

### 4. 配置放置位置

本地配置文件为：

```text
server/config.yaml
```

配置结构和环境变量覆盖逻辑放在：

```text
server/internal/config/config.go
```

规则：

- 新增配置项时，先在 `Config` 结构体中添加字段。
- 再在 `applyFile()` 中支持 YAML 读取。
- 如需环境变量覆盖，再在 `applyEnv()` 中添加覆盖逻辑。
- 配置默认值应写在 `Load()` 的默认配置中。
- 不要把数据库地址、Redis 地址、JWT 密钥等写死在业务代码中。

### 5. 错误码放置位置

统一错误码放在：

```text
server/internal/platform/httperr/errors.go
```

响应工具放在：

```text
server/internal/platform/httperr/respond.go
```

规则：

- 新增业务错误时，先补充 `ErrorCode`。
- 再补充错误码对应的 HTTP 状态码、是否可重试、默认消息。
- Handler 返回错误时优先使用 `httperr.Respond()`。
- 不要在不同接口里手写不一致的错误响应结构。

### 6. 状态机放置位置

任务状态机放在：

```text
server/internal/tasks/state_machine.go
```

缺陷状态流转当前放在：

```text
server/internal/defects/service.go
```

规则：

- 任务领取、开始、节点提交、任务提交、完成、取消、逾期必须走统一状态判断。
- 新增状态时，需要同步修改状态机、Service 校验、测试和前端展示。
- 不要只在前端限制按钮，后端必须再次校验状态合法性。

### 7. 定时任务和事件放置位置

调度器放在：

```text
server/internal/events/scheduler.go
```

事件发布放在：

```text
server/internal/events/publisher.go
```

规则：

- 计划生成任务必须依赖数据库唯一索引保证幂等。
- 重复 Tick 不应生成重复任务。
- 后续如增加 RabbitMQ 消费者，应放在 `server/internal/events/` 或独立业务域下。
- 事件副作用必须设计为可重试、幂等。

### 8. 后端启动入口

API 服务入口：

```text
server/cmd/api/main.go
```

数据库初始化入口：

```text
server/cmd/initdb/main.go
```

规则：

- 服务启动、配置加载、数据库连接、路由注册放在 `cmd/api/main.go`。
- 初始化数据库、种子数据写入放在 `cmd/initdb/main.go`。
- 不要在 `cmd/` 中堆业务逻辑。

## 三、后台管理端开发规范

后台管理端采用 Vue 3 + Element Plus，页面风格保持 likeadmin 风格和国网绿主题。

### 1. 页面放置位置

页面组件放在：

```text
admin/src/views/
```

现有示例：

```text
admin/src/views/Login.vue       登录页
admin/src/views/Dashboard.vue   工作台
admin/src/views/Templates.vue   巡检模板
admin/src/views/Plans.vue       任务计划
admin/src/views/Tasks.vue       任务管理
admin/src/views/Defects.vue     缺陷管理
admin/src/views/Devices.vue     设备管理
```

规则：

- 一个业务页面对应一个 `.vue` 文件。
- 页面内只写展示状态、表单状态和页面交互。
- 与后端通信应调用 `admin/src/api/client.ts`。
- 不要在页面中硬编码大量接口路径和重复 fetch 逻辑。

### 2. 前端路由放置位置

路由规则放在：

```text
admin/src/router/index.ts
```

规则：

- 新增页面时，需要在路由表中添加 path、name、component。
- 登录页保持 `/login`。
- 登录后的业务页面放在主布局内。
- 路由 path 应与菜单 path 保持一致。

### 3. 菜单放置位置

左侧菜单配置当前放在：

```text
admin/src/App.vue
```

重点维护：

```text
menuItems
```

规则：

- 新增后台页面后，需要在 `menuItems` 中添加菜单项。
- 菜单的 `path` 必须与 `admin/src/router/index.ts` 中的路由一致。
- 菜单图标应使用 Element Plus 图标。
- 菜单名称使用中文业务名。
- 多 Tab 页签依赖菜单路径识别，菜单路径错误会导致页签显示异常。

### 4. 接口调用放置位置

后台接口调用统一放在：

```text
admin/src/api/client.ts
```

规则：

- 通用请求函数放在 `request()`。
- 登录 token 读取、请求头设置、错误处理统一在该文件中处理。
- 新增接口时，优先封装成具名函数。
- 页面组件调用具名函数，不直接散落 `fetch()`。

### 5. 样式放置位置

全局样式放在：

```text
admin/src/styles/index.css
```

规则：

- 国网绿主题变量放在全局样式中统一维护。
- 页面局部样式写在对应 `.vue` 文件的 `<style scoped>` 中。
- 通用布局、菜单、Tab、顶部栏风格应保持一致。
- 不要在多个页面重复定义主题色。

### 6. 后台入口文件

前端入口：

```text
admin/src/main.ts
```

根组件：

```text
admin/src/App.vue
```

规则：

- 插件注册、Element Plus 注册、路由挂载放在 `main.ts`。
- 主布局、菜单、顶部栏、多 Tab、用户菜单放在 `App.vue`。

## 四、Android 眼镜端开发规范

Android baseline 代码放在：

```text
android/app/src/main/java/com/aiglasses/inspection/
```

现有核心文件：

```text
TaskModels.kt    任务、节点、节点提交模型
MainActivity.kt  演示任务详情界面
```

规则：

- 与任务执行相关的数据模型放在 `TaskModels.kt` 或新的模型文件中。
- 页面和交互逻辑可按功能新增 Activity、Fragment 或 Compose 页面。
- 必填照片、必填文字等提交校验应保留在可测试的模型函数中。
- 后续接入相机、上传队列、弱网重试时，应避免把网络重试逻辑写死在 Activity 中。

## 五、测试规范

### 1. 后端测试

后端测试文件与被测代码放在同级目录，命名为：

```text
*_test.go
```

现有示例：

```text
server/internal/tasks/state_machine_test.go
server/internal/platform/httperr/errors_test.go
```

规则：

- 状态机、错误码、核心业务校验必须有单元测试。
- 新增任务状态、错误码、设备生命周期规则时，需要同步补测试。
- 运行命令：

```bash
cd server
go test ./...
```

### 2. 前端验证

后台构建命令：

```bash
cd admin
npm run build
```

规则：

- 新增页面后至少执行构建验证。
- 涉及 UI 变化时，需要在浏览器中实际打开页面检查。
- 登录、菜单跳转、多 Tab、退出登录是基础回归路径。

### 3. Android 验证

Android 工程使用 Android Studio 或 Gradle 验证。

规则：

- 新增模型校验逻辑时优先补单元测试。
- 涉及真实硬件能力时，需要说明当前是否只完成 baseline。

## 六、注释和命名规范

### 1. 注释规范

- 文件顶部应使用中文说明该文件职责。
- 每个函数或方法前应添加中文注释。
- 注释说明“为什么存在、负责什么”，不要只重复函数名。
- 新增复杂状态规则时，应说明允许和禁止的业务场景。

### 2. 命名规范

- 后端 Go 包名使用小写英文，例如 `tasks`、`defects`、`devices`。
- Service 方法使用清晰动词，例如 `Create`、`List`、`Cancel`、`SubmitNode`。
- 前端页面文件使用业务名英文复数，例如 `Templates.vue`、`Devices.vue`。
- 路由 path 使用小写英文，例如 `/templates`、`/plans`。
- 数据库字段使用 GORM 默认命名风格，业务含义保持清晰。

## 七、新增功能推荐流程

新增一个完整业务功能时，建议按以下顺序修改：

1. 在 `server/internal/platform/database/models.go` 添加数据模型。
2. 在 `server/internal/platform/database/migrate.go` 添加迁移模型。
3. 在 `server/internal/<业务域>/service.go` 添加业务逻辑。
4. 在 `server/internal/httpapi/handlers.go` 添加接口和路由。
5. 在 `server/internal/platform/httperr/errors.go` 补充需要的错误码。
6. 在 `admin/src/api/client.ts` 添加前端接口函数。
7. 在 `admin/src/views/` 添加页面。
8. 在 `admin/src/router/index.ts` 添加前端路由。
9. 在 `admin/src/App.vue` 的 `menuItems` 添加菜单。
10. 在对应目录补充测试或构建验证。
11. 如涉及眼镜端执行流程，在 `android/app/src/main/java/com/aiglasses/inspection/` 添加模型和界面。

## 八、不要这样做

- 不要把数据库连接、密码、端口写死在业务代码中。
- 不要在前端页面中复制大量重复请求逻辑。
- 不要绕过后端状态机直接改任务状态。
- 不要只做前端按钮限制而缺少后端权限和状态校验。
- 不要新增页面后忘记添加路由和菜单。
- 不要新增数据模型后忘记加入 AutoMigrate。
- 不要新增错误场景时返回随意格式的 JSON。
- 不要把 Android Activity 写成同时负责 UI、网络、重试、存储的“大文件”。

## 九、VS Code 开发环境

项目已预配置 `.vscode/launch.json` 和 `.vscode/tasks.json`，在 VS Code 中打开项目文件夹即可使用。

### 1. 推荐扩展

安装以下扩展可获得最佳开发体验：

- **Go**（`golang.go`）：Go 语言支持、调试、代码补全
- **Vue - Official**（`Vue.volar`）：Vue 3 语法高亮、类型检查
- **ESLint**（`dbaeumer.vscode-eslint`）：前端代码规范检查

### 2. 启动和调试

按 `F5` 打开运行面板，可选择：

| 配置名称 | 说明 |
| --- | --- |
| 启动前后端 (全部) | 同时启动后端和前端，支持断点调试，停止时一并关闭 |
| 启动后端 (Go API) | 仅启动后端，可在 Go 代码中设置断点 |
| 启动前端 (Vite) | 仅启动前端开发服务 |
| 初始化数据库 | 执行数据库迁移和种子数据写入 |

调试 Go 后端前，需要安装 `dlv` 调试器：

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

### 3. 任务命令

按 `Cmd+Shift+B`（macOS）或 `Ctrl+Shift+B`（Windows/Linux）可执行默认构建任务（并行启动前后端）。

通过 `Cmd+Shift+P` → `Tasks: Run Task` 可以运行更多任务：启动后端、启动前端、初始化数据库、安装前端依赖、后端测试、前端构建等。

### 4. 一键启动脚本

如果不在 VS Code 中开发，也可以使用终端脚本：

```bash
./scripts/dev.sh            # 启动前后端
./scripts/dev.sh backend    # 仅后端
./scripts/dev.sh frontend   # 仅前端
```

脚本会自动检测前端依赖、启动服务并显示访问地址，按 `Ctrl+C` 统一停止。

## 十、常用命令

初始化数据库：

```bash
cd server
go run ./cmd/initdb
```

启动后端：

```bash
cd server
go run ./cmd/api
```

后端测试：

```bash
cd server
go test ./...
```

启动后台管理端：

```bash
cd admin
npm run dev
```

后台构建：

```bash
cd admin
npm run build
```
