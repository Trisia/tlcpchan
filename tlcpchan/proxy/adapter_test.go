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
			errMsg:  "协议类型为TLCP，但未提供有效的TLCP配置（需要keystore配置）",
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
			errMsg:  "协议类型为TLS，但未提供有效的TLS配置（需要keystore配置）",
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
			errMsg:  "协议类型为 auto，但未配置任何 keystore（至少需要配置 tlcp.keystore 或 tls.keystore）",
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
							"sign-cert": "test-sign-cert.pem",
							"sign-key":  "test-sign-key.pem",
							"enc-cert":  "test-enc-cert.pem",
							"enc-key":   "test-enc-key.pem",
						},
					},
				},
			},
			wantErr: false,
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
							"sign-cert": "test-sign-cert.pem",
							"sign-key":  "test-sign-key.pem",
						},
					},
				},
			},
			wantErr: false,
			errMsg:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
