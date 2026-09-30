package config

import (
	"path/filepath"
	"testing"
)

// TestTLCPIBCSuiteNames 验证 4 个 IBC/IBSDH 套件名称能解析为约定的编号。
func TestTLCPIBCSuiteNames(t *testing.T) {
	tests := []struct {
		name  string
		suite uint16
	}{
		{"IBC_SM4_GCM_SM3", 0xE057},
		{"IBC_SM4_CBC_SM3", 0xE017},
		{"IBSDH_SM4_GCM_SM3", 0xE055},
		{"IBSDH_SM4_CBC_SM3", 0xE015},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := TLCPCipherSuiteNames[tt.name]
			if !ok {
				t.Fatalf("TLCPCipherSuiteNames 缺少 %s", tt.name)
			}
			if got != tt.suite {
				t.Fatalf("%s = 0x%04X, want 0x%04X", tt.name, got, tt.suite)
			}

			parsed, err := ParseCipherSuites([]string{tt.name}, true)
			if err != nil {
				t.Fatalf("ParseCipherSuites(%s) error = %v", tt.name, err)
			}
			if len(parsed) != 1 || parsed[0] != tt.suite {
				t.Fatalf("ParseCipherSuites(%s) = %v, want [0x%04X]", tt.name, parsed, tt.suite)
			}

			// 套件在 TLS 协议下不可用
			if _, err := ParseCipherSuites([]string{tt.name}, false); err == nil {
				t.Fatalf("TLS 协议下解析 %s 应当失败", tt.name)
			}
		})
	}
}

// TestIsTLCPIBCSuite 验证 IBC/IBSDH 套件的分类判断。
func TestIsTLCPIBCSuite(t *testing.T) {
	ibcSuites := []uint16{0xE057, 0xE017}
	ibsdhSuites := []uint16{0xE055, 0xE015}
	otherSuites := []uint16{0xE011, 0xE013, 0xE051, 0xE053}

	for _, suite := range ibcSuites {
		if !IsTLCPIBCSuite(suite) {
			t.Fatalf("IsTLCPIBCSuite(0x%04X) = false, want true", suite)
		}
		if IsTLCPIBSDHSuite(suite) {
			t.Fatalf("IsTLCPIBSDHSuite(0x%04X) = true, want false", suite)
		}
	}
	for _, suite := range ibsdhSuites {
		if !IsTLCPIBCSuite(suite) {
			t.Fatalf("IsTLCPIBCSuite(0x%04X) = false, want true", suite)
		}
		if !IsTLCPIBSDHSuite(suite) {
			t.Fatalf("IsTLCPIBSDHSuite(0x%04X) = false, want true", suite)
		}
	}
	for _, suite := range otherSuites {
		if IsTLCPIBCSuite(suite) {
			t.Fatalf("IsTLCPIBCSuite(0x%04X) = true, want false", suite)
		}
		if IsTLCPIBSDHSuite(suite) {
			t.Fatalf("IsTLCPIBSDHSuite(0x%04X) = true, want false", suite)
		}
	}
}

// TestGetIBCParamDir 验证信任池目录位于工作目录下的 ibcparams。
func TestGetIBCParamDir(t *testing.T) {
	cfg := &Config{WorkDir: "/tmp/tlcpchan-test"}
	want := filepath.Join("/tmp/tlcpchan-test", "ibcparams")
	if got := cfg.GetIBCParamDir(); got != want {
		t.Fatalf("GetIBCParamDir() = %s, want %s", got, want)
	}
}

// TestValidateCipherSuites 验证配置校验能拒绝非法套件名与协议不匹配的套件。
func TestValidateCipherSuites(t *testing.T) {
	tests := []struct {
		name     string
		protocol string
		tlcp     []string
		tls      []string
		wantErr  bool
	}{
		{"空列表", "tlcp", nil, nil, false},
		{"合法 IBC 套件", "tlcp", []string{"IBC_SM4_GCM_SM3", "IBSDH_SM4_CBC_SM3"}, nil, false},
		{"合法证书套件", "tlcp", []string{"ECC_SM4_GCM_SM3"}, nil, false},
		{"非法套件名", "tlcp", []string{"NOT_A_SUITE"}, nil, true},
		{"TLS 套件用于 TLCP", "tlcp", []string{"TLS_RSA_WITH_AES_128_CBC_SHA"}, nil, true},
		{"IBC 套件用于 TLS", "tls", nil, []string{"IBC_SM4_GCM_SM3"}, true},
		{"合法 TLS 套件", "tls", nil, []string{"TLS_RSA_WITH_AES_128_CBC_SHA"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Instances: []InstanceConfig{
					{
						Name:     "test",
						Type:     "server",
						Protocol: tt.protocol,
						Listen:   ":20443",
						Target:   "127.0.0.1:20080",
						TLCP:     TLCPConfig{CipherSuites: tt.tlcp},
						TLS:      TLSConfig{CipherSuites: tt.tls},
					},
				},
			}
			err := Validate(cfg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
