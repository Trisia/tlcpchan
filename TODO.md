# TLCP Channel 开发进度

## 最后更新时间
2026-02-13

## 项目概述
TLCP Channel 是一款 TLCP/TLS 协议代理工具，支持双协议并行工作。

## 已完成任务 ✅

### 文档 (100%)
- [x] 设计文档 (`docs/design.md`)
- [x] 需求文档 (`docs/requirements.md`)
- [x] API接口文档 (`docs/api.md`)
- [x] 使用指南 (`docs/README.md`)
- [x] 安装指南 (`docs/installation.md`)
- [x] 证书管理文档 (`docs/certificates.md`)
- [x] API使用指南 (`docs/api-usage.md`)
- [x] 配置示例 (`docs/config-examples.md`)

### Go核心模块 (100%)
- [x] 配置管理 (`tlcpchan/config/`)
  - YAML配置解析
  - 密码套件/版本字符串解析
  - 配置验证
- [x] 日志管理 (`tlcpchan/logger/`)
  - 多级别日志
  - 文件+控制台输出
  - 日志轮转
- [x] 证书管理 (`tlcpchan/cert/`)
  - PEM证书加载
  - SM2/RSA证书支持
  - 证书热更新
  - 证书生成器
- [x] 代理核心 (`tlcpchan/proxy/`)
  - 服务端代理 (TLCP/TLS → TCP)
  - 客户端代理 (TCP → TLCP/TLS)
  - HTTP代理
  - 协议自动适配
- [x] 实例管理 (`tlcpchan/instance/`)
  - 实例创建/删除
  - 启动/停止/重载
  - 状态管理
- [x] API控制器 (`tlcpchan/controller/`)
  - RESTful API
  - 实例管理API
  - 配置管理API
  - 证书管理API
  - 系统信息API
- [x] 流量统计 (`tlcpchan/stats/`)
  - 连接/流量统计
  - 延迟统计
  - 历史快照

### CLI工具 (100%)
- [x] 命令行工具 (`tlcpchan-cli/`)
  - 实例管理命令
  - 配置管理命令
  - 证书管理命令
  - 系统命令

### 部署支持 (80%)
- [x] systemd服务文件
- [x] 安装/卸载脚本
- [x] GitHub CI/CD配置
- [x] 默认配置文件

### UI静态资源服务器 (100%)
- [x] Go静态资源服务器 (`tlcpchan-ui/`)
  - SPA路由支持
  - API代理

### 测试 (100%)
- [x] 配置模块测试
- [x] 证书模块测试
- [x] 统计模块测试
- [x] 代理模块测试

## 待完成任务 ⏳

### 前端UI (0%)
- [ ] Vue3项目初始化
- [ ] 仪表盘页面
- [ ] 实例管理页面
- [ ] 证书管理页面
- [ ] 日志查看页面
- [ ] 系统设置页面

### Docker支持 (0%)
- [ ] Dockerfile
- [ ] docker-compose.yml

### 打包发布 (0%)
- [ ] .deb 安装包
- [ ] .rpm 安装包
- [ ] .msi 安装包
- [ ] .pkg 安装包

## 项目结构

```
tlcpchan/
├── .github/              # GitHub CI/CD
│   └── workflows/
│       ├── ci.yml
│       └── release.yml
├── docs/                 # 文档
│   ├── design.md
│   ├── requirements.md
│   ├── api.md
│   ├── README.md
│   ├── installation.md
│   ├── certificates.md
│   ├── api-usage.md
│   └── config-examples.md
├── tlcpchan/            # 主程序
│   ├── main.go
│   ├── config/          # 配置管理
│   ├── logger/          # 日志管理
│   ├── cert/            # 证书管理
│   ├── proxy/           # 代理核心
│   ├── instance/        # 实例管理
│   ├── controller/      # API控制器
│   ├── stats/           # 流量统计
│   └── release/         # 发布文件
├── tlcpchan-cli/        # CLI工具
│   ├── main.go
│   ├── client/
│   └── commands/
└── tlcpchan-ui/         # UI服务
    ├── main.go
    ├── server/
    └── proxy/
```

## 编译命令

```bash
# 编译主程序
cd tlcpchan && go build -o bin/tlcpchan .

# 编译CLI工具
cd tlcpchan-cli && go build -o bin/tlcpchan-cli .

# 编译UI服务
cd tlcpchan-ui && go build -o bin/tlcpchan-ui .

# 运行测试
cd tlcpchan && go test ./...
```

## 下次启动待办

1. 完成前端UI开发
   - 初始化Vue3项目
   - 实现仪表盘
   - 实现实例管理界面

2. 创建Docker支持
   - 编写Dockerfile
   - 编写docker-compose.yml

3. 完善打包发布
   - 配置goreleaser
   - 生成各平台安装包

## 注意事项

- 配置文件路径: `./config/config.yaml`
- 证书目录: `./certs/tlcp/` 和 `./certs/tls/`
- 日志目录: `./logs/`
- API默认端口: 8080
- UI默认端口: 3000
