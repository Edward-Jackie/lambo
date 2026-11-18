# Docker 部署

本目录统一维护 Go 后端与 Vben Element Plus 前端的镜像和 Compose 配置。默认只发布一个 HTTP 入口，后端通过容器网络访问；不启动公共 Mock，也不安装开发 CLI。

主服务已有 PostgreSQL、真实账号登录、分类权限和分类 CRUD；运行前必须提供私有数据库与初始化配置。完整权限后台、字典后台及生产迁移仍待实现，本机页面联调不等于容器部署验证。

## 构建与运行

需要 Docker Engine、BuildKit 和 Docker Compose v2.20+。所有命令从仓库根目录执行：

```bash
# 可选：调整端口、镜像名和后端配置文件；.env 不提交
cp deploy/.env.example deploy/.env

# 只校验配置，不启动容器
docker compose -f deploy/compose.yaml config --quiet

# 构建镜像，不启动业务服务
docker compose -f deploy/compose.yaml build

# 由开发者或运维明确执行启动
docker compose -f deploy/compose.yaml up -d --wait
```

默认访问 `http://127.0.0.1:8080`。远程部署时按实际入口调整 `HTTP_BIND` 和 `HTTP_PORT`；HTTPS 由外部网关终止，本配置不生成或管理域名证书。

```bash
# 前端静态服务和 Go HTTP 进程的健康检查
curl --fail http://127.0.0.1:8080/healthz
curl --fail http://127.0.0.1:8080/api/healthz

docker compose -f deploy/compose.yaml ps
docker compose -f deploy/compose.yaml logs --tail=100 -f
```

启动时会检查 PostgreSQL，并按配置执行迁移和初始化；`/healthz` 本身仅表示进程可响应，不持续检查数据库或鉴权。Compose 等待后端健康后再启动前端，不会替代后续业务验收。

## 文件与配置

| 文件 | 用途 |
| --- | --- |
| `backend.Dockerfile` | Go 多阶段编译，运行镜像不含 Go 工具和 CLI，使用 UID/GID 10001 |
| `frontend.Dockerfile` | 固定 pnpm、锁文件安装、内部包编译、类型检查，只构建 `@vben/web-ele`，以非 root Nginx 提供静态资源 |
| `*.Dockerfile.dockerignore` | 按镜像限制根目录构建上下文，排除依赖、产物、本地配置、工具和无关服务代码 |
| `compose.yaml` | 镜像构建、容器网络、健康依赖、端口、日志卷、只读根文件系统和退出宽限时间 |
| `backend.yml` | 公开的容器基础配置，固定容器端口 8000，日志写 stdout 和 `/app/storage/logs/` |
| `frontend.env` | 构建期公开配置：Lambo 标题、同域 API、history 路由、关闭 Mock 和额外压缩包 |
| `nginx.conf` | 静态资源、SPA 回退、API 代理、WebSocket/SSE 转发与 Docker DNS 重解析 |
| `.env.example` | Compose 运行参数示例，复制为 `.env` 后按环境修改 |

后端通过既有 `APP_CONF` 加载配置。自定义文件可放 `deploy/backend.local.yml`，并在 `.env` 中设置 `BACKEND_CONFIG_FILE=./backend.local.yml`；此类本地配置已加入忽略规则，不会进入构建上下文。绑定文件必须已存在，并对容器 UID 10001 可读；其中 HTTP 端口仍需为 8000。默认公开配置不包含数据库 DSN 或账号密码；需在私有配置中填写 `data.postgres.dsn`。首次测试初始化可启用 `data.auto_migrate`、`bootstrap.enabled` 并提供管理员和只读账号密码，详见 [后端配置](../backend/README.md)。公开配置默认关闭迁移和账号初始化。

`frontend.env` 会由构建器复制为 `apps/web-ele/.env.production.local`。Vben 的全局配置插件直接读取环境文件，因此仅给运行中的 Nginx 设置 `VITE_*` 不会生效，修改此文件后需要重建前端镜像。前端配置会下发浏览器，不存放服务端密钥。

## API 与缓存约定

浏览器统一使用同域 `/api` 入口。Nginx 保留完整 `/api/` 路径交给 Go，例如 `/api/categories?page=1` 对应后端同名路径；API 的 401、404 和 5xx 不会回退成 `index.html`。后续业务接口与生成器的 `api` 定义需使用相同的后端路径。

`/api/healthz` 单独代理到后端 `/healthz`，仅用于进程连通性验证。登录、菜单、权限码与分类接口均由真实主服务提供，客户端 baseURL 为 `/`，不会重复拼接 `/api`。反向代理关闭接口缓存、缓冲和自动重试；静态版本文件长期缓存，入口 HTML 和 `_app-config*.js` 不长期缓存，SPA 深层路径可回退到首页。

## 更新与停止

```bash
# 从当前源码重建并更新；不会删除日志卷
docker compose -f deploy/compose.yaml build
docker compose -f deploy/compose.yaml up -d --wait --remove-orphans

# 停止，保留日志卷
docker compose -f deploy/compose.yaml down
```

默认镜像名为 `lambo-backend:local`、`lambo-frontend:local`。需要发布镜像时先在 `.env` 指定带仓库和不可变版本号的 `BACKEND_IMAGE`、`FRONTEND_IMAGE`，由运维执行 `docker compose push`；本配置不会自动发布镜像或清理旧版本。使用既有镜像的服务器可执行 `docker compose pull` 后 `docker compose up -d --no-build --wait`，服务器仍需要这份 Compose 和后端配置文件。

日志同时保留容器日志轮转和后端文件轮转。Compose 不包含数据库、Redis 或备份服务；PostgreSQL 通过私有 DSN 连接外部实例。生产迁移、备份及恢复流程仍需另行落实。

## 验证边界

本轮已验证 Compose 配置解析、后端健康接口、Linux amd64/arm64 编译、前端同等构建流程及产物 API 配置。当前机器 Docker 引擎未运行，尚未完成镜像构建、Nginx 容器语法检查和实际容器网络验证；需要在 Docker 可用的环境执行上述 `build` 与健康检查。

配置语义参考：[构建上下文与专用忽略文件](https://docs.docker.com/build/concepts/context/)、[Compose 健康依赖](https://docs.docker.com/compose/how-tos/startup-order)、[Nginx 代理路径规则](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_pass)。
