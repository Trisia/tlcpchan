# 证书管理

本文档说明 TLCP Channel 中的证书配置和管理方法。

## 证书类型

### TLCP 证书（SM2）

TLCP 协议使用国密算法，证书特点：

| 项目 | 说明 |
|------|------|
| 公钥算法 | SM2（椭圆曲线） |
| 签名算法 | SM2-SM3 |
| 加密算法 | SM4 |
| 哈希算法 | SM3 |

TLCP 证书通常包含两个证书文件：

```
certs/tlcp/
├── server-sm2.crt    # 签名证书
├── server-sm2.key    # 签名私钥
├── server-sm2.enc    # 加密证书（可选）
└── server-sm2.enckey # 加密私钥（可选）
```

### TLS 证书（RSA/ECC）

标准 TLS 协议证书，支持：

| 算法 | 说明 |
|------|------|
| RSA | 传统 RSA 算法 |
| ECDSA | 椭圆曲线签名算法 |
| Ed25519 | Edwards 曲线算法 |

证书文件结构：

```
certs/tls/
├── server-rsa.crt    # 证书文件
└── server-rsa.key    # 私钥文件
```

## 证书生成

### 自动生成（首次启动）

首次启动时，程序会自动生成测试证书：

```bash
tlcpchan
```

生成的证书包括：

| 证书名称 | 类型 | 用途 |
|----------|------|------|
| `ca-sm2` | TLCP CA | TLCP 根证书 |
| `server-sm2` | TLCP 服务端 | TLCP 服务端证书 |
| `client-sm2` | TLCP 客户端 | TLCP 客户端证书 |
| `ca-rsa` | TLS CA | TLS 根证书 |
| `server-rsa` | TLS 服务端 | TLS 服务端证书 |

### 使用 CLI 生成

```bash
# 生成 TLCP 证书
tlcpchan-cli cert generate -name my-server -type tlcp

# 生成 TLS 证书
tlcpchan-cli cert generate -name my-server -type tls
```

### 使用 API 生成

```bash
curl -X POST http://localhost:8080/api/v1/certificates/generate \
  -H "Content-Type: application/json" \
  -d '{
    "type": "tlcp",
    "name": "my-server",
    "common_name": "example.com",
    "dns_names": ["example.com", "*.example.com"],
    "ip_addresses": ["192.168.1.1"],
    "days": 365,
    "ca_name": "ca-sm2"
  }'
```

### 使用 OpenSSL 生成（TLS）

```bash
# 生成私钥
openssl genrsa -out server.key 2048

# 生成 CSR
openssl req -new -key server.key -out server.csr \
  -subj "/CN=example.com"

# 生成自签名证书
openssl x509 -req -in server.csr -signkey server.key \
  -out server.crt -days 365

# 或使用 CA 签名
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key \
  -CAcreateserial -out server.crt -days 365
```

### 使用国密工具生成（TLCP）

推荐使用 [gmssl](https://github.com/guanzhi/GmSSL) 或项目内置工具：

```bash
# 使用 gmssl 生成 SM2 密钥对
gmssl sm2 -genkey -out server.key

# 生成证书签名请求
gmssl sm2 -new -key server.key -out server.csr \
  -subj "/CN=example.com"

# 使用 CA 签发证书
gmssl sm2 -req -in server.csr -CA ca.crt -CAkey ca.key \
  -out server.crt -days 365
```

## 证书配置

### 目录结构

```
/opt/tlcpchan/
├── config/
│   └── config.yaml
├── certs/
│   ├── tlcp/                    # TLCP 证书目录
│   │   ├── ca-sm2.crt
│   │   ├── ca-sm2.key
│   │   ├── server-sm2.crt
│   │   ├── server-sm2.key
│   │   ├── client-sm2.crt
│   │   └── client-sm2.key
│   └── tls/                     # TLS 证书目录
│       ├── ca-rsa.crt
│       ├── ca-rsa.key
│       ├── server-rsa.crt
│       └── server-rsa.key
└── logs/
```

### 文件命名规则

证书通过名称引用，程序会自动查找对应文件：

| 引用名称 | 查找文件 |
|----------|----------|
| `server-sm2` | `server-sm2.crt`, `server-sm2.key` |
| `server-sm2` (TLCP) | `server-sm2.crt`, `server-sm2.key`, `server-sm2.enc`（可选） |

### 配置示例

#### 单向认证（服务端）

```yaml
instances:
  - name: "tlcp-server"
    type: "server"
    protocol: "tlcp"
    auth: "one-way"
    listen: ":443"
    target: "127.0.0.1:8080"
    certificates:
      tlcp:
        cert: "server-sm2"
        key: "server-sm2"
```

#### 双向认证（服务端）

```yaml
instances:
  - name: "tlcp-server"
    type: "server"
    protocol: "tlcp"
    auth: "mutual"
    listen: ":443"
    target: "127.0.0.1:8080"
    certificates:
      tlcp:
        cert: "server-sm2"
        key: "server-sm2"
    client_ca:
      - "ca-sm2"
```

#### 客户端代理

```yaml
instances:
  - name: "tlcp-client"
    type: "client"
    protocol: "tlcp"
    auth: "one-way"
    listen: ":9000"
    target: "server.example.com:443"
    server_ca:
      - "ca-sm2"
```

#### 客户端代理（双向认证）

```yaml
instances:
  - name: "tlcp-client"
    type: "client"
    protocol: "tlcp"
    auth: "mutual"
    listen: ":9000"
    target: "server.example.com:443"
    certificates:
      tlcp:
        cert: "client-sm2"
        key: "client-sm2"
    server_ca:
      - "ca-sm2"
```

#### 双协议自动检测

```yaml
instances:
  - name: "auto-server"
    type: "server"
    protocol: "auto"
    auth: "one-way"
    listen: ":443"
    target: "127.0.0.1:8080"
    certificates:
      tlcp:
        cert: "server-sm2"
        key: "server-sm2"
      tls:
        cert: "server-rsa"
        key: "server-rsa"
```

## 证书热更新

TLCP Channel 支持证书热更新，无需重启服务。

### 方式一：API 更新

```bash
# 替换证书文件后，调用热更新接口
curl -X POST http://localhost:8080/api/v1/certificates/reload
```

响应：

```json
{
  "reloaded": true,
  "updated": ["server-sm2", "client-sm2"]
}
```

### 方式二：CLI 更新

```bash
tlcpchan-cli cert reload
```

### 方式三：信号触发

```bash
# 发送 SIGHUP 信号
kill -HUP <pid>

# 或使用 systemctl
systemctl reload tlcpchan
```

### 更新流程

```
1. 替换证书文件
   ├── 备份旧证书
   ├── 复制新证书到 certs/ 目录
   └── 确保文件权限正确

2. 触发热更新
   ├── API 调用 /certificates/reload
   └── 或发送 SIGHUP 信号

3. 验证更新
   ├── 检查 API 响应
   └── 测试新证书是否生效
```

### 自动化脚本

```bash
#!/bin/bash
# cert-update.sh - 自动更新证书

CERT_DIR="/opt/tlcpchan/certs/tlcp"
BACKUP_DIR="/opt/tlcpchan/certs/backup"
API_URL="http://localhost:8080/api/v1"

# 备份旧证书
mkdir -p "$BACKUP_DIR"
cp "$CERT_DIR"/*.crt "$BACKUP_DIR/"
cp "$CERT_DIR"/*.key "$BACKUP_DIR/"

# 复制新证书（从 Let's Encrypt 或其他来源）
cp /etc/letsencrypt/live/example.com/fullchain.pem "$CERT_DIR/server-sm2.crt"
cp /etc/letsencrypt/live/example.com/privkey.pem "$CERT_DIR/server-sm2.key"

# 触发热更新
curl -X POST "$API_URL/certificates/reload"

echo "证书更新完成"
```

## 使用外部 CA

### Let's Encrypt（TLS）

```bash
# 安装 certbot
sudo apt install certbot

# 申请证书
sudo certbot certonly --standalone -d example.com

# 证书位置
# /etc/letsencrypt/live/example.com/fullchain.pem
# /etc/letsencrypt/live/example.com/privkey.pem

# 复制到 tlcpchan 目录
cp /etc/letsencrypt/live/example.com/fullchain.pem /opt/tlcpchan/certs/tls/server-rsa.crt
cp /etc/letsencrypt/live/example.com/privkey.pem /opt/tlcpchan/certs/tls/server-rsa.key

# 触发热更新
curl -X POST http://localhost:8080/api/v1/certificates/reload
```

### 国密 CA（TLCP）

使用国内认可的 CA 机构签发国密证书：

1. 生成 CSR 文件
2. 提交给 CA 机构
3. 获取签发的证书
4. 配置到 tlcpchan

```bash
# 使用 gmssl 生成 CSR
gmssl sm2 -new -key server.key -out server.csr \
  -subj "/C=CN/ST=Beijing/L=Beijing/O=Example/CN=example.com"

# 将 CSR 提交给 CA 机构
# 获取签发的证书后
cp issued_cert.crt /opt/tlcpchan/certs/tlcp/server-sm2.crt

# 触发热更新
curl -X POST http://localhost:8080/api/v1/certificates/reload
```

### 企业内部 CA

```yaml
# 配置内部 CA 证书作为信任链
instances:
  - name: "internal-proxy"
    type: "client"
    protocol: "tls"
    auth: "mutual"
    # ...
    server_ca:
      - "internal-ca"  # 内部 CA 根证书
    certificates:
      tls:
        cert: "internal-client"
        key: "internal-client"
```

## 证书管理操作

### 查看证书列表

```bash
# CLI 方式
tlcpchan-cli cert list

# API 方式
curl http://localhost:8080/api/v1/certificates
```

### 查看证书详情

```bash
# CLI 方式
tlcpchan-cli cert show server-sm2

# API 方式
curl http://localhost:8080/api/v1/certificates/server-sm2
```

### 删除证书

```bash
# CLI 方式
tlcpchan-cli cert delete old-cert

# API 方式
curl -X DELETE http://localhost:8080/api/v1/certificates/old-cert
```

## 证书文件格式

### PEM 格式（推荐）

```
-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAJC1HiIAZAiUMA0G...
-----END CERTIFICATE-----

-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgw...
-----END PRIVATE KEY-----
```

### PKCS#12 格式

需要先转换为 PEM：

```bash
# 提取证书
openssl pkcs12 -in cert.p12 -clcerts -nokeys -out cert.crt

# 提取私钥
openssl pkcs12 -in cert.p12 -nocerts -nodes -out cert.key
```

## 安全建议

### 文件权限

```bash
# 证书文件可读
chmod 644 certs/tlcp/*.crt
chmod 644 certs/tls/*.crt

# 私钥文件仅所有者可读
chmod 600 certs/tlcp/*.key
chmod 600 certs/tls/*.key

# 设置所有者
chown -R tlcpchan:tlcpchan /opt/tlcpchan/certs
```

### 证书有效期监控

```bash
# 检查证书过期时间
openssl x509 -in server.crt -noout -dates

# 通过 API 检查
curl http://localhost:8080/api/v1/certificates | jq '.certificates[] | {name, not_after}'
```

### 密钥保护

- 私钥文件权限设为 600
- 定期轮换密钥
- 使用硬件安全模块（HSM）存储密钥
- 备份私钥到安全位置

## 常见问题

### Q: 证书加载失败？

检查：
1. 文件路径是否正确
2. 文件格式是否为 PEM
3. 私钥是否与证书匹配
4. 文件权限是否正确

### Q: TLCP 握手失败？

确保：
1. 证书使用 SM2 算法
2. 密码套件配置正确
3. CA 证书链完整

### Q: 如何验证证书？

```bash
# 验证证书链
openssl verify -CAfile ca.crt server.crt

# 查看证书信息
openssl x509 -in server.crt -text -noout

# 验证私钥匹配
openssl x509 -noout -modulus -in server.crt | md5sum
openssl rsa -noout -modulus -in server.key | md5sum
```
