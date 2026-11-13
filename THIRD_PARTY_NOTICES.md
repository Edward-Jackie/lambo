# 上游来源与许可

Lambo 使用 MIT 许可。导入源码保留原作者的版权和许可，根目录许可不替代上游声明；第三方依赖继续遵循各自许可。

导入日期：2026-11-13。下列提交是本仓库的导入基线，不表示两个项目已经完成集成。

| 目录 | 上游 | 分支与提交 | 上游许可 |
| --- | --- | --- | --- |
| `frontend/` | https://github.com/vbenjs/vue-vben-admin | `main` · `53b47f8ce99ecea11aa04c9c876a2f18c479838f` | [MIT，Vben](frontend/LICENSE) |
| `backend/` | https://github.com/go-nunu/nunu-layout-basic | `main` · `7eb2e6da206f4f0da89ab1aea68dece2d976c01f` | [MIT，Nunu](backend/LICENSE) |

前端导入时的包版本为 `5.7.0`。后端使用 Nunu Basic 模板；[Nunu CLI](https://github.com/go-nunu/nunu) 是配套开发工具，没有作为运行服务或第三套源码导入。

## 导入与更新方式

- 两个上游通过 `git clone --depth 1` 拉取，随后将受控源码导入 Lambo 单一 Git 仓库，没有使用子模块。
- 初始化机器保留了原始 Git 元数据于根仓库的 `.git/upstream-imports/`，该目录仅供本地追溯，不会发布。
- 后续核对上游变化时，以此处提交为比较基线，在独立临时目录读取差异，再按需要移植并更新本文件。
- 不在 `frontend/` 或 `backend/` 内对上游执行直接覆盖式拉取，也不把上游更新与业务改造混在同一提交中。
- 上游测试、示例、文档及锁文件随基线源码保留；它们的存在不表示 Lambo 已执行或通过这些检查。
