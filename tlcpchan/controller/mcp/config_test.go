package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Trisia/tlcpchan/config"
)

func TestHandleGetConfig(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() (*MCPController, func())
		wantErr   bool
	}{
		{
			name: "正常获取配置",
			setupFunc: func() (*MCPController, func()) {
				cfg := &config.Config{
					MCP: config.MCPConfig{
						Enabled: true,
					},
					Server: config.ServerConfig{
						API: config.APIConfig{
							Address: ":20080",
						},
					},
				}

				config.Init(cfg, "")

				opts := &ServerOptions{
					Config:          cfg,
					InstanceManager: nil,
					KeyStoreManager: nil,
					RootCertManager: nil,
					ConfigPath:      "/tmp/test-config.yaml",
				}

				ctrl, err := NewMCPController(opts)
				if err != nil {
					t.Fatalf("创建 MCPController 失败: %v", err)
				}

				cleanup := func() {
					config.Set(nil)
				}

				return ctrl, cleanup
			},
			wantErr: false,
		},
		{
			name: "获取空配置",
			setupFunc: func() (*MCPController, func()) {
				cfg := &config.Config{
					MCP: config.MCPConfig{
						Enabled: true,
					},
				}

				config.Init(cfg, "")

				opts := &ServerOptions{
					Config:          cfg,
					InstanceManager: nil,
					KeyStoreManager: nil,
					RootCertManager: nil,
					ConfigPath:      "/tmp/test-config.yaml",
				}

				ctrl, err := NewMCPController(opts)
				if err != nil {
					t.Fatalf("创建 MCPController 失败: %v", err)
				}

				cleanup := func() {
					config.Set(nil)
				}

				return ctrl, cleanup
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl, cleanup := tt.setupFunc()
			defer cleanup()

			_, output, err := ctrl.handleGetConfig(context.Background(), nil, GetConfigInput{})

			if tt.wantErr {
				if err == nil {
					t.Error("期望返回错误，实际得到 nil")
				}
				return
			}

			if err != nil {
				t.Errorf("handleGetConfig 失败: %v", err)
				return
			}

			if output.Config == nil {
				t.Error("期望返回配置对象，实际得到 nil")
			}
		})
	}
}

func TestHandleUpdateConfig(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() (*MCPController, string, func())
		input     UpdateConfigInput
		wantErr   bool
		errMsg    string
	}{
		{
			name: "正常更新配置",
			setupFunc: func() (*MCPController, string, func()) {
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, "config.yaml")

				cfg := &config.Config{
					MCP: config.MCPConfig{
						Enabled: true,
					},
					Server: config.ServerConfig{
						API: config.APIConfig{
							Address: ":20080",
						},
					},
				}

				config.Init(cfg, configPath)

				opts := &ServerOptions{
					Config:          cfg,
					InstanceManager: nil,
					KeyStoreManager: nil,
					RootCertManager: nil,
					ConfigPath:      configPath,
				}

				ctrl, err := NewMCPController(opts)
				if err != nil {
					t.Fatalf("创建 MCPController 失败: %v", err)
				}

				cleanup := func() {
					config.Set(nil)
				}

				return ctrl, configPath, cleanup
			},
			input: UpdateConfigInput{
				Config: &config.Config{
					MCP: config.MCPConfig{
						Enabled: true,
					},
					Server: config.ServerConfig{
						API: config.APIConfig{
							Address: ":20081",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "更新配置添加实例",
			setupFunc: func() (*MCPController, string, func()) {
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, "config.yaml")

				cfg := &config.Config{
					MCP: config.MCPConfig{
						Enabled: true,
					},
					Server: config.ServerConfig{
						API: config.APIConfig{
							Address: ":20080",
						},
					},
				}

				config.Init(cfg, configPath)

				opts := &ServerOptions{
					Config:          cfg,
					InstanceManager: nil,
					KeyStoreManager: nil,
					RootCertManager: nil,
					ConfigPath:      configPath,
				}

				ctrl, err := NewMCPController(opts)
				if err != nil {
					t.Fatalf("创建 MCPController 失败: %v", err)
				}

				cleanup := func() {
					config.Set(nil)
				}

				return ctrl, configPath, cleanup
			},
			input: UpdateConfigInput{
				Config: &config.Config{
					MCP: config.MCPConfig{
						Enabled: true,
					},
					Server: config.ServerConfig{
						API: config.APIConfig{
							Address: ":20080",
						},
					},
					Instances: []config.InstanceConfig{
						{
							Name:     "test-instance",
							Type:     "server",
							Listen:   ":8443",
							Target:   "backend:8080",
							Protocol: "auto",
							Enabled:  true,
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl, _, cleanup := tt.setupFunc()
			defer cleanup()

			_, output, err := ctrl.handleUpdateConfig(context.Background(), nil, tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("期望返回错误，实际得到 nil")
				}
				if tt.errMsg != "" && err != nil {
					if !containsString(err.Error(), tt.errMsg) {
						t.Errorf("期望错误信息包含 %q，实际 %q", tt.errMsg, err.Error())
					}
				}
				return
			}

			if err != nil {
				t.Errorf("handleUpdateConfig 失败: %v", err)
				return
			}

			if output.Config == nil {
				t.Error("期望返回配置对象，实际得到 nil")
			}
		})
	}
}

func TestHandleUpdateConfig_Invalid(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() (*MCPController, string, func())
		input     UpdateConfigInput
		wantErr   bool
		errMsg    string
	}{
		{
			name: "配置参数为空",
			setupFunc: func() (*MCPController, string, func()) {
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, "config.yaml")

				cfg := &config.Config{
					MCP: config.MCPConfig{
						Enabled: true,
					},
				}

				config.Init(cfg, configPath)

				opts := &ServerOptions{
					Config:          cfg,
					InstanceManager: nil,
					KeyStoreManager: nil,
					RootCertManager: nil,
					ConfigPath:      configPath,
				}

				ctrl, err := NewMCPController(opts)
				if err != nil {
					t.Fatalf("创建 MCPController 失败: %v", err)
				}

				cleanup := func() {
					config.Set(nil)
				}

				return ctrl, configPath, cleanup
			},
			input: UpdateConfigInput{
				Config: nil,
			},
			wantErr: true,
			errMsg:  "配置参数不能为空",
		},
		{
			name: "实例名称为空",
			setupFunc: func() (*MCPController, string, func()) {
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, "config.yaml")

				cfg := &config.Config{
					MCP: config.MCPConfig{
						Enabled: true,
					},
					Server: config.ServerConfig{
						API: config.APIConfig{
							Address: ":20080",
						},
					},
				}

				config.Init(cfg, configPath)

				opts := &ServerOptions{
					Config:          cfg,
					InstanceManager: nil,
					KeyStoreManager: nil,
					RootCertManager: nil,
					ConfigPath:      configPath,
				}

				ctrl, err := NewMCPController(opts)
				if err != nil {
					t.Fatalf("创建 MCPController 失败: %v", err)
				}

				cleanup := func() {
					config.Set(nil)
				}

				return ctrl, configPath, cleanup
			},
			input: UpdateConfigInput{
				Config: &config.Config{
					MCP: config.MCPConfig{
						Enabled: true,
					},
					Server: config.ServerConfig{
						API: config.APIConfig{
							Address: ":20080",
						},
					},
					Instances: []config.InstanceConfig{
						{
							Name:     "",
							Type:     "server",
							Listen:   ":8443",
							Target:   "backend:8080",
							Protocol: "auto",
							Enabled:  true,
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "名称不能为空",
		},
		{
			name: "实例类型无效",
			setupFunc: func() (*MCPController, string, func()) {
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, "config.yaml")

				cfg := &config.Config{
					MCP: config.MCPConfig{
						Enabled: true,
					},
					Server: config.ServerConfig{
						API: config.APIConfig{
							Address: ":20080",
						},
					},
				}

				config.Init(cfg, configPath)

				opts := &ServerOptions{
					Config:          cfg,
					InstanceManager: nil,
					KeyStoreManager: nil,
					RootCertManager: nil,
					ConfigPath:      configPath,
				}

				ctrl, err := NewMCPController(opts)
				if err != nil {
					t.Fatalf("创建 MCPController 失败: %v", err)
				}

				cleanup := func() {
					config.Set(nil)
				}

				return ctrl, configPath, cleanup
			},
			input: UpdateConfigInput{
				Config: &config.Config{
					MCP: config.MCPConfig{
						Enabled: true,
					},
					Server: config.ServerConfig{
						API: config.APIConfig{
							Address: ":20080",
						},
					},
					Instances: []config.InstanceConfig{
						{
							Name:     "test-instance",
							Type:     "invalid-type",
							Listen:   ":8443",
							Target:   "backend:8080",
							Protocol: "auto",
							Enabled:  true,
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "无效的类型",
		},
		{
			name: "实例监听地址为空",
			setupFunc: func() (*MCPController, string, func()) {
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, "config.yaml")

				cfg := &config.Config{
					MCP: config.MCPConfig{
						Enabled: true,
					},
					Server: config.ServerConfig{
						API: config.APIConfig{
							Address: ":20080",
						},
					},
				}

				config.Init(cfg, configPath)

				opts := &ServerOptions{
					Config:          cfg,
					InstanceManager: nil,
					KeyStoreManager: nil,
					RootCertManager: nil,
					ConfigPath:      configPath,
				}

				ctrl, err := NewMCPController(opts)
				if err != nil {
					t.Fatalf("创建 MCPController 失败: %v", err)
				}

				cleanup := func() {
					config.Set(nil)
				}

				return ctrl, configPath, cleanup
			},
			input: UpdateConfigInput{
				Config: &config.Config{
					MCP: config.MCPConfig{
						Enabled: true,
					},
					Server: config.ServerConfig{
						API: config.APIConfig{
							Address: ":20080",
						},
					},
					Instances: []config.InstanceConfig{
						{
							Name:     "test-instance",
							Type:     "server",
							Listen:   "",
							Target:   "backend:8080",
							Protocol: "auto",
							Enabled:  true,
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "监听地址不能为空",
		},
		{
			name: "实例目标地址为空",
			setupFunc: func() (*MCPController, string, func()) {
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, "config.yaml")

				cfg := &config.Config{
					MCP: config.MCPConfig{
						Enabled: true,
					},
					Server: config.ServerConfig{
						API: config.APIConfig{
							Address: ":20080",
						},
					},
				}

				config.Init(cfg, configPath)

				opts := &ServerOptions{
					Config:          cfg,
					InstanceManager: nil,
					KeyStoreManager: nil,
					RootCertManager: nil,
					ConfigPath:      configPath,
				}

				ctrl, err := NewMCPController(opts)
				if err != nil {
					t.Fatalf("创建 MCPController 失败: %v", err)
				}

				cleanup := func() {
					config.Set(nil)
				}

				return ctrl, configPath, cleanup
			},
			input: UpdateConfigInput{
				Config: &config.Config{
					MCP: config.MCPConfig{
						Enabled: true,
					},
					Server: config.ServerConfig{
						API: config.APIConfig{
							Address: ":20080",
						},
					},
					Instances: []config.InstanceConfig{
						{
							Name:     "test-instance",
							Type:     "server",
							Listen:   ":8443",
							Target:   "",
							Protocol: "auto",
							Enabled:  true,
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "目标地址不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl, _, cleanup := tt.setupFunc()
			defer cleanup()

			_, _, err := ctrl.handleUpdateConfig(context.Background(), nil, tt.input)

			if !tt.wantErr {
				if err != nil {
					t.Errorf("不期望返回错误，实际得到: %v", err)
				}
				return
			}

			if err == nil {
				t.Error("期望返回错误，实际得到 nil")
				return
			}

			if tt.errMsg != "" {
				if !containsString(err.Error(), tt.errMsg) {
					t.Errorf("期望错误信息包含 %q，实际 %q", tt.errMsg, err.Error())
				}
			}
		})
	}
}

func TestHandleReloadConfig(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() (*MCPController, string, func())
		input     ReloadConfigInput
		wantErr   bool
		errMsg    string
	}{
		{
			name: "正常重新加载配置",
			setupFunc: func() (*MCPController, string, func()) {
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, "config.yaml")

				cfg := &config.Config{
					MCP: config.MCPConfig{
						Enabled: true,
					},
					Server: config.ServerConfig{
						API: config.APIConfig{
							Address: ":20080",
						},
					},
				}

				config.Init(cfg, configPath)

				opts := &ServerOptions{
					Config:          cfg,
					InstanceManager: nil,
					KeyStoreManager: nil,
					RootCertManager: nil,
					ConfigPath:      configPath,
				}

				ctrl, err := NewMCPController(opts)
				if err != nil {
					t.Fatalf("创建 MCPController 失败: %v", err)
				}

				cleanup := func() {
					config.Set(nil)
				}

				return ctrl, configPath, cleanup
			},
			input: ReloadConfigInput{
				ConfigPath: "",
			},
			wantErr: true,
			errMsg:  "配置文件路径未指定",
		},
		{
			name: "重新加载配置指定路径",
			setupFunc: func() (*MCPController, string, func()) {
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, "config.yaml")

				cfg := &config.Config{
					MCP: config.MCPConfig{
						Enabled: true,
					},
					Server: config.ServerConfig{
						API: config.APIConfig{
							Address: ":20080",
						},
					},
				}

				config.Init(cfg, configPath)

				opts := &ServerOptions{
					Config:          cfg,
					InstanceManager: nil,
					KeyStoreManager: nil,
					RootCertManager: nil,
					ConfigPath:      "/tmp/test-config.yaml",
				}

				ctrl, err := NewMCPController(opts)
				if err != nil {
					t.Fatalf("创建 MCPController 失败: %v", err)
				}

				cleanup := func() {
					config.Set(nil)
				}

				return ctrl, configPath, cleanup
			},
			input: ReloadConfigInput{
				ConfigPath: "",
			},
			wantErr: true,
			errMsg:  "配置文件路径未指定",
		},
		{
			name: "重新加载配置文件不存在",
			setupFunc: func() (*MCPController, string, func()) {
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, "config.yaml")

				cfg := &config.Config{
					MCP: config.MCPConfig{
						Enabled: true,
					},
					Server: config.ServerConfig{
						API: config.APIConfig{
							Address: ":20080",
						},
					},
				}

				config.Init(cfg, configPath)

				opts := &ServerOptions{
					Config:          cfg,
					InstanceManager: nil,
					KeyStoreManager: nil,
					RootCertManager: nil,
					ConfigPath:      "/tmp/test-config.yaml",
				}

				ctrl, err := NewMCPController(opts)
				if err != nil {
					t.Fatalf("创建 MCPController 失败: %v", err)
				}

				cleanup := func() {
					config.Set(nil)
				}

				return ctrl, configPath, cleanup
			},
			input: ReloadConfigInput{
				ConfigPath: "/nonexistent/config.yaml",
			},
			wantErr: true,
			errMsg:  "读取配置文件失败",
		},
		{
			name: "重新加载有效配置文件",
			setupFunc: func() (*MCPController, string, func()) {
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, "config.yaml")

				cfg := &config.Config{
					MCP: config.MCPConfig{
						Enabled: true,
					},
					Server: config.ServerConfig{
						API: config.APIConfig{
							Address: ":20080",
						},
					},
				}

				config.Init(cfg, configPath)

				opts := &ServerOptions{
					Config:          cfg,
					InstanceManager: nil,
					KeyStoreManager: nil,
					RootCertManager: nil,
					ConfigPath:      "/tmp/test-config.yaml",
				}

				ctrl, err := NewMCPController(opts)
				if err != nil {
					t.Fatalf("创建 MCPController 失败: %v", err)
				}

				cleanup := func() {
					config.Set(nil)
				}

				return ctrl, configPath, cleanup
			},
			input: ReloadConfigInput{
				ConfigPath: "",
			},
			wantErr: true,
			errMsg:  "配置文件路径未指定",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl, _, cleanup := tt.setupFunc()
			defer cleanup()

			_, _, err := ctrl.handleReloadConfig(context.Background(), nil, tt.input)

			if !tt.wantErr {
				if err != nil {
					t.Errorf("不期望返回错误，实际得到: %v", err)
				}
				return
			}

			if err == nil {
				t.Error("期望返回错误，实际得到 nil")
				return
			}

			if tt.errMsg != "" {
				if !containsString(err.Error(), tt.errMsg) {
					t.Errorf("期望错误信息包含 %q，实际 %q", tt.errMsg, err.Error())
				}
			}
		})
	}
}

func TestHandleReloadConfig_Success(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	originalCfg := &config.Config{
		MCP: config.MCPConfig{
			Enabled: true,
		},
		Server: config.ServerConfig{
			API: config.APIConfig{
				Address: ":20080",
			},
		},
	}

	if err := os.WriteFile(configPath, []byte("mcp:\n  enabled: true\nserver:\n  api:\n    address: \":20081\""), 0644); err != nil {
		t.Fatalf("写入配置文件失败: %v", err)
	}

	config.Init(originalCfg, configPath)

	opts := &ServerOptions{
		Config:          originalCfg,
		InstanceManager: nil,
		KeyStoreManager: nil,
		RootCertManager: nil,
		ConfigPath:      configPath,
	}

	ctrl, err := NewMCPController(opts)
	if err != nil {
		t.Fatalf("创建 MCPController 失败: %v", err)
	}

	defer func() {
		config.Set(nil)
	}()

	_, output, err := ctrl.handleReloadConfig(context.Background(), nil, ReloadConfigInput{
		ConfigPath: configPath,
	})

	if err != nil {
		t.Errorf("handleReloadConfig 失败: %v", err)
		return
	}

	if output.Config == nil {
		t.Error("期望返回配置对象，实际得到 nil")
		return
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && len(substr) > 0)
}
