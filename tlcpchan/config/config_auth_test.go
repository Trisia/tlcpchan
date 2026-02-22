package config

import (
	"testing"
)

func TestAuthDefaultValues(t *testing.T) {
	tests := []struct {
		name         string
		cfg          *Config
		expectedTLCP string
		expectedTLS  string
	}{
		{
			name: "空认证模式应默认为单向认证",
			cfg: &Config{
				Instances: []InstanceConfig{
					{
						Name:     "test-1",
						Type:     "server",
						Listen:   ":8443",
						Target:   "127.0.0.1:8080",
						Protocol: "auto",
						TLCP:     TLCPConfig{Auth: ""},
						TLS:      TLSConfig{Auth: ""},
					},
				},
			},
			expectedTLCP: "one-way",
			expectedTLS:  "one-way",
		},
		{
			name: "已设置认证模式应保持不变",
			cfg: &Config{
				Instances: []InstanceConfig{
					{
						Name:     "test-2",
						Type:     "server",
						Listen:   ":8443",
						Target:   "127.0.0.1:8080",
						Protocol: "auto",
						TLCP:     TLCPConfig{Auth: "mutual"},
						TLS:      TLSConfig{Auth: "none"},
					},
				},
			},
			expectedTLCP: "mutual",
			expectedTLS:  "none",
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
						TLCP:     TLCPConfig{Auth: ""},
						TLS:      TLSConfig{Auth: "mutual"},
					},
					{
						Name:     "test-3-2",
						Type:     "client",
						Listen:   ":8444",
						Target:   "127.0.0.1:8081",
						Protocol: "tlcp",
						TLCP:     TLCPConfig{Auth: "none"},
						TLS:      TLSConfig{Auth: ""},
					},
				},
			},
			expectedTLCP: "one-way",
			expectedTLS:  "one-way",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.cfg)
			if err != nil {
				t.Fatalf("验证失败: %v", err)
			}

			for i, inst := range tt.cfg.Instances {
				if inst.TLCP.Auth != tt.expectedTLCP && (i == 0 || tt.expectedTLCP == "") {
					t.Errorf("实例 %s TLCP 认证模式错误: 期望 %s, 实际 %s", inst.Name, tt.expectedTLCP, inst.TLCP.Auth)
				}
			}
		})
	}
}

func TestValidateAuthMode(t *testing.T) {
	tests := []struct {
		name        string
		auth        string
		expectError bool
	}{
		{"有效: none", "none", false},
		{"有效: one-way", "one-way", false},
		{"有效: mutual", "mutual", false},
		{"无效: invalid", "invalid", true},
		{"无效: two-way", "two-way", true},
		{"无效: 空字符串", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Instances: []InstanceConfig{
					{
						Name:     "test",
						Type:     "server",
						Listen:   ":8443",
						Target:   "127.0.0.1:8080",
						Protocol: "auto",
						TLCP:     TLCPConfig{Auth: tt.auth},
						TLS:      TLSConfig{Auth: tt.auth},
					},
				},
			}

			err := Validate(cfg)
			if tt.expectError && err == nil {
				t.Errorf("期望错误但没有错误")
			}
			if !tt.expectError && err != nil {
				t.Errorf("不期望错误但有错误: %v", err)
			}
		})
	}
}
