package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
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
}

func testInitAndGet(t *testing.T) {
	cfg := Default()
	cfg.Server.API.Address = ":12345"

	Init(cfg)

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
	Init(cfg1)

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
	Init(cfg)

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

	if err := Save(configPath, cfg1); err != nil {
		t.Fatalf("保存配置失败: %v", err)
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
	if err := SaveAndUpdate(configPath, cfg2); err != nil {
		t.Fatalf("SaveAndUpdate 失败: %v", err)
	}

	result2 := Get()
	if result2.Server.API.Address != ":65432" {
		t.Errorf("期望地址 :65432, 实际 %s", result2.Server.API.Address)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}
	if !strings.Contains(string(data), ":65432") {
		t.Error("配置文件中未找到更新后的地址")
	}
}

func TestParseCipherSuite(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		isTLCP  bool
		want    uint16
		wantErr bool
	}{
		{
			name:    "TLCP ECC_SM4_CBC_SM3",
			input:   "ECC_SM4_CBC_SM3",
			isTLCP:  true,
			want:    0xE013,
			wantErr: false,
		},
		{
			name:    "TLCP ECC_SM4_GCM_SM3",
			input:   "ECC_SM4_GCM_SM3",
			isTLCP:  true,
			want:    0xE053,
			wantErr: false,
		},
		{
			name:    "TLCP ECDHE_SM4_CBC_SM3",
			input:   "ECDHE_SM4_CBC_SM3",
			isTLCP:  true,
			want:    0xE011,
			wantErr: false,
		},
		{
			name:    "TLCP ECDHE_SM4_GCM_SM3",
			input:   "ECDHE_SM4_GCM_SM3",
			isTLCP:  true,
			want:    0xE051,
			wantErr: false,
		},
		{
			name:    "TLCP hex value",
			input:   "0xE013",
			isTLCP:  true,
			want:    0xE013,
			wantErr: false,
		},
		{
			name:    "TLCP decimal value",
			input:   "57427",
			isTLCP:  true,
			want:    0xE053,
			wantErr: false,
		},
		{
			name:    "TLS AES_128_GCM_SHA256",
			input:   "TLS_AES_128_GCM_SHA256",
			isTLCP:  false,
			want:    0x1301,
			wantErr: false,
		},
		{
			name:    "unknown cipher suite",
			input:   "UNKNOWN_SUITE",
			isTLCP:  true,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			isTLCP:  true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCipherSuite(tt.input, tt.isTLCP)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseCipherSuite(%q, %v) 期望错误，但未返回错误", tt.input, tt.isTLCP)
				}
			} else {
				if err != nil {
					t.Errorf("ParseCipherSuite(%q, %v) 意外错误: %v", tt.input, tt.isTLCP, err)
				}
				if got != tt.want {
					t.Errorf("ParseCipherSuite(%q, %v) = 0x%04X, 期望 0x%04X", tt.input, tt.isTLCP, got, tt.want)
				}
			}
		})
	}
}

func TestParseCipherSuites(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		isTLCP  bool
		want    []uint16
		wantErr bool
	}{
		{
			name:    "empty slice",
			input:   []string{},
			isTLCP:  true,
			want:    nil,
			wantErr: false,
		},
		{
			name:    "TLCP multiple suites",
			input:   []string{"ECC_SM4_GCM_SM3", "ECDHE_SM4_GCM_SM3"},
			isTLCP:  true,
			want:    []uint16{0xE053, 0xE051},
			wantErr: false,
		},
		{
			name:    "TLCP with hex values",
			input:   []string{"0xE013", "0xE051"},
			isTLCP:  true,
			want:    []uint16{0xE013, 0xE051},
			wantErr: false,
		},
		{
			name:    "TLS multiple suites",
			input:   []string{"TLS_AES_128_GCM_SHA256", "TLS_AES_256_GCM_SHA384"},
			isTLCP:  false,
			want:    []uint16{0x1301, 0x1302},
			wantErr: false,
		},
		{
			name:    "contains unknown suite",
			input:   []string{"ECC_SM4_GCM_SM3", "UNKNOWN_SUITE"},
			isTLCP:  true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCipherSuites(tt.input, tt.isTLCP)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseCipherSuites(%v, %v) 期望错误，但未返回错误", tt.input, tt.isTLCP)
				}
			} else {
				if err != nil {
					t.Errorf("ParseCipherSuites(%v, %v) 意外错误: %v", tt.input, tt.isTLCP, err)
				}
				if len(got) != len(tt.want) {
					t.Errorf("ParseCipherSuites(%v, %v) 长度不匹配: 得到 %d, 期望 %d", tt.input, tt.isTLCP, len(got), len(tt.want))
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("ParseCipherSuites(%v, %v)[%d] = 0x%04X, 期望 0x%04X", tt.input, tt.isTLCP, i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}
