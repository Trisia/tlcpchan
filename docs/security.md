# 安全参数管理

本文档说明 TLCP Channel 中的安全参数（Keystore、根证书）配置和管理方法。

## 概述

TLCP Channel 采用新的安全参数管理体系，统一管理 Keystore 和根证书。

### 核心概念

| 概念 | 说明 |
|------|------|
| **Keystore** | 密钥存储，包含签名/加密证书和密钥 |
| **RootCert** | 根证书，用于验证对端证书 |
| **Loader** | Keystore 加载器，支持多种加载方式 |

---

## Keystore 管理

### Keystore 类型

| 类型 | 说明 |
|------|------|
| `tlcp` | 国密 TLCP 证书（签名+加密双证书） |
| `tls` | 标准 TLS 证书（单证书） |

### 加载器类型

| 类型 | 说明 |
|------|------|
| `file` | 从文件系统加载（默认） |
| `named` | 通过名称引用已存在的 keystore |
| `skf` | SKF 硬件接口（预留） |
| `sdf` | SDF 硬件接口（预留） |

---

## 配置方式

### 方式一：通过名称引用（推荐）

```yaml
instances:
  - name: "tlcp-server"
    type: "server"
    protocol: "tlcp"
    listen: ":443"
    target: "127.0.0.1:8080"
    keystore: "my-server-keystore"
```

### 方式二：直接配置文件路径

```yaml
instances:
  - name: "tlcp-server"
    type: "server"
    protocol: "tlcp"
    listen: ":443"
    target: "127.0.0.1:8080"
    keystore:
      sign-cert: "/path/to/sign.crt"
      sign-key: "/path/to/sign.key"
      enc-cert: "/path/to/enc.crt"  # TLCP 可选
      enc-key: "/path/to/enc.key"    # TLCP 可选
```

### 方式三：完整配置（扩展用）

```yaml
instances:
  - name: "tlcp-server"
    type: "server"
    protocol: "tlcp"
    listen: ":443"
    target: "127.0.0.1:8080"
    keystore:
      type: "file"
      params:
        sign-cert-path: "/path/to/sign.crt"
        sign-key-path: "/path/to/sign.key"
        enc-cert-path: "/path/to/enc.crt"
        enc-key-path: "/path/to/enc.key"
```

---

## 根证书管理

### 根证书配置

```yaml
instances:
  - name: "tlcp-server"
    type: "server"
    protocol: "tlcp"
    listen: ":443"
    target: "127.0.0.1:8080"
    keystore: "my-server-keystore"
    root-certs: ["ca1", "ca2"]
```

### 根证书优先级

1. 实例配置 `root-certs` 有值时，使用指定的根证书
2. 否则使用全局默认根证书池

---

## API 接口

### Keystore API

#### 列出所有 Keystore

```bash
GET /api/security/keystores
```

#### 创建 Keystore

```bash
POST /api/security/keystores
Content-Type: application/json

{
  "name": "my-keystore",
  "loaderType": "file",
  "params": {
    "sign-cert": "/path/to/sign.crt",
    "sign-key": "/path/to/sign.key",
    "enc-cert": "/path/to/enc.crt",
    "enc-key": "/path/to/enc.key"
  },
  "protected": false
}
```

#### 获取 Keystore 详情

```bash
GET /api/security/keystores/:name
```

#### 删除 Keystore

```bash
DELETE /api/security/keystores/:name
```

#### 重载 Keystore

```bash
POST /api/security/keystores/:name/reload
```

---

### 根证书 API

#### 列出所有根证书

```bash
GET /api/security/rootcerts
```

#### 添加根证书

```bash
POST /api/security/rootcerts
Content-Type: multipart/form-data

name=my-ca&cert=@ca.crt
```

#### 获取根证书详情

```bash
GET /api/security/rootcerts/:name
```

#### 删除根证书

```bash
DELETE /api/security/rootcerts/:name
```

#### 重载所有根证书

```bash
POST /api/security/rootcerts/reload
```

---

## 目录结构

```
/etc/tlcpchan/
├── config/
│   └── config.yaml        # 主配置文件（包含 keystores 配置）
├── rootcerts/             # 根证书存储目录
│   ├── ca1.pem            # 根证书文件（直接存放，无 yaml 元数据）
│   ├── ca2.crt            # 支持 .pem, .cer, .crt, .der 格式
│   └── ...
└── logs/
```

## Keystore 持久化说明

Keystore 记录持久化在主配置文件 `config.yaml` 的 `keystores` 字段中，不再使用独立的 `keystores/` 目录。

### 配置文件中的 Keystore 格式

```yaml
keystores:
  - name: "my-keystore"
    type: "file"
    params:
      sign-cert: "/path/to/sign.crt"
      sign-key: "/path/to/sign.key"
      enc-cert: "/path/to/enc.crt"  # TLCP 可选
      enc-key: "/path/to/enc.key"    # TLCP 可选
```

### 职责划分

- **Keystore Manager**：仅负责内存中的 keystore 管理，不关心持久化
- **Security Controller**：负责更新配置文件中的 `keystores` 字段
- **Config 模块**：负责配置文件的读写

---

## 工作原理

### Keystore 加载流程

```
1. 实例启动时，根据配置加载 keystore
   ├─ 字符串类型：通过名称查找已存在的 keystore
   ├─ map 类型：使用 FileLoader 从文件加载
   └─ 对象类型：使用指定的加载器加载

2. 加载器创建 KeyStore 实例
   └─ KeyStore 接口提供 TLCP/TLS 证书

3. 实例使用证书进行 TLCP/TLS 握手
```

### 根证书选择流程

```
1. 检查实例配置 root-certs
   ├─ 有值：从 RootCertManager 获取指定的根证书池
   └─ 无值：使用 RootCertManager 的默认根证书池

2. 使用根证书池验证对端证书
```

---

## 受保护的 Keystore

实例配置直接创建的 keystore 会被标记为 `protected: true`，不允许删除。

命名规则：`instance-<实例名>`

---

## 热更新

### Keystore 热更新

```bash
# 重载单个 keystore
curl -X POST http://localhost:30080/api/security/keystores/my-keystore/reload

# 重载所有 keystore（暂未实现）
```

### 根证书热更新

```bash
# 重载所有根证书
curl -X POST http://localhost:30080/api/security/rootcerts/reload
```

---

## 扩展性

### 自定义加载器

实现 `Loader` 接口即可添加新的加载器：

```go
type Loader interface {
    Load(config LoaderConfig) (KeyStore, error)
}
```

注册到 `KeyStoreManager`：

```go
manager.RegisterLoader(LoaderTypeSKF, &SKFLoader{})
```

---

## 安全建议

1. **文件权限**
   ```bash
   chmod 600 /etc/tlcpchan/config/config.yaml
   chmod 600 /etc/tlcpchan/rootcerts/certs/*.pem
   ```

2. **定期更新证书**
3. **使用硬件安全模块（HSM）** - 未来可通过 SKF/SDF 加载器支持
4. **监控证书有效期**
