# ============================================
# 阶段 1: 构建 Go 核心服务 和 CLI 工具（合并构建）
# ============================================
FROM golang:1.26.0-alpine3.23 AS builder-go

# 复制所有源码
COPY tlcpchan/ /tlcpchan/
COPY tlcpchan-cli/ /tlcpchan-cli/

# 编译 tlcpchan（二进制服务）
WORKDIR /tlcpchan
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o tlcpchan .

# 编译 tlcpchan-cli（命令行工具）
WORKDIR /tlcpchan-cli
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o tlcpchan-cli .

# ============================================
# 阶段 2: 构建前端静态资源
# ============================================
FROM node:24.14.0-alpine3.23 AS builder-frontend

WORKDIR /tlcpchan-ui
COPY tlcpchan-ui/ /tlcpchan-ui
RUN npm ci
RUN npm run build

# ============================================
# 最终运行镜像
# ============================================
FROM alpine:3.23

LABEL maintainer="TLCP Channel Team"
LABEL description="TLCP/TLS 协议代理工具"

# 安装必要的工具（ca-certificates 用于 TLS 验证）
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建工作目录和必要的目录结构（使用 /etc/tlcpchan）
WORKDIR /etc/tlcpchan
RUN mkdir -p keystores rootcerts logs ui

# 复制编译好的二进制文件到 /etc/tlcpchan/
COPY --from=builder-go /tlcpchan/tlcpchan ./
COPY --from=builder-go /tlcpchan-cli/tlcpchan-cli ./

# 复制前端静态资源
COPY --from=builder-frontend /tlcpchan-ui/ui ./ui/

# 复制 trustedcerts 目录中的所有证书到 rootcerts（支持多种格式）
COPY trustedcerts/ ./rootcerts/

# 将 /etc/tlcpchan 添加到 PATH 环境变量
ENV PATH="/etc/tlcpchan:${PATH}"

# 创建软链接到 /usr/bin/ 以便兼容习惯用法
RUN ln -sf /etc/tlcpchan/tlcpchan /usr/bin/tlcpchan && \
    ln -sf /etc/tlcpchan/tlcpchan-cli /usr/bin/tlcpchan-cli && \
    ln -sf /etc/tlcpchan/tlcpchan-cli /usr/bin/tlcpc

# 暴露端口（仅 20080 和 20443，不再需要 30000）
EXPOSE 20080 20443

# 数据卷挂载点（持久化数据）
# 注意：rootcerts 不使用 volume，因为我们预置了证书
VOLUME ["/etc/tlcpchan/keystores", "/etc/tlcpchan/logs"]

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:20080/api/system/health || exit 1

# 默认启动命令
ENTRYPOINT ["tlcpchan"]
