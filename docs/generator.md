# Lambo 代码生成器

## 命令与产物

在根目录 `go install ./cmd/lambo` 后，命令可从项目根目录或任意子目录使用；`--root` 可显式指定另一个 Lambo 项目。字段文件路径和自定义模板路径相对当前工作目录解析。

| 命令 | 产物 |
| --- | --- |
| `lambo create model category` | 后端模型 |
| `lambo create repository category` | 后端仓储起始组件 |
| `lambo create service category` | 后端服务起始组件 |
| `lambo create handler category` | 后端处理器起始组件 |
| `lambo create backend category` | 以上四个组件 |
| `lambo create frontend category` | Vben Element Plus API、字段配置、列表、详情/编辑弹窗与路由，共 5 个文件 |
| `lambo create all category` | 后端 4 个文件 + 前端 5 个文件 |

通用参数：`--schema <文件>`、`--dry-run`、`--check`、`--root <项目>`、`--tpl-path <模板目录>`。`--dry-run` 与 `--check` 互斥。名称采用小写英文、数字及中间的 `-` 或 `_`，不接受任意输出路径。单个后端层可能引用尚未创建的相邻层；需要完整可编译组件时使用 `create backend` 或 `create all`。

## 开发期字段定义

完整示例见 [category.json](../examples/category.json)。这是当前 CLI 的本地输入契约，不代表已经决定字典后台或业务枚举的唯一真源。

| 属性 | 说明 |
| --- | --- |
| `name` | 可省略；提供时必须与命令名称一致 |
| `title` | 中文页面标题，默认使用名称 |
| `api` | 静态 REST 路径，默认 `/api/<名称>`；不自动处理单复数 |
| `permission` | 权限前缀，默认 `business:<名称>` |
| `idType` | `number`（默认）或 `string`；大整数 ID 应以字符串传输 |
| `fields` | 至少一个业务字段，不能重复或覆盖 `id` 等系统字段 |
| `fields[].name / label` | 小驼峰字段名及中文标签 |
| `fields[].type` | `string`、`number`、`boolean` |
| `fields[].required` | 表单必填；`0`、`false` 仍是有效值 |
| `fields[].search` | 在查询表单和分页请求中使用此字段 |
| `fields[].readOnly` | 列表、详情可见，不进入编辑表单或提交体；关键业务状态应设为只读 |
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

记录使用小写 `id`，字段名与定义一致。未填写的可选字段可能省略；需要清空可选字段时，后端应按完整替换约定处理 `PUT`，或调整生成的客户端与业务接口同时使用明确的清空协议。后端分页、数据范围、索引、授权和写入校验需实现后再联调。

按钮使用 `<前缀>:create`、`:update`、`:delete`、`:query`。需要在后端实际下发这些权限码；生成器不会添加管理员绕过，也不自动写入权限表。路由 `/business/<名称>` 是本地路由起点，动态菜单接入后可引用 `business/<名称>/index.vue` 组件。

## 模板与更新策略

模板内嵌于 CLI，无需联网生成。`tpl/create/*.tpl` 是后端模板，`tpl/frontend/*.tpl` 是前端模板。使用 `--tpl-path` 时保留这两个子目录，可只提供需要替换的文件，其余回落到内嵌模板。

生成器先渲染并检查整批文件，内容一致则跳过，有冲突则整批停止；新文件以独占创建方式写入，写入失败撤销本批新文件。拒绝经过符号链接的目标路径。当前不尝试自动合并手写区，也不提供 `--force`：修改后的代码（含自动格式化）需要人工合并，缺少或不同的产物会让 `--check` 失败。

## 新建、运行与 Wire

- `lambo new shop --module github.com/example/shop`：从 Lambo 公共仓库克隆前后端，替换 CLI 和后端模块；拒绝已有目录，不自动装依赖、初始化新 Git 历史或启动服务。`-r` 可以指定本地或其他 Lambo 模板仓库；目标必须包含前端及两个 Go 模块。
- `lambo run`：默认在 `backend/` 发现主程序，监控后端文件；例如 `lambo run backend/cmd/server -- -conf config/local.yml`。可使用 `--excludeDir`、`--includeExt`、`--buildFlags`。自动重载时先退出旧进程组，再启动新进程。
- `lambo wire` / `lambo wire all`：默认只扫描后端，以固定 Wire `v0.7.0` 生成，不依赖全局工具版本。
- `lambo upgrade`：更新 Lambo CLI，不升级业务项目或后端依赖；新增 CLI 源码发布到远端后，此命令才可获取。
