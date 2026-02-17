
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
# 阶段 3: 构建前端静态资源（输出到 tlcpchan-ui/ui）
# ============================================
FROM node:20-alpine3.18 AS builder-frontend

WORKDIR /build
COPY tlcpchan-ui/web/package.json tlcpchan-ui/web/package-lock.json ./
RUN npm ci

COPY tlcpchan-ui/web/ ./
RUN npm run build

# ============================================
# 阶段 4: 构建 UI 服务
# ============================================
FROM golang:1.21-alpine3.18 AS builder-ui

WORKDIR /build
COPY tlcpchan-ui/go.mod tlcpchan-ui/go.sum ./
RUN go mod download

COPY tlcpchan-ui/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/tlcpchan-ui .

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
COPY --from=builder-ui /build/bin/tlcpchan-ui ./

# 复制前端静态资源 - 从 builder-frontend 的 /ui（即 tlcpchan-ui/ui）
COPY --from=builder-frontend /ui ./ui/

# 复制 trustedcerts 目录中的所有证书到 rootcerts（支持多种格式）
COPY trustedcerts/ ./rootcerts/

# 将 /etc/tlcpchan 添加到 PATH 环境变量
ENV PATH="/etc/tlcpchan:${PATH}"

# 创建软链接到 /usr/bin/ 以便兼容习惯用法
RUN ln -sf /etc/tlcpchan/tlcpchan /usr/bin/tlcpchan && \
    ln -sf /etc/tlcpchan/tlcpchan-cli /usr/bin/tlcpchan-cli && \
    ln -sf /etc/tlcpchan/tlcpchan-ui /usr/bin/tlcpchan-ui && \
    ln -sf /etc/tlcpchan/tlcpchan-cli /usr/bin/tlcpc

# 暴露端口
EXPOSE 30080 30000 30443

# 数据卷挂载点（持久化数据）
# 注意：rootcerts 不使用 volume，因为我们预置了证书
VOLUME ["/etc/tlcpchan/keystores", "/etc/tlcpchan/logs"]

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:30080/api/system/health || exit 1

# 默认启动命令（使用相对路径或 PATH 中的命令）
ENTRYPOINT ["tlcpchan"]
# 默认参数：启动 UI 服务
CMD ["-ui"]
