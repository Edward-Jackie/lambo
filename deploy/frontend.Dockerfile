# syntax=docker/dockerfile:1

ARG NODE_IMAGE=node:24.22.3-bookworm-slim
ARG NGINX_IMAGE=nginx:1.30.6-alpine3.24

FROM ${NODE_IMAGE} AS build
ENV CI=true \
    PNPM_HOME=/pnpm \
    PATH=/pnpm:$PATH \
    NODE_OPTIONS=--max-old-space-size=8192
WORKDIR /app

RUN npm install --global pnpm@11.16.0 \
    && pnpm config set store-dir /pnpm/store --global
COPY frontend/ ./
# Vben 的全局配置插件从文件读取配置，仅设置构建环境变量并不生效。
COPY deploy/frontend.env ./apps/web-ele/.env.production.local
RUN --mount=type=cache,id=lambo-pnpm-store,target=/pnpm/store,sharing=locked \
    pnpm install --frozen-lockfile --ignore-scripts --filter @vben/web-ele... --filter @vben/vite-config...
RUN pnpm -r --filter @vben/vite-config... --filter @vben-core/shared --if-present run stub \
    && pnpm --filter @vben/web-ele run typecheck \
    && pnpm --filter @vben/web-ele run build

FROM ${NGINX_IMAGE} AS runtime
COPY deploy/nginx.conf /etc/nginx/nginx.conf
COPY --from=build /app/apps/web-ele/dist/ /usr/share/nginx/html/
COPY LICENSE /usr/share/licenses/lambo/LICENSE
COPY frontend/LICENSE /usr/share/licenses/lambo/vben.LICENSE

USER nginx
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD ["wget", "-q", "-T", "2", "-O", "/dev/null", "http://127.0.0.1:8080/healthz"]
# 直接运行 Nginx，避免上游入口脚本修改只读的配置目录。
ENTRYPOINT ["nginx"]
CMD ["-g", "daemon off;"]
