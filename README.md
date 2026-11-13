# Lambo

基于 Vben Admin + Gin/GORM 的 Go / Vue 业务脚手架，后端从 Nunu Basic 演进。对外开发工具统一使用 **`lambo`**。

**当前能力：** CLI 新建项目、后端分层组件生成、Vben Element Plus 前端代码生成、后端热重载、Wire 生成与工具升级。认证、权限后台、字典后台、真实数据库初始化和业务状态流转仍按 [plan.md](plan.md) 开发，尚未完成前后端联调。

## 安装本地命令

需要 Go `1.25.8+`，本次使用 Go `1.27.1` 验证。在仓库根目录执行：

```bash
go install ./cmd/lambo
lambo --help
```

确保 `$(go env GOPATH)/bin` 在 `PATH` 中。也可以不安装，使用 `go run ./cmd/lambo create frontend category`。

## 生成前端业务代码

```bash
# 先预览，不写入文件
lambo create frontend category --schema examples/category.json --dry-run

# 生成 API、表单配置、列表、详情/编辑弹窗与路由
lambo create frontend category --schema examples/category.json

# 校验产物一致性，不写文件
lambo create frontend category --schema examples/category.json --check
```

生成位置为 `frontend/apps/web-ele/src/` 下的 `api/business/`、`views/business/` 和 `router/routes/modules/business/`。复用 Vben 请求客户端、表格、表单、弹窗、图标及 Element Plus 组件，支持分页、字段筛选、选项回显、表单校验、只读字段、删除确认、加载失败和保存后刷新。

不传 `--schema` 时生成一个包含「名称」字段的起始页面；字段定义和接口约定见 [生成器说明](docs/generator.md)。这是一份开发期输入格式，字典后台与业务枚举的最终发布机制仍待确定。

## 其他命令

```bash
lambo create model category
lambo create repository category
lambo create service category
lambo create handler category
lambo create backend category --schema examples/category.json
lambo create all category --schema examples/category.json

# 以下命令由开发者按需执行，本次交付没有启动服务
lambo run
lambo run backend/cmd/server -- -conf config/local.yml
lambo wire
lambo wire all
lambo new shop --module github.com/example/shop
lambo upgrade
```

`create backend` 生成四层起始组件；`create all` 在此基础上增加前端产物。后端组件需补齐真实业务并注册 Wire/路由；未实现的接口明确返回错误，不伪装成可用 CRUD。前端生成器只生成代码，不会自动创建数据库、权限记录或 HTTP 接口。

`run` 默认定位 `backend/`，支持文件热重载及 `--` 后透传应用参数。`wire` 使用固定的 Wire `v0.7.0`，首次执行可能下载工具依赖，无需全局安装 Wire。`new` 只克隆 Lambo 前后端模板并替换两个 Go 模块路径，不覆盖已有目录、不安装依赖。`new` 和 `upgrade` 默认使用公开仓库，因此新增工具需发布后才可从远端获得。

## 生成安全与开发边界

- 相同内容重复生成不改动文件；有不同内容时整批停止，不提供强制覆盖。可先 `--dry-run`，再手动合并。自动格式化过的起始代码也可能与原模板产生差异，属于应保留的修改。
- 拒绝非法名称、重复字段、类型不匹配、目录越界和符号链接目标。`--check` 对缺失或有差异的文件返回非零状态。
- `--tpl-path` 支持按相对路径替换部分模板；自定义模板和字段定义属于受信任的本地开发输入。
- 后端模块为 `github.com/Edward-Jackie/lambo/backend`；CLI 是独立的根 Go 模块，工具依赖不进入业务服务。
- 前端使用 Element Plus 作为首套业务应用；暂时保留其他上游应用，不维护其生成模板。
- 前端要求 Node.js `^22.18.0 || ^24.12.0`、pnpm `>=11.0.0`，锁定 `pnpm@11.16.0`。在 `frontend/` 内执行 `pnpm install --frozen-lockfile`。
- 继承的数据库初始化和用户查询仍是占位实现；公开演示配置不得用于真实部署。前端按钮隐藏不代替后端鉴权。

## 目录与文档

```text
lambo/
├── cmd/lambo/、internal/、tpl/   # Lambo CLI 与内嵌模板
├── config/                         # CLI 工具版本和默认值
├── examples/category.json          # 开发期字段定义示例
├── backend/                        # 独立的 Go 业务模块
├── frontend/                       # Vben 多包结构，业务首选 web-ele
├── docs/generator.md               # 命令、字段、API 契约与限制
└── plan.md                         # 功能阶段和验收
```

[工程约定](AGENTS.md) · [技术决策](docs/architecture/0001-foundation.md) · [上游来源与版权](THIRD_PARTY_NOTICES.md)

[MIT](LICENSE)。保留 Vben、Nunu 和其他依赖的原版权声明。
