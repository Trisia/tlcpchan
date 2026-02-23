package proxy

import (
	"testing"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/security"
	"github.com/Trisia/tlcpchan/security/keystore"
)

// TestReloadServerConfigWithMissingKeystore 测试协议类型与 keystore 配置的一致性校验
func TestReloadServerConfigWithMissingKeystore(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.InstanceConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "TLCP协议但未配置keystore",
			cfg: &config.InstanceConfig{
				Name:     "test-tlcp-server",
				Type:     "server",
				Protocol: "tlcp",
				TLCP: config.TLCPConfig{
					Keystore: nil,
				},
			},
			wantErr: true,
			errMsg:  "服务端配置错误: 协议类型为 TLCP，但未配置 tlcp.keystore",
		},
		{
			name: "TLS协议但未配置keystore",
			cfg: &config.InstanceConfig{
				Name:     "test-tls-server",
				Type:     "server",
				Protocol: "tls",
				TLS: config.TLSConfig{
					Keystore: nil,
				},
			},
			wantErr: true,
			errMsg:  "服务端配置错误: 协议类型为 TLS，但未配置 tls.keystore",
		},
		{
			name: "Auto协议但未配置任何keystore",
			cfg: &config.InstanceConfig{
				Name:     "test-auto-server",
				Type:     "server",
				Protocol: "auto",
				TLCP: config.TLCPConfig{
					Keystore: nil,
				},
				TLS: config.TLSConfig{
					Keystore: nil,
				},
			},
			wantErr: true,
			errMsg:  "服务端配置错误: 协议类型为 auto，但未配置任何 keystore（至少需要配置 tlcp.keystore 或 tls.keystore）",
		},
		{
			name: "Auto协议配置了TLCP keystore",
			cfg: &config.InstanceConfig{
				Name:     "test-auto-tlcp-server",
				Type:     "server",
				Protocol: "auto",
				TLCP: config.TLCPConfig{
					Keystore: &config.KeyStoreConfig{
						Type: keystore.LoaderTypeFile,
						Params: map[string]string{
							"cert-file": "test-cert.pem",
							"key-file":  "test-key.pem",
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "",
		},
		{
			name: "Auto协议配置了TLS keystore",
			cfg: &config.InstanceConfig{
				Name:     "test-auto-tls-server",
				Type:     "server",
				Protocol: "auto",
				TLS: config.TLSConfig{
					Keystore: &config.KeyStoreConfig{
						Type: keystore.LoaderTypeFile,
						Params: map[string]string{
							"cert-file": "test-cert.pem",
							"key-file":  "test-key.pem",
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 KeyStoreManager 和（RootCertManager
			keyStoreMgr := security.NewKeyStoreManager()
			rootCertMgr := security.NewRootCertManager(".")

			adapter, err := NewTLCPAdapter(keyStoreMgr, rootCertMgr)
			if err != nil {
				t.Fatalf("NewTLCPAdapter() error = %v", err)
			}

			err = adapter.ReloadConfig(tt.cfg)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ReloadConfig() 期望返回错误，但没有返回")
				} else if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("ReloadConfig() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil && tt.errMsg != "" && err.Error() == tt.errMsg {
					t.Errorf("ReloadConfig() 不应返回错误: %v", err)
				}
			}
		})
	}
}
