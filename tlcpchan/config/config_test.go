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
