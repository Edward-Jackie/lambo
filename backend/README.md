# Lambo 后端

模块：`github.com/Edward-Jackie/lambo/backend`。沿用 Gin、GORM、Viper、Zap 和 Wire 的显式分层，源自 Nunu Basic；原版权保留于 [LICENSE](LICENSE)。

统一命令从项目根目录安装，详见 [根目录 README](../README.md) 和 [代码生成器说明](../docs/generator.md)：

```bash
# 在仓库根目录执行
go install ./cmd/lambo
lambo create backend category --schema examples/category.json
lambo create all category --schema examples/category.json
lambo wire
lambo run
```

后端组件按 `internal/model → repository → service → handler` 分层生成。生成后的组件需补齐业务实现、Wire 装配和路由；当前模板的真实数据库初始化及用户查询仍待实现，不能直接作为完整后台部署。

本目录可独立执行 `go build ./...` 和 `go vet ./...`。命令中的 `run` 只在开发者明确需要运行时使用。
