# DeviceKeyStoreImpl.md

## 硬件 KeyStore 实现方案

### 硬件接口

- https://github.com/guanzhi/GmSSL/blob/master/src/skf/skf.h
- https://github.com/guanzhi/GmSSL/blob/master/src/sdf/sdf.h

### 1. 架构概述

本方案实现了基于国密硬件密钥存储接口的 KeyStore，支持 SKF（密码设备接口规范）和 SDF（密码设备功能接口）两种标准，使用 Go 语言开发，完全避免了 CGO 依赖。

### 2. 设计原则

- **跨平台支持**：通过 build tags 实现 Linux、Windows 平台的动态库加载
- **动态加载**：使用系统 dlopen / LoadLibrary 接口，禁止引入任何 CGO 绑定
- **包隔离**：SKF 和 SDF 接口完全独立实现，便于维护
- **模拟支持**：提供完整的 Mock 实现，便于测试和开发
- **安全性**：密钥在硬件中存储，不导出到内存

### 3. 目录结构

```
security/keystore/
├── hardware/                          # 硬件模块实现
│   ├── dlopen/                        # 动态库加载器
│   │   ├── dlopen.go                  # 接口定义
│   │   ├── dlopen_unix.go            # Linux/macOS 实现 (#+build linux || darwin)
│   │   └── dlopen_windows.go         # Windows 实现 (#+build windows)
│   ├── skf/                          # SKF 接口实现
│   │   ├── types.go                  # 数据结构定义
│   │   ├── api.go                    # API 函数声明
│   │   ├── api_loader.go             # 动态库函数加载器
│   │   ├── adapter.go                # 适配器
│   │   └── keystore.go               # KeyStore 实现
│   ├── sdf/                          # SDF 接口实现
│   │   ├── types.go                  # 数据结构定义
│   │   ├── api.go                    # API 函数声明
│   │   ├── api_loader.go             # 动态库函数加载器
│   │   ├── adapter.go                # 适配器
│   │   └── keystore.go               # KeyStore 实现
│   ├── errors.go                     # 错误处理
│   ├── util.go                       # 工具函数
│   ├── loader.go                     # 加载器接口
│   └── loader_register.go            # 加载器注册
├── testmocks/                        # 测试 Mock 实现
│   ├── mock_skf/                      # SKF 模拟库
│   │   ├── mock_skf.go               # 模拟实现（CGO 用于测试）
│   │   ├── build.sh                  # 编译脚本（Linux）
│   │   └── build.bat                 # 编译脚本（Windows）
│   ├── mock_sdf/                      # SDF 模拟库
│   │   ├── mock_sdf.go               # 模拟实现（CGO 用于测试）
│   │   ├── build.sh                  # 编译脚本（Linux）
│   │   └── build.bat                 # 编译脚本（Windows）
│   └── certs/                        # 测试证书数据
└── 现有文件                           # 保持原样
```

### 4. 核心模块设计

#### 4.1 动态库加载器（dlopen）

**功能**：统一跨平台的动态库加载接口

**实现方式**：
- Linux/macOS：使用系统 libc 的 dlopen、dlsym、dlclose 函数
- Windows：使用 kernel32 的 LoadLibrary、GetProcAddress、FreeLibrary 函数

**设计要点**：
```go
// 通用接口
type Library interface {
	Close() error
	Sym(name string) (uintptr, error)
}

// 平台特定实现
func OpenLibrary(path string) (Library, error) {
	// 每个平台实现不同
}
```

**架构**：
- 使用 build tags 实现平台隔离
- 内部符号查找
- 错误处理

#### 4.2 SKF 接口

**功能**：实现密码设备接口规范

SKF支持在容器内管理证书详见skf.h

**核心功能**：
- 设备管理（连接、断开、获取信息）
- 应用管理（打开、关闭、创建应用）
- 容器管理（打开、关闭、创建容器）
- 密钥操作（生成、导出公钥、导入证书、导出证书）
- 签名/验证
- PIN 管理

**实现结构**：
```go
// 配置
type Config struct {
	LibraryPath string
	DeviceName  string
	AppName     string
	ContainerName string
	UserPIN     string
}

// 适配器
type Adapter struct {
	loader   *apiLoader
	devHandle     uint32
	appHandle     uint32
	containerHandle uint32
	config   *Config
}

// KeyStore 实现
type SKFKeyStore struct {
	config  *Config
	adapter *Adapter
	tlsCert  *tls.Certificate
	tlcpCert *tlcp.Certificate
	mu sync.RWMutex
}
```

**API 加载器**：
```go
type apiLoader struct {
	lib    dlopen.Library
	funcs  map[string]uintptr
}

func newAPILoader(libPath string) (*apiLoader, error) {
	// 加载所需的 SKF API 函数
	funcNames := []string{
		"SKF_ConnectDev",
		"SKF_DisConnectDev",
		"SKF_OpenApplication",
		"SKF_CloseApplication",
		"SKF_OpenContainer",
		"SKF_CloseContainer",
		"SKF_VerifyPIN",
		"SKF_GenECCKeyPair",
		"SKF_ECCSignData",
		"SKF_ECCVerify",
		"SKF_ExportPublicKey",
		"SKF_ImportCertificate",
		"SKF_ExportCertificate",
	}
}
```

#### 4.3 SDF 接口

**功能**：实现密码设备功能接口规范

SDF接口不支持管理证书，只能通过用户文件类操作函数实现，详见sdf.h

**核心功能**：
- 设备管理（打开、关闭）
- 会话管理（打开、关闭会话）
- 密钥操作（生成、导出）
- 签名/验证
- 数据加密/解密

**实现结构**：
```go
// 配置
type Config struct {
	LibraryPath string
	SignKeyIndex uint32
	EncKeyIndex  uint32
}

// 适配器
type Adapter struct {
	config *Config
}

// KeyStore 实现
type SDFKeyStore struct {
	config  *Config
	adapter *Adapter
	tlsCert  *tls.Certificate
	tlcpCert *tlcp.Certificate
	mu sync.RWMutex
}
```

#### 4.4 Mock 实现

**功能**：提供完整的模拟接口，用于开发和测试

**实现特点**：
- 使用 CGO 开发，但仅用于测试
- 预编译的模拟库，包含固定的证书和密钥
- 30 年有效期的自签名证书
- 简单的命令编译脚本

**API 支持**：
- SKF：ConnectDev、DisConnectDev、OpenApplication、CloseApplication、OpenContainer、CloseContainer、VerifyPIN、GenECCKeyPair、ECCSignData、ECCVerify、ExportPublicKey、ImportCertificate、ExportCertificate
- SDF：OpenDevice、CloseDevice、OpenSession、CloseSession、GetDeviceInfo、GenerateKeyPair_ECC、ExportSignPublicKey_ECC、InternalSign_ECC、InternalVerify_ECC

### 5. 实施任务

#### Phase 1: 基础框架（3 天）
- [ ] 创建目录结构
- [ ] 实现跨平台动态库加载器
- [ ] 实现错误处理和工具函数
- [ ] 编写单元测试

#### Phase 2: SKF 接口（5 天）
- [ ] 定义数据结构
- [ ] 实现 API 加载器
- [ ] 实现适配器
- [ ] 实现 KeyStore
- [ ] 编写测试

#### Phase 3: SDF 接口（5 天）
- [ ] 定义数据结构
- [ ] 实现 API 加载器
- [ ] 实现适配器
- [ ] 实现 KeyStore
- [ ] 编写测试

#### Phase 4: 集成与注册（3 天）
- [ ] 实现硬件加载器注册
- [ ] 修改 keystore Manager
- [ ] 编写集成测试
- [ ] 更新文档

#### Phase 5: 跨平台测试（2 天）
- [ ] Linux 平台测试
- [ ] Windows 平台测试
- [ ] 性能测试

#### Phase 6: 文档与优化（2 天）
- [ ] 完善代码注释
- [ ] 编写使用文档
- [ ] 故障排查指南

### 6. 配置示例

#### SKF 配置
```yaml
keystores:
  - name: "skf-hardware"
    type: "skf"
    params:
      library-path: "/usr/local/lib/libskf.so"
      device-name: ""                    # 自动选择
      app-name: "tlcpchan"
      container-name: "tlcp-container"
      user-pin: "12345678"
```

#### SDF 配置
```yaml
keystores:
  - name: "sdf-hardware"
    type: "sdf"
    params:
      library-path: "/usr/local/lib/libsdf.so"
      sign-key-index: "1"
      enc-key-index: "2"
```

### 7. 使用说明

#### 1. 编译模拟库（仅测试时需要）
```bash
cd security/keystore/testmocks/mock_skf
chmod +x build.sh
./build.sh  # 生成 mock_skf.so

cd ../mock_sdf
./build.sh  # 生成 mock_sdf.so
```

#### 2. 运行测试
```bash
go test -v ./security/keystore/hardware/skf
go test -v ./security/keystore/hardware/sdf
```

#### 3. 启动服务器
```bash
cd tlcpchan
go run main.go -config config.yaml
```

### 8. 技术挑战与解决方案

#### 8.1 无 CGO 调用系统函数
**问题**：需要调用系统函数但不能使用 CGO
**解决方案**：使用 syscall 包直接调用系统 API

#### 8.2 动态符号查找
**问题**：SKF/SDF API 在编译时未知
**解决方案**：使用 dlopen / dlsym 动态加载

#### 8.3 数据结构对齐
**问题**：硬件数据结构可能与 Go 对齐方式不同
**解决方案**：使用 #pragma pack(1)，Go struct 字节对齐

#### 8.4 内存管理
**问题**：系统调用返回的内存需要管理
**解决方案**：确保资源正确释放，使用 sync.RWMutex 保护

### 9. 代码规范

- 包命名：skf, sdf, dlopen, hardware
- 文件命名：小写字母 + 下划线
- 变量命名：驼峰命名
- 接口方法：动词开头
- 错误处理：使用 errors 包，返回 error 接口

### 10. 测试覆盖

**目标**：70% 以上

**测试类型**：
- 单元测试：验证单个函数/方法正确性
- 集成测试：验证整个流程
- 模拟测试：测试边界条件

**测试工具**：
- Go testing 框架
- 模拟实现（Mock 库）
- 固定证书和密钥用于测试

### 11. 部署注意事项

#### 11.1 硬件要求
- 支持 SKF 或 SDF 接口的 USB Key 或 PCIe 卡
- 管理员权限可能需要

#### 11.2 证书管理
- 证书导出到文件系统存储（如果硬件不支持）
- 证书文件定期备份
- 证书有效性检查

#### 11.3 安全注意事项
- 配置文件中的 PIN 码加密
- 定期更新证书
- 硬件设备物理安全

### 12. 故障排查

**常见问题**：

1. **动态库加载失败**：
   - 检查 library-path 配置
   - 验证库是否存在
   - 检查库权限

2. **设备连接失败**：
   - 检查设备是否插好
   - 验证设备名称
   - 检查设备驱动

3. **PIN 验证失败**：
   - 确认 PIN 码正确性
   - 检查 PIN 重试次数
   - 解锁 PIN（如果锁定）

4. **证书加载失败**：
   - 检查证书路径
   - 验证证书格式
   - 检查证书权限

### 13. 性能优化

**建议**：
- 缓存已加载的密钥
- 减少证书验证次数
- 使用会话复用

### 14. 扩展计划

- **性能分析**：使用 Go 性能分析工具
- **内存优化**：减少分配
- **并发优化**：使用 goroutine 提高性能
- **硬件支持**：添加更多硬件接口