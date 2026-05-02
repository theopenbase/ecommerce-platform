# ============================================================
# 电商平台 — 前端 (Node.js / Vite + React)
# 多阶段构建，优化体积
# ============================================================
# 基础阶段：依赖安装
# 构建阶段：代码构建
# 运行阶段：最小化运行时镜像
# ============================================================

# ── Stage 1: Dependencies ──────────────────────────────────
FROM node:20-alpine AS deps
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production && npm cache clean --force

# ── Stage 2: Build ──────────────────────────────────────────
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
COPY --from=deps /app/node_modules ./node_modules
COPY . .
ENV NODE_ENV=production
RUN npm run build

# ── Stage 3: Runtime (nginx) ────────────────────────────────
FROM nginx:1.25-alpine AS runner

# 安全加固：非 root 运行
RUN addgroup -g 1001 -S nginx && adduser -S nginx -u 1001

# 复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制 nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 开放端口
EXPOSE 80

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost/health || exit 1

USER nginx
CMD ["nginx", "-g", "daemon off;"]
