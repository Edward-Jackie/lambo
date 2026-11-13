# 上游来源与许可

Lambo 使用 MIT 许可。导入源码保留原作者的版权和许可，根目录许可不替代上游声明；第三方依赖继续遵循各自许可。

导入日期：2026-11-13。下列提交是本仓库的导入基线，不表示业务后台已完成联调。

| 目录 | 上游 | 分支与提交 | 上游许可 |
| --- | --- | --- | --- |
| `frontend/` | https://github.com/vbenjs/vue-vben-admin | `main` · `53b47f8ce99ecea11aa04c9c876a2f18c479838f` | [MIT，Vben](frontend/LICENSE) |
| `backend/` | https://github.com/go-nunu/nunu-layout-basic | `main` · `7eb2e6da206f4f0da89ab1aea68dece2d976c01f` | [MIT，Nunu](backend/LICENSE) |
| `backend/tools/lambo/` CLI | https://github.com/go-nunu/nunu | `main` · `f20c775987517384ad2c931222a4bbd7b2dffbdb` | [MIT，Nunu CLI](backend/tools/lambo/LICENSE) |

前端导入时的根包版本为 `5.7.0`。后端使用 Nunu Basic 模板，已修改为 Lambo 模块路径。后续导入 Nunu CLI 作为开发工具起点，统一命令为 `lambo`，保留和适配热重载、Wire 等实现，并新增 Lambo 的项目初始化与前端生成能力；CLI 不作为业务运行服务。

## 导入与更新方式

- 两个上游通过 `git clone --depth 1` 拉取，随后将受控源码导入 Lambo 单一 Git 仓库，没有使用子模块。
- 初始化机器保留了原始 Git 元数据于根仓库的 `.git/upstream-imports/`，该目录仅供本地追溯，不会发布。
- 后续核对上游变化时，以此处提交为比较基线，在独立临时目录读取差异，再按需要移植并更新本文件。
- 不在 `frontend/` 或 `backend/` 内对上游执行直接覆盖式拉取，也不把上游更新与业务改造混在同一提交中。
- 后端与前端基线的原始许可持续保留；后续功能文档以 Lambo 实际能力为准。CLI 继承的测试在本地适配并验证，是否提交遵循用户指令。测试文件存在不代替实际验证。
