package config

import (
	"crypto/tls"
	"os"
	"path/filepath"
	"testing"

	"gitee.com/Trisia/gotlcp/tlcp"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg == nil {
		t.Fatal("Default() 返回 nil")
	}

	if cfg.Server.API.Address != ":30080" {
		t.Errorf("默认API地址应为 :30080, 实际为 %s", cfg.Server.API.Address)
	}

	if !cfg.Server.UI.Enabled {
		t.Error("默认UI应启用")
	}

	if cfg.Server.UI.Address != ":30000" {
		t.Errorf("默认UI地址应为 :30000, 实际为 %s", cfg.Server.UI.Address)
	}

	if cfg.Server.Log == nil {
		t.Fatal("默认日志配置不应为空")
	}

	if cfg.Server.Log.Level != "info" {
		t.Errorf("默认日志级别应为 info, 实际为 %s", cfg.Server.Log.Level)
	}
}

func TestLoadYAMLConfig(t *testing.T) {
	yamlContent := `
server:
  api:
    address: ":9090"
  ui:
    enabled: false
    address: ":4000"
    path: "./custom-ui"
instances:
  - name: "test-instance"
    type: "server"
    listen: ":8443"
    target: "localhost:443"
    protocol: "tlcp"
    auth: "one-way"
    enabled: true
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("写入配置文件失败: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() 失败: %v", err)
	}

	if cfg.Server.API.Address != ":9090" {
		t.Errorf("API地址应为 :9090, 实际为 %s", cfg.Server.API.Address)
	}

	if cfg.Server.UI.Enabled {
		t.Error("UI应禁用")
	}

	if len(cfg.Instances) != 1 {
		t.Fatalf("应有1个实例, 实际有 %d 个", len(cfg.Instances))
	}

	inst := cfg.Instances[0]
	if inst.Name != "test-instance" {
		t.Errorf("实例名称应为 test-instance, 实际为 %s", inst.Name)
	}

	if inst.Type != "server" {
		t.Errorf("实例类型应为 server, 实际为 %s", inst.Type)
	}

	if inst.Protocol != "tlcp" {
		t.Errorf("协议应为 tlcp, 实际为 %s", inst.Protocol)
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("加载不存在的文件应返回错误")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	yamlContent := `
server:
  api:
    address: ":9090"
  [invalid yaml
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("写入配置文件失败: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("加载无效YAML应返回错误")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "有效配置",
			cfg: &Config{
				Server: ServerConfig{
					API: APIConfig{Address: ":30080"},
				},
				Instances: []InstanceConfig{
					{
						Name:     "test",
						Type:     "server",
						Listen:   ":8443",
						Target:   "localhost:443",
						Protocol: "auto",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "空实例名称",
			cfg: &Config{
				Server: ServerConfig{
					API: APIConfig{Address: ":30080"},
				},
				Instances: []InstanceConfig{
					{
						Name:   "",
						Type:   "server",
						Listen: ":8443",
						Target: "localhost:443",
					},
				},
			},
			wantErr: true,
			errMsg:  "名称不能为空",
		},
		{
			name: "重复实例名称",
			cfg: &Config{
				Server: ServerConfig{
					API: APIConfig{Address: ":30080"},
				},
				Instances: []InstanceConfig{
					{Name: "test", Type: "server", Listen: ":8443", Target: "localhost:443", Protocol: "auto"},
					{Name: "test", Type: "client", Listen: ":8444", Target: "localhost:443", Protocol: "auto"},
				},
			},
			wantErr: true,
			errMsg:  "名称重复",
		},
		{
			name: "空监听地址",
			cfg: &Config{
				Server: ServerConfig{
					API: APIConfig{Address: ":30080"},
				},
				Instances: []InstanceConfig{
					{
						Name:   "test",
						Type:   "server",
						Listen: "",
						Target: "localhost:443",
					},
				},
			},
			wantErr: true,
			errMsg:  "监听地址不能为空",
		},
		{
			name: "空目标地址",
			cfg: &Config{
				Server: ServerConfig{
					API: APIConfig{Address: ":30080"},
				},
				Instances: []InstanceConfig{
					{
						Name:   "test",
						Type:   "server",
						Listen: ":8443",
						Target: "",
					},
				},
			},
			wantErr: true,
			errMsg:  "目标地址不能为空",
		},
		{
			name: "空类型",
			cfg: &Config{
				Server: ServerConfig{
					API: APIConfig{Address: ":30080"},
				},
				Instances: []InstanceConfig{
					{
						Name:   "test",
						Type:   "",
						Listen: ":8443",
						Target: "localhost:443",
					},
				},
			},
			wantErr: true,
			errMsg:  "类型不能为空",
		},
		{
			name: "无效类型",
			cfg: &Config{
				Server: ServerConfig{
					API: APIConfig{Address: ":30080"},
				},
				Instances: []InstanceConfig{
					{
						Name:   "test",
						Type:   "invalid",
						Listen: ":8443",
						Target: "localhost:443",
					},
				},
			},
			wantErr: true,
			errMsg:  "无效的类型",
		},
		{
			name: "无效协议",
			cfg: &Config{
				Server: ServerConfig{
					API: APIConfig{Address: ":30080"},
				},
				Instances: []InstanceConfig{
					{
						Name:     "test",
						Type:     "server",
						Listen:   ":8443",
						Target:   "localhost:443",
						Protocol: "invalid",
					},
				},
			},
			wantErr: true,
			errMsg:  "无效的协议",
		},
		{
			name: "无效TLCP认证模式",
			cfg: &Config{
				Server: ServerConfig{
					API: APIConfig{Address: ":30080"},
				},
				Instances: []InstanceConfig{
					{
						Name:     "test",
						Type:     "server",
						Listen:   ":8443",
						Target:   "localhost:443",
						Protocol: "tlcp",
						TLCP:     TLCPConfig{Auth: "invalid"},
					},
				},
			},
			wantErr: true,
			errMsg:  "无效的TLCP认证模式",
		},
		{
			name: "默认API地址",
			cfg: &Config{
				Server: ServerConfig{
					API: APIConfig{Address: ""},
				},
				Instances: []InstanceConfig{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Errorf("期望错误但未返回")
				} else if tt.errMsg != "" && !containsString(err.Error(), tt.errMsg) {
					t.Errorf("错误信息应包含 %q, 实际为 %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("不期望错误但返回: %v", err)
				}
			}
		})
	}
}

func TestValidateDefaultValueFill(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			API: APIConfig{Address: ""},
		},
		Instances: []InstanceConfig{
			{
				Name:     "test",
				Type:     "server",
				Listen:   ":8443",
				Target:   "localhost:443",
				Protocol: "auto",
			},
		},
	}

	if err := Validate(cfg); err != nil {
		t.Fatalf("Validate() 失败: %v", err)
	}

	if cfg.Server.API.Address != ":30080" {
		t.Errorf("API地址应默认为 :30080, 实际为 %s", cfg.Server.API.Address)
	}
}

func TestParseCipherSuite(t *testing.T) {
	tests := []struct {
		input    string
		isTLCP   bool
		expected uint16
		wantErr  bool
	}{
		{"ECC_SM4_GCM_SM3", true, 0xC012, false},
		{"ECC_SM4_CBC_SM3", true, 0xC011, false},
		{"ECDHE_SM4_GCM_SM3", true, 0xC014, false},
		{"TLS_AES_128_GCM_SHA256", false, 0x1301, false},
		{"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256", false, 0xC02F, false},
		{"0xC012", true, 0xC012, false},
		{"0xc012", true, 0xC012, false},
		{"49170", true, 0xC012, false},
		{"", true, 0, true},
		{"INVALID_SUITE", true, 0, true},
		{"0xGGGG", true, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseCipherSuite(tt.input, tt.isTLCP)
			if tt.wantErr {
				if err == nil {
					t.Errorf("期望错误但未返回")
				}
			} else {
				if err != nil {
					t.Errorf("不期望错误但返回: %v", err)
				}
				if result != tt.expected {
					t.Errorf("结果应为 0x%04X, 实际为 0x%04X", tt.expected, result)
				}
			}
		})
	}
}

func TestParseCipherSuites(t *testing.T) {
	tests := []struct {
		name     string
		suites   []string
		isTLCP   bool
		expected []uint16
		wantErr  bool
	}{
		{
			name:     "空列表",
			suites:   nil,
			isTLCP:   true,
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "TLCP密码套件",
			suites:   []string{"ECC_SM4_GCM_SM3", "ECC_SM4_CBC_SM3"},
			isTLCP:   true,
			expected: []uint16{0xC012, 0xC011},
			wantErr:  false,
		},
		{
			name:     "TLS密码套件",
			suites:   []string{"TLS_AES_128_GCM_SHA256", "TLS_AES_256_GCM_SHA384"},
			isTLCP:   false,
			expected: []uint16{0x1301, 0x1302},
			wantErr:  false,
		},
		{
			name:     "混合格式",
			suites:   []string{"ECC_SM4_GCM_SM3", "0xC011", "49168"},
			isTLCP:   true,
			expected: []uint16{0xC012, 0xC011, 0xC010},
			wantErr:  false,
		},
		{
			name:     "包含无效项",
			suites:   []string{"ECC_SM4_GCM_SM3", "INVALID"},
			isTLCP:   true,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseCipherSuites(tt.suites, tt.isTLCP)
			if tt.wantErr {
				if err == nil {
					t.Errorf("期望错误但未返回")
				}
			} else {
				if err != nil {
					t.Errorf("不期望错误但返回: %v", err)
				}
				if len(result) != len(tt.expected) {
					t.Errorf("结果长度应为 %d, 实际为 %d", len(tt.expected), len(result))
					return
				}
				for i, v := range result {
					if v != tt.expected[i] {
						t.Errorf("第%d项应为 0x%04X, 实际为 0x%04X", i, tt.expected[i], v)
					}
				}
			}
		})
	}
}

func TestParseTLSVersion(t *testing.T) {
	tests := []struct {
		input    string
		isTLCP   bool
		expected uint16
		wantErr  bool
	}{
		{"1.0", false, tls.VersionTLS10, false},
		{"1.1", false, tls.VersionTLS11, false},
		{"1.2", false, tls.VersionTLS12, false},
		{"1.3", false, tls.VersionTLS13, false},
		{"1.1", true, tlcp.VersionTLCP, false},
		{"0x0303", false, 0x0303, false},
		{"771", false, 0x0303, false},
		{"", false, 0, true},
		{"2.0", false, 0, true},
		{"invalid", false, 0, true},
		{"0xGGGG", false, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseTLSVersion(tt.input, tt.isTLCP)
			if tt.wantErr {
				if err == nil {
					t.Errorf("期望错误但未返回")
				}
			} else {
				if err != nil {
					t.Errorf("不期望错误但返回: %v", err)
				}
				if result != tt.expected {
					t.Errorf("结果应为 0x%04X, 实际为 0x%04X", tt.expected, result)
				}
			}
		})
	}
}

func TestSaveConfig(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			API: APIConfig{Address: ":9090"},
			UI: UIConfig{
				Enabled: true,
				Address: ":4000",
				Path:    "./ui",
			},
		},
		Instances: []InstanceConfig{
			{
				Name:     "test",
				Type:     "server",
				Listen:   ":8443",
				Target:   "localhost:443",
				Protocol: "tlcp",
				TLCP:     TLCPConfig{Auth: "one-way"},
				Enabled:  true,
			},
		},
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := Save(configPath, cfg); err != nil {
		t.Fatalf("Save() 失败: %v", err)
	}

	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() 失败: %v", err)
	}

	if loaded.Server.API.Address != cfg.Server.API.Address {
		t.Errorf("API地址不匹配: 期望 %s, 实际 %s", cfg.Server.API.Address, loaded.Server.API.Address)
	}

	if len(loaded.Instances) != len(cfg.Instances) {
		t.Errorf("实例数量不匹配: 期望 %d, 实际 %d", len(cfg.Instances), len(loaded.Instances))
	}
}

func TestTLCPCipherSuiteNames(t *testing.T) {
	expectedSuites := map[string]uint16{
		"ECC_SM4_CBC_SM3":   0xC011,
		"ECC_SM4_GCM_SM3":   0xC012,
		"ECC_SM4_CCM_SM3":   0xC019,
		"ECDHE_SM4_CBC_SM3": 0xC013,
		"ECDHE_SM4_GCM_SM3": 0xC014,
		"ECDHE_SM4_CCM_SM3": 0xC01A,
	}

	for name, expected := range expectedSuites {
		if v, ok := TLCPCipherSuiteNames[name]; !ok {
			t.Errorf("TLCP密码套件 %s 不存在", name)
		} else if v != expected {
			t.Errorf("TLCP密码套件 %s 值应为 0x%04X, 实际为 0x%04X", name, expected, v)
		}
	}
}

func TestTLSCipherSuiteNames(t *testing.T) {
	expectedSuites := map[string]uint16{
		"TLS_RSA_WITH_AES_128_GCM_SHA256":         0x009C,
		"TLS_RSA_WITH_AES_256_GCM_SHA384":         0x009D,
		"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256": 0xC02B,
		"TLS_AES_128_GCM_SHA256":                  0x1301,
	}

	for name, expected := range expectedSuites {
		if v, ok := TLSCipherSuiteNames[name]; !ok {
			t.Errorf("TLS密码套件 %s 不存在", name)
		} else if v != expected {
			t.Errorf("TLS密码套件 %s 值应为 0x%04X, 实际为 0x%04X", name, expected, v)
		}
	}
}

func TestTLSVersionNames(t *testing.T) {
	expectedVersions := map[string]uint16{
		"1.0": tls.VersionTLS10,
		"1.1": tls.VersionTLS11,
		"1.2": tls.VersionTLS12,
		"1.3": tls.VersionTLS13,
	}

	for name, expected := range expectedVersions {
		if v, ok := TLSVersionNames[name]; !ok {
			t.Errorf("TLS版本 %s 不存在", name)
		} else if v != expected {
			t.Errorf("TLS版本 %s 值应为 0x%04X, 实际为 0x%04X", name, expected, v)
		}
	}
}

func TestTLCPVersionNames(t *testing.T) {
	if v, ok := TLCPVersionNames["1.1"]; !ok {
		t.Error("TLCP版本 1.1 不存在")
	} else if v != tlcp.VersionTLCP {
		t.Errorf("TLCP版本 1.1 值应为 0x%04X, 实际为 0x%04X", tlcp.VersionTLCP, v)
	}
}

func TestValidTypes(t *testing.T) {
	validTypes := []string{"server", "client", "http-server", "http-client"}

	for _, typ := range validTypes {
		t.Run(typ, func(t *testing.T) {
			cfg := &Config{
				Server: ServerConfig{
					API: APIConfig{Address: ":30080"},
				},
				Instances: []InstanceConfig{
					{
						Name:     "test",
						Type:     typ,
						Listen:   ":8443",
						Target:   "localhost:443",
						Protocol: "auto",
					},
				},
			}
			if err := Validate(cfg); err != nil {
				t.Errorf("类型 %s 应有效, 但返回错误: %v", typ, err)
			}
		})
	}
}

func TestValidProtocols(t *testing.T) {
	validProtocols := []string{"auto", "tlcp", "tls"}

	for _, protocol := range validProtocols {
		t.Run(protocol, func(t *testing.T) {
			cfg := &Config{
				Server: ServerConfig{
					API: APIConfig{Address: ":30080"},
				},
				Instances: []InstanceConfig{
					{
						Name:     "test",
						Type:     "server",
						Listen:   ":8443",
						Target:   "localhost:443",
						Protocol: protocol,
					},
				},
			}
			if err := Validate(cfg); err != nil {
				t.Errorf("协议 %s 应有效, 但返回错误: %v", protocol, err)
			}
		})
	}
}

func TestValidAuthModes(t *testing.T) {
	validAuths := []string{"none", "one-way", "mutual"}

	for _, auth := range validAuths {
		t.Run(auth, func(t *testing.T) {
			cfg := &Config{
				Server: ServerConfig{
					API: APIConfig{Address: ":30080"},
				},
				Instances: []InstanceConfig{
					{
						Name:     "test",
						Type:     "server",
						Listen:   ":8443",
						Target:   "localhost:443",
						Protocol: "auto",
						TLCP:     TLCPConfig{Auth: auth},
					},
				},
			}
			if err := Validate(cfg); err != nil {
				t.Errorf("认证模式 %s 应有效, 但返回错误: %v", auth, err)
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestParseTLCPClientAuth(t *testing.T) {
	tests := []struct {
		input    string
		expected tlcp.ClientAuthType
		wantErr  bool
	}{
		{"", tlcp.NoClientCert, false},
		{"no-client-cert", tlcp.NoClientCert, false},
		{"request-client-cert", tlcp.RequestClientCert, false},
		{"require-any-client-cert", tlcp.RequireAnyClientCert, false},
		{"verify-client-cert-if-given", tlcp.VerifyClientCertIfGiven, false},
		{"require-and-verify-client-cert", tlcp.RequireAndVerifyClientCert, false},
		{"invalid", tlcp.NoClientCert, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseTLCPClientAuth(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("期望错误但未返回")
				}
			} else {
				if err != nil {
					t.Errorf("不期望错误但返回: %v", err)
				}
				if result != tt.expected {
					t.Errorf("结果应为 %d, 实际为 %d", tt.expected, result)
				}
			}
		})
	}
}

func TestParseTLSClientAuth(t *testing.T) {
	tests := []struct {
		input    string
		expected tls.ClientAuthType
		wantErr  bool
	}{
		{"", tls.NoClientCert, false},
		{"no-client-cert", tls.NoClientCert, false},
		{"request-client-cert", tls.RequestClientCert, false},
		{"require-any-client-cert", tls.RequireAnyClientCert, false},
		{"verify-client-cert-if-given", tls.VerifyClientCertIfGiven, false},
		{"require-and-verify-client-cert", tls.RequireAndVerifyClientCert, false},
		{"invalid", tls.NoClientCert, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseTLSClientAuth(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("期望错误但未返回")
				}
			} else {
				if err != nil {
					t.Errorf("不期望错误但返回: %v", err)
				}
				if result != tt.expected {
					t.Errorf("结果应为 %d, 实际为 %d", tt.expected, result)
				}
			}
		})
	}
}

func TestValidClientAuthValues(t *testing.T) {
	values := ValidClientAuthValues()
	if len(values) != 5 {
		t.Errorf("应有5个有效的客户端认证值, 实际有 %d 个", len(values))
	}
}

func TestResolveCertPath(t *testing.T) {
	tests := []struct {
		name     string
		certDir  string
		path     string
		expected string
	}{
		{"空路径", "/certs", "", ""},
		{"绝对路径", "/certs", "/etc/certs/cert.pem", "/etc/certs/cert.pem"},
		{"相对路径有certDir", "/certs", "cert.pem", "/certs/cert.pem"},
		{"相对路径无certDir", "", "cert.pem", "cert.pem"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{CertDir: tt.certDir}
			result := cfg.ResolveCertPath(tt.path)
			if result != tt.expected {
				t.Errorf("结果应为 %s, 实际为 %s", tt.expected, result)
			}
		})
	}
}

func TestGetCertDir(t *testing.T) {
	tests := []struct {
		name     string
		certDir  string
		expected string
	}{
		{"使用配置的certDir", "/custom/certs", "/custom/certs"},
		{"使用默认certDir", "", "./certs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{CertDir: tt.certDir}
			result := cfg.GetCertDir()
			if result != tt.expected {
				t.Errorf("结果应为 %s, 实际为 %s", tt.expected, result)
			}
		})
	}
}

func TestConfigWithCertDir(t *testing.T) {
	yamlContent := `
cert-dir: "/etc/tlcpchan/certs"
server:
  api:
    address: ":9090"
instances:
  - name: "test-instance"
    type: "server"
    listen: ":8443"
    target: "localhost:443"
    protocol: "tlcp"
    auth: "one-way"
    enabled: true
    tlcp:
      client-auth: "require-and-verify-client-cert"
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("写入配置文件失败: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() 失败: %v", err)
	}

	if cfg.CertDir != "/etc/tlcpchan/certs" {
		t.Errorf("CertDir 应为 /etc/tlcpchan/certs, 实际为 %s", cfg.CertDir)
	}

	if cfg.GetCertDir() != "/etc/tlcpchan/certs" {
		t.Errorf("GetCertDir() 应为 /etc/tlcpchan/certs, 实际为 %s", cfg.GetCertDir())
	}

	if len(cfg.Instances) != 1 {
		t.Fatalf("应有1个实例, 实际有 %d 个", len(cfg.Instances))
	}

	inst := cfg.Instances[0]
	if inst.TLCP.Auth != "mutual" {
		t.Errorf("Auth 应为 mutual, 实际为 %s", inst.TLCP.Auth)
	}
}
