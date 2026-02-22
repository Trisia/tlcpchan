package config

import (
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/Trisia/tlcpchan/security/keystore"
	"gopkg.in/yaml.v3"
)

// ProtocolType 协议类型
type ProtocolType string

const (
	// ProtocolAuto 自动检测，同时支持TLCP和TLS
	ProtocolAuto ProtocolType = "auto"
	// ProtocolTLCP 仅使用TLCP协议（国密）
	ProtocolTLCP ProtocolType = "tlcp"
	// ProtocolTLS 仅使用TLS协议
	ProtocolTLS ProtocolType = "tls"
)

// AuthType 认证类型
type AuthType string

const (
	// AuthNone 无认证
	AuthNone AuthType = "none"
	// AuthOneWay 单向认证（验证对端证书）
	AuthOneWay AuthType = "one-way"
	// AuthMutual 双向认证（双方互相验证证书）
	AuthMutual AuthType = "mutual"
)

// ParseProtocolType 解析协议类型字符串
// 参数:
//   - s: 协议类型字符串，如 "auto", "tlcp", "tls"
//
// 返回:
//   - ProtocolType: 协议类型，无法识别时返回 ProtocolAuto
func ParseProtocolType(s string) ProtocolType {
	switch s {
	case "tlcp":
		return ProtocolTLCP
	case "tls":
		return ProtocolTLS
	case "auto":
		fallthrough
	default:
		return ProtocolAuto
	}
}

// ParseAuthType 解析认证类型字符串
// 参数:
//   - s: 认证类型字符串，如 "none", "one-way", "mutual"
//
// 返回:
//   - AuthType: 认证类型，无法识别时返回 AuthNone
func ParseAuthType(s string) AuthType {
	switch s {
	case "one-way":
		return AuthOneWay
	case "mutual":
		return AuthMutual
	case "none":
		fallthrough
	default:
		return AuthNone
	}
}

// MCPConfig MCP服务配置
type MCPConfig struct {
	// Enabled 是否启用MCP服务
	Enabled bool `yaml:"enabled" json:"enabled"`
	// APIKey MCP服务API密钥，为空表示无需认证
	APIKey string `yaml:"api_key,omitempty" json:"api_key,omitempty"`
}

// Config 主配置结构，包含服务端配置、代理实例列表和证书目录
type Config struct {
	// Server 服务端配置，包含API、UI和日志配置
	Server ServerConfig `yaml:"server" json:"server"`
	// KeyStores 密钥存储配置列表
	KeyStores []KeyStoreConfig `yaml:"keystores" json:"keystores"`
	// Instances 代理实例配置列表，每个实例代表一个独立的代理服务
	Instances []InstanceConfig `yaml:"instances" json:"instances"`
	// MCP MCP服务配置
	MCP MCPConfig `yaml:"mcp,omitempty" json:"mcp,omitempty"`
	// WorkDir 工作目录（运行时设置，不从配置文件读取）
	// Linux默认: /etc/tlcpchan
	// Windows默认: 程序所在目录
	WorkDir string `yaml:"-" json:"-"`
}

// KeyStoreConfig 密钥存储配置
type KeyStoreConfig struct {
	// Name 密钥存储名称，唯一标识符（可选）
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// Type 加载器类型
	Type keystore.LoaderType `yaml:"type" json:"type"`
	// Params 加载器参数
	Params map[string]string `yaml:"params" json:"params"`
}

// ServerConfig 服务端配置，定义管理界面和日志设置
type ServerConfig struct {
	// API API服务配置
	API APIConfig `yaml:"api" json:"api"`
	// Log 日志配置，nil表示使用默认配置
	Log *LogConfig `yaml:"log,omitempty" json:"log,omitempty"`
}

// APIConfig API服务配置
type APIConfig struct {
	// Address API服务监听地址
	// 格式: "host:port" 或 ":port"
	// 示例: ":8080" 表示监听所有网卡的8080端口
	// 示例: "127.0.0.1:8080" 表示仅监听本地回环地址
	Address string `yaml:"address" json:"address"`
}

// LogConfig 日志配置
type LogConfig struct {
	// Level 日志级别，可选值: "debug", "info", "warn", "error"
	Level string `yaml:"level" json:"level"`
	// File 日志文件路径，为空则仅输出到控制台
	// 示例: "./logs/tlcpchan.log"
	File string `yaml:"file" json:"file"`
	// MaxSize 单个日志文件最大大小，单位: MB
	MaxSize int `yaml:"max-size" json:"maxSize"`
	// MaxBackups 保留的旧日志文件最大数量，单位: 个
	MaxBackups int `yaml:"max-backups" json:"maxBackups"`
	// MaxAge 保留旧日志文件的最大天数，单位: 天
	MaxAge int `yaml:"max-age" json:"maxAge"`
	// Compress 是否压缩旧日志文件
	Compress bool `yaml:"compress" json:"compress"`
	// Enabled 是否启用日志
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// InstanceConfig 代理实例配置，定义单个代理服务的所有参数
type InstanceConfig struct {
	// Name 实例名称，全局唯一标识符
	// 示例: "proxy-1", "tlcp-server"
	Name string `yaml:"name" json:"name"`
	// Type 代理类型，可选值:
	// - "server": TCP服务端代理，接收TLCP/TLS连接并转发到目标
	// - "client": TCP客户端代理，接收普通TCP连接并以TLCP/TLS连接目标
	// - "http-server": HTTP服务端代理，处理HTTP/HTTPS请求
	// - "http-client": HTTP客户端代理，发起HTTP/HTTPS请求
	Type string `yaml:"type" json:"type"`
	// Listen 监听地址，格式: "host:port" 或 ":port"
	// 示例: ":8443" 表示监听所有网卡的8443端口
	Listen string `yaml:"listen" json:"listen"`
	// Target 目标地址，格式: "host:port"
	// 示例: "192.168.1.100:443"
	Target string `yaml:"target" json:"target"`
	// Protocol 协议类型，可选值:
	// - "auto": 自动检测，同时支持TLCP和TLS
	// - "tlcp": 仅使用TLCP协议（国密）
	// - "tls": 仅使用TLS协议
	Protocol string `yaml:"protocol" json:"protocol"`
	// Enabled 是否启用该实例
	Enabled bool `yaml:"enabled" json:"enabled"`
	// TLCP TLCP协议专用配置
	TLCP TLCPConfig `yaml:"tlcp,omitempty" json:"tlcp,omitempty"`
	// TLS TLS协议专用配置
	TLS TLSConfig `yaml:"tls,omitempty" json:"tls,omitempty"`
	// ClientCA 客户端CA证书路径列表，用于验证客户端证书
	// 示例: ["ca1.crt", "ca2.crt"]
	ClientCA []string `yaml:"client-ca,omitempty" json:"clientCa,omitempty"`
	// ServerCA 服务端CA证书路径列表，用于验证服务端证书
	// 示例: ["server-ca.crt"]
	ServerCA []string `yaml:"server-ca,omitempty" json:"serverCa,omitempty"`
	// HTTP HTTP协议专用配置，用于HTTP代理
	HTTP *HTTPConfig `yaml:"http,omitempty" json:"http,omitempty"`
	// Log 实例级别日志配置，nil表示使用全局配置
	Log *LogConfig `yaml:"log,omitempty" json:"log,omitempty"`
	// Stats 统计信息配置
	Stats *StatsConfig `yaml:"stats,omitempty" json:"stats,omitempty"`
	// SNI 服务器名称指示，用于TLS/TLCP握手时的SNI扩展
	// 示例: "example.com"
	SNI string `yaml:"sni,omitempty" json:"sni,omitempty"`
	// Timeout 连接超时配置
	Timeout *TimeoutConfig `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	// BufferSize 缓冲区大小，单位字节，默认 4096
	BufferSize int `yaml:"buffer-size,omitempty" json:"bufferSize,omitempty"`
}

// TLCPConfig TLCP协议配置（国密协议）
type TLCPConfig struct {
	// Auth 认证模式，可选值:
	// - "none": 无认证
	// - "one-way": 单向认证（验证对端证书）
	// - "mutual": 双向认证（双方互相验证证书）
	Auth string `yaml:"auth,omitempty" json:"auth,omitempty"`
	// ClientAuth 客户端认证类型（兼容性字段，会转换为Auth）
	ClientAuth string `yaml:"client-auth,omitempty" json:"-"`
	// MinVersion 最低协议版本，可选值: "1.1"（TLCP仅有1.1版本）
	MinVersion string `yaml:"min-version,omitempty" json:"minVersion,omitempty"`
	// MaxVersion 最高协议版本，可选值: "1.1"
	MaxVersion string `yaml:"max-version,omitempty" json:"maxVersion,omitempty"`
	// CipherSuites 密码套件列表，可选值:
	// - "ECC_SM4_CBC_SM3": ECC签名 + SM4 CBC模式 + SM3哈希
	// - "ECC_SM4_GCM_SM3": ECC签名 + SM4 GCM模式 + SM3哈希
	// - "ECDHE_SM4_CBC_SM3": ECDHE密钥交换 + SM4 CBC + SM3
	// - "ECDHE_SM4_GCM_SM3": ECDHE密钥交换 + SM4 GCM + SM3
	CipherSuites []string `yaml:"cipher-suites,omitempty" json:"cipherSuites,omitempty"`
	// CurvePreferences 椭圆曲线偏好，TLCP通常使用 "SM2"
	CurvePreferences []string `yaml:"curve-preferences,omitempty" json:"curvePreferences,omitempty"`
	// SessionTickets 是否启用会话票据
	SessionTickets bool `yaml:"session-tickets,omitempty" json:"sessionTickets,omitempty"`
	// SessionCache 是否启用会话缓存
	SessionCache bool `yaml:"session-cache,omitempty" json:"sessionCache,omitempty"`
	// InsecureSkipVerify 是否跳过证书验证（不安全，仅用于测试）
	InsecureSkipVerify bool `yaml:"insecure-skip-verify,omitempty" json:"insecureSkipVerify,omitempty"`
	// Keystore 密钥存储配置
	Keystore *KeyStoreConfig `yaml:"keystore,omitempty" json:"keystore,omitempty"`
}

// TLSConfig TLS协议配置
type TLSConfig struct {
	// Auth 认证模式，可选值:
	// - "none": 无认证
	// - "one-way": 单向认证（验证对端证书）
	// - "mutual": 双向认证（双方互相验证证书）
	Auth string `yaml:"auth,omitempty" json:"auth,omitempty"`
	// MinVersion 最低协议版本，可选值: "1.0", "1.1", "1.2", "1.3"
	MinVersion string `yaml:"min-version,omitempty" json:"minVersion,omitempty"`
	// MaxVersion 最高协议版本，可选值: "1.0", "1.1", "1.2", "1.3"
	MaxVersion string `yaml:"max-version,omitempty" json:"maxVersion,omitempty"`
	// CipherSuites 密码套件列表，可选值:
	// - "TLS_RSA_WITH_AES_128_CBC_SHA"
	// - "TLS_RSA_WITH_AES_256_CBC_SHA"
	// - "TLS_RSA_WITH_AES_128_CBC_SHA256"
	// - "TLS_RSA_WITH_AES_128_GCM_SHA256"
	// - "TLS_RSA_WITH_AES_256_GCM_SHA384"
	// - "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA"
	// - "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA"
	// - "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256"
	// - "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"
	// - "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"
	// - "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256"
	// - "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA"
	// - "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA"
	// - "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256"
	// - "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
	// - "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
	// - "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256"
	// - "TLS_AES_128_GCM_SHA256" (TLS 1.3)
	// - "TLS_AES_256_GCM_SHA384" (TLS 1.3)
	// - "TLS_CHACHA20_POLY1305_SHA256" (TLS 1.3)
	CipherSuites []string `yaml:"cipher-suites,omitempty" json:"cipherSuites,omitempty"`
	// CurvePreferences 椭圆曲线偏好，可选值: "P256", "P38", "P521", "X25519"
	CurvePreferences []string `yaml:"curve-preferences,omitempty" json:"curvePreferences,omitempty"`
	// SessionTickets 是否启用会话票据
	SessionTickets bool `yaml:"session-tickets,omitempty" json:"sessionTickets,omitempty"`
	// SessionCache 是否启用会话缓存
	SessionCache bool `yaml:"session-cache,omitempty" json:"sessionCache,omitempty"`
	// InsecureSkipVerify 是否跳过证书验证（不安全，仅用于测试）
	InsecureSkipVerify bool `yaml:"insecure-skip-verify,omitempty" json:"insecureSkipVerify,omitempty"`
	// Keystore 密钥存储配置
	Keystore *KeyStoreConfig `yaml:"keystore,omitempty" json:"keystore,omitempty"`
}

// HTTPConfig HTTP代理配置
type HTTPConfig struct {
	// RequestHeaders 请求头处理配置
	RequestHeaders HeadersConfig `yaml:"request-headers,omitempty" json:"requestHeaders,omitempty"`
	// ResponseHeaders 响应头处理配置
	ResponseHeaders HeadersConfig `yaml:"response-headers,omitempty" json:"responseHeaders,omitempty"`
}

// HeadersConfig HTTP头处理配置
type HeadersConfig struct {
	// Add 添加HTTP头，不会覆盖已存在的头
	// 示例: {"X-Proxy": "tlcpchan"}
	Add map[string]string `yaml:"add,omitempty" json:"add,omitempty"`
	// Remove 删除指定的HTTP头
	// 示例: ["X-Powered-By", "Server"]
	Remove []string `yaml:"remove,omitempty" json:"remove,omitempty"`
	// Set 设置HTTP头，会覆盖已存在的头
	// 示例: {"X-Frame-Options": "DENY"}
	Set map[string]string `yaml:"set,omitempty" json:"set,omitempty"`
}

// StatsConfig 统计信息配置
type StatsConfig struct {
	// Enabled 是否启用统计信息收集
	Enabled bool `yaml:"enabled" json:"enabled"`
	// Interval 统计信息收集间隔，单位: 纳秒（支持时间格式如 "10s", "1m"）
	// 示例: "10s" 表示每10秒收集一次
	Interval time.Duration `yaml:"interval" json:"interval"`
}

// TimeoutConfig 连接超时配置
type TimeoutConfig struct {
	// Dial 连接建立超时，默认: 10s
	Dial time.Duration `yaml:"dial,omitempty" json:"dial,omitempty"`
	// Read 读取超时，默认: 30s
	Read time.Duration `yaml:"read,omitempty" json:"read,omitempty"`
	// Write 写入超时，默认: 30s
	Write time.Duration `yaml:"write,omitempty" json:"write,omitempty"`
	// Handshake TLS/TLCP握手超时，默认: 15s
	Handshake time.Duration `yaml:"handshake,omitempty" json:"handshake,omitempty"`
}

// DefaultTimeout 返回默认超时配置
// 返回:
//   - *TimeoutConfig: 默认超时配置实例
func DefaultTimeout() *TimeoutConfig {
	return &TimeoutConfig{
		Dial:      10 * time.Second,
		Read:      30 * time.Second,
		Write:     30 * time.Second,
		Handshake: 15 * time.Second,
	}
}

// Default 返回默认配置
// 返回:
//   - *Config: 默认配置实例，API监听:20080
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			API: APIConfig{
				Address: ":20080",
			},
			Log: &LogConfig{
				Level:      "info",
				File:       "./logs/tlcpchan.log",
				MaxSize:    100,
				MaxBackups: 5,
				MaxAge:     30,
				Compress:   true,
				Enabled:    true,
			},
		},
		KeyStores: []KeyStoreConfig{},
		Instances: []InstanceConfig{},
	}
}

// Load 从文件加载配置
// 参数:
//   - path: 配置文件路径，支持YAML格式
//
// 返回:
//   - *Config: 配置实例
//   - error: 文件不存在、解析失败或验证失败时返回错误
//
// 注意: 加载后会自动填充默认值并验证配置
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	if err := Validate(cfg); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return cfg, nil
}

// Save 保存配置到文件
// 参数:
//   - path: 配置文件路径
//   - cfg: 配置实例
//
// 返回:
//   - error: 序列化或写入失败时返回错误
func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// Validate 验证配置有效性
// 参数:
//   - cfg: 待验证的配置实例
//
// 返回:
//   - error: 配置无效时返回错误，包含具体原因
//
// 注意: 该方法会自动填充缺失的默认值
func Validate(cfg *Config) error {
	if cfg.Server.API.Address == "" {
		cfg.Server.API.Address = ":20080"
	}

	// 验证 keystores
	ksNames := make(map[string]bool)
	for i, ks := range cfg.KeyStores {
		if ks.Name == "" {
			return fmt.Errorf("keystore %d: 名称不能为空", i)
		}
		if ksNames[ks.Name] {
			return fmt.Errorf("keystore名称重复: %s", ks.Name)
		}
		ksNames[ks.Name] = true

		if ks.Type == "" {
			return fmt.Errorf("keystore %s: 加载器类型不能为空", ks.Name)
		}
	}

	instanceNames := make(map[string]bool)
	for i, inst := range cfg.Instances {
		if inst.Name == "" {
			return fmt.Errorf("实例 %d: 名称不能为空", i)
		}
		if instanceNames[inst.Name] {
			return fmt.Errorf("实例名称重复: %s", inst.Name)
		}
		instanceNames[inst.Name] = true

		if inst.Listen == "" {
			return fmt.Errorf("实例 %s: 监听地址不能为空", inst.Name)
		}
		if inst.Target == "" {
			return fmt.Errorf("实例 %s: 目标地址不能为空", inst.Name)
		}

		if inst.Type == "" {
			return fmt.Errorf("实例 %s: 类型不能为空", inst.Name)
		}
		validTypes := map[string]bool{
			"server":      true,
			"client":      true,
			"http-server": true,
			"http-client": true,
		}
		if !validTypes[inst.Type] {
			return fmt.Errorf("实例 %s: 无效的类型 %s", inst.Name, inst.Type)
		}

		if inst.Protocol == "" {
			cfg.Instances[i].Protocol = string(ProtocolAuto)
		}
		validProtocols := map[string]bool{
			string(ProtocolAuto): true,
			string(ProtocolTLCP): true,
			string(ProtocolTLS):  true,
		}
		if !validProtocols[inst.Protocol] {
			return fmt.Errorf("实例 %s: 无效的协议 %s", inst.Name, inst.Protocol)
		}

		// 验证TLCP Auth
		if inst.TLCP.Auth == "" {
			cfg.Instances[i].TLCP.Auth = string(AuthNone)
		} else if inst.TLCP.Auth != string(AuthNone) && inst.TLCP.Auth != string(AuthOneWay) && inst.TLCP.Auth != string(AuthMutual) {
			return fmt.Errorf("实例 %s: 无效的TLCP认证模式 %s", inst.Name, inst.TLCP.Auth)
		}

		// 验证TLS Auth
		if inst.TLS.Auth == "" {
			cfg.Instances[i].TLS.Auth = string(AuthNone)
		} else if inst.TLS.Auth != string(AuthNone) && inst.TLS.Auth != string(AuthOneWay) && inst.TLS.Auth != string(AuthMutual) {
			return fmt.Errorf("实例 %s: 无效的TLS认证模式 %s", inst.Name, inst.TLS.Auth)
		}

		// 设置默认超时配置
		if inst.Timeout == nil {
			cfg.Instances[i].Timeout = DefaultTimeout()
		}

		// 设置默认缓冲区大小
		if inst.BufferSize <= 0 {
			cfg.Instances[i].BufferSize = 4096
		}
	}

	return nil
}

var TLCPCipherSuiteNames = map[string]uint16{
	"ECC_SM4_CBC_SM3":   tlcp.ECC_SM4_CBC_SM3,
	"ECC_SM4_GCM_SM3":   tlcp.ECC_SM4_GCM_SM3,
	"ECDHE_SM4_CBC_SM3": tlcp.ECDHE_SM4_CBC_SM3,
	"ECDHE_SM4_GCM_SM3": tlcp.ECDHE_SM4_GCM_SM3,
}

var TLSCipherSuiteNames = map[string]uint16{
	"TLS_RSA_WITH_AES_128_CBC_SHA":                  tls.TLS_RSA_WITH_AES_128_CBC_SHA,
	"TLS_RSA_WITH_AES_256_CBC_SHA":                  tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	"TLS_RSA_WITH_AES_128_CBC_SHA256":               tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
	"TLS_RSA_WITH_AES_128_GCM_SHA256":               tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
	"TLS_RSA_WITH_AES_256_GCM_SHA384":               tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
	"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA":          tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
	"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA":          tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
	"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256":       tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
	"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256":       tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384":       tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256": tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
	"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA":            tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA":            tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256":         tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":         tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":         tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256":   tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
	"TLS_AES_128_GCM_SHA256":                        tls.TLS_AES_128_GCM_SHA256,
	"TLS_AES_256_GCM_SHA384":                        tls.TLS_AES_256_GCM_SHA384,
	"TLS_CHACHA20_POLY1305_SHA256":                  tls.TLS_CHACHA20_POLY1305_SHA256,
}

var TLSVersionNames = map[string]uint16{
	"1.0": tls.VersionTLS10,
	"1.1": tls.VersionTLS11,
	"1.2": tls.VersionTLS12,
	"1.3": tls.VersionTLS13,
}

var TLCPVersionNames = map[string]uint16{
	"1.1": tlcp.VersionTLCP,
}

// ParseCipherSuite 解析密码套件名称为数值
// 参数:
//   - s: 密码套件名称或16进制值，如 "ECC_SM4_GCM_SM3" 或 "0xC012"
//   - isTLCP: 是否为TLCP协议，true则使用TLCP密码套件映射
//
// 返回:
//   - uint16: 密码套件数值
//   - error: 无法识别的密码套件名称时返回错误
func ParseCipherSuite(s string, isTLCP bool) (uint16, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("密码套件不能为空")
	}

	if isTLCP {
		if v, ok := TLCPCipherSuiteNames[s]; ok {
			return v, nil
		}
	} else {
		if v, ok := TLSCipherSuiteNames[s]; ok {
			return v, nil
		}
	}

	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		v, err := strconv.ParseUint(s[2:], 16, 16)
		if err != nil {
			return 0, fmt.Errorf("无效的16进制密码套件: %s", s)
		}
		return uint16(v), nil
	}

	v, err := strconv.ParseUint(s, 10, 16)
	if err == nil {
		return uint16(v), nil
	}

	return 0, fmt.Errorf("未知的密码套件: %s", s)
}

// ParseTLSVersion 解析协议版本名称为数值
// 参数:
//   - s: 版本名称或16进制值，如 "1.2" 或 "1.3"
//   - isTLCP: 是否为TLCP协议，true则使用TLCP版本映射
//
// 返回:
//   - uint16: 版本数值
//   - error: 无法识别的版本名称时返回错误
func ParseTLSVersion(s string, isTLCP bool) (uint16, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("版本不能为空")
	}

	if isTLCP {
		if v, ok := TLCPVersionNames[s]; ok {
			return v, nil
		}
	} else {
		if v, ok := TLSVersionNames[s]; ok {
			return v, nil
		}
	}

	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		v, err := strconv.ParseUint(s[2:], 16, 16)
		if err != nil {
			return 0, fmt.Errorf("无效的16进制版本: %s", s)
		}
		return uint16(v), nil
	}

	v, err := strconv.ParseUint(s, 10, 16)
	if err == nil {
		return uint16(v), nil
	}

	return 0, fmt.Errorf("未知的版本: %s", s)
}

// ParseCipherSuites 批量解析密码套件名称列表
// 参数:
//   - suites: 密码套件名称列表
//   - isTLCP: 是否为TLCP协议
//
// 返回:
//   - []uint16: 密码套件数值列表
//   - error: 任意一个密码套件解析失败时返回错误
func ParseCipherSuites(suites []string, isTLCP bool) ([]uint16, error) {
	if len(suites) == 0 {
		return nil, nil
	}

	result := make([]uint16, 0, len(suites))
	for _, s := range suites {
		v, err := ParseCipherSuite(s, isTLCP)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}

var TLCPClientAuthNames = map[string]tlcp.ClientAuthType{
	"no-client-cert":                 tlcp.NoClientCert,
	"request-client-cert":            tlcp.RequestClientCert,
	"require-any-client-cert":        tlcp.RequireAnyClientCert,
	"verify-client-cert-if-given":    tlcp.VerifyClientCertIfGiven,
	"require-and-verify-client-cert": tlcp.RequireAndVerifyClientCert,
}

var TLSClientAuthNames = map[string]tls.ClientAuthType{
	"no-client-cert":                 tls.NoClientCert,
	"request-client-cert":            tls.RequestClientCert,
	"require-any-client-cert":        tls.RequireAnyClientCert,
	"verify-client-cert-if-given":    tls.VerifyClientCertIfGiven,
	"require-and-verify-client-cert": tls.RequireAndVerifyClientCert,
}

// ParseTLCPClientAuth 解析TLCP客户端认证类型
// 参数:
//   - s: 认证类型名称，如 "no-client-cert", "require-and-verify-client-cert"
//
// 返回:
//   - tlcp.ClientAuthType: 客户端认证类型
//   - error: 无法识别的认证类型时返回错误
func ParseTLCPClientAuth(s string) (tlcp.ClientAuthType, error) {
	if s == "" {
		return tlcp.NoClientCert, nil
	}

	if v, ok := TLCPClientAuthNames[s]; ok {
		return v, nil
	}

	return tlcp.NoClientCert, fmt.Errorf("未知的TLCP客户端认证类型: %s", s)
}

// ParseTLSClientAuth 解析TLS客户端认证类型
// 参数:
//   - s: 认证类型名称，如 "no-client-cert", "require-and-verify-client-cert"
//
// 返回:
//   - tls.ClientAuthType: 客户端认证类型
//   - error: 无法识别的认证类型时返回错误
func ParseTLSClientAuth(s string) (tls.ClientAuthType, error) {
	if s == "" {
		return tls.NoClientCert, nil
	}

	if v, ok := TLSClientAuthNames[s]; ok {
		return v, nil
	}

	return tls.NoClientCert, fmt.Errorf("未知的TLS客户端认证类型: %s", s)
}

// ValidClientAuthValues 返回所有有效的客户端认证类型名称
// 返回:
//   - []string: 认证类型名称列表
func ValidClientAuthValues() []string {
	return []string{
		"no-client-cert",
		"request-client-cert",
		"require-any-client-cert",
		"verify-client-cert-if-given",
		"require-and-verify-client-cert",
	}
}

// GetKeyStoreStoreDir 获取 keystore 存储目录路径
// 返回:
//   - string: keystore 存储目录路径
func (c *Config) GetKeyStoreStoreDir() string {
	if c.WorkDir != "" {
		return filepath.Join(c.WorkDir, "keystores")
	}
	return "./keystores"
}

// GetRootCertDir 获取根证书存储目录路径
// 返回:
//   - string: 根证书存储目录路径
func (c *Config) GetRootCertDir() string {
	if c.WorkDir != "" {
		return filepath.Join(c.WorkDir, "rootcerts")
	}
	return "./rootcerts"
}

var (
	globalConfig *Config
	configMutex  sync.RWMutex
)

// Init 初始化全局配置单例
// 参数:
//   - cfg: 配置实例
//
// 注意: 该方法应该在程序启动时调用一次
func Init(cfg *Config) {
	configMutex.Lock()
	defer configMutex.Unlock()
	globalConfig = cfg
}

// Get 获取当前全局配置
// 返回:
//   - *Config: 当前全局配置实例，如果未初始化则返回 nil
func Get() *Config {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return globalConfig
}

// Set 更新全局配置
// 参数:
//   - cfg: 新的配置实例
func Set(cfg *Config) {
	configMutex.Lock()
	defer configMutex.Unlock()
	globalConfig = cfg
}

// LoadAndInit 从文件加载配置并初始化全局单例
// 参数:
//   - path: 配置文件路径
//
// 返回:
//   - error: 加载失败时返回错误
func LoadAndInit(path string) error {
	cfg, err := Load(path)
	if err != nil {
		return err
	}
	Init(cfg)
	return nil
}

// SaveAndUpdate 保存配置到文件并更新全局单例
// 参数:
//   - path: 配置文件路径
//   - cfg: 配置实例
//
// 返回:
//   - error: 保存失败时返回错误
func SaveAndUpdate(path string, cfg *Config) error {
	if err := Save(path, cfg); err != nil {
		return err
	}
	Set(cfg)
	return nil
}
