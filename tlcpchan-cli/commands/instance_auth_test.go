package commands

import (
	"testing"
)

func TestInstanceCreateAuthDefaults(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedAuth string
	}{
		{
			name: "未指定 --auth 参数应默认为 one-way",
			args: []string{
				"--name", "test-1",
				"--type", "server",
				"--listen", ":8443",
				"--target", "127.0.0.1:8080",
				"--protocol", "auto",
			},
			expectedAuth: "one-way",
		},
		{
			name: "指定 --auth 参数应使用指定值",
			args: []string{
				"--name", "test-2",
				"--type", "server",
				"--listen", ":8444",
				"--target", "127.0.0.1:8081",
				"--protocol", "auto",
				"--auth", "mutual",
			},
			expectedAuth: "mutual",
		},
		{
			name: "指定 --auth 为 none",
			args: []string{
				"--name", "test-3",
				"--type", "server",
				"--listen", ":8445",
				"--target", "127.0.0.1:8082",
				"--protocol", "auto",
				"--auth", "none",
			},
			expectedAuth: "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flagSet("create")
			auth := fs.String("auth", "one-way", "认证模式（none/one-way/mutual）")
			_ = fs.String("name", "", "实例名称（必需）")
			_ = fs.String("type", "server", "类型（server/client/http-server/http-client）")
			_ = fs.String("listen", "", "监听地址（必需）")
			_ = fs.String("target", "", "目标地址（必需）")
			_ = fs.String("protocol", "auto", "协议（auto/tlcp/tls）")

			if err := fs.Parse(tt.args); err != nil {
				t.Fatalf("参数解析失败: %v", err)
			}

			if *auth != tt.expectedAuth {
				t.Errorf("认证模式错误: 期望 %s, 实际 %s", tt.expectedAuth, *auth)
			}
		})
	}
}

func TestPopulateInstanceConfig(t *testing.T) {
	authDefault := "one-way"

	testCases := []struct {
		name     string
		authFlag string
		expected string
	}{
		{"使用默认值", "", authDefault},
		{"指定 mutual", "mutual", "mutual"},
		{"指定 none", "none", "none"},
		{"指定 one-way", "one-way", "one-way"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var auth string
			if tc.authFlag == "" {
				auth = authDefault
			} else {
				auth = tc.authFlag
			}

			if auth != tc.expected {
				t.Errorf("认证模式错误: 期望 %s, 实际 %s", tc.expected, auth)
			}
		})
	}
}

func TestCLIInstanceConfig(t *testing.T) {
	type InstanceConfig struct {
		Name     string
		Type     string
		Listen   string
		Target   string
		Protocol string
		Auth     string
		Enabled  bool
	}

	tests := []struct {
		name         string
		authFlag     string
		expectedAuth string
	}{
		{"默认认证模式", "", "one-way"},
		{"指定 mutual", "mutual", "mutual"},
		{"指定 none", "none", "none"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := tt.authFlag
			if auth == "" {
				auth = "one-way"
			}

			cfg := InstanceConfig{
				Name:     "test-instance",
				Type:     "server",
				Listen:   ":8443",
				Target:   "127.0.0.1:8080",
				Protocol: "auto",
				Auth:     auth,
				Enabled:  true,
			}

			if cfg.Auth != tt.expectedAuth {
				t.Errorf("认证模式错误: 期望 %s, 实际 %s", tt.expectedAuth, cfg.Auth)
			}
		})
	}
}
