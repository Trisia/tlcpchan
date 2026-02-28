package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestConfigSingleton(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "Init and Get",
			fn:   testInitAndGet,
		},
		{
			name: "Set and Get",
			fn:   testSetAndGet,
		},
		{
			name: "Concurrent Access",
			fn:   testConcurrentAccess,
		},
		{
			name: "LoadAndInit and SaveAndUpdate",
			fn:   testLoadAndSave,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetSingleton()
			tt.fn(t)
		})
	}
}

func resetSingleton() {
	configMutex.Lock()
	defer configMutex.Unlock()
	globalConfig = nil
	globalConfigPath = ""
}

func testInitAndGet(t *testing.T) {
	cfg := Default()
	cfg.Server.API.Address = ":12345"

	Init(cfg, "")

	result := Get()
	if result == nil {
		t.Fatal("Get() 返回 nil")
	}
	if result.Server.API.Address != ":12345" {
		t.Errorf("期望地址 :12345, 实际 %s", result.Server.API.Address)
	}
}

func testSetAndGet(t *testing.T) {
	cfg1 := Default()
	cfg1.Server.API.Address = ":11111"
	Init(cfg1, "")

	cfg2 := Default()
	cfg2.Server.API.Address = ":22222"
	Set(cfg2)

	result := Get()
	if result.Server.API.Address != ":22222" {
		t.Errorf("期望地址 :22222, 实际 %s", result.Server.API.Address)
	}
}

func testConcurrentAccess(t *testing.T) {
	cfg := Default()
	cfg.Server.API.Address = ":99999"
	Init(cfg, "")

	var wg sync.WaitGroup
	readCount := 100
	writeCount := 10

	wg.Add(readCount + writeCount)

	for i := 0; i < readCount; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				Get()
			}
		}()
	}

	for i := 0; i < writeCount; i++ {
		go func(idx int) {
			defer wg.Done()
			newCfg := Default()
			newCfg.Server.API.Address = fmt.Sprintf(":%d", 10000+idx)
			Set(newCfg)
		}(i)
	}

	wg.Wait()
}

func testLoadAndSave(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	cfg1 := Default()
	cfg1.Server.API.Address = ":54321"
	cfg1.Instances = []InstanceConfig{
		{
			Name:     "test-instance",
			Type:     "server",
			Listen:   ":8443",
			Target:   "backend:8080",
			Protocol: "auto",
			Enabled:  true,
		},
	}

	data, err := cfgMarshal(cfg1)
	if err != nil {
		t.Fatalf("序列化配置失败: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("写入配置文件失败: %v", err)
	}

	if err := LoadAndInit(configPath); err != nil {
		t.Fatalf("LoadAndInit 失败: %v", err)
	}

	result := Get()
	if result.Server.API.Address != ":54321" {
		t.Errorf("期望地址 :54321, 实际 %s", result.Server.API.Address)
	}
	if len(result.Instances) != 1 {
		t.Errorf("期望 1 个实例, 实际 %d", len(result.Instances))
	}

	cfg2 := Default()
	cfg2.Server.API.Address = ":65432"
	if err := SaveAndUpdate(cfg2); err != nil {
		t.Fatalf("SaveAndUpdate 失败: %v", err)
	}

	result2 := Get()
	if result2.Server.API.Address != ":65432" {
		t.Errorf("期望地址 :65432, 实际 %s", result2.Server.API.Address)
	}

	data, err = os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}
	if !strings.Contains(string(data), ":65432") {
		t.Error("配置文件中未找到更新后的地址")
	}
}

func cfgMarshal(cfg *Config) ([]byte, error) {
	return yaml.Marshal(cfg)
}

func TestMCPConfig(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		wantErr  bool
		validate func(*testing.T, *Config)
	}{
		{
			name: "MCP配置禁用",
			yaml: `
server:
  api:
    address: ":20080"
mcp:
  enabled: false
  api_key: ""
  server_info:
    name: "tlcpchan-mcp"
    version: "1.0.0"
`,
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.MCP.Enabled {
					t.Error("期望 MCP 为禁用状态")
				}
				if cfg.MCP.ServerInfo.Name != "tlcpchan-mcp" {
					t.Errorf("期望服务器名称 'tlcpchan-mcp', 实际 '%s'", cfg.MCP.ServerInfo.Name)
				}
			},
		},
		{
			name: "MCP配置启用带API Key",
			yaml: `
server:
  api:
    address: ":20080"
mcp:
  enabled: true
  api_key: "test-secret-key-12345"
  server_info:
    name: "my-mcp-server"
    version: "2.0.0"
`,
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				if !cfg.MCP.Enabled {
					t.Error("期望 MCP 为启用状态")
				}
				if cfg.MCP.APIKey != "test-secret-key-12345" {
					t.Errorf("期望 API Key 'test-secret-key-12345', 实际 '%s'", cfg.MCP.APIKey)
				}
				if cfg.MCP.ServerInfo.Name != "my-mcp-server" {
					t.Errorf("期望服务器名称 'my-mcp-server', 实际 '%s'", cfg.MCP.ServerInfo.Name)
				}
				if cfg.MCP.ServerInfo.Version != "2.0.0" {
					t.Errorf("期望版本 '2.0.0', 实际 '%s'", cfg.MCP.ServerInfo.Version)
				}
			},
		},
		{
			name: "MCP配置启用开放访问",
			yaml: `
server:
  api:
    address: ":20080"
mcp:
  enabled: true
  api_key: ""
`,
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				if !cfg.MCP.Enabled {
					t.Error("期望 MCP 为启用状态")
				}
				if cfg.MCP.APIKey != "" {
					t.Error("期望 API Key 为空（开放访问）")
				}
			},
		},
		{
			name: "完整配置包含MCP",
			yaml: `
server:
  api:
    address: ":20080"
keystores:
  - name: "keystore1"
    type: "pkcs12"
    params:
      path: "/path/to/keystore.p12"
instances:
  - name: "proxy-1"
    type: "server"
    listen: ":8443"
    target: "backend:8080"
    protocol: "auto"
    enabled: true
mcp:
  enabled: true
  api_key: "secret-key"
  server_info:
    name: "tlcpchan-mcp"
    version: "1.0.0"
`,
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				if len(cfg.Instances) != 1 {
					t.Errorf("期望 1 个实例, 实际 %d", len(cfg.Instances))
				}
				if len(cfg.KeyStores) != 1 {
					t.Errorf("期望 1 个 keystore, 实际 %d", len(cfg.KeyStores))
				}
				if !cfg.MCP.Enabled {
					t.Error("期望 MCP 为启用状态")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default()
			if err := yaml.Unmarshal([]byte(tt.yaml), cfg); err != nil {
				if tt.wantErr {
					return
				}
				t.Fatalf("解析 YAML 失败: %v", err)
			}

			if err := Validate(cfg); err != nil {
				if tt.wantErr {
					return
				}
				t.Fatalf("验证配置失败: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestMCPConfigLoadAndSave(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	yamlContent := `
server:
  api:
    address: ":20080"
mcp:
  enabled: true
  api_key: "test-api-key-123"
  server_info:
    name: "test-server"
    version: "1.5.0"
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("写入配置文件失败: %v", err)
	}

	resetSingleton()
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	if !cfg.MCP.Enabled {
		t.Error("期望 MCP 为启用状态")
	}
	if cfg.MCP.APIKey != "test-api-key-123" {
		t.Errorf("期望 API Key 'test-api-key-123', 实际 '%s'", cfg.MCP.APIKey)
	}
	if cfg.MCP.ServerInfo.Name != "test-server" {
		t.Errorf("期望服务器名称 'test-server', 实际 '%s'", cfg.MCP.ServerInfo.Name)
	}

	Init(cfg, configPath)

	updatedCfg := Get()
	updatedCfg.MCP.APIKey = "new-api-key-456"

	if err := SaveAndUpdate(updatedCfg); err != nil {
		t.Fatalf("保存配置失败: %v", err)
	}

	reloadedCfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("重新加载配置失败: %v", err)
	}

	if reloadedCfg.MCP.APIKey != "new-api-key-456" {
		t.Errorf("期望保存的 API Key 'new-api-key-456', 实际 '%s'", reloadedCfg.MCP.APIKey)
	}
}
