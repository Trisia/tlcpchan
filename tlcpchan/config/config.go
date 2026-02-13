package config

import (
	"crypto/tls"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gitee.com/Trisia/gotlcp/tlcp"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig     `yaml:"server"`
	Instances []InstanceConfig `yaml:"instances"`
}

type ServerConfig struct {
	API APIConfig  `yaml:"api"`
	UI  UIConfig   `yaml:"ui"`
	Log *LogConfig `yaml:"log,omitempty"`
}

type APIConfig struct {
	Address string `yaml:"address"`
}

type UIConfig struct {
	Enabled bool   `yaml:"enabled"`
	Address string `yaml:"address"`
	Path    string `yaml:"path"`
}

type LogConfig struct {
	Level      string `yaml:"level"`
	File       string `yaml:"file"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
	Enabled    bool   `yaml:"enabled"`
}

type InstanceConfig struct {
	Name     string             `yaml:"name"`
	Type     string             `yaml:"type"`
	Listen   string             `yaml:"listen"`
	Target   string             `yaml:"target"`
	Protocol string             `yaml:"protocol"`
	Auth     string             `yaml:"auth"`
	Enabled  bool               `yaml:"enabled"`
	TLCP     TLCPConfig         `yaml:"tlcp,omitempty"`
	TLS      TLSConfig          `yaml:"tls,omitempty"`
	Certs    CertificatesConfig `yaml:"certificates,omitempty"`
	ClientCA []string           `yaml:"client_ca,omitempty"`
	ServerCA []string           `yaml:"server_ca,omitempty"`
	HTTP     *HTTPConfig        `yaml:"http,omitempty"`
	Log      *LogConfig         `yaml:"log,omitempty"`
	Stats    *StatsConfig       `yaml:"stats,omitempty"`
	SNI      string             `yaml:"sni,omitempty"`
}

type TLCPConfig struct {
	MinVersion         string   `yaml:"min_version,omitempty"`
	MaxVersion         string   `yaml:"max_version,omitempty"`
	CipherSuites       []string `yaml:"cipher_suites,omitempty"`
	CurvePreferences   []string `yaml:"curve_preferences,omitempty"`
	SessionTickets     bool     `yaml:"session_tickets,omitempty"`
	SessionCache       bool     `yaml:"session_cache,omitempty"`
	InsecureSkipVerify bool     `yaml:"insecure_skip_verify,omitempty"`
}

type TLSConfig struct {
	MinVersion         string   `yaml:"min_version,omitempty"`
	MaxVersion         string   `yaml:"max_version,omitempty"`
	CipherSuites       []string `yaml:"cipher_suites,omitempty"`
	CurvePreferences   []string `yaml:"curve_preferences,omitempty"`
	SessionTickets     bool     `yaml:"session_tickets,omitempty"`
	SessionCache       bool     `yaml:"session_cache,omitempty"`
	InsecureSkipVerify bool     `yaml:"insecure_skip_verify,omitempty"`
}

type CertificatesConfig struct {
	TLCP CertConfig `yaml:"tlcp,omitempty"`
	TLS  CertConfig `yaml:"tls,omitempty"`
}

type CertConfig struct {
	Cert string `yaml:"cert,omitempty"`
	Key  string `yaml:"key,omitempty"`
}

type HTTPConfig struct {
	RequestHeaders  HeadersConfig `yaml:"request_headers,omitempty"`
	ResponseHeaders HeadersConfig `yaml:"response_headers,omitempty"`
}

type HeadersConfig struct {
	Add    map[string]string `yaml:"add,omitempty"`
	Remove []string          `yaml:"remove,omitempty"`
	Set    map[string]string `yaml:"set,omitempty"`
}

type StatsConfig struct {
	Enabled  bool          `yaml:"enabled"`
	Interval time.Duration `yaml:"interval"`
}

func Default() *Config {
	return &Config{
		Server: ServerConfig{
			API: APIConfig{
				Address: ":8080",
			},
			UI: UIConfig{
				Enabled: true,
				Address: ":3000",
				Path:    "./ui",
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
		Instances: []InstanceConfig{},
	}
}

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

func Validate(cfg *Config) error {
	if cfg.Server.API.Address == "" {
		cfg.Server.API.Address = ":8080"
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
			cfg.Instances[i].Protocol = "auto"
		}
		validProtocols := map[string]bool{
			"auto": true,
			"tlcp": true,
			"tls":  true,
		}
		if !validProtocols[inst.Protocol] {
			return fmt.Errorf("实例 %s: 无效的协议 %s", inst.Name, inst.Protocol)
		}

		if inst.Auth == "" {
			cfg.Instances[i].Auth = "none"
		}
		validAuth := map[string]bool{
			"none":    true,
			"one-way": true,
			"mutual":  true,
		}
		if !validAuth[inst.Auth] {
			return fmt.Errorf("实例 %s: 无效的认证模式 %s", inst.Name, inst.Auth)
		}
	}

	return nil
}

var TLCPCipherSuiteNames = map[string]uint16{
	"ECC_SM4_CBC_SM3":   0xC011,
	"ECC_SM4_GCM_SM3":   0xC012,
	"ECC_SM4_CCM_SM3":   0xC019,
	"ECDHE_SM4_CBC_SM3": 0xC013,
	"ECDHE_SM4_GCM_SM3": 0xC014,
	"ECDHE_SM4_CCM_SM3": 0xC01A,
}

var TLSCipherSuiteNames = map[string]uint16{
	"TLS_RSA_WITH_AES_128_GCM_SHA256":         0x009C,
	"TLS_RSA_WITH_AES_256_GCM_SHA384":         0x009D,
	"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256": 0xC02B,
	"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384": 0xC02C,
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":   0xC02F,
	"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":   0xC030,
	"TLS_AES_128_GCM_SHA256":                  0x1301,
	"TLS_AES_256_GCM_SHA384":                  0x1302,
	"TLS_CHACHA20_POLY1305_SHA256":            0x1303,
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
