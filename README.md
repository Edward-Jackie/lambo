# Lambo

基于 Vben Admin + Gin/GORM 的 Go / Vue 业务脚手架，后端从 Nunu Basic 演进。对外开发工具统一使用 **`lambo`**。

**当前能力：** CLI 新建项目、后端 CRUD 生成、Vben Element Plus 前端代码生成、后端热重载、Wire 生成与工具升级。生成的前后端 API 已验证普通 CRUD；认证、权限后台、字典后台、业务主数据库初始化和业务状态流转仍按 [plan.md](plan.md) 开发。

## 安装本地命令

需要 Go `1.25.8+`，本次使用 Go `1.27.1` 验证。在仓库根目录执行（`-C` 指定 CLI 所在模块）：

```bash
go -C tools/lambo install ./cmd/lambo
lambo --help
```

确保 `go env GOBIN` 指定的目录在 `PATH` 中；未设置时使用 `$(go env GOPATH)/bin`。也可以不安装，从仓库根目录使用 `go -C tools/lambo run ./cmd/lambo create frontend category`。

CLI 已迁至项目级 `tools/lambo/`。已有本地命令请先按上面的安装命令重新安装，使后续 `lambo upgrade` 使用新的模块地址。

## 生成前端业务代码

```bash
# 先预览，不写入文件
lambo create frontend category --schema tools/lambo/examples/category.json --dry-run

# 生成 API、表单配置、列表、详情/编辑弹窗与路由
lambo create frontend category --schema tools/lambo/examples/category.json

# 校验产物一致性，不写文件
lambo create frontend category --schema tools/lambo/examples/category.json --check
```

生成位置为 `frontend/apps/web-ele/src/` 下的 `api/business/`、`views/business/` 和 `router/routes/modules/business/`。复用 Vben 请求客户端、表格、表单、弹窗、图标及 Element Plus 组件，支持分页、字段筛选、选项回显、表单校验、只读字段、删除确认、加载失败和保存后刷新。

不传 `--schema` 时生成一个包含「名称」字段的起始页面；字段定义和接口约定见 [生成器说明](docs/generator.md)。这是一份开发期输入格式，字典后台与业务枚举的最终发布机制仍待确定。

## 其他命令

```bash
lambo create model category
lambo create repository category
lambo create service category
lambo create handler category
lambo create backend category --schema tools/lambo/examples/category.json
lambo create all category --schema tools/lambo/examples/category.json

# 以下命令由开发者按需执行，本次交付没有启动服务
lambo run
lambo run backend/cmd/server -- -conf config/local.yml
lambo wire
lambo wire all
lambo new shop --module github.com/example/shop
lambo upgrade
```

`create backend` 生成模型与校验类型、GORM 仓储、服务和五个 CRUD 接口；`create all` 在此基础上增加前端产物。注入数据库、迁移表并调用生成处理器的 `RegisterRoutes(router, authorize)`，即可显式接入现有应用。请求与响应和前端 API 对齐，支持分页筛选、零值更新、可选字段清空、只读字段保护及权限拒绝；装配示例见[生成器说明](docs/generator.md)。具体业务规则和按身份的数据范围由业务模块补充。

`run` 默认定位 `backend/`，支持文件热重载及 `--` 后透传应用参数。`wire` 使用固定的 Wire `v0.7.0`，首次执行可能下载工具依赖，无需全局安装 Wire。`new` 只克隆 Lambo 前后端模板并替换两个 Go 模块路径，不覆盖已有目录、不安装依赖。`new` 和 `upgrade` 默认使用公开仓库，因此新增工具需发布后才可从远端获得。

## 生成安全与开发边界

- 相同内容重复生成不改动文件；有不同内容时整批停止，不提供强制覆盖。可先 `--dry-run`，再手动合并。自动格式化过的起始代码也可能与原模板产生差异，属于应保留的修改。
- 拒绝非法名称、重复字段、类型不匹配、目录越界和符号链接目标。`--check` 对缺失或有差异的文件返回非零状态。
- `--tpl-path` 支持按相对路径替换部分模板；自定义模板和字段定义属于受信任的本地开发输入。
- 后端模块为 `github.com/Edward-Jackie/lambo/backend`；CLI 位于 `tools/lambo/`，保留独立 Go 模块，工具依赖不进入业务服务。
- 前端使用 Element Plus 作为首套业务应用；暂时保留其他上游应用，不维护其生成模板。
- 前端要求 Node.js `^22.18.0 || ^24.12.0`、pnpm `>=11.0.0`，锁定 `pnpm@11.16.0`。在 `frontend/` 内执行 `pnpm install --frozen-lockfile`。
- 继承的数据库初始化和用户查询仍是占位实现；公开演示配置不得用于真实部署。前端按钮隐藏不代替后端鉴权。

## 目录与文档

```text
lambo/
├── backend/                        # Go 业务服务
├── tools/lambo/                    # 跨前后端的独立 CLI 模块：命令、配置、模板、字段示例
├── docs/                           # 使用说明和架构决策
├── frontend/                       # Vben 多包结构，业务首选 web-ele
├── deploy/                         # 部署配置与脚本
├── .skills/                        # 项目专用 AI 技能
├── AGENTS.md                       # 工程约定
├── README.md
├── LICENSE
├── THIRD_PARTY_NOTICES.md
└── plan.md                         # 功能阶段和验收
```

[工程约定](AGENTS.md) · [技术决策](docs/architecture/0001-foundation.md) · [上游来源与版权](THIRD_PARTY_NOTICES.md)

[MIT](LICENSE)。保留 Vben、Nunu 和其他依赖的原版权声明。
