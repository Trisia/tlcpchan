
# ============================================
# 阶段 1: 构建 Go 核心服务 (tlcpchan)
# ============================================
FROM golang:1.21-alpine3.18 AS builder-tlcpchan

WORKDIR /build
COPY tlcpchan/go.mod tlcpchan/go.sum ./
RUN go mod download

COPY tlcpchan/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/tlcpchan .

# ============================================
# 阶段 2: 构建 CLI 工具 (tlcpchan-cli)
# ============================================
FROM golang:1.21-alpine3.18 AS builder-cli

WORKDIR /build
COPY tlcpchan-cli/go.mod tlcpchan-cli/go.sum ./
RUN go mod download

COPY tlcpchan-cli/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/tlcpchan-cli .

# ============================================
# 阶段 3: 构建前端静态资源
# ============================================
FROM node:20-alpine3.18 AS builder-frontend

WORKDIR /build
COPY tlcpchan-ui/package.json tlcpchan-ui/package-lock.json ./
RUN npm ci

COPY tlcpchan-ui/ ./
RUN npm run build

# ============================================
# 最终运行镜像
# ============================================
FROM alpine:3.18

LABEL maintainer="TLCP Channel Team"
LABEL description="TLCP/TLS 协议代理工具"

# 安装必要的工具（ca-certificates 用于 TLS 验证）
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建工作目录和必要的目录结构（使用 /etc/tlcpchan）
WORKDIR /etc/tlcpchan
RUN mkdir -p keystores rootcerts logs ui

# 复制可执行文件到 /etc/tlcpchan/（工作目录）
COPY --from=builder-tlcpchan /build/bin/tlcpchan ./
COPY --from=builder-cli /build/bin/tlcpchan-cli ./

# 复制前端静态资源
COPY --from=builder-frontend /build/ui ./ui/

# 复制 trustedcerts 目录中的所有证书到 rootcerts（支持多种格式）
COPY trustedcerts/ ./rootcerts/

# 将 /etc/tlcpchan 添加到 PATH 环境变量
ENV PATH="/etc/tlcpchan:${PATH}"

# 创建软链接到 /usr/bin/ 以便兼容习惯用法
RUN ln -sf /etc/tlcpchan/tlcpchan /usr/bin/tlcpchan && \
    ln -sf /etc/tlcpchan/tlcpchan-cli /usr/bin/tlcpchan-cli && \
    ln -sf /etc/tlcpchan/tlcpchan-cli /usr/bin/tlcpc

# 暴露端口（仅 30080 和 30443，不再需要 30000）
EXPOSE 30080 30443

# 数据卷挂载点（持久化数据）
# 注意：rootcerts 不使用 volume，因为我们预置了证书
VOLUME ["/etc/tlcpchan/keystores", "/etc/tlcpchan/logs"]

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:30080/api/system/health || exit 1

# 默认启动命令
ENTRYPOINT ["tlcpchan"]
