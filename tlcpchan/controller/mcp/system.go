package mcp

import (
	"context"
	"runtime"
	"time"

	"github.com/Trisia/tlcpchan/instance"
	"github.com/Trisia/tlcpchan/version"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var startTime = time.Now()

// GetSystemInfoInput 获取系统信息输入（无参数）
type GetSystemInfoInput struct{}

// GetSystemInfoOutput 获取系统信息输出
type GetSystemInfoOutput struct {
	// Version 版本号
	Version string `json:"version"`
	// GoVersion Go运行时版本
	GoVersion string `json:"goVersion"`
	// OS 操作系统
	OS string `json:"os"`
	// Arch 系统架构
	Arch string `json:"arch"`
	// Uptime 运行时长，单位：秒
	Uptime float64 `json:"uptime"`
}

// GetSystemStatsInput 获取系统统计输入（无参数）
type GetSystemStatsInput struct{}

// GetSystemStatsOutput 获取系统统计输出
type GetSystemStatsOutput struct {
	// CPUUsage CPU使用率，百分比（当前不支持，返回0）
	CPUUsage float64 `json:"cpuUsage"`
	// MemoryUsage 内存使用量，单位：字节
	MemoryUsage int64 `json:"memoryUsage"`
	// TotalConnections 总连接数（所有实例的累计连接数之和）
	TotalConnections int64 `json:"totalConnections"`
	// ActiveInstances 活跃实例数（运行中的实例数量）
	ActiveInstances int `json:"activeInstances"`
}

// handleGetSystemInfo 处理获取系统信息请求
//
// 参数:
//   - ctx: 上下文
//   - req: MCP 工具调用请求
//   - input: 获取系统信息输入参数
//
// 返回:
//   - *mcpsdk.CallToolResult: MCP 工具调用结果（可以为 nil，SDK 自动处理）
//   - GetSystemInfoOutput: 获取系统信息输出参数
//   - error: 处理失败时返回错误
//
// 注意:
//   - 版本号从 version.Version 获取
//   - Go 版本从 runtime.Version() 获取
//   - 操作系统和架构从 runtime 包获取
//   - 运行时长从全局启动时间计算
func (c *MCPController) handleGetSystemInfo(_ context.Context, _ *mcpsdk.CallToolRequest, input GetSystemInfoInput) (
	*mcpsdk.CallToolResult,
	GetSystemInfoOutput,
	error,
) {
	info := GetSystemInfoOutput{
		Version:   version.Version,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Uptime:    time.Since(startTime).Seconds(),
	}

	return nil, info, nil
}

// handleGetSystemStats 处理获取系统统计信息请求
//
// 参数:
//   - ctx: 上下文
//   - req: MCP 工具调用请求
//   - input: 获取系统统计输入参数
//
// 返回:
//   - *mcpsdk.CallToolResult: MCP 工具调用结果（可以为 nil，SDK 自动处理）
//   - GetSystemStatsOutput: 获取系统统计输出参数
//   - error: 处理失败时返回错误
//
// 注意:
//   - CPU 使用率暂不支持，返回 0
//   - 内存使用量从 runtime 获取已分配内存
//   - 总连接数为所有实例的累计连接数之和
//   - 活跃实例数为运行中的实例数量
func (c *MCPController) handleGetSystemStats(_ context.Context, _ *mcpsdk.CallToolRequest, input GetSystemStatsInput) (
	*mcpsdk.CallToolResult,
	GetSystemStatsOutput,
	error,
) {
	// 获取内存使用量
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memoryUsage := int64(m.Alloc)

	// 统计总连接数和活跃实例数
	var totalConnections int64
	var activeInstances int

	if c.instanceMgr != nil {
		instances := c.instanceMgr.List()
		for _, inst := range instances {
			// 累计连接数
			stats := inst.Stats()
			if stats != nil {
				totalConnections += stats.TotalConnections
			}

			// 统计活跃实例数
			if inst.Status() == instance.StatusRunning {
				activeInstances++
			}
		}
	}

	stats := GetSystemStatsOutput{
		CPUUsage:         0,
		MemoryUsage:      memoryUsage,
		TotalConnections: totalConnections,
		ActiveInstances:  activeInstances,
	}

	return nil, stats, nil
}

// registerSystemTools 注册系统信息工具到 MCP 服务器
//
// 注意:
//   - 注册 get_system_info 工具，获取系统基本信息
//   - 注册 get_system_stats 工具，获取系统统计信息
func (c *MCPController) registerSystemTools() {
	mcpsdk.AddTool(c.server, &mcpsdk.Tool{
		Name:        "get_system_info",
		Description: "获取系统信息（版本、Go版本、操作系统、架构、运行时长）",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"version": map[string]any{
					"description": "版本号",
					"type":        "string",
				},
				"goVersion": map[string]any{
					"description": "Go运行时版本",
					"type":        "string",
				},
				"os": map[string]any{
					"description": "操作系统",
					"type":        "string",
				},
				"arch": map[string]any{
					"description": "系统架构",
					"type":        "string",
				},
				"uptime": map[string]any{
					"description": "运行时长，单位：秒",
					"type":        "number",
				},
			},
		},
	}, c.handleGetSystemInfo)

	mcpsdk.AddTool(c.server, &mcpsdk.Tool{
		Name:        "get_system_stats",
		Description: "获取系统统计信息（CPU使用率、内存使用、总连接数、活跃实例数）",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"cpuUsage": map[string]any{
					"description": "CPU使用率，百分比（暂不支持，返回0）",
					"type":        "number",
				},
				"memoryUsage": map[string]any{
					"description": "内存使用量，单位：字节",
					"type":        "integer",
				},
				"totalConnections": map[string]any{
					"description": "总连接数（所有实例的累计连接数之和）",
					"type":        "integer",
				},
				"activeInstances": map[string]any{
					"description": "活跃实例数（运行中的实例数量）",
					"type":        "integer",
				},
			},
		},
	}, c.handleGetSystemStats)
}
