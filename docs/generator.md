# Lambo 代码生成器

## 命令与产物

在根目录 `go -C tools/lambo install ./cmd/lambo` 后，命令可从项目根目录或任意子目录使用；`--root` 可显式指定另一个 Lambo 项目。字段文件路径和自定义模板路径相对当前工作目录解析。

CLI 是跨前后端的项目开发工具，独立模块为 `github.com/Edward-Jackie/lambo/tools/lambo`，业务后端仍使用 `github.com/Edward-Jackie/lambo/backend`。已有本地命令需先重新安装，随后 `lambo upgrade` 才会使用新的模块地址。

| 命令 | 产物 |
| --- | --- |
| `lambo create model category` | 后端模型、请求/查询/分页类型、字段校验 |
| `lambo create repository category` | GORM 分页、详情、新增、编辑、删除 |
| `lambo create service category` | 参数与 ID 校验、新增默认值、业务调用 |
| `lambo create handler category` | 五个 HTTP 接口、统一响应、权限路由注册、操作日志 |
| `lambo create backend category` | 以上四个组件 |
| `lambo create frontend category` | Vben Element Plus API、字段配置、列表、详情/编辑弹窗与路由，共 5 个文件；另加菜单初始化 SQL 1 个 |
| `lambo create all category` | 后端 4 个文件 + 前端 5 个文件 + 菜单 SQL 1 个 |
| `lambo db seed category` | 读取第三版种子定义，通过后端 IAM 安装菜单与按钮权限 |

菜单真源是数据库。前端生成同时产出 `backend/internal/<domain>/seed/<名称>.sql`，其中保存第三版种子定义和 SQL 预览。首次安装创建一条 `:query` 页面菜单和三条操作按钮，并将四项权限授权给 `admin`。已识别的安装保留现有菜单和授权，冲突修正后可重试。升级旧版种子时，先在空目录预览新版产物，再合并到现有项目。

`lambo db seed` 依次读取 `--dsn`、环境变量 `LAMBO_POSTGRES_DSN` 和 `backend/config/local.yml` 的 `data.postgres.dsn`，YAML 键兼容大小写。命令借助 Go 工具链运行当前项目后端的一次性 IAM 运维入口，使用该版本的注册目录完成校验和写入。`--dry-run` 只显示种子文件。旧库缺少安装历史时，仅在核实模块从未安装后使用 `--new`。系统 IAM 骨架仍由 `InitializeIAM` 在首次启动时内嵌执行。

通用参数：`--schema <文件>`、`--dry-run`、`--check`、`--root <项目>`、`--tpl-path <模板目录>`。`--dry-run` 与 `--check` 互斥。名称采用小写英文、数字及中间的 `-` 或 `_`，不接受任意输出路径。单个后端层可能引用尚未创建的相邻层；需要完整可编译组件时使用 `create backend` 或 `create all`。

## 开发期字段定义

完整示例见 [category.json](../tools/lambo/examples/category.json)。这是当前 CLI 的本地输入契约，不代表已经决定字典后台或业务枚举的唯一真源。

| 属性 | 说明 |
| --- | --- |
| `name` | 可省略；提供时必须与命令名称一致 |
| `domain` | `business`（默认）或 `system`，决定后端输出到 `internal/business/` 还是 `internal/system/` |
| `title` | 中文页面标题，默认使用名称 |
| `api` | 静态 REST 路径，默认 `/api/<名称>`；不自动处理单复数 |
| `permission` | 权限前缀，默认 `business:<名称>` |
| `idType` | `number`（默认）或 `string`；大整数 ID 应以字符串传输 |
| `fields` | 至少一个业务字段，不能重复或覆盖 `id` 等系统字段 |
| `fields[].name / label` | 小驼峰字段名及中文标签 |
| `fields[].type` | `string`、`number`、`boolean` |
| `fields[].required` | 表单必填；`0`、`false` 仍是有效值 |
| `fields[].search` | 在查询表单和分页请求中使用此字段 |
| `fields[].readOnly` | 列表、详情可见，不进入编辑表单或提交体；关键业务状态应设为只读。生成完整 CRUD 时，只读必填字段必须提供显式默认值 |
| `fields[].maxLength` | 字符串最大长度 |
| `fields[].default` | 默认值，需匹配类型和选项；不使用真假值判断替代类型判断 |
| `fields[].options` | `[{"label":"启用","value":1}]`，类型必须匹配；回显使用同一份选项，未知值保留显示 |

目前不处理关联表、树表、文件上传字段、日期专用控件、远程字典或状态动作生成。选项是静态生成输入；业务枚举、数据库字典迁移和动态字典获取仍属 P2。数字字段按普通 JS 数字/Go `float64` 生成，不适合金额或高精度编号，定义阶段应选择字符串并补充对应业务类型。

## API 契约

以 `api: /api/categories` 为例，前端使用现有 `requestClient` 解包 `{code: 0, data, message}`，错误反馈也由其统一处理：

| 操作 | 请求 | `data` 内容 |
| --- | --- | --- |
| 分页 | `GET /api/categories?page=1&pageSize=20&name=...` | `{items: Row[], total: number}` |
| 详情 | `GET /api/categories/:id` | `Row` |
| 新增 | `POST /api/categories`，可编辑字段 JSON | 保存后的 `Row` |
| 编辑 | `PUT /api/categories/:id`，可编辑字段 JSON | 保存后的 `Row` |
| 删除 | `DELETE /api/categories/:id` | 任意成功结果 |

记录使用小写 `id`，字段名与定义一致。`POST` 对缺省字段应用显式默认值；`PUT` 使用完整替换语义，必须提供所有可编辑的必填字段，缺省或 `null` 的可选字段清空为数据库 `NULL`，响应省略这些字段。必填字段显式传 `null` 返回 400，缺省与显式清空区分处理。数值 `0`、布尔 `false` 均是有效值；只读字段、ID、时间戳和未知字段不能写入请求。只读字段的显式默认值在新增时写入，编辑时保留原值。

查询默认 `page=1`、`pageSize=20`，页大小最大 100、页码最大 1000000；非法范围返回 400。普通字符串按字面包含筛选，静态选项、数字和布尔值按相等筛选，多字段条件以 AND 组合；按 ID 倒序稳定分页。查询字段生成索引，包含匹配和总数统计仍需按实际数据量评估。所有查询参数化，更新使用字段白名单与事务，不存在的记录不会自动新增。

数字 ID 由数据库生成，限制在 JS 安全整数范围；超过范围的新增事务回滚。字符串 ID 使用 UUID。缺失记录返回 HTTP 404，非法参数返回 400，超过 1 MiB 的请求体返回 413；数据库错误记录到 Zap，响应返回通用 500 信息。响应仍统一为 `{code, data, message}`。

按钮使用 `<前缀>:create`、`:update`、`:delete`、`:query`。生成器不添加管理员绕过，权限码必须由数据库菜单行承载：`lambo db seed` 生成的四条菜单已覆盖这些码，额外动作权限需在菜单管理页补 `button` 菜单。路由 `/business/<名称>` 对应菜单 `path` 与组件 `business/<名称>/index.vue`；Vben 以 `accessMode: 'backend'` 运行，侧边栏与路由完全由 `/api/menu/all` 返回的数据库菜单决定，`router/routes/modules/business/*.ts` 只是生成痕迹，不参与实际路由。

## 后端装配与数据库

后端使用独立注入的 GORM 连接。主服务的 `platform/database.NewDb` 已接入 PostgreSQL，迁移清单由装配层给出；Wire 已装配示例分类 CRUD 及账号权限路由。配置和初始化步骤见 [后端 README](../backend/README.md)。生成器只输出通用 GORM 组件，不替其他应用选择数据库或自动修改迁移入口。

后端产物按 `domain` 落在两个平行区：`backend/internal/business/<层>/`（业务）与 `backend/internal/system/<层>/`（系统）。两区都按 `model`、`repository`、`service`、`handler` 分目录，包名与层名一致，同层内的 import 路径由域前缀区分：`internal/business/model` 与 `internal/system/model`。生成器不会输出到 `internal/platform/`，该目录只放数据库连接与 HTTP 底座。

以生成的 `category`（`domain: business`）为例，在已有应用初始化入口中显式装配，或把相同构造函数加入现有 Wire provider set：

```go
// db 是应用已初始化的 GORM 连接，logger 沿用现有日志实例。
if err := db.AutoMigrate(&model.Category{}); err != nil {
    return err
}
repo := repository.NewCategoryRepository(repository.NewRepository(logger, db))
svc := service.NewCategoryService(service.NewService(logger), repo)
categoryHandler := handler.NewCategoryHandler(handler.NewHandler(logger), svc)
if err := categoryHandler.RegisterRoutes(router, authorize); err != nil {
    return err
}
```

`router` 使用应用根路由器，生成方法按字段定义中的完整 API 路径注册。`authorize` 的签名为 `func(*gin.Context, string) bool`，由现有认证/权限模块根据当前身份与权限码返回结果。未提供该函数时拒绝注册；返回 `false` 时中止请求并返回 403。列表和详情检查 `:query`，其余操作分别检查 `:create`、`:update`、`:delete`。

前端 `web-ele` 当前已关闭 Mock，登录、账号信息、权限码、菜单和分类均使用真实后端。开发环境 API baseURL 为 `/`，Vite 将 `/api` 原样代理到 `http://127.0.0.1:8000`；反向代理也需保留该路径。字段定义中的 `api` 是完整路径，请求为 `/api/categories`，不要重复添加 `/api` 或移除后端所需的前缀。需要直接连接其他后端时，可在 `.env.development.local` 设置 `VITE_GLOB_API_URL` 为对应 origin。

新模块需要显式加入迁移、Wire 和受会话保护的路由组。在 Wire 的 `provideMenuCatalog` 中加入该模块生成的 `<模块>MenuRegistration()`，再使用同一版本的后端源码运行菜单种子。当前用户、角色、目录/页面/按钮授权通过数据库配置，管理模块使用生成基线扩展，见 [IAM 契约](iam.md)。按部门的数据范围、计算字段和状态流转仍需接入具体业务规则。

生成器行为验证使用隔离 SQLite 项目；主服务集成另外使用 独立 PostgreSQL 测试库完成真实登录、动态菜单、权限拒绝、浏览器页面增删改查及重启持久化验证。这不代表生产迁移与部署已完成。

## 模板与更新策略

模板内嵌于 CLI，无需联网生成。`tools/lambo/tpl/create/*.tpl` 是后端模板，`tools/lambo/tpl/frontend/*.tpl` 是前端模板。使用 `--tpl-path` 时保留这两个子目录，可只提供需要替换的文件，其余回落到内嵌模板。

生成器先渲染并检查整批文件，拒绝与其他 Go 文件的包级名称或生成 API 路径冲突；内容一致则跳过，有冲突则整批停止；新文件以独占创建方式写入，写入失败撤销本批新文件。拒绝经过符号链接的目标路径。当前不尝试自动合并手写区，也不提供 `--force`：修改后的代码（含自动格式化）需要人工合并，缺少或不同的产物会让 `--check` 失败。

## 新建、运行与 Wire

- `lambo new shop --module github.com/example/shop`：从 Lambo 公共仓库克隆前后端，替换 CLI 和后端模块；拒绝已有目录，不自动装依赖、初始化新 Git 历史或启动服务。`-r` 可以指定本地或其他 Lambo 模板仓库；目标必须包含前端、`backend/go.mod` 和 `tools/lambo/go.mod`。
- `lambo run`：默认在 `backend/` 发现主程序，监控后端文件；例如 `lambo run backend/cmd/server -- -conf config/local.yml`。可使用 `--excludeDir`、`--includeExt`、`--buildFlags`。自动重载时先退出旧进程组，再启动新进程。
- `lambo wire` / `lambo wire all`：默认只扫描后端，以固定 Wire `v0.7.0` 生成，不依赖全局工具版本。
- `lambo upgrade`：更新 Lambo CLI，不升级业务项目或后端依赖；新增 CLI 源码发布到远端后，此命令才可获取。
