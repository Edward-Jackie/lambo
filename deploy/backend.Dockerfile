# syntax=docker/dockerfile:1

ARG GO_IMAGE=golang:1.27.4-alpine3.24
ARG RUNTIME_IMAGE=alpine:3.24.4

FROM --platform=$BUILDPLATFORM ${GO_IMAGE} AS build
WORKDIR /src

COPY backend/go.mod backend/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked go mod download

COPY backend/ ./
ARG TARGETOS
ARG TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -mod=readonly -trimpath -buildvcs=false -ldflags="-s -w" -o /out/lambo-server ./cmd/server

FROM ${RUNTIME_IMAGE} AS runtime
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 10001 lambo \
    && adduser -S -D -H -u 10001 -G lambo lambo \
    && mkdir -p /app/config /app/storage/logs \
    && chown -R 10001:10001 /app

WORKDIR /app
ENV APP_CONF=/app/config/container.yml \
    TZ=Asia/Shanghai
COPY --from=build --chown=10001:10001 /out/lambo-server ./lambo-server
COPY --chown=10001:10001 deploy/backend.yml ./config/container.yml
COPY LICENSE /usr/share/licenses/lambo/LICENSE
COPY backend/LICENSE /usr/share/licenses/lambo/nunu.LICENSE

USER 10001:10001
EXPOSE 8000
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD ["wget", "-q", "-T", "2", "-O", "/dev/null", "http://127.0.0.1:8000/healthz"]
ENTRYPOINT ["/app/lambo-server"]
