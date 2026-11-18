# Lambo

基于 Vben Admin、Element Plus 和 Nunu Basic 的 Go / Vue 业务脚手架。内置 PostgreSQL、账号会话、角色与菜单管理，以及接入真实接口的分类 CRUD。使用 `lambo` 生成前后端起始代码、管理菜单种子、生成 Wire 装配和运行后端热重载。

## 本地使用

需要 Go `1.25.8+`、PostgreSQL、Node.js `^22.18.0 || ^24.12.0` 和 `pnpm@11.16.0`。以下命令从仓库根目录执行。

```bash
go -C tools/lambo install ./cmd/lambo
lambo --help
(cd frontend && pnpm install --frozen-lockfile)
```

将 `go env GOBIN` 指向的目录加入 `PATH`。未指定 `GOBIN` 时使用 `$(go env GOPATH)/bin`。

1. 创建本地开发数据库，将 [后端开发配置](backend/config/local.yml) 复制到仓库外的私有路径，填写 `data.postgres.dsn`。首次启动启用 `data.auto_migrate` 和 `bootstrap.enabled`，填写管理员和只读账号的密码，每个密码为 12–72 字节。
2. 在两个终端分别运行后端和前端。后端默认端口为 8000，前端默认地址为 [http://localhost:5777](http://localhost:5777)，开发代理转发 `/api`。
3. 使用配置中指定的账号登录。完成初始化后关闭 `bootstrap.enabled`。数据库配置、环境变量和 HTTP 接口见 [后端说明](backend/README.md)。

```bash
APP_CONF=/绝对路径/lambo.local.yml lambo run
(cd frontend && pnpm --filter @vben/web-ele run dev)
```

### 使用 init.sql 初始化基础数据

本地空库默认由首次启动完成建表、账号和 IAM 初始化。需要手工装入基础数据时，在 **已建表的新库** 中执行 [docs/init.sql](docs/init.sql)，再按上面的账号配置启动后端。

```bash
psql -X -v ON_ERROR_STOP=1 -d lambo_dev -f docs/init.sql
```

`init.sql` 写入管理员与只读角色以及目录、页面骨架。首次启动继续创建账号、按钮和授权。脚本有事务锁和安装标记，对已完成基础初始化的库重复执行会保留数据。表结构迁移与业务模块菜单分别走各自的入口，详见 [IAM 说明](docs/iam.md)。

## 开发流程

1. **定义并验收需求**。确认角色、成功与失败场景，梳理业务流程。定义 Domain 契约中的字段、枚举、动作、状态、权限和数据范围。派生字段 Schema，审核通过后冻结第一版。完整检查标准见 [开发规范](plan.md#121-流程与阶段审核)。
2. **预览并生成**。先运行 `lambo create all <模块> --schema <字段文件> --dry-run`，确认产物和修改范围后去掉 `--dry-run`。可分别使用 `create backend`、`create frontend` 或单层命令。字段格式和选项见 [生成器说明](docs/generator.md)，`--check` 可核对产物与模板的一致性。
3. **接入真实业务**。补齐领域规则和数据范围，在 Wire 的迁移清单和会话路由中加入模块，并在 `provideMenuCatalog` 注册生成的 `<模块>MenuRegistration()`。运行 `lambo wire` 更新装配。
4. **安装业务菜单**。运行 `lambo db seed <模块> --dry-run` 查看生成的第三版种子，再执行 `lambo db seed <模块>`。命令使用当前项目后端的注册目录校验，在同一事务安装菜单、按钮和默认管理员授权。对安装历史不完整的旧库，仅在核实模块从未安装后使用 `--new`。DSN 依次来自 `--dsn`、`LAMBO_POSTGRES_DSN`、`backend/config/local.yml` 中的 `data.postgres.dsn`。
5. **验证并交付**。运行下列验证，在真实开发库和页面验证操作成功、越权拒绝与持久化。生成器产物有手写修改时先预览再合并。完成后记录验证证据及剩余边界，更新相关 Codemap 源码锚点。

```bash
go -C backend test ./...
go -C backend vet ./...
go -C backend build ./...
go -C tools/lambo test ./...
(cd frontend && pnpm --filter @vben/web-ele run typecheck)
(cd frontend && pnpm --filter @vben/web-ele run build)
```

### 常用命令

| 命令 | 用途 |
| --- | --- |
| `lambo new shop --module github.com/example/shop` | 拉取 Lambo 模板并替换两个 Go 模块路径 |
| `lambo create all category --schema tools/lambo/examples/category.json --check` | 检查现有分类模块的生成产物 |
| `lambo create handler <模块>` | 只生成后端处理器 |
| `lambo run backend/cmd/server -- -conf config/local.yml` | 运行热重载并透传后端配置参数 |
| `lambo wire all` | 用固定版本的 Wire 更新后端装配 |
| `lambo upgrade` | 更新 CLI，业务项目仍由自己的仓库维护 |

## 部署与项目文档

- [Docker 部署](deploy/README.md) 包含镜像构建、外部 PostgreSQL 配置、Compose 运行与验证边界。
- [工程约定](AGENTS.md) 描述代码边界和验收规则。[架构说明](docs/architecture/0001-foundation.md) 记录技术决策。
- [项目计划](plan.md) 记录阶段进度与验收。部门数据范围、动态字典、业务状态流转和生产版本化迁移仍在计划中。

主要目录为 `backend/` 业务服务、`frontend/` 前端工作区、`tools/lambo/` CLI 与模板、`deploy/` 部署配置和 `docs/` 技术文档。

[MIT 许可证](LICENSE) · [上游来源与版权](THIRD_PARTY_NOTICES.md)
