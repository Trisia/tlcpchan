package config

import (
	"testing"
)

func TestClientAuthTypeDefaultValues(t *testing.T) {
	tests := []struct {
		name         string
		cfg          *Config
		expectedTLCP string
		expectedTLS  string
	}{
		{
			name: "空客户端认证类型应默认为 no-client-cert",
			cfg: &Config{
				Instances: []InstanceConfig{
					{
						Name:     "test-1",
						Type:     "server",
						Listen:   ":8443",
						Target:   "127.0.0.1:8080",
						Protocol: "auto",
						TLCP:     TLCPConfig{ClientAuthType: ""},
						TLS:      TLSConfig{ClientAuthType: ""},
					},
				},
			},
			expectedTLCP: "no-client-cert",
			expectedTLS:  "no-client-cert",
		},
		{
			name: "已设置客户端认证类型应保持不变",
			cfg: &Config{
				Instances: []InstanceConfig{
					{
						Name:     "test-2",
						Type:     "server",
						Listen:   ":8443",
						Target:   "127.0.0.1:8080",
						Protocol: "auto",
						TLCP:     TLCPConfig{ClientAuthType: "require-and-verify-client-cert"},
						TLS:      TLSConfig{ClientAuthType: "no-client-cert"},
					},
				},
			},
			expectedTLCP: "require-and-verify-client-cert",
			expectedTLS:  "no-client-cert",
		},
		{
			name: "混合情况",
			cfg: &Config{
				Instances: []InstanceConfig{
					{
						Name:     "test-3-1",
						Type:     "server",
						Listen:   ":8443",
						Target:   "127.0.0.1:8080",
						Protocol: "auto",
						TLCP:     TLCPConfig{ClientAuthType: ""},
						TLS:      TLSConfig{ClientAuthType: "request-client-cert"},
					},
					{
						Name:     "test-3-2",
						Type:     "client",
						Listen:   ":8444",
						Target:   "127.0.0.1:8081",
						Protocol: "tlcp",
						TLCP:     TLCPConfig{ClientAuthType: "no-client-cert"},
						TLS:      TLSConfig{ClientAuthType: ""},
					},
				},
			},
			expectedTLCP: "no-client-cert",
			expectedTLS:  "no-client-cert",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.cfg)
			if err != nil {
				t.Fatalf("验证失败: %v", err)
			}

			for i, inst := range tt.cfg.Instances {
				if inst.TLCP.ClientAuthType != tt.expectedTLCP && (i == 0 || tt.expectedTLCP == "") {
					t.Errorf("实例 %s TLCP 客户端认证类型错误: 期望 %s, 实际 %s", inst.Name, tt.expectedTLCP, inst.TLCP.ClientAuthType)
				}
			}
		})
	}
}

func TestParseTLCPClientAuth(t *testing.T) {
	tests := []struct {
		name        string
		auth        string
		expectError bool
	}{
		{"有效: no-client-cert", "no-client-cert", false},
		{"有效: request-client-cert", "request-client-cert", false},
		{"有效: require-any-client-cert", "require-any-client-cert", false},
		{"有效: verify-client-cert-if-given", "verify-client-cert-if-given", false},
		{"有效: require-and-verify-client-cert", "require-and-verify-client-cert", false},
		{"无效: invalid", "invalid", true},
		{"无效: one-way", "one-way", true},
		{"无效: 空字符串", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseTLCPClientAuth(tt.auth)
			if tt.expectError && err == nil {
				t.Errorf("期望错误但没有错误")
			}
			if !tt.expectError && err != nil {
				t.Errorf("不期望错误但有错误: %v", err)
			}
		})
	}
}

func TestParseTLSClientAuth(t *testing.T) {
	tests := []struct {
		name        string
		auth        string
		expectError bool
	}{
		{"有效: no-client-cert", "no-client-cert", false},
		{"有效: request-client-cert", "request-client-cert", false},
		{"有效: require-any-client-cert", "require-any-client-cert", false},
		{"有效: verify-client-cert-if-given", "verify-client-cert-if-given", false},
		{"有效: require-and-verify-client-cert", "require-and-verify-client-cert", false},
		{"无效: invalid", "invalid", true},
		{"无效: one-way", "one-way", true},
		{"无效: 空字符串", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseTLSClientAuth(tt.auth)
			if tt.expectError && err == nil {
				t.Errorf("期望错误但没有错误")
			}
			if !tt.expectError && err != nil {
				t.Errorf("不期望错误但有错误: %v", err)
			}
		})
	}
}
