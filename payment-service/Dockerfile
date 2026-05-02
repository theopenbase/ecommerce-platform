# ============================================================
# 电商平台 — 后端服务 (Go)
# 多阶段构建，优化体积
# ============================================================
# Stage 1: Builder — 编译
# Stage 2: Runtime — 最小化运行时
# ============================================================

# ── Stage 1: Builder ───────────────────────────────────────
FROM golang:1.22-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# 依赖缓存层
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 编译（关闭 CGO，使用静态二进制）
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

ARG GIT_SHA=unknown
ARG BUILD_TIME=unknown
ARG SERVICE_NAME=unknown

LABEL git_sha=${GIT_SHA}
LABEL build_time=${BUILD_TIME}

RUN go build -ldflags=" \
    -s -w \
    -X main.GitSHA=${GIT_SHA} \
    -X main.BuildTime=${BUILD_TIME} \
    -X main.ServiceName=${SERVICE_NAME}" \
    -o /app ./cmd

# ── Stage 2: Runtime ───────────────────────────────────────
FROM gcr.io/distroless/static-debian12 AS runner

# 时区数据
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 复制二进制
COPY --from=builder /app /app

WORKDIR /app

# 非 root 运行
USER nonroot

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

EXPOSE 8080

ENTRYPOINT ["/app"]
