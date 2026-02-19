# tlcpchan-cli 用户手册

tlcpchan-cli 是 TLCP/TLS 代理服务的命令行管理工具，用于与 tlcpchan 核心服务的 REST API 交互。

---

## 1. 快速入门

### 1.1 连接服务

tlcpchan-cli 默认连接到 `http://localhost:30080`。如果服务运行在其他地址，使用 `--api` 或 `-a` 选项指定：

**调用示例：**
```bash
# 连接默认地址
tlcpchan-cli system info

# 连接自定义地址
tlcpchan-cli --api http://192.168.1.100:30080 system info
```

### 1.2 验证连接

使用 `system health` 命令验证与服务的连接：

**调用示例：**
```bash
tlcpchan-cli system health
```

**响应示例：**
```
状态: ok
版本: 1.0.0
```

---

## 2. 全局选项

| 选项 | 缩写 | 说明 | 默认值 |
|------|------|------|--------|
| `--api` | `-a` | API 服务地址 | `http://localhost:30080` |
| `--output` | `-o` | 输出格式（`table` 或 `json`） | `table` |

**说明：**
- 默认所有命令的输出格式为 `table`（文本格式），适合人类阅读
- 只有在指定 `--output json` 或 `-o json` 时才输出 JSON 格式，适合脚本处理

**调用示例：**
```bash
# 默认 table 格式输出
tlcpchan-cli instance list

# JSON 格式输出
tlcpchan-cli -o json instance list

# 自定义 API 地址
tlcpchan-cli -a http://10.0.0.1:30080 system info
```

---

## 3. 实例管理

### 3.1 列出所有实例

**调用示例：**
```bash
tlcpchan-cli instance list
```

**响应示例：**
```
名称      状态    类型    监听    目标          启用
my-proxy  running server  :8443  localhost:8080  true
test-proxy stopped  client  :8080  example.com:443 false
```

### 3.2 查看实例详情

**调用示例：**
```bash
tlcpchan-cli instance show my-proxy
```

**响应示例：**
```
名称: my-proxy
状态: running
类型: server
监听: :8443
目标: localhost:8080
协议: tlcp
认证: one-way
启用: true
```

### 3.3 创建实例

**参数说明：**

| 参数 | 说明 | 是否必需 | 默认值 |
|------|------|---------|--------|
| `--name` | 实例名称 | 是 | - |
| `--type` | 类型（server/client/http-server/http-client） | 否 | server |
| `--listen` | 监听地址 | 是 | - |
| `--target` | 目标地址 | 是 | - |
| `--protocol` | 协议（auto/tlcp/tls） | 否 | auto |
| `--auth` | 认证模式（none/one-way/mutual） | 否 | one-way |
| `--keystore` | keystore 名称（引用已创建的 keystore） | 否 | - |
| `--enabled` | 是否启用 | 否 | true |
| `--sni` | SNI 名称 | 否 | - |
| `--buffer-size` | 缓冲区大小 | 否 | 0 |
| **TLCP 直接文件参数** | | | |
| `--tlcp-sign-cert` | TLCP 签名证书路径 | 否 | - |
| `--tlcp-sign-key` | TLCP 签名密钥路径 | 否 | - |
| `--tlcp-enc-cert` | TLCP 加密证书路径 | 否 | - |
| `--tlcp-enc-key` | TLCP 加密密钥路径 | 否 | - |
| **TLS 直接文件参数** | | | |
| `--tls-cert` | TLS 证书路径 | 否 | - |
| `--tls-key` | TLS 密钥路径 | 否 | - |

**说明：**
- 使用 `--keystore` 参数引用已创建的 keystore
- 或使用 `--tlcp-*` / `--tls-*` 参数直接指定文件路径
- 所有文件路径支持**绝对路径**和**相对路径**
- TLCP 需要双证书（签名+加密），TLS 只需要单证书

**示例 1：创建 TLCP 服务端代理（使用已创建的 keystore）**

```bash
tlcpchan-cli instance create \
  --name tlcp-server \
  --type server \
  --listen :8443 \
  --target localhost:8080 \
  --protocol tlcp \
  --keystore my-keystore
```

**响应示例：**
```
实例 tlcp-server 创建成功
```

**示例 2：创建 TLS 服务端代理**

```bash
tlcpchan-cli instance create \
  --name tls-server \
  --type server \
  --listen :8443 \
  --target localhost:8080 \
  --protocol tls \
  --keystore tls-keystore
```

**响应示例：**
```
实例 tls-server 创建成功
```

**示例 3：创建 AUTO 模式（双协议）代理**

```bash
tlcpchan-cli instance create \
  --name auto-proxy \
  --type server \
  --listen :8443 \
  --target localhost:8080 \
  --protocol auto \
  --keystore my-keystore
```

**响应示例：**
```
实例 auto-proxy 创建成功
```

**示例 4：创建 TLCP 服务端代理（直接指定文件路径）**

```bash
tlcpchan-cli instance create \
  --name tlcp-file-server \
  --type server \
  --listen :8443 \
  --target localhost:8080 \
  --protocol tlcp \
  --tlcp-sign-cert /etc/tlcpchan/certs/sign.crt \
  --tlcp-sign-key /etc/tlcpchan/certs/sign.key \
  --tlcp-enc-cert /etc/tlcpchan/certs/enc.crt \
  --tlcp-enc-key /etc/tlcpchan/certs/enc.key
```

**响应示例：**
```
实例 tlcp-file-server 创建成功
```

**示例 5：创建 TLS 服务端代理（直接指定文件路径）**

```bash
tlcpchan-cli instance create \
  --name tls-file-server \
  --type server \
  --listen :8443 \
  --target localhost:8080 \
  --protocol tls \
  --tls-cert /etc/tlcpchan/certs/server.crt \
  --tls-key /etc/tlcpchan/certs/server.key
```

**响应示例：**
```
实例 tls-file-server 创建成功
```

**示例 6：创建 AUTO 模式代理（同时指定 TLCP 和 TLS 文件路径）**

```bash
tlcpchan-cli instance create \
  --name auto-file-proxy \
  --type server \
  --listen :8443 \
  --target localhost:8080 \
  --protocol auto \
  --tlcp-sign-cert /etc/tlcpchan/certs/sign.crt \
  --tlcp-sign-key /etc/tlcpchan/certs/sign.key \
  --tlcp-enc-cert /etc/tlcpchan/certs/enc.crt \
  --tlcp-enc-key /etc/tlcpchan/certs/enc.key \
  --tls-cert /etc/tlcpchan/certs/server.crt \
  --tls-key /etc/tlcpchan/certs/server.key
```

**响应示例：**
```
实例 auto-file-proxy 创建成功
```

### 3.4 更新实例配置

**参数说明：**

| 参数 | 说明 |
|------|------|
| `<name>` | 实例名称（位置参数） |
| `--type` | 类型（server/client/http-server/http-client） |
| `--listen` | 监听地址 |
| `--target` | 目标地址 |
| `--protocol` | 协议（auto/tlcp/tls） |
| `--auth` | 认证模式（none/one-way/mutual） |
| `--keystore` | keystore 名称（引用已创建的 keystore） |
| `--enabled` | 是否启用 |
| `--sni` | SNI 名称 |
| `--buffer-size` | 缓冲区大小 |
| **TLCP 直接文件参数** | |
| `--tlcp-sign-cert` | TLCP 签名证书路径 |
| `--tlcp-sign-key` | TLCP 签名密钥路径 |
| `--tlcp-enc-cert` | TLCP 加密证书路径 |
| `--tlcp-enc-key` | TLCP 加密密钥路径 |
| **TLS 直接文件参数** | |
| `--tls-cert` | TLS 证书路径 |
| `--tls-key` | TLS 密钥路径 |

**调用示例：**
```bash
tlcpchan-cli instance update my-proxy --target localhost:9090
```

**响应示例：**
```
实例 my-proxy 更新成功
```

### 3.5 删除实例

**调用示例：**
```bash
tlcpchan-cli instance delete my-proxy
```

**响应示例：**
```
实例 my-proxy 已删除
```

### 3.6 启动实例

**调用示例：**
```bash
tlcpchan-cli instance start my-proxy
```

**响应示例：**
```
实例 my-proxy 已启动
```

### 3.7 停止实例

**调用示例：**
```bash
tlcpchan-cli instance stop my-proxy
```

**响应示例：**
```
实例 my-proxy 已停止
```

### 3.8 重启实例

**调用示例：**
```bash
tlcpchan-cli instance restart my-proxy
```

**响应示例：**
```
实例 my-proxy 已重启
```

### 3.9 重载实例

**调用示例：**
```bash
tlcpchan-cli instance reload my-proxy
```

**响应示例：**
```
实例 my-proxy 已重载
```

### 3.10 查看实例统计

**调用示例：**
```bash
tlcpchan-cli instance stats my-proxy
```

**响应示例：**
```
connections: 10
bytes_in: 1048576
bytes_out: 2097152
total_requests: 100
```

### 3.11 查看实例日志

**调用示例：**
```bash
tlcpchan-cli instance logs my-proxy
```

**响应示例：**
```
[info] 2024-01-01T00:00:00Z: 实例启动
[info] 2024-01-01T00:00:01Z: 监听端口 :8443
[info] 2024-01-01T00:00:02Z: 新连接 192.168.1.100:54321
```

### 3.12 实例健康检查

**调用示例：**
```bash
tlcpchan-cli instance health my-proxy
```

**响应示例：**
```
实例: my-proxy

协议: tlcp
  状态: 成功
  延迟: 5ms

协议: tls
  状态: 成功
  延迟: 3ms
```

带超时的健康检查：

**调用示例：**
```bash
tlcpchan-cli instance health my-proxy -t 10
```

**响应示例：**
```
实例: my-proxy

协议: tlcp
  状态: 成功
  延迟: 5ms
```

---

## 4. 配置管理

### 4.1 查询当前配置

**调用示例：**
```bash
tlcpchan-cli config show
```

**响应示例：**
```
当前配置:
server:
  api:
    address: :30080
  ui:
    enabled: true
    address: :30000
    path: ./ui
  log:
    level: info
    file: ./logs/tlcpchan.log
    maxSize: 100
    maxBackups: 5
    maxAge: 30
    compress: true
    enabled: true
keystores: []
instances: []
```

### 4.2 重载配置

**调用示例：**
```bash
tlcpchan-cli config reload
```

**响应示例：**
```
配置已重新加载
```

### 4.3 验证配置文件

配置验证由服务端加载文件并检测，CLI 仅负责调用。

**参数说明：**

| 参数 | 说明 | 是否必需 |
|------|------|---------|
| `-f, --file` | 配置文件路径，可选，不提供则使用默认配置文件 | 否 |

**示例 1：验证默认配置文件**

```bash
tlcpchan-cli config validate
```

**响应示例：**
```
配置文件 默认配置文件 格式有效
```

**示例 2：验证指定配置文件**

```bash
tlcpchan-cli config validate -f /etc/tlcpchan/config/config.yaml
```

**响应示例：**
```
配置文件 /etc/tlcpchan/config/config.yaml 格式有效
```

---

## 5. keystore 管理

### 5.1 查询 keystore 列表

**调用示例：**
```bash
tlcpchan-cli keystore list
```

**响应示例：**
```
名称        类型  加载器  保护  创建时间
my-keystore tlcp file    否    2024-01-01T00:00:00Z
tls-keystore tls  file    否    2024-01-01T00:00:00Z
```

### 5.2 查看 keystore 详情

**调用示例：**
```bash
tlcpchan-cli keystore show my-keystore
```

**响应示例：**
```
名称: my-keystore
类型: tlcp
加载器: file
受保护: false
创建时间: 2024-01-01T00:00:00Z
更新时间: 2024-01-01T00:00:00Z
参数:
  sign-cert: ./keystores/my-keystore-sign.crt
  sign-key: ./keystores/my-keystore-sign.key
  enc-cert: ./keystores/my-keystore-enc.crt
  enc-key: ./keystores/my-keystore-enc.key
```

### 5.3 创建 keystore（导入已有证书）

**参数说明：**

| 参数 | 说明 | 是否必需 | 默认值 |
|------|------|---------|--------|
| `--name` | keystore 名称 | 是 | - |
| `--loader-type` | 加载器类型（file/named/skf/sdf） | 否 | file |
| `--sign-cert` | 签名证书文件路径 | 否 | - |
| `--sign-key` | 签名密钥文件路径 | 否 | - |
| `--enc-cert` | 加密证书文件路径（TLCP） | 否 | - |
| `--enc-key` | 加密密钥文件路径（TLCP） | 否 | - |
| `--protected` | 是否受保护 | 否 | false |

**说明：**
- 所有文件路径参数支持**绝对路径**和**相对路径**
- TLCP 类型需要双证书（`--sign-cert`、`--sign-key`、`--enc-cert`、`--enc-key`）
- TLS 类型只需要单证书（`--sign-cert`、`--sign-key`）

**示例 1：创建 TLCP keystore（双证书）**

```bash
tlcpchan-cli keystore create \
  --name tlcp-keystore \
  --sign-cert /path/to/sign.crt \
  --sign-key /path/to/sign.key \
  --enc-cert /path/to/enc.crt \
  --enc-key /path/to/enc.key
```

**响应示例：**
```
keystore tlcp-keystore 创建成功
```

**示例 2：创建 TLS keystore（单证书）**

```bash
tlcpchan-cli keystore create \
  --name tls-keystore \
  --sign-cert /path/to/cert.crt \
  --sign-key /path/to/cert.key
```

**响应示例：**
```
keystore tls-keystore 创建成功
```

### 5.4 生成 keystore（含自签证书）

**参数说明：**

| 参数 | 说明 | 是否必需 | 默认值 |
|------|------|---------|--------|
| `--name` | keystore 名称 | 是 | - |
| `--type` | 类型（tlcp/tls） | 否 | tlcp |
| `--cn` | 通用名称（CN） | 是 | - |
| `--c` | 国家（C，2字母代码） | 否 | - |
| `--st` | 省/州（ST） | 否 | - |
| `--l` | 地区/城市（L） | 否 | - |
| `--org` | 组织名称（O） | 否 | tlcpchan |
| `--org-unit` | 组织单位（OU） | 否 | - |
| `--email` | 邮箱地址 | 否 | - |
| `--years` | 有效期（年） | 否 | 0 |
| `--days` | 有效期（天，优先级高于years） | 否 | 0 |
| `--key-alg` | 密钥算法（ecdsa/rsa，仅TLS有效） | 否 | ecdsa |
| `--key-bits` | 密钥位数（仅RSA有效） | 否 | 2048 |
| `--dns` | DNS名称，多个用逗号分隔 | 否 | - |
| `--ip` | IP地址，多个用逗号分隔 | 否 | - |
| `--protected` | 是否受保护 | 否 | false |

**示例 1：生成 TLCP keystore**

```bash
tlcpchan-cli keystore generate \
  --name my-tlcp-keystore \
  --type tlcp \
  --cn "proxy.example.com" \
  --org "My Company" \
  --days 1825
```

**响应示例：**
```
keystore my-tlcp-keystore 生成成功
```

**示例 2：生成 TLS keystore**

```bash
tlcpchan-cli keystore generate \
  --name my-tls-keystore \
  --type tls \
  --cn "tls.example.com" \
  --org "My Company" \
  --days 1095 \
  --key-alg rsa \
  --key-bits 2048
```

**响应示例：**
```
keystore my-tls-keystore 生成成功
```

**示例 3：带完整参数的 TLCP keystore**

```bash
tlcpchan-cli keystore generate \
  --name my-keystore \
  --type tlcp \
  --cn "proxy.example.com" \
  --c CN \
  --st "Beijing" \
  --l "Beijing" \
  --org "My Company" \
  --org-unit "IT" \
  --email "admin@example.com" \
  --days 1825 \
  --dns "proxy.example.com,*.example.com" \
  --ip "192.168.1.100,10.0.0.1"
```

**响应示例：**
```
keystore my-keystore 生成成功
```

### 5.5 删除 keystore

**调用示例：**
```bash
tlcpchan-cli keystore delete my-keystore
```

**响应示例：**
```
keystore my-keystore 已删除
```

### 5.6 重载 keystore

**调用示例：**
```bash
tlcpchan-cli keystore reload my-keystore
```

**响应示例：**
```
keystore my-keystore 已重载
```

---

## 6. 根证书管理

### 6.1 查询根证书列表

**调用示例：**
```bash
tlcpchan-cli rootcert list
```

**响应示例：**
```
文件名        主题                      颁发者                    过期时间
root-ca.crt   CN=My Root CA,O=My Company  CN=My Root CA,O=My Company  2034-01-01
```

### 6.2 查看根证书详情

**调用示例：**
```bash
tlcpchan-cli rootcert show root-ca.crt
```

**响应示例：**
```
文件名: root-ca.crt
主题: CN=My Root CA,O=My Company
颁发者: CN=My Root CA,O=My Company
过期时间: 2034-01-01
```

### 6.3 生成根 CA 证书

**参数说明：**

| 参数 | 说明 | 是否必需 | 默认值 |
|------|------|---------|--------|
| `--cn` | 通用名称（CN） | 否 | tlcpchan-root-ca |
| `--c` | 国家（C，2字母代码） | 否 | - |
| `--st` | 省/州（ST） | 否 | - |
| `--l` | 地区/城市（L） | 否 | - |
| `--org` | 组织名称（O） | 否 | tlcpchan |
| `--org-unit` | 组织单位（OU） | 否 | - |
| `--email` | 邮箱地址 | 否 | - |
| `--years` | 有效期（年） | 否 | 0 |
| `--days` | 有效期（天，优先级高于years） | 否 | 0 |

**调用示例：**
```bash
tlcpchan-cli rootcert generate \
  --cn "My Root CA" \
  --org "My Company" \
  --years 10
```

**响应示例：**
```
根 CA 证书 tlcpchan-root-ca.crt 生成成功
```

带完整参数的示例：

**调用示例：**
```bash
tlcpchan-cli rootcert generate \
  --cn "My Root CA" \
  --c CN \
  --st "Beijing" \
  --l "Beijing" \
  --org "My Company" \
  --org-unit "IT" \
  --email "admin@example.com" \
  --days 3650
```

**响应示例：**
```
根 CA 证书 tlcpchan-root-ca.crt 生成成功
```

### 6.4 添加根证书

**参数说明：**

| 参数 | 说明 | 是否必需 |
|------|------|---------|
| `--filename` | 保存文件名 | 是 |
| `--cert` | 证书文件路径 | 是 |

**调用示例：**
```bash
tlcpchan-cli rootcert add \
  --filename external-ca.crt \
  --cert /path/to/external-ca.crt
```

**响应示例：**
```
根证书 external-ca.crt 添加成功
```

### 6.5 删除根证书

**调用示例：**
```bash
tlcpchan-cli rootcert delete external-ca.crt
```

**响应示例：**
```
根证书 external-ca.crt 已删除
```

### 6.6 重载根证书

**调用示例：**
```bash
tlcpchan-cli rootcert reload
```

**响应示例：**
```
根证书已重新加载
```

---

## 7. 系统信息

### 7.1 显示系统信息

**调用示例：**
```bash
tlcpchan-cli system info
```

**响应示例：**
```
操作系统:    linux
架构:      amd64
CPU核心数:   8
Goroutine数: 50
内存分配:    128 MB
内存总量:    8192 MB
系统内存:    256 MB
启动时间:    2024-01-01T00:00:00Z
运行时长:    24h0m0s
```

### 7.2 系统健康检查

**调用示例：**
```bash
tlcpchan-cli system health
```

**响应示例：**
```
状态: ok
版本: 1.0.0
```

---

## 8. 版本信息

**调用示例：**
```bash
tlcpchan-cli version
```

**响应示例：**
```
CLI版本:    1.0.0
服务端版本:  1.0.0
```

---

## 9. 完整命令参考

### 9.1 instance 命令组

| 子命令 | 说明 | 用法 |
|--------|------|------|
| `list` | 列出所有实例 | `instance list` |
| `show` | 显示实例详情 | `instance show <name>` |
| `create` | 创建实例 | `instance create [选项]` |
| `update` | 更新实例配置 | `instance update <name> [选项]` |
| `delete` | 删除实例 | `instance delete <name>` |
| `start` | 启动实例 | `instance start <name>` |
| `stop` | 停止实例 | `instance stop <name>` |
| `reload` | 重载实例 | `instance reload <name>` |
| `restart` | 重启实例 | `instance restart <name>` |
| `stats` | 查看统计信息 | `instance stats <name>` |
| `logs` | 查看日志 | `instance logs <name>` |
| `health` | 健康检查 | `instance health <name> [-t timeout]` |

### 9.2 config 命令组

| 子命令 | 说明 | 用法 |
|--------|------|------|
| `show` | 显示当前配置 | `config show` |
| `reload` | 重载配置 | `config reload` |
| `validate` | 验证配置文件 | `config validate [-f file]` |

### 9.3 keystore 命令组

| 子命令 | 说明 | 用法 |
|--------|------|------|
| `list` | 列出所有 keystore | `keystore list` |
| `show` | 显示 keystore 详情 | `keystore show <name>` |
| `create` | 创建 keystore | `keystore create [选项]` |
| `generate` | 生成 keystore（含证书） | `keystore generate [选项]` |
| `delete` | 删除 keystore | `keystore delete <name>` |
| `reload` | 重载 keystore | `keystore reload <name>` |

### 9.4 rootcert 命令组

| 子命令 | 说明 | 用法 |
|--------|------|------|
| `list` | 列出所有根证书 | `rootcert list` |
| `show` | 显示根证书详情 | `rootcert show <filename>` |
| `add` | 添加根证书 | `rootcert add [选项]` |
| `generate` | 生成根 CA 证书 | `rootcert generate [选项]` |
| `delete` | 删除根证书 | `rootcert delete <filename>` |
| `reload` | 重载所有根证书 | `rootcert reload` |

### 9.5 system 命令组

| 子命令 | 说明 | 用法 |
|--------|------|------|
| `info` | 显示系统信息 | `system info` |
| `health` | 健康检查 | `system health` |

### 9.6 version 命令

| 命令 | 说明 | 用法 |
|------|------|------|
| `version` | 显示版本 | `version` |

---

## 10. 完整示例流程

### 示例：创建自签证书密钥并创建实例

#### 步骤 1：生成根 CA 证书

**调用示例：**
```bash
tlcpchan-cli rootcert generate \
  --cn "My Root CA" \
  --org "My Company" \
  --days 3650
```

**响应示例：**
```
根 CA 证书 tlcpchan-root-ca.crt 生成成功
```

#### 步骤 2：生成 keystore（含自签证书）

**调用示例：**
```bash
tlcpchan-cli keystore generate \
  --name my-keystore \
  --type tlcp \
  --cn "proxy.example.com" \
  --org "My Company" \
  --days 1825
```

**响应示例：**
```
keystore my-keystore 生成成功
```

#### 步骤 3：创建实例

**调用示例：**
```bash
tlcpchan-cli instance create \
  --name my-proxy \
  --type server \
  --listen :8443 \
  --target localhost:8080 \
  --protocol auto \
  --keystore my-keystore
```

**响应示例：**
```
实例 my-proxy 创建成功
```

#### 步骤 4：启动实例

**调用示例：**
```bash
tlcpchan-cli instance start my-proxy
```

**响应示例：**
```
实例 my-proxy 已启动
```

#### 步骤 5：查看实例列表

**调用示例：**
```bash
tlcpchan-cli instance list
```

**响应示例：**
```
名称      状态    类型    监听    目标          启用
my-proxy  running server  :8443  localhost:8080  true
```

#### 步骤 6：健康检查

**调用示例：**
```bash
tlcpchan-cli instance health my-proxy
```

**响应示例：**
```
实例: my-proxy

协议: tlcp
  状态: 成功
  延迟: 5ms

协议: tls
  状态: 成功
  延迟: 3ms
```

---

## 11. 常见问题解答

**Q: 如何查看命令帮助？**

A: 使用 `tlcpchan-cli <命令> --help` 查看命令帮助。

**Q: 创建实例时提示配置错误怎么办？**

A: 使用 `config validate -f <配置文件>` 验证配置文件格式。

**Q: 实例启动失败如何排查？**

A: 查看实例日志：`tlcpchan-cli instance logs <实例名>`。

**Q: 如何在脚本中使用 tlcpchan-cli？**

A: 使用 `--output json` 选项获取 JSON 格式输出，便于脚本解析。

**Q: 修改了证书后需要重启实例吗？**

A: 不需要，使用 `instance reload` 或 `keystore reload` 命令即可热重载。

**Q: 配置验证是如何工作的？**

A: 配置验证由服务端加载文件并检测，CLI 仅负责调用 API。
