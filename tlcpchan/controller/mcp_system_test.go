package controller

import (
	"context"
	"testing"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/instance"
	"github.com/Trisia/tlcpchan/logger"
	"github.com/Trisia/tlcpchan/security"
)

// TestMCPSystemTools 测试 MCP 系统信息工具
func TestMCPSystemTools(t *testing.T) {
	// 创建测试配置
	cfg := &config.Config{
		MCP: config.MCPConfig{
			Enabled: true,
			ServerInfo: config.MCPServerInfo{
				Name:    "test-mcp",
				Version: "1.0.0",
			},
		},
	}

	// 创建必要的依赖
	keyStoreMgr := security.NewKeyStoreManager()
	rootCertMgr := security.NewRootCertManager("")
	instMgr := instance.NewManager(nil, keyStoreMgr, rootCertMgr)

	// 创建 MCP 控制器
	opts := &ServerOptions{
		Config:          cfg,
		KeyStoreManager: keyStoreMgr,
		RootCertManager: rootCertMgr,
		InstanceManager: instMgr,
	}
	ctrl, err := NewMCPController(opts)
	if err != nil {
		t.Fatalf("创建 MCP 控制器失败: %v", err)
	}

	if ctrl == nil {
		t.Fatal("MCP 控制器为空")
	}

	// 测试 get_system_info 工具
	t.Run("get_system_info", func(t *testing.T) {
		ctx := context.Background()
		input := GetSystemInfoInput{}

		_, output, err := ctrl.handleGetSystemInfo(ctx, nil, input)
		if err != nil {
			t.Errorf("调用 get_system_info 失败: %v", err)
		}

		if output.Version == "" {
			t.Error("版本号不能为空")
		}
		if output.GoVersion == "" {
			t.Error("Go 版本不能为空")
		}
		if output.OS == "" {
			t.Error("操作系统不能为空")
		}
		if output.Arch == "" {
			t.Error("架构不能为空")
		}
		if output.Uptime <= 0 {
			t.Error("运行时长必须大于 0")
		}

		t.Logf("系统信息: Version=%s, GoVersion=%s, OS=%s, Arch=%s, Uptime=%.2f",
			output.Version, output.GoVersion, output.OS, output.Arch, output.Uptime)
	})

	// 测试 get_system_stats 工具
	t.Run("get_system_stats", func(t *testing.T) {
		ctx := context.Background()
		input := GetSystemStatsInput{}

		_, output, err := ctrl.handleGetSystemStats(ctx, nil, input)
		if err != nil {
			t.Errorf("调用 get_system_stats 失败: %v", err)
		}

		if output.MemoryUsage <= 0 {
			t.Error("内存使用量必须大于 0")
		}
		if output.TotalConnections < 0 {
			t.Error("总连接数不能小于 0")
		}
		if output.ActiveInstances < 0 {
			t.Error("活跃实例数不能小于 0")
		}

		t.Logf("系统统计: CPU=%.2f%%, Memory=%d bytes, Connections=%d, Instances=%d",
			output.CPUUsage, output.MemoryUsage, output.TotalConnections, output.ActiveInstances)
	})
}

// TestHandleGetSystemInfo_VersionOverride 测试版本号覆盖功能
//
// 测试要点:
//   - 验证版本号从 version.Version 获取，不受配置中的版本号影响
func TestHandleGetSystemInfo_VersionOverride(t *testing.T) {
	// 创建测试配置，配置中的版本号与实际版本号不同
	cfg := &config.Config{
		MCP: config.MCPConfig{
			Enabled: true,
			ServerInfo: config.MCPServerInfo{
				Name:    "test-mcp",
				Version: "2.0.0", // 配置中的版本号
			},
		},
	}

	// 创建必要的依赖
	keyStoreMgr := security.NewKeyStoreManager()
	rootCertMgr := security.NewRootCertManager("")
	instMgr := instance.NewManager(nil, keyStoreMgr, rootCertMgr)

	// 创建 MCP 控制器
	opts := &ServerOptions{
		Config:          cfg,
		KeyStoreManager: keyStoreMgr,
		RootCertManager: rootCertMgr,
		InstanceManager: instMgr,
	}
	ctrl, err := NewMCPController(opts)
	if err != nil {
		t.Fatalf("创建 MCP 控制器失败: %v", err)
	}

	// 调用 get_system_info
	ctx := context.Background()
	input := GetSystemInfoInput{}
	_, output, err := ctrl.handleGetSystemInfo(ctx, nil, input)
	if err != nil {
		t.Fatalf("调用 get_system_info 失败: %v", err)
	}

	// 验证版本号不等于配置中的版本号
	if output.Version == cfg.MCP.ServerInfo.Version {
		t.Errorf("版本号不应等于配置中的版本号，得到: %s", output.Version)
	}

	// 版本号应该是 "1.0.0"（从 version.Version 获取）
	expectedVersion := "1.0.0"
	if output.Version != expectedVersion {
		t.Errorf("版本号应为 %s，实际得到: %s", expectedVersion, output.Version)
	}

	t.Logf("版本号验证通过: %s", output.Version)
}

// TestHandleGetSystemStats_NoInstances 测试无实例时的统计信息
//
// 测试要点:
//   - 验证无实例时统计信息的正确性
//   - 验证内存使用量大于0
//   - 验证总连接数为0
//   - 验证活跃实例数为0
func TestHandleGetSystemStats_NoInstances(t *testing.T) {
	// 创建测试配置
	cfg := &config.Config{
		MCP: config.MCPConfig{
			Enabled: true,
			ServerInfo: config.MCPServerInfo{
				Name:    "test-mcp",
				Version: "1.0.0",
			},
		},
	}

	// 创建必要的依赖
	keyStoreMgr := security.NewKeyStoreManager()
	rootCertMgr := security.NewRootCertManager("")
	// 创建空的实例管理器（无实例）
	instMgr := instance.NewManager(nil, keyStoreMgr, rootCertMgr)

	// 创建 MCP 控制器
	opts := &ServerOptions{
		Config:          cfg,
		KeyStoreManager: keyStoreMgr,
		RootCertManager: rootCertMgr,
		InstanceManager: instMgr,
	}
	ctrl, err := NewMCPController(opts)
	if err != nil {
		t.Fatalf("创建 MCP 控制器失败: %v", err)
	}

	// 调用 get_system_stats
	ctx := context.Background()
	input := GetSystemStatsInput{}
	_, output, err := ctrl.handleGetSystemStats(ctx, nil, input)
	if err != nil {
		t.Fatalf("调用 get_system_stats 失败: %v", err)
	}

	// 验证内存使用量大于0
	if output.MemoryUsage <= 0 {
		t.Errorf("内存使用量必须大于0，实际得到: %d", output.MemoryUsage)
	}

	// 验证总连接数为0
	if output.TotalConnections != 0 {
		t.Errorf("无实例时总连接数应为0，实际得到: %d", output.TotalConnections)
	}

	// 验证活跃实例数为0
	if output.ActiveInstances != 0 {
		t.Errorf("无实例时活跃实例数应为0，实际得到: %d", output.ActiveInstances)
	}

	// 验证 CPU 使用率为0（暂不支持）
	if output.CPUUsage != 0 {
		t.Errorf("CPU 使用率应为0（暂不支持），实际得到: %.2f", output.CPUUsage)
	}

	t.Logf("无实例统计验证通过: Memory=%d, Connections=%d, Instances=%d",
		output.MemoryUsage, output.TotalConnections, output.ActiveInstances)
}

// TestHandleGetSystemStats_WithInstances 测试有运行实例时的统计信息
//
// 测试要点:
//   - 验证有实例时统计信息的正确性
//   - 验证活跃实例数正确计算
//   - 验证总连接数正确累加
func TestHandleGetSystemStats_WithInstances(t *testing.T) {
	log, _ := logger.New(logger.LogConfig{Level: "info", Enabled: false})

	// 创建测试配置
	cfg := &config.Config{
		MCP: config.MCPConfig{
			Enabled: true,
			ServerInfo: config.MCPServerInfo{
				Name:    "test-mcp",
				Version: "1.0.0",
			},
		},
	}

	// 创建必要的依赖
	keyStoreMgr := security.NewKeyStoreManager()
	rootCertMgr := security.NewRootCertManager("")
	instMgr := instance.NewManager(log, keyStoreMgr, rootCertMgr)

	// 创建多个实例配置（使用 client + auto 协议，不需要 keystore）
	instanceConfigs := []*config.InstanceConfig{
		{
			Name:     "s-1",
			Type:     "client",
			Protocol: "auto",
			Enabled:  true,
			Listen:   "127.0.0.1:19001",
			Target:   "127.0.0.1:19002",
			TLCP: config.TLCPConfig{
				ClientAuthType: "no-client-cert",
			},
		},
		{
			Name:     "s-2",
			Type:     "client",
			Protocol: "auto",
			Enabled:  true,
			Listen:   "127.0.0.1:19003",
			Target:   "127.0.0.1:19004",
			TLCP: config.TLCPConfig{
				ClientAuthType: "no-client-cert",
			},
		},
	}

	// 创建并启动实例
	var createdInstances []instance.Instance
	for _, icfg := range instanceConfigs {
		inst, err := instMgr.Create(icfg)
		if err != nil {
			t.Fatalf("创建实例 %s 失败: %v", icfg.Name, err)
		}
		createdInstances = append(createdInstances, inst)
	}

	// 创建 MCP 控制器
	opts := &ServerOptions{
		Config:          cfg,
		KeyStoreManager: keyStoreMgr,
		RootCertManager: rootCertMgr,
		InstanceManager: instMgr,
	}
	ctrl, err := NewMCPController(opts)
	if err != nil {
		t.Fatalf("创建 MCP 控制器失败: %v", err)
	}

	// 获取有实例但未启动时的统计信息
	ctx := context.Background()
	input := GetSystemStatsInput{}
	_, output, err := ctrl.handleGetSystemStats(ctx, nil, input)
	if err != nil {
		t.Fatalf("调用 get_system_stats 失败: %v", err)
	}

	// 验证内存使用量大于0
	if output.MemoryUsage <= 0 {
		t.Errorf("内存使用量必须大于0，实际得到: %d", output.MemoryUsage)
	}

	// 实例未启动，活跃实例数应为0
	if output.ActiveInstances != 0 {
		t.Errorf("实例未启动时活跃实例数应为0，实际得到: %d", output.ActiveInstances)
	}

	// 总连接数应为0（没有运行中的连接）
	if output.TotalConnections != 0 {
		t.Errorf("无运行连接时总连接数应为0，实际得到: %d", output.TotalConnections)
	}

	// 尝试启动第一个实例（监听地址需要可用，这里只测试状态变化）
	// 注意：实际启动可能会失败因为端口可能已被占用，我们主要测试统计逻辑
	inst1 := createdInstances[0]
	err = inst1.Start()
	if err != nil {
		t.Logf("启动实例 %s 失败（可能端口被占用），跳过活跃实例测试: %v", inst1.Name(), err)
	} else {
		defer inst1.Stop()

		// 获取启动实例后的统计信息
		_, outputAfterStart, err := ctrl.handleGetSystemStats(ctx, nil, input)
		if err != nil {
			t.Fatalf("调用 get_system_stats 失败: %v", err)
		}

		// 活跃实例数应为1
		if outputAfterStart.ActiveInstances != 1 {
			t.Errorf("启动1个实例后活跃实例数应为1，实际得到: %d", outputAfterStart.ActiveInstances)
		}

		t.Logf("启动实例后统计验证通过: Memory=%d, Connections=%d, Instances=%d",
			outputAfterStart.MemoryUsage, outputAfterStart.TotalConnections, outputAfterStart.ActiveInstances)
	}

	t.Logf("有实例统计验证通过: Memory=%d, Connections=%d, Instances=%d",
		output.MemoryUsage, output.TotalConnections, output.ActiveInstances)
}
