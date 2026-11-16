# Lambo 后端

模块：`github.com/Edward-Jackie/lambo/backend`。沿用 Gin、GORM、Viper、Zap 和 Wire 的显式分层，源自 Nunu Basic；原版权保留于 [LICENSE](LICENSE)。

统一命令从项目根目录安装，详见 [根目录 README](../README.md) 和 [代码生成器说明](../docs/generator.md)：

```bash
# 在仓库根目录执行
go -C tools/lambo install ./cmd/lambo
lambo create backend category --schema tools/lambo/examples/category.json
lambo create all category --schema tools/lambo/examples/category.json
lambo wire
lambo run
```

后端组件按 `internal/model → repository → service → handler` 分层生成，包含字段校验、分页/详情/新增/编辑/删除及权限路由注册方法。连接数据库并迁移表后，显式装配或加入 Wire，再调用处理器的 `RegisterRoutes(router, authorize)`；具体步骤见生成器说明。主服务已通过 Wire 装配分类 CRUD、PostgreSQL、真实账号会话及权限校验；旧用户示例不再公开注册。其他业务模块和数据范围仍需显式接入。

本目录可独立执行 `go build ./...` 和 `go vet ./...`。命令中的 `run` 只在开发者明确需要运行时使用。

## 真实接口运行配置

连接位置为 `config/local.yml` 的 `data.postgres.dsn`，可由 `LAMBO_POSTGRES_DSN` 覆盖；也可通过 `APP_CONF` 指向私有 YAML/JSON 文件。DSN、初始化密码与运行产物不得提交。本机联调配置保存在 Git 忽略的 `tools/lambo/.artifacts/main-service/runtime.json`（权限 0600）；使用独立 PostgreSQL 测试库及非超级用户账号，不接入 Redis。

首次初始化需要 `data.auto_migrate: true` 和 `bootstrap.enabled: true`，并在私有配置中提供 `bootstrap.admin_password`、`bootstrap.reader_password`（12—72 字节）。用户名默认 `admin`、`reader`，也可配置；已有账号不会被重复创建或覆盖密码。初始化后关闭 bootstrap；生产迁移需另行管理，公开生产配置默认不自动迁移。迁移只涉及分类、后台账号和会话三张表。

所有初始化项也支持环境变量：`LAMBO_AUTO_MIGRATE`、`LAMBO_BOOTSTRAP_ENABLED`、`LAMBO_ADMIN_USERNAME`、`LAMBO_ADMIN_PASSWORD`、`LAMBO_READER_USERNAME`、`LAMBO_READER_PASSWORD`。例如按需在 `backend/` 执行 `APP_CONF=/绝对路径/私有配置.json go run ./cmd/server`；程序会在连接、迁移、账号初始化和路由装配成功后报告就绪。

账号密码使用 bcrypt；随机访问令牌有效期 8 小时，数据库只保存 SHA-256 摘要，退出登录立即撤销。每次请求检查会话有效期与账号启用状态。管理员显式拥有分类 query/create/update/delete 四个权限，只读账号仅有 query；前端按钮和后端路由使用同一组权限码。

| 接口 | 用途 |
| --- | --- |
| `POST /api/auth/login` | 真实账号登录，返回 accessToken |
| `POST /api/auth/logout` | 幂等撤销当前令牌 |
| `GET /api/user/info` | 当前账号信息 |
| `GET /api/auth/codes` | 当前账号权限码 |
| `GET /api/menu/all` | 当前账号动态菜单 |
| `/api/categories`、`/api/categories/:id` | 生成的五个 CRUD 接口，需 Bearer 令牌及相应权限 |

前端 `@vben/web-ele` 已关闭 Mock，默认 API baseURL 为 `/`；开发代理将完整 `/api/...` 路径交给 `127.0.0.1:8000`。页面入口为 `/business/category`。不提供公开演示密码、注册、短信或刷新令牌接口；完整账号/角色管理后台尚未实现。
