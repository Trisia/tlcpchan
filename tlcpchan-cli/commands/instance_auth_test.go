package commands

import (
	"testing"
)

func TestInstanceCreateClientAuthDefaults(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedTLCPAuth string
		expectedTLSAuth  string
	}{
		{
			name: "未指定 --client-auth-type 参数应默认为 no-client-cert",
			args: []string{
				"--name", "test-1",
				"--type", "server",
				"--listen", ":8443",
				"--target", "127.0.0.1:8080",
				"--protocol", "auto",
			},
			expectedTLCPAuth: "no-client-cert",
			expectedTLSAuth:  "no-client-cert",
		},
		{
			name: "指定 --tlcp-client-auth-type 参数应使用指定值",
			args: []string{
				"--name", "test-2",
				"--type", "server",
				"--listen", ":8444",
				"--target", "127.0.0.1:8081",
				"--protocol", "auto",
				"--tlcp-client-auth-type", "require-and-verify-client-cert",
			},
			expectedTLCPAuth: "require-and-verify-client-cert",
			expectedTLSAuth:  "no-client-cert",
		},
		{
			name: "指定 --tls-client-auth-type 参数应使用指定值",
			args: []string{
				"--name", "test-3",
				"--type", "server",
				"--listen", ":8445",
				"--target", "127.0.0.1:8082",
				"--protocol", "auto",
				"--tls-client-auth-type", "request-client-cert",
			},
			expectedTLCPAuth: "no-client-cert",
			expectedTLSAuth:  "request-client-cert",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flagSet("create")
			tlcpClientAuthType := fs.String("tlcp-client-auth-type", "no-client-cert", "TLCP客户端认证类型")
			tlsClientAuthType := fs.String("tls-client-auth-type", "no-client-cert", "TLS客户端认证类型")
			_ = fs.String("name", "", "实例名称（必需）")
			_ = fs.String("type", "server", "类型（server/client/http-server/http-client）")
			_ = fs.String("listen", "", "监听地址（必需）")
			_ = fs.String("target", "", "目标地址（必需）")
			_ = fs.String("protocol", "auto", "协议（auto/tlcp/tls）")

			if err := fs.Parse(tt.args); err != nil {
				t.Fatalf("参数解析失败: %v", err)
			}

			if *tlcpClientAuthType != tt.expectedTLCPAuth {
				t.Errorf("TLCP客户端认证类型错误: 期望 %s, 实际 %s", tt.expectedTLCPAuth, *tlcpClientAuthType)
			}
			if *tlsClientAuthType != tt.expectedTLSAuth {
				t.Errorf("TLS客户端认证类型错误: 期望 %s, 实际 %s", tt.expectedTLSAuth, *tlsClientAuthType)
			}
		})
	}
}

func TestPopulateInstanceConfig(t *testing.T) {
	tlcpClientAuthDefault := "no-client-cert"
	tlsClientAuthDefault := "no-client-cert"

	testCases := []struct {
		name               string
		tlcpClientAuthFlag string
		tlsClientAuthFlag  string
		expectedTLCPAuth   string
		expectedTLSAuth    string
	}{
		{"使用默认值", "", "", tlcpClientAuthDefault, tlsClientAuthDefault},
		{"指定 TLCP require-and-verify", "require-and-verify-client-cert", "", "require-and-verify-client-cert", tlsClientAuthDefault},
		{"指定 TLS request", "", "request-client-cert", tlcpClientAuthDefault, "request-client-cert"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tlcpClientAuth := tc.tlcpClientAuthFlag
			tlsClientAuth := tc.tlsClientAuthFlag
			if tlcpClientAuth == "" {
				tlcpClientAuth = tlcpClientAuthDefault
			}
			if tlsClientAuth == "" {
				tlsClientAuth = tlsClientAuthDefault
			}

			if tlcpClientAuth != tc.expectedTLCPAuth {
				t.Errorf("TLCP客户端认证类型错误: 期望 %s, 实际 %s", tc.expectedTLCPAuth, tlcpClientAuth)
			}
			if tlsClientAuth != tc.expectedTLSAuth {
				t.Errorf("TLS客户端认证类型错误: 期望 %s, 实际 %s", tc.expectedTLSAuth, tlsClientAuth)
			}
		})
	}
}
