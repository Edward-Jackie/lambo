# Lambo

基于 **Vben Admin + Nunu Basic** 的 Go / Vue 业务脚手架，目标是以轻量单体承载基础管理能力，并打通字段、字典、显式枚举与前后端代码生成。

**当前阶段：上游源码导入与方案规划。** 前后端尚未联调，用户权限、字典、枚举生成等 Lambo 功能尚未实现；Nunu Basic 示例中的数据库初始化与用户查询仍为占位实现。

- [开发计划与待定项](plan.md)
- [工程约定](AGENTS.md)
- [技术决策](docs/architecture/0001-foundation.md)
- [上游提交与版权声明](THIRD_PARTY_NOTICES.md)

## 目录

```text
lambo/
├── frontend/                 # Vben 5.7.0 上游源码，当前保留其多包结构
├── backend/                  # Nunu Basic 上游源码，尚未改为 Lambo 模块路径
├── docs/architecture/        # 已确认技术决策与取舍
├── plan.md                   # 基础功能、改造阶段、验收与未决项
├── AGENTS.md
├── THIRD_PARTY_NOTICES.md
└── LICENSE
```

## 本地开发边界

- 前端导入基线要求 Node.js `^22.18.0 || ^24.12.0`、pnpm `>=11.0.0`，锁定 `pnpm@11.16.0`；依赖应在 `frontend/` 内管理。
- 后端仍保留上游 Go 模块名和依赖，正式 Go 版本及依赖升级在 P0 核定，不能把其历史 `go 1.16` 声明当作 Lambo 的运行基线。
- 上游配置中的账号、数据库地址、JWT 密钥是公开演示值，尚未成为 Lambo 配置约定，不可用于真实部署。
- 本次没有安装依赖、启动服务、执行前后端构建或验证真实数据库连接。
- 后续先完成 [P0](plan.md#p0-工程基线与依赖治理)，再接入选定的 Vben 应用与真实 Go 接口。

## 许可

[MIT](LICENSE)。保留 Vben、Nunu 及其他依赖的原版权声明，详见 [上游来源与许可](THIRD_PARTY_NOTICES.md)。
