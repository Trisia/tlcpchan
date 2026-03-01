package mcp

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/instance"
	"github.com/Trisia/tlcpchan/proxy"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListInstancesInput 列出实例输入（无参数）
type ListInstancesInput struct{}

// ListInstancesOutput 列出实例输出
type ListInstancesOutput struct {
	// Instances 实例列表
	Instances []InstanceInfo `json:"instances"`
}

// InstanceInfo 实例信息
type InstanceInfo struct {
	// Name 实例名称
	Name string `json:"name"`
	// Status 实例状态
	Status string `json:"status"`
	// Config 实例配置
	Config *config.InstanceConfig `json:"config"`
	// Enabled 是否启用
	Enabled bool `json:"enabled"`
}

// GetInstanceInput 获取实例输入
type GetInstanceInput struct {
	// Name 实例名称
	Name string `json:"name"`
}

// GetInstanceOutput 获取实例输出
type GetInstanceOutput struct {
	// Name 实例名称
	Name string `json:"name"`
	// Status 实例状态
	Status string `json:"status"`
	// Config 实例配置
	Config *config.InstanceConfig `json:"config"`
}

// CreateInstanceInput 创建实例输入
type CreateInstanceInput struct {
	// Config 实例配置
	Config config.InstanceConfig `json:"config"`
}

// CreateInstanceOutput 创建实例输出
type CreateInstanceOutput struct {
	// Name 实例名称
	Name string `json:"name"`
	// Status 实例状态
	Status string `json:"status"`
}

// UpdateInstanceInput 更新实例输入
type UpdateInstanceInput struct {
	// Name 实例名称
	Name string `json:"name"`
	// Config 实例配置
	Config config.InstanceConfig `json:"config"`
}

// UpdateInstanceOutput 更新实例输出
type UpdateInstanceOutput struct {
	// Name 实例名称
	Name string `json:"name"`
	// Status 实例状态
	Status string `json:"status"`
	// Config 实例配置
	Config config.InstanceConfig `json:"config"`
}

// DeleteInstanceInput 删除实例输入
type DeleteInstanceInput struct {
	// Name 实例名称
	Name string `json:"name"`
}

// DeleteInstanceOutput 删除实例输出
type DeleteInstanceOutput struct {
	// Success 是否成功
	Success bool `json:"success"`
}

// StartInstanceInput 启动实例输入
type StartInstanceInput struct {
	// Name 实例名称
	Name string `json:"name"`
}

// InstanceStatusOutput 实例状态输出
type InstanceStatusOutput struct {
	// Status 实例状态
	Status instance.Status `json:"status"`
}

// StopInstanceInput 停止实例输入
type StopInstanceInput struct {
	// Name 实例名称
	Name string `json:"name"`
}

// RestartInstanceInput 重启实例输入
type RestartInstanceInput struct {
	// Name 实例名称
	Name string `json:"name"`
}

// GetInstanceStatsInput 获取实例统计输入
type GetInstanceStatsInput struct {
	// Name 实例名称
	Name string `json:"name"`
}

// GetInstanceStatsOutput 获取实例统计输出
type GetInstanceStatsOutput struct {
	// Stats 统计信息
	Stats interface{} `json:"stats"`
}

// CheckInstanceHealthInput 检查实例健康输入
type CheckInstanceHealthInput struct {
	// Name 实例名称
	Name string `json:"name"`
	// Timeout 超时时间（秒）
	Timeout int `json:"timeout,omitempty"`
}

// CheckInstanceHealthOutput 检查实例健康输出
type CheckInstanceHealthOutput struct {
	// Instance 实例名称
	Instance string `json:"instance"`
	// Results 健康检查结果
	Results []*proxy.HealthCheckResult `json:"results"`
}

// handleListInstances 处理 list_instances 工具调用
//
// 参数:
//   - ctx: 上下文
//   - req: MCP 工具调用请求
//   - input: 列出实例输入参数（无参数）
//
// 返回:
//   - *mcpsdk.CallToolResult: MCP 工具调用结果（可以为 nil，SDK 自动处理）
//   - ListInstancesOutput: 实例列表
//   - error: 操作失败时返回错误
func (c *MCPController) handleListInstances(_ context.Context, _ *mcpsdk.CallToolRequest, input ListInstancesInput) (
	*mcpsdk.CallToolResult,
	ListInstancesOutput,
	error,
) {
	instances := c.instanceMgr.List()
	data := make([]InstanceInfo, len(instances))
	for i, inst := range instances {
		cfg := inst.Config()
		data[i] = InstanceInfo{
			Name:    inst.Name(),
			Status:  string(inst.Status()),
			Config:  cfg,
			Enabled: cfg.Enabled,
		}
	}
	return nil, ListInstancesOutput{Instances: data}, nil
}

// handleGetInstance 处理 get_instance 工具调用
//
// 参数:
//   - ctx: 上下文
//   - req: MCP 工具调用请求
//   - input: 获取实例输入参数
//
// 返回:
//   - *mcpsdk.CallToolResult: MCP 工具调用结果（可以为 nil，SDK 自动处理）
//   - GetInstanceOutput: 实例详细信息
//   - error: 实例不存在时返回错误
func (c *MCPController) handleGetInstance(_ context.Context, _ *mcpsdk.CallToolRequest, input GetInstanceInput) (
	*mcpsdk.CallToolResult,
	GetInstanceOutput,
	error,
) {
	inst, ok := c.instanceMgr.Get(input.Name)
	if !ok {
		return nil, GetInstanceOutput{}, fmt.Errorf("实例不存在: %s", input.Name)
	}

	return nil, GetInstanceOutput{
		Name:   inst.Name(),
		Status: string(inst.Status()),
		Config: inst.Config(),
	}, nil
}

// handleCreateInstance 处理 create_instance 工具调用
//
// 参数:
//   - ctx: 上下文
//   - req: MCP 工具调用请求
//   - input: 创建实例输入参数
//
// 返回:
//   - *mcpsdk.CallToolResult: MCP 工具调用结果（可以为 nil，SDK 自动处理）
//   - CreateInstanceOutput: 创建的实例信息
//   - error: 创建失败时返回错误
func (c *MCPController) handleCreateInstance(_ context.Context, _ *mcpsdk.CallToolRequest, input CreateInstanceInput) (
	*mcpsdk.CallToolResult,
	CreateInstanceOutput,
	error,
) {
	cfg := input.Config
	if cfg.Name == "" {
		return nil, CreateInstanceOutput{}, fmt.Errorf("实例名称不能为空")
	}

	// 检查实例是否已存在
	currentCfg := config.Get()
	for _, inst := range currentCfg.Instances {
		if inst.Name == cfg.Name {
			return nil, CreateInstanceOutput{}, fmt.Errorf("实例已存在: %s", cfg.Name)
		}
	}

	// 检查端口冲突
	if err := checkPortConflict(cfg.Listen, cfg.Enabled, "", currentCfg.Instances); err != nil {
		return nil, CreateInstanceOutput{}, err
	}

	// 验证 file 类型 keystore 的文件是否存在
	if err := validateInstanceFileKeystores(config.Get().WorkDir, &cfg); err != nil {
		return nil, CreateInstanceOutput{}, err
	}

	// 创建实例
	inst, err := c.instanceMgr.Create(&cfg)
	if err != nil {
		return nil, CreateInstanceOutput{}, fmt.Errorf("创建实例失败: %w", err)
	}

	// 保存配置
	currentCfg.Instances = append(currentCfg.Instances, cfg)
	if err := config.Save(currentCfg); err != nil {
		return nil, CreateInstanceOutput{}, fmt.Errorf("保存配置失败: %w", err)
	}
	config.Set(currentCfg)

	c.log.Info("创建实例: %s 协议: %s", cfg.Name, cfg.Protocol)
	return nil, CreateInstanceOutput{
		Name:   inst.Name(),
		Status: string(inst.Status()),
	}, nil
}

// handleUpdateInstance 处理 update_instance 工具调用
//
// 参数:
//   - ctx: 上下文
//   - req: MCP 工具调用请求
//   - input: 更新实例输入参数
//
// 返回:
//   - *mcpsdk.CallToolResult: MCP 工具调用结果（可以为 nil，SDK 自动处理）
//   - UpdateInstanceOutput: 更新后的实例信息
//   - error: 更新失败时返回错误
func (c *MCPController) handleUpdateInstance(_ context.Context, _ *mcpsdk.CallToolRequest, input UpdateInstanceInput) (
	*mcpsdk.CallToolResult,
	UpdateInstanceOutput,
	error,
) {
	// 获取实例
	inst, ok := c.instanceMgr.Get(input.Name)
	if !ok {
		return nil, UpdateInstanceOutput{}, fmt.Errorf("实例不存在: %s", input.Name)
	}

	newCfg := input.Config
	newCfg.Name = input.Name

	currentCfg := config.Get()

	// 检查端口冲突
	if err := checkPortConflict(newCfg.Listen, newCfg.Enabled, input.Name, currentCfg.Instances); err != nil {
		return nil, UpdateInstanceOutput{}, err
	}

	// 验证 file 类型 keystore 的文件是否存在
	if err := validateInstanceFileKeystores(config.Get().WorkDir, &newCfg); err != nil {
		return nil, UpdateInstanceOutput{}, err
	}

	// 更新配置
	found := false
	for i, instance := range currentCfg.Instances {
		if instance.Name == input.Name {
			currentCfg.Instances[i] = newCfg
			found = true
			break
		}
	}
	if !found {
		return nil, UpdateInstanceOutput{}, fmt.Errorf("实例不存在: %s", input.Name)
	}

	if err := config.Save(currentCfg); err != nil {
		return nil, UpdateInstanceOutput{}, fmt.Errorf("保存配置失败: %w", err)
	}
	config.Set(currentCfg)

	c.log.Info("实例配置已保存: %s", input.Name)

	// 如果实例运行中，热重载
	if inst.Status() == instance.StatusRunning {
		if err := inst.Reload(&newCfg); err != nil {
			return nil, UpdateInstanceOutput{}, fmt.Errorf("热重载失败: %w", err)
		}
		c.log.Info("实例已热重载: %s", input.Name)
	}

	return nil, UpdateInstanceOutput{
		Name:   input.Name,
		Status: string(inst.Status()),
		Config: newCfg,
	}, nil
}

// handleDeleteInstance 处理 delete_instance 工具调用
//
// 参数:
//   - ctx: 上下文
//   - req: MCP 工具调用请求
//   - input: 删除实例输入参数
//
// 返回:
//   - *mcpsdk.CallToolResult: MCP 工具调用结果（可以为 nil，SDK 自动处理）
//   - DeleteInstanceOutput: 删除确认
//   - error: 删除失败时返回错误
func (c *MCPController) handleDeleteInstance(_ context.Context, _ *mcpsdk.CallToolRequest, input DeleteInstanceInput) (
	*mcpsdk.CallToolResult,
	DeleteInstanceOutput,
	error,
) {
	// 删除实例
	if err := c.instanceMgr.Delete(input.Name); err != nil {
		return nil, DeleteInstanceOutput{}, fmt.Errorf("删除实例失败: %w", err)
	}

	// 更新配置
	currentCfg := config.Get()
	found := false
	for i, inst := range currentCfg.Instances {
		if inst.Name == input.Name {
			currentCfg.Instances = append(currentCfg.Instances[:i], currentCfg.Instances[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return nil, DeleteInstanceOutput{}, fmt.Errorf("实例不存在: %s", input.Name)
	}

	if err := config.Save(currentCfg); err != nil {
		return nil, DeleteInstanceOutput{}, fmt.Errorf("保存配置失败: %w", err)
	}
	config.Set(currentCfg)

	c.log.Info("删除实例: %s", input.Name)
	return nil, DeleteInstanceOutput{Success: true}, nil
}

// handleStartInstance 处理 start_instance 工具调用
//
// 参数:
//   - ctx: 上下文
//   - req: MCP 工具调用请求
//   - input: 启动实例输入参数
//
// 返回:
//   - *mcpsdk.CallToolResult: MCP 工具调用结果（可以为 nil，SDK 自动处理）
//   - InstanceStatusOutput: 启动状态
//   - error: 启动失败时返回错误
func (c *MCPController) handleStartInstance(_ context.Context, _ *mcpsdk.CallToolRequest, input StartInstanceInput) (
	*mcpsdk.CallToolResult,
	InstanceStatusOutput,
	error,
) {
	inst, ok := c.instanceMgr.Get(input.Name)
	if !ok {
		return nil, InstanceStatusOutput{}, fmt.Errorf("实例不存在: %s", input.Name)
	}

	if err := inst.Start(); err != nil {
		return nil, InstanceStatusOutput{}, fmt.Errorf("启动实例失败: %w", err)
	}

	c.log.Info("启动实例: %s", input.Name)
	return nil, InstanceStatusOutput{Status: inst.Status()}, nil
}

// handleStopInstance 处理 stop_instance 工具调用
//
// 参数:
//   - ctx: 上下文
//   - req: MCP 工具调用请求
//   - input: 停止实例输入参数
//
// 返回:
//   - *mcpsdk.CallToolResult: MCP 工具调用结果（可以为 nil，SDK 自动处理）
//   - InstanceStatusOutput: 停止状态
//   - error: 停止失败时返回错误
func (c *MCPController) handleStopInstance(_ context.Context, _ *mcpsdk.CallToolRequest, input StopInstanceInput) (
	*mcpsdk.CallToolResult,
	InstanceStatusOutput,
	error,
) {
	inst, ok := c.instanceMgr.Get(input.Name)
	if !ok {
		return nil, InstanceStatusOutput{}, fmt.Errorf("实例不存在: %s", input.Name)
	}

	if err := inst.Stop(); err != nil {
		return nil, InstanceStatusOutput{}, fmt.Errorf("停止实例失败: %w", err)
	}

	c.log.Info("停止实例: %s", input.Name)
	return nil, InstanceStatusOutput{Status: inst.Status()}, nil
}

// handleRestartInstance 处理 restart_instance 工具调用
//
// 参数:
//   - ctx: 上下文
//   - req: MCP 工具调用请求
//   - input: 重启实例输入参数
//
// 返回:
//   - *mcpsdk.CallToolResult: MCP 工具调用结果（可以为 nil，SDK 自动处理）
//   - InstanceStatusOutput: 重启状态
//   - error: 重启失败时返回错误
func (c *MCPController) handleRestartInstance(_ context.Context, _ *mcpsdk.CallToolRequest, input RestartInstanceInput) (
	*mcpsdk.CallToolResult,
	InstanceStatusOutput,
	error,
) {
	inst, ok := c.instanceMgr.Get(input.Name)
	if !ok {
		return nil, InstanceStatusOutput{}, fmt.Errorf("实例不存在: %s", input.Name)
	}

	cfg := inst.Config()
	if err := inst.Restart(cfg); err != nil {
		return nil, InstanceStatusOutput{}, fmt.Errorf("重启实例失败: %w", err)
	}

	c.log.Info("重启实例: %s", input.Name)
	return nil, InstanceStatusOutput{Status: inst.Status()}, nil
}

// handleGetInstanceStats 处理 get_instance_stats 工具调用
//
// 参数:
//   - ctx: 上下文
//   - req: MCP 工具调用请求
//   - input: 获取实例统计输入参数
//
// 返回:
//   - *mcpsdk.CallToolResult: MCP 工具调用结果（可以为 nil，SDK 自动处理）
//   - GetInstanceStatsOutput: 统计信息
//   - error: 获取失败时返回错误
func (c *MCPController) handleGetInstanceStats(_ context.Context, _ *mcpsdk.CallToolRequest, input GetInstanceStatsInput) (
	*mcpsdk.CallToolResult,
	GetInstanceStatsOutput,
	error,
) {
	inst, ok := c.instanceMgr.Get(input.Name)
	if !ok {
		return nil, GetInstanceStatsOutput{}, fmt.Errorf("实例不存在: %s", input.Name)
	}

	stats := inst.Stats()
	return nil, GetInstanceStatsOutput{Stats: stats}, nil
}

// handleCheckInstanceHealth 处理 check_instance_health 工具调用
//
// 参数:
//   - ctx: 上下文
//   - req: MCP 工具调用请求
//   - input: 检查实例健康输入参数
//
// 返回:
//   - *mcpsdk.CallToolResult: MCP 工具调用结果（可以为 nil，SDK 自动处理）
//   - CheckInstanceHealthOutput: 健康检查结果
//   - error: 检查失败时返回错误
func (c *MCPController) handleCheckInstanceHealth(_ context.Context, _ *mcpsdk.CallToolRequest, input CheckInstanceHealthInput) (
	*mcpsdk.CallToolResult,
	CheckInstanceHealthOutput,
	error,
) {
	inst, ok := c.instanceMgr.Get(input.Name)
	if !ok {
		return nil, CheckInstanceHealthOutput{}, fmt.Errorf("实例不存在: %s", input.Name)
	}

	timeout := 10 * time.Second
	if input.Timeout > 0 {
		timeout = time.Duration(input.Timeout) * time.Second
	}

	instanceProtocol := inst.Protocol()

	var results []*proxy.HealthCheckResult

	// 根据协议类型处理
	if instanceProtocol == string(config.ProtocolAuto) {
		// auto 模式：检查 TLCP 和 TLS
		tlcpResult := inst.CheckHealth(proxy.ProtocolTLCP, timeout)
		tlsResult := inst.CheckHealth(proxy.ProtocolTLS, timeout)
		results = append(results, tlcpResult, tlsResult)
	} else if instanceProtocol == string(config.ProtocolTLCP) {
		// TLCP 模式
		result := inst.CheckHealth(proxy.ProtocolTLCP, timeout)
		results = append(results, result)
	} else if instanceProtocol == string(config.ProtocolTLS) {
		// TLS 模式
		result := inst.CheckHealth(proxy.ProtocolTLS, timeout)
		results = append(results, result)
	}

	return nil, CheckInstanceHealthOutput{
		Instance: input.Name,
		Results:  results,
	}, nil
}

// registerInstanceTools 注册实例管理工具到 MCP 服务器
//
// 注意:
//   - 注册 10 个实例管理工具
//   - 在 NewMCPController 中调用此函数
func (c *MCPController) registerInstanceTools() {
	// 1. list_instances - 获取所有代理实例的列表信息
	mcpsdk.AddTool(c.server, &mcpsdk.Tool{
		Name:        "list_instances",
		Description: "获取所有代理实例的列表信息",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"instances": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"name": map[string]any{
								"description": "实例名称",
								"type":        "string",
							},
							"status": map[string]any{
								"description": "实例状态",
								"type":        "string",
							},
							"config": map[string]any{
								"description": "实例配置",
								"type":        "object",
							},
							"enabled": map[string]any{
								"description": "是否启用",
								"type":        "boolean",
							},
						},
					},
				},
			},
		},
	}, c.handleListInstances)

	// 2. get_instance - 获取指定实例的详细信息
	mcpsdk.AddTool(c.server, &mcpsdk.Tool{
		Name:        "get_instance",
		Description: "获取指定实例的详细信息",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "实例名称",
				},
			},
			"required": []string{"name"},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"description": "实例名称",
					"type":        "string",
				},
				"status": map[string]any{
					"description": "实例状态",
					"type":        "string",
				},
				"config": map[string]any{
					"description": "实例配置",
					"type":        "object",
				},
			},
		},
	}, c.handleGetInstance)

	// 3. create_instance - 创建新的代理实例
	mcpsdk.AddTool(c.server, &mcpsdk.Tool{
		Name:        "create_instance",
		Description: "创建新的代理实例",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"config": map[string]any{
					"type":        "object",
					"description": "实例配置",
				},
			},
			"required": []string{"config"},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"description": "实例名称",
					"type":        "string",
				},
				"status": map[string]any{
					"description": "实例状态",
					"type":        "string",
				},
			},
		},
	}, c.handleCreateInstance)

	// 4. update_instance - 更新实例配置，支持热重载
	mcpsdk.AddTool(c.server, &mcpsdk.Tool{
		Name:        "update_instance",
		Description: "更新实例配置，支持热重载",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "实例名称",
				},
				"config": map[string]any{
					"type":        "object",
					"description": "实例配置",
				},
			},
			"required": []string{"name", "config"},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"description": "实例名称",
					"type":        "string",
				},
				"status": map[string]any{
					"description": "实例状态",
					"type":        "string",
				},
				"config": map[string]any{
					"description": "实例配置",
					"type":        "object",
				},
			},
		},
	}, c.handleUpdateInstance)

	// 5. delete_instance - 删除指定实例
	mcpsdk.AddTool(c.server, &mcpsdk.Tool{
		Name:        "delete_instance",
		Description: "删除指定实例",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "实例名称",
				},
			},
			"required": []string{"name"},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"success": map[string]any{
					"description": "是否成功",
					"type":        "boolean",
				},
			},
		},
	}, c.handleDeleteInstance)

	// 6. start_instance - 启动指定实例
	mcpsdk.AddTool(c.server, &mcpsdk.Tool{
		Name:        "start_instance",
		Description: "启动指定实例",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "实例名称",
				},
			},
			"required": []string{"name"},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"status": map[string]any{
					"description": "实例状态",
					"type":        "string",
				},
			},
		},
	}, c.handleStartInstance)

	// 7. stop_instance - 停止指定实例
	mcpsdk.AddTool(c.server, &mcpsdk.Tool{
		Name:        "stop_instance",
		Description: "停止指定实例",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "实例名称",
				},
			},
			"required": []string{"name"},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"status": map[string]any{
					"description": "实例状态",
					"type":        "string",
				},
			},
		},
	}, c.handleStopInstance)

	// 8. restart_instance - 重启指定实例
	mcpsdk.AddTool(c.server, &mcpsdk.Tool{
		Name:        "restart_instance",
		Description: "重启指定实例",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "实例名称",
				},
			},
			"required": []string{"name"},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"status": map[string]any{
					"description": "实例状态",
					"type":        "string",
				},
			},
		},
	}, c.handleRestartInstance)

	// 9. get_instance_stats - 获取实例运行统计信息
	mcpsdk.AddTool(c.server, &mcpsdk.Tool{
		Name:        "get_instance_stats",
		Description: "获取实例运行统计信息",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "实例名称",
				},
			},
			"required": []string{"name"},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"stats": map[string]any{
					"description": "统计信息",
					"type":        "object",
				},
			},
		},
	}, c.handleGetInstanceStats)

	// 10. check_instance_health - 检查实例健康状态
	mcpsdk.AddTool(c.server, &mcpsdk.Tool{
		Name:        "check_instance_health",
		Description: "检查实例健康状态",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "实例名称",
				},
				"timeout": map[string]any{
					"type":        "integer",
					"description": "超时时间（秒），默认 10 秒",
				},
			},
			"required": []string{"name"},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"instance": map[string]any{
					"description": "实例名称",
					"type":        "string",
				},
				"results": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"protocol": map[string]any{
								"description": "协议类型",
								"type":        "string",
							},
							"success": map[string]any{
								"description": "是否成功",
								"type":        "boolean",
							},
							"latencyMs": map[string]any{
								"description": "延迟（毫秒）",
								"type":        "integer",
							},
							"error": map[string]any{
								"description": "错误信息",
								"type":        "string",
							},
						},
					},
				},
			},
		},
	}, c.handleCheckInstanceHealth)

	c.log.Info("已已注册 10 个实例管理 MCP 工具")
}

// checkPortConflict 检查实例端口是否与已启用的其他实例冲突
//
// 参数:
//   - listen: 监听地址，格式为 ":port" 或 "ip:port"
//   - enabled: 实例是否启用
//   - excludeName: 需要排除的实例名称（编辑时使用），创建时传空字符串
//   - instances: 所有实例配置列表
//
// 返回:
//   - error: 端口冲突时返回错误信息，包含冲突的实例名称；无冲突返回 nil
func checkPortConflict(listen string, enabled bool, excludeName string, instances []config.InstanceConfig) error {
	if !enabled {
		return nil
	}

	targetPort, err := parseListenPort(listen)
	if err != nil {
		return fmt.Errorf("无效的监听地址: %w", err)
	}

	for _, inst := range instances {
		if inst.Name == excludeName {
			continue
		}
		if !inst.Enabled {
			continue
		}

		instPort, err := parseListenPort(inst.Listen)
		if err != nil {
			continue
		}

		if instPort == targetPort {
			return fmt.Errorf("端口 %d 已被实例 '%s' 使用", targetPort, inst.Name)
		}
	}

	return nil
}

// parseListenPort 解析监听地址，提取端口号
//
// 参数:
//   - listen: 监听地址，格式为 ":port" 或 "ip:port"，例如 ":443" 或 "127.0.0.1:8443"
//
// 返回:
//   - int: 端口号
//   - error: 解析失败时返回错误
func parseListenPort(listen string) (int, error) {
	if listen == "" {
		return 0, fmt.Errorf("监听地址不能为空")
	}

	_, port, err := net.SplitHostPort(listen)
	if err != nil {
		return 0, fmt.Errorf("解析监听地址失败: %w", err)
	}

	if strings.HasPrefix(port, ":") {
		port = port[1:]
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		return 0, fmt.Errorf("无效的端口号: %w", err)
	}

	if portNum <= 0 || portNum > 65535 {
		return 0, fmt.Errorf("端口号超出有效范围 (1-65535): %d", portNum)
	}

	return portNum, nil
}

// validateInstanceFileKeystores 验证实例配置中 file 类型 keystore 的文件是否存在
//
// 参数:
//   - workDir: 工作目录，用于解析相对路径
//   - cfg: 实例配置
//
// 返回:
//   - error: 如果文件不存在则返回错误信息，否则返回 nil
func validateInstanceFileKeystores(workDir string, cfg *config.InstanceConfig) error {
	// 验证 TLCP keystore
	if cfg.TLCP.Keystore != nil && cfg.TLCP.Keystore.Type == "file" {
		for _, filePath := range cfg.TLCP.Keystore.Params {
			if filePath == "" {
				continue
			}

			var fullPath string
			if !filepath.IsAbs(filePath) {
				fullPath = filepath.Join(workDir, filePath)
			} else {
				fullPath = filePath
			}

			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				return fmt.Errorf("TLCP keystore 文件 %s 不存在", filePath)
			}
		}
	}

	// 验证 TLS keystore
	if cfg.TLS.Keystore != nil && cfg.TLS.Keystore.Type == "file" {
		for _, filePath := range cfg.TLS.Keystore.Params {
			if filePath == "" {
				continue
			}

			var fullPath string
			if !filepath.IsAbs(filePath) {
				fullPath = filepath.Join(workDir, filePath)
			} else {
				fullPath = filePath
			}

			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				return fmt.Errorf("TLS keystore 文件 %s 不存在", filePath)
			}
		}
	}

	return nil
}
