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

## 2. 查询和修改配置

### 2.1 查询默认配置

**调用示例：**
```bash
tlcpchan-cli config show
```

**响应示例：**
```json
{
  "api": {
    "address": ":30080"
  },
  "web": {
    "address": ":30000"
  },
  "workDir": "/etc/tlcpchan"
}
```

### 2.2 重载配置

**调用示例：**
```bash
tlcpchan-cli config reload
```

**响应示例：**
```
配置已重新加载
```

### 2.3 验证配置文件

**调用示例：**
```bash
tlcpchan-cli config validate -f config.yaml
```

**响应示例：**
```
配置文件 config.yaml 格式有效
```

---

## 3. 查询实例

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

---

## 4. 创建实例

### 4.1 指定路径类型创建实例

首先创建配置文件 `instance.json`：

```json
{
  "name": "path-proxy",
  "type": "server",
  "listen": ":8443",
  "target": "localhost:8080",
  "protocol": "tlcp",
  "auth": "one-way",
  "enabled": true
}
```

**调用示例：**
```bash
tlcpchan-cli instance create -f instance.json
```

**响应示例：**
```
实例 path-proxy 创建成功
```

或者通过标准输入：

**调用示例：**
```bash
cat instance.json | tlcpchan-cli instance create
```

**响应示例：**
```
实例 path-proxy 创建成功
```

### 4.2 指定 keystore 类型创建实例

创建包含 keystore 配置的 `keystore-proxy.json`：

```json
{
  "name": "keystore-proxy",
  "type": "server",
  "listen": ":8444",
  "target": "localhost:8081",
  "protocol": "tlcp",
  "auth": "one-way",
  "enabled": true,
  "tlcp": {
    "keyStore": "my-keystore"
  }
}
```

**调用示例：**
```bash
tlcpchan-cli instance create -f keystore-proxy.json
```

**响应示例：**
```
实例 keystore-proxy 创建成功
```

### 4.3 完整示例：创建自签证书密钥并创建实例

#### 步骤 1：生成密钥库（含自签证书）

**调用示例：**
```bash
tlcpchan-cli keystore generate --name my-keystore --type tlcp --cn "proxy.example.com" --org "My Company" --days 1825
```

**响应示例：**
```
keystore my-keystore 生成成功
```

带完整参数的示例：

**调用示例：**
```bash
tlcpchan-cli keystore generate --name my-keystore --type tlcp \
  --cn "proxy.example.com" \
  --c CN --st "Beijing" --l "Beijing" \
  --org "My Company" --org-unit "IT" \
  --email "admin@example.com" \
  --days 1825 \
  --dns "proxy.example.com,*.example.com" \
  --ip "192.168.1.100,10.0.0.1"
```

**响应示例：**
```
keystore my-keystore 生成成功
```

#### 步骤 2：创建实例配置文件

创建 `my-instance.json`：

```json
{
  "name": "my-proxy",
  "type": "server",
  "listen": ":8443",
  "target": "localhost:8080",
  "protocol": "tlcp",
  "auth": "one-way",
  "enabled": true
}
```

#### 步骤 3：创建实例

**调用示例：**
```bash
tlcpchan-cli instance create -f my-instance.json
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

---

## 5. 实例的管理

### 5.1 启动实例

**调用示例：**
```bash
tlcpchan-cli instance start my-proxy
```

**响应示例：**
```
实例 my-proxy 已启动
```

### 5.2 停止实例

**调用示例：**
```bash
tlcpchan-cli instance stop my-proxy
```

**响应示例：**
```
实例 my-proxy 已停止
```

### 5.3 重启实例

**调用示例：**
```bash
tlcpchan-cli instance restart my-proxy
```

**响应示例：**
```
实例 my-proxy 已重启
```

### 5.4 重载实例

**调用示例：**
```bash
tlcpchan-cli instance reload my-proxy
```

**响应示例：**
```
实例 my-proxy 已重载
```

### 5.5 测试实例（健康检查）

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

### 5.6 查看实例统计

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

### 5.7 修改实例配置

创建 `updated-instance.json`：

```json
{
  "name": "my-proxy",
  "type": "server",
  "listen": ":8443",
  "target": "localhost:9090",
  "protocol": "tlcp",
  "auth": "one-way",
  "enabled": true
}
```

**调用示例：**
```bash
tlcpchan-cli instance update my-proxy -f updated-instance.json
```

**响应示例：**
```
实例 my-proxy 更新成功
```

### 5.8 删除实例

**调用示例：**
```bash
tlcpchan-cli instance delete my-proxy
```

**响应示例：**
```
实例 my-proxy 已删除
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

**调用示例：**
```bash
tlcpchan-cli rootcert generate --cn "My Root CA" --org "My Company" --years 10
```

**响应示例：**
```
根 CA 证书 tlcpchan-root-ca.crt 生成成功
```

带完整参数的示例：

**调用示例：**
```bash
tlcpchan-cli rootcert generate --cn "My Root CA" \
  --c CN --st "Beijing" --l "Beijing" \
  --org "My Company" --org-unit "IT" \
  --email "admin@example.com" \
  --days 3650
```

**响应示例：**
```
根 CA 证书 tlcpchan-root-ca.crt 生成成功
```

### 6.4 添加根证书

**调用示例：**
```bash
tlcpchan-cli rootcert add --filename external-ca.crt --cert /path/to/external-ca.crt
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

## 7. keystore 的管理

### 7.1 查询 keystore 列表

**调用示例：**
```bash
tlcpchan-cli keystore list
```

**响应示例：**
```
名称        类型  加载器  保护  创建时间
my-keystore tlcp file    否    2024-01-01T00:00:00Z
```

### 7.2 查看 keystore 详情

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
  signCert: /etc/tlcpchan/certs/sign.crt
  signKey: /etc/tlcpchan/certs/sign.key
  encCert: /etc/tlcpchan/certs/enc.crt
  encKey: /etc/tlcpchan/certs/enc.key
```

### 7.3 创建密钥对和自签证书（generate）

**调用示例：**
```bash
tlcpchan-cli keystore generate --name my-keystore --type tlcp --cn "proxy.example.com" --org "My Company" --days 1825
```

**响应示例：**
```
keystore my-keystore 生成成功
```

TLS 类型示例：

**调用示例：**
```bash
tlcpchan-cli keystore generate --name tls-keystore --type tls \
  --cn "tls.example.com" --org "My Company" \
  --days 1095 \
  --key-alg rsa --key-bits 2048
```

**响应示例：**
```
keystore tls-keystore 生成成功
```

带完整参数的示例：

**调用示例：**
```bash
tlcpchan-cli keystore generate --name my-keystore --type tlcp \
  --cn "proxy.example.com" \
  --c CN --st "Beijing" --l "Beijing" \
  --org "My Company" --org-unit "IT" \
  --email "admin@example.com" \
  --days 1825 \
  --dns "proxy.example.com,*.example.com" \
  --ip "192.168.1.100,10.0.0.1"
```

**响应示例：**
```
keystore my-keystore 生成成功
```

### 7.4 导入已有证书创建 keystore（create）

**调用示例：**
```bash
tlcpchan-cli keystore create --name import-keystore --loader-type file \
  --sign-cert sign.crt --sign-key sign.key \
  --enc-cert enc.crt --enc-key enc.key
```

**响应示例：**
```
keystore import-keystore 创建成功
```

### 7.5 重载 keystore

**调用示例：**
```bash
tlcpchan-cli keystore reload my-keystore
```

**响应示例：**
```
keystore my-keystore 已重载
```

### 7.6 删除 keystore

**调用示例：**
```bash
tlcpchan-cli keystore delete my-keystore
```

**响应示例：**
```
keystore my-keystore 已删除
```

---

## 8. 日志管理

### 8.1 查看实例日志

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

### 8.2 系统信息

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

### 8.3 系统健康检查

**调用示例：**
```bash
tlcpchan-cli system health
```

**响应示例：**
```
状态: ok
版本: 1.0.0
```

### 8.4 版本信息

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

### 9.1 全局选项

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

**响应示例（table）：**
```
名称      状态    类型    监听    目标          启用
my-proxy  running server  :8443  localhost:8080  true
```

**响应示例（JSON）：**
```json
[
  {
    "name": "my-proxy",
    "status": "running",
    "config": {
      "name": "my-proxy",
      "type": "server",
      "listen": ":8443",
      "target": "localhost:8080",
      "protocol": "tlcp",
      "auth": "one-way",
      "enabled": true
    },
    "enabled": true
  }
]
```

**操作命令 JSON 响应示例：**
```json
{
  "success": true,
  "message": "实例创建成功",
  "name": "my-proxy"
}
```

---

### 9.2 instance 命令组

实例管理命令。

| 子命令 | 说明 | 用法 |
|--------|------|------|
| `list` | 列出所有实例 | `instance list` |
| `show` | 显示实例详情 | `instance show <name>` |
| `create` | 创建实例 | `instance create [-f file]` |
| `update` | 更新实例配置 | `instance update <name> [-f file]` |
| `delete` | 删除实例 | `instance delete <name>` |
| `start` | 启动实例 | `instance start <name>` |
| `stop` | 停止实例 | `instance stop <name>` |
| `reload` | 重载实例 | `instance reload <name>` |
| `restart` | 重启实例 | `instance restart <name>` |
| `stats` | 查看统计信息 | `instance stats <name>` |
| `logs` | 查看日志 | `instance logs <name>` |
| `health` | 健康检查 | `instance health <name> [-t timeout]` |

#### instance list

**调用示例：**
```bash
tlcpchan-cli instance list
```

**响应示例：**
```
名称      状态    类型    监听    目标          启用
my-proxy  running server  :8443  localhost:8080  true
```

#### instance show

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

#### instance create

**调用示例：**
```bash
tlcpchan-cli instance create -f instance.json
```

**响应示例：**
```
实例 my-proxy 创建成功
```

#### instance update

**调用示例：**
```bash
tlcpchan-cli instance update my-proxy -f updated.json
```

**响应示例：**
```
实例 my-proxy 更新成功
```

#### instance delete

**调用示例：**
```bash
tlcpchan-cli instance delete my-proxy
```

**响应示例：**
```
实例 my-proxy 已删除
```

#### instance start

**调用示例：**
```bash
tlcpchan-cli instance start my-proxy
```

**响应示例：**
```
实例 my-proxy 已启动
```

#### instance stop

**调用示例：**
```bash
tlcpchan-cli instance stop my-proxy
```

**响应示例：**
```
实例 my-proxy 已停止
```

#### instance reload

**调用示例：**
```bash
tlcpchan-cli instance reload my-proxy
```

**响应示例：**
```
实例 my-proxy 已重载
```

#### instance restart

**调用示例：**
```bash
tlcpchan-cli instance restart my-proxy
```

**响应示例：**
```
实例 my-proxy 已重启
```

#### instance stats

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

#### instance logs

**调用示例：**
```bash
tlcpchan-cli instance logs my-proxy
```

**响应示例：**
```
[info] 2024-01-01T00:00:00Z: 实例启动
[info] 2024-01-01T00:00:01Z: 监听端口 :8443
```

#### instance health

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
```

---

### 9.3 config 命令组

配置管理命令。

| 子命令 | 说明 | 用法 |
|--------|------|------|
| `show` | 显示当前配置 | `config show` |
| `reload` | 重载配置 | `config reload` |
| `validate` | 验证配置文件 | `config validate [-f file]` |

#### config show

**调用示例：**
```bash
tlcpchan-cli config show
```

**响应示例（table 格式）：**
```
当前配置:
api:
  address: :30080
web:
  address: :30000
workDir: /etc/tlcpchan
```

**响应示例（JSON 格式）：**
```json
{
  "api": {
    "address": ":30080"
  },
  "web": {
    "address": ":30000"
  },
  "workDir": "/etc/tlcpchan"
}
```

#### config reload

**调用示例：**
```bash
tlcpchan-cli config reload
```

**响应示例：**
```
配置已重新加载
```

#### config validate

**调用示例：**
```bash
tlcpchan-cli config validate -f config.yaml
```

**响应示例：**
```
配置文件 config.yaml 格式有效
```

---

### 9.4 keystore 命令组

密钥库管理命令。

| 子命令 | 说明 | 用法 |
|--------|------|------|
| `list` | 列出所有 keystore | `keystore list` |
| `show` | 显示 keystore 详情 | `keystore show <name>` |
| `create` | 创建 keystore | `keystore create [选项]` |
| `generate` | 生成 keystore（含证书） | `keystore generate [选项]` |
| `delete` | 删除 keystore | `keystore delete <name>` |
| `reload` | 重载 keystore | `keystore reload <name>` |

#### keystore list

**调用示例：**
```bash
tlcpchan-cli keystore list
```

**响应示例：**
```
名称        类型  加载器  保护  创建时间
my-keystore tlcp file    否    2024-01-01T00:00:00Z
```

#### keystore show

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
  signCert: /etc/tlcpchan/certs/sign.crt
  signKey: /etc/tlcpchan/certs/sign.key
  encCert: /etc/tlcpchan/certs/enc.crt
  encKey: /etc/tlcpchan/certs/enc.key
```

#### keystore create 选项

| 选项 | 说明 |
|------|------|
| `--name` | 密钥库名称（必需） |
| `--loader-type` | 加载器类型（`file`/`named`/`skf`/`sdf`） |
| `--sign-cert` | 签名证书路径 |
| `--sign-key` | 签名密钥路径 |
| `--enc-cert` | 加密证书路径 |
| `--enc-key` | 加密密钥路径 |
| `--protected` | 是否受密码保护 |

**调用示例：**
```bash
tlcpchan-cli keystore create --name import-keystore --loader-type file \
  --sign-cert sign.crt --sign-key sign.key
```

**响应示例：**
```
keystore import-keystore 创建成功
```

#### keystore generate 选项

| 选项 | 说明 |
|------|------|
| `--name` | 密钥库名称（必需） |
| `--type` | 类型（`tlcp`/`tls`） |
| `--cn` | 通用名称（必需） |
| `--c` | 国家（2字母代码） |
| `--st` | 省/州 |
| `--l` | 地区/城市 |
| `--org` | 组织名称 |
| `--org-unit` | 组织单位 |
| `--email` | 邮箱地址 |
| `--years` | 有效期（年） |
| `--days` | 有效期（天，优先级高于years） |
| `--key-alg` | 密钥算法（ecdsa/rsa，仅TLS有效） |
| `--key-bits` | 密钥位数（仅RSA有效） |
| `--dns` | DNS名称，多个用逗号分隔 |
| `--ip` | IP地址，多个用逗号分隔 |
| `--protected` | 是否受密码保护 |

**调用示例：**
```bash
tlcpchan-cli keystore generate --name my-keystore --type tlcp \
  --cn "proxy.example.com" \
  --c CN --st "Beijing" --l "Beijing" \
  --org "My Company" --org-unit "IT" \
  --email "admin@example.com" \
  --days 1825 \
  --dns "proxy.example.com,*.example.com" \
  --ip "192.168.1.100,10.0.0.1"
```

**响应示例：**
```
keystore my-keystore 生成成功
```

#### keystore delete

**调用示例：**
```bash
tlcpchan-cli keystore delete my-keystore
```

**响应示例：**
```
keystore my-keystore 已删除
```

#### keystore reload

**调用示例：**
```bash
tlcpchan-cli keystore reload my-keystore
```

**响应示例：**
```
keystore my-keystore 已重载
```

---

### 9.5 rootcert 命令组

根证书管理命令。

| 子命令 | 说明 | 用法 |
|--------|------|------|
| `list` | 列出所有根证书 | `rootcert list` |
| `show` | 显示根证书详情 | `rootcert show <filename>` |
| `add` | 添加根证书 | `rootcert add [选项]` |
| `generate` | 生成根 CA 证书 | `rootcert generate [选项]` |
| `delete` | 删除根证书 | `rootcert delete <filename>` |
| `reload` | 重载所有根证书 | `rootcert reload` |

#### rootcert list

**调用示例：**
```bash
tlcpchan-cli rootcert list
```

**响应示例：**
```
文件名        主题                      颁发者                    过期时间
root-ca.crt   CN=My Root CA,O=My Company  CN=My Root CA,O=My Company  2034-01-01
```

#### rootcert show

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

#### rootcert add 选项

| 选项 | 说明 |
|------|------|
| `--filename` | 保存文件名（必需） |
| `--cert` | 证书文件路径（必需） |

**调用示例：**
```bash
tlcpchan-cli rootcert add --filename external-ca.crt --cert /path/to/external-ca.crt
```

**响应示例：**
```
根证书 external-ca.crt 添加成功
```

#### rootcert generate 选项

| 选项 | 说明 |
|------|------|
| `--cn` | 通用名称 |
| `--c` | 国家（2字母代码） |
| `--st` | 省/州 |
| `--l` | 地区/城市 |
| `--org` | 组织名称 |
| `--org-unit` | 组织单位 |
| `--email` | 邮箱地址 |
| `--years` | 有效期（年） |
| `--days` | 有效期（天，优先级高于years） |

**调用示例：**
```bash
tlcpchan-cli rootcert generate --cn "My Root CA" \
  --c CN --st "Beijing" --l "Beijing" \
  --org "My Company" --org-unit "IT" \
  --email "admin@example.com" \
  --days 3650
```

**响应示例：**
```
根 CA 证书 tlcpchan-root-ca.crt 生成成功
```

#### rootcert delete

**调用示例：**
```bash
tlcpchan-cli rootcert delete external-ca.crt
```

**响应示例：**
```
根证书 external-ca.crt 已删除
```

#### rootcert reload

**调用示例：**
```bash
tlcpchan-cli rootcert reload
```

**响应示例：**
```
根证书已重新加载
```

---

### 9.6 system 命令组

系统信息命令。

| 子命令 | 说明 | 用法 |
|--------|------|------|
| `info` | 显示系统信息 | `system info` |
| `health` | 健康检查 | `system health` |

#### system info

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

#### system health

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

### 9.7 version 命令

显示 CLI 和服务端版本信息。

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

## 10. 附录

### 10.1 实例配置 JSON 示例

#### TLCP 服务端代理

```json
{
  "name": "tlcp-server",
  "type": "server",
  "listen": ":8443",
  "target": "localhost:8080",
  "protocol": "tlcp",
  "auth": "one-way",
  "enabled": true,
  "tlcp": {
    "auth": "one-way",
    "minVersion": "1.1",
    "maxVersion": "1.1"
  }
}
```

#### TLS 服务端代理

```json
{
  "name": "tls-server",
  "type": "server",
  "listen": ":8443",
  "target": "localhost:8080",
  "protocol": "tls",
  "auth": "one-way",
  "enabled": true,
  "tls": {
    "auth": "one-way",
    "minVersion": "1.2",
    "maxVersion": "1.3"
  }
}
```

#### 客户端代理

```json
{
  "name": "client-proxy",
  "type": "client",
  "listen": ":8080",
  "target": "tlcp-server.example.com:8443",
  "protocol": "tlcp",
  "auth": "one-way",
  "enabled": true
}
```

#### HTTP 服务端代理

```json
{
  "name": "http-server",
  "type": "http-server",
  "listen": ":8443",
  "target": "http://localhost:8080",
  "protocol": "tlcp",
  "auth": "one-way",
  "enabled": true
}
```

### 10.2 常见问题解答

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

### 10.3 输出格式说明

tlcpchan-cli 支持两种输出格式：

1. **table 格式**（默认）：适合人类阅读的文本格式
2. **JSON 格式**：适合脚本处理

**重要说明：**
- 默认所有命令的输入和输出都是 table 格式（文本格式）
- 只有在明确指定 `--output json` 或 `-o json` 时才输出 JSON 格式
- 所有查询和操作命令都支持这两种输出格式

**调用示例：**
```bash
# 默认 table 格式
tlcpchan-cli instance list

# JSON 格式
tlcpchan-cli -o json instance list
```

**响应示例（table）：**
```
名称      状态    类型    监听    目标          启用
my-proxy  running server  :8443  localhost:8080  true
```

**响应示例（JSON）：**
```json
[
  {
    "name": "my-proxy",
    "status": "running",
    "config": {
      "name": "my-proxy",
      "type": "server",
      "listen": ":8443",
      "target": "localhost:8080",
      "protocol": "tlcp",
      "auth": "one-way",
      "enabled": true
    },
    "enabled": true
  }
]
```

**操作命令 JSON 响应示例：**
```json
{
  "success": true,
  "message": "实例创建成功",
  "name": "my-proxy"
}
```
