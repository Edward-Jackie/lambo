# Lambo 工程约定

[README.md](README.md) 是安装、初始化、使用和开发流程的入口。功能范围及验收按 [plan.md](plan.md)，技术决策见 [架构说明](docs/architecture/0001-foundation.md)。

## 代码边界

- 业务前端使用 `frontend/apps/web-ele`，复用 Vben 和 Element Plus 的表格、表单、图标、按钮及主题变量。联调连接真实 API 和 PostgreSQL。
- `backend/internal/` 分为 `platform`、`system` 和 `business`。系统与业务均按 `model`、`repository`、`service`、`handler` 平铺。业务可引用系统，两者可引用平台和 `pkg`。平台和 `pkg` 保持独立。
- 在 `backend/cmd/server/wire/` 装配路由、迁移清单和菜单目录。业务授权、状态流转、会话撤销和最后管理员保护由后端执行，系统规则由 `system` 维护。
- `tools/lambo/` 是独立的 CLI Go 模块。生成位置由字段 JSON 的 `domain` 决定。系统骨架和业务菜单种子采用各自的初始化入口，业务种子复用 IAM 校验。
- Docker 配置集中在 `deploy/`，构建上下文为仓库根目录。运行镜像仅包含选定的业务前端和服务端，`/healthz` 用于检查进程可达。

## 修改与验证

- 修改前参考 1–2 个同类实现。先确定字段、权限与业务动作，再生成起始代码并保留手写修改。新增和修改的注释使用简体中文。
- 后端按范围执行测试、构建和静态检查。前端执行类型检查与构建，交互变更补查加载、空态、错误、权限和刷新。
- 模板变更验证实际生成产物、重复生成、冲突拒绝和路径边界。权限、迁移、并发和状态流转覆盖必要的失败路径。汇报区分源码检查、数据库联调、浏览器验收和实际部署。
- 本地验收产物放在 `tools/lambo/.artifacts/`。公开配置只使用演示值，凭据、运行数据和构建产物留在本地。保留上游来源和许可证。
