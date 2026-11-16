# Lambo 工程约定

## 当前范围

- 已确定使用 Vben Admin + Nunu Basic；一个仓库、前后端分目录、后端模块化单体、REST JSON、GORM。
- 当前已有 Lambo CLI、后端组件与 Vben Element Plus 前端生成器；业务后台仍在开发。功能范围和验收以根目录 `plan.md` 为准，技术取舍见 `docs/architecture/0001-foundation.md`。
- 业务 UI 已确定为 Element Plus（`frontend/apps/web-ele`）；开发联调使用 PostgreSQL；生产数据库迁移、字段元数据真源和发布机制尚未定稿，不得将候选方案写成已实现事实。

## 开发纪律

- 对话使用简体中文，称呼用户为大当家；新增或修改的注释使用简体中文，不为统一语言批量改动上游历史注释。
- 修改功能前先读同一应用或模块的 1—2 个既有实现；保留其有效工程规范，示例占位实现不得直接当生产基准。
- 前端复用选定的 Vben 组件、图标、按钮和主题变量；首版只维护 Element Plus 一套业务 UI 应用，其他上游应用暂保留，裁剪在 P0 完成。
- 后端沿用 Nunu 的显式调用关系，优先复用已有库和标准库；不为每个简单类型增设接口、工厂或空层。
- 业务规则由后端执行；菜单或按钮隐藏不构成授权。流程状态通过业务动作修改，通用编辑接口不能任意设置关键状态。
- 字段、字典和枚举应有明确引用关系及单一真源；先定义模型关系与动作契约，再用现有生成器生成前后端基线，在基线上补充业务规则并记录差异。生成器不得覆盖已有业务修改。
- 不提交真实密钥、生产 DSN、真实业务数据、运行日志或构建产物。开发示例配置使用明确的演示值。
- 不自动启动长期服务。提交按内容分批、使用中文说明，推送需用户授权；本次创建公开仓库及初始发布已获授权，不代表后续持续推送授权。
- 上游源码自带的测试和文档保留。本次明确要求的 `plan.md` 随项目归档；后续新增测试或计划文件是否提交遵循用户当次要求。

## 验证与交付

- Go 变更按范围执行格式、构建、静态检查及必要行为验证；并发、资金和授权规则需要覆盖失败路径。
- Vue 变更使用已有类型检查和构建命令，交互改动验证加载、空态、错误、权限及数据刷新。
- 没有执行的检查明确记录，不把上游有测试文件、源码已导入或计划任务已列出当作验证通过。
- 交付自查：后端查询能力、性能、日志、注释、常量及封装；前端组件复用、图标、按钮、文案与主题一致性。

## CLI 与生成器

- `backend/` 承载业务 Go 模块，`tools/lambo/` 是独立的 CLI Go 模块；两者分别验证，不把工具依赖引入服务。
- 一级目录按职责划分为 `backend/`、`docs/`、`frontend/`、`deploy/`、`tools/`、`.skills/`；`.git/` 由 Git 管理。跨前后端的命令、模板、配置和字段示例统一归入 `tools/lambo/`。
- CLI 安装：从仓库根目录执行 `go -C tools/lambo install ./cmd/lambo`；模块移动需同步修正新建项目、升级、扫描与文档路径。
- 本地验证产物放在 `tools/lambo/.artifacts/`，IDE 配置归入对应子项目，不提交这些本地文件。
- 对外命令统一为 `lambo`，保留许可证和来源中的原作者信息。
- 字段 JSON 是当前开发期生成契约，尚不代表字典后台与枚举发布真源已经确定。
- 修改模板时，以实际生成产物进行 Go 构建、前端类型检查与构建，并验证冲突拒绝、重复生成和路径边界；不要只检查模板文本。

## 后端分层

- `backend/internal/` 按职责分三区：`platform/`（数据库连接、HTTP 引擎与中间件）、`system/`（认证、授权、账号、角色、菜单）、`business/`（业务模块）。`pkg/` 只放与项目无关的通用工具。
- `system/` 与 `business/` 都按 `model`、`repository`、`service`、`handler` 四层平铺，包名与层名一致；同层靠域前缀区分，即 `internal/business/model` 与 `internal/system/model`，不用目录嵌套表达模块。
- 依赖方向单向：`business` 可引用 `system`，`system` 不得引用 `business`；两者都可引用 `platform` 与 `pkg`，`platform`、`pkg` 不得引用 `system` 或 `business`。
- 路由挂载、数据库迁移清单与 Wire 装配只出现在 `cmd/server/wire/`。`platform/httpserver` 只提供引擎、CORS 与 `/healthz`；`platform/database` 只提供连接和通用迁移，表清单由装配层传入。
- 系统表的模型、事务与安全不变式（最后一个管理员保护、越权拒绝、会话撤销）留在 `system/`；`business/` 不得复制这些规则，也不得绕过 `system` 的 service 直接写系统表。
- 生成位置由字段 JSON 的 `domain` 决定：`business`（默认）写 `internal/business/<层>/`，`system` 写 `internal/system/<层>/`。调整目录必须同步修改 `tools/lambo/internal/command/create/files.go` 与 `tpl/create/*.tpl` 的 import，否则下次生成会写回旧路径。
- 已知待解耦项：`system/model/iam.go` 的 `PermissionCatalog`、`PageComponents` 与 `system/repository/iam_seed.go` 仍硬编码 `catalog:category` 权限码和"分类管理"菜单。目标是由业务模块注册权限与菜单种子，系统区不出现业务标识。

## Docker 部署

- Dockerfile、Compose、Nginx 与容器配置以 `deploy/` 为唯一入口；构建上下文为仓库根目录，使用各 Dockerfile 对应的 `.dockerignore`。
- 后端运行镜像不包含开发 CLI；前端只构建 `@vben/web-ele`，通过部署环境文件覆盖公共 Mock 配置。
- `/healthz` 只表示 HTTP 进程可响应，不等同于数据库、鉴权和业务就绪；不得据此宣称完整后台可部署。
- 不自动启动 Docker 引擎或业务容器。Docker 不可用时明确区分 Compose 解析、源码构建与实际镜像/容器验证。
