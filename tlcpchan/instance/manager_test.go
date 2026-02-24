package instance

import (
	"testing"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/logger"
	"github.com/Trisia/tlcpchan/security"
	"github.com/Trisia/tlcpchan/security/rootcert"
)

func TestInstanceManagerCreateAuthDefaults(t *testing.T) {
	log, _ := logger.New(logger.LogConfig{Level: "info", Enabled: false})
	ksMgr := security.NewKeyStoreManager()
	rcMgr := rootcert.NewManager("")
	mgr := NewManager(log, ksMgr, rcMgr)

	tests := []struct {
		name         string
		cfg          *config.InstanceConfig
		expectedTLCP string
		expectedTLS  string
	}{
		{
			name: "空认证模式应默认为单向认证",
			cfg: &config.InstanceConfig{
				Name:     "test-1",
				Type:     "client",
				Protocol: "auto",
				Listen:   ":8443",
				Target:   "127.0.0.1:8080",
				TLCP: config.TLCPConfig{
					ClientAuthType: "",
				},
				TLS: config.TLSConfig{
					ClientAuthType: "",
				},
			},
			expectedTLCP: "no-client-cert",
			expectedTLS:  "no-client-cert",
		},
		{
			name: "已设置认证模式应保持不变",
			cfg: &config.InstanceConfig{
				Name:     "test-2",
				Type:     "client",
				Protocol: "auto",
				Listen:   ":8444",
				Target:   "127.0.0.1:8081",
				TLCP: config.TLCPConfig{
					ClientAuthType: "require-and-verify-client-cert",
				},
				TLS: config.TLSConfig{
					ClientAuthType: "no-client-cert",
				},
			},
			expectedTLCP: "require-and-verify-client-cert",
			expectedTLS:  "no-client-cert",
		},
		{
			name: "部分设置",
			cfg: &config.InstanceConfig{
				Name:     "test-3",
				Type:     "client",
				Protocol: "tlcp",
				Listen:   ":8445",
				Target:   "127.0.0.1:8082",
				TLCP: config.TLCPConfig{
					ClientAuthType: "no-client-cert",
				},
				TLS: config.TLSConfig{},
			},
			expectedTLCP: "no-client-cert",
			expectedTLS:  "no-client-cert",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst, err := mgr.Create(tt.cfg)
			if err != nil {
				t.Fatalf("创建实例失败: %v", err)
			}
			defer mgr.Delete(tt.cfg.Name)

			cfg := inst.Config()
			if cfg.TLCP.ClientAuthType != tt.expectedTLCP {
				t.Errorf("TLCP 认证模式错误: 期望 %s, 实际 %s", tt.expectedTLCP, cfg.TLCP.ClientAuthType)
			}
			if cfg.TLS.ClientAuthType != tt.expectedTLS {
				t.Errorf("TLS 认证模式错误: 期望 %s, 实际 %s", tt.expectedTLS, cfg.TLS.ClientAuthType)
			}
		})
	}
}

func TestInstanceManagerCreateDuplicate(t *testing.T) {
	log, _ := logger.New(logger.LogConfig{Level: "info", Enabled: false})
	ksMgr := security.NewKeyStoreManager()
	rcMgr := rootcert.NewManager("")
	mgr := NewManager(log, ksMgr, rcMgr)

	cfg := &config.InstanceConfig{
		Name:     "duplicate-test",
		Type:     "client",
		Protocol: "auto",
		Listen:   ":8446",
		Target:   "127.0.0.1:8083",
		TLCP: config.TLCPConfig{
			ClientAuthType: "no-client-cert",
		},
	}

	_, err := mgr.Create(cfg)
	if err != nil {
		t.Fatalf("第一次创建失败: %v", err)
	}
	defer mgr.Delete(cfg.Name)

	_, err = mgr.Create(cfg)
	if err == nil {
		t.Error("期望创建重复实例失败，但成功了")
	}
}
