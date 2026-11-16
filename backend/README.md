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

后端组件按 `internal/model → repository → service → handler` 分层生成，包含字段校验、分页/详情/新增/编辑/删除及权限路由注册方法。连接数据库并迁移表后，显式装配或加入 Wire，再调用处理器的 `RegisterRoutes(router, authorize)`；具体步骤见生成器说明。项目现有数据库初始化及用户查询仍待实现，认证与数据范围需接入业务规则。

本目录可独立执行 `go build ./...` 和 `go vet ./...`。命令中的 `run` 只在开发者明确需要运行时使用。
