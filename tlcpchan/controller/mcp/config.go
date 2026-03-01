package mcp

import (
	"context"
	"fmt"

	"github.com/Trisia/tlcpchan/config"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetConfigInput 获取配置输入参数（无参数）
type GetConfigInput struct{}

// GetConfigOutput 获取配置输出参数
type GetConfigOutput struct {
	// Config 当前系统配置
	Config *config.Config `json:"config"`
}

// UpdateConfigInput 更新配置输入参数
type UpdateConfigInput struct {
	// Config 新的配置
	Config *config.Config `json:"config"`
}

// UpdateConfigOutput 更新配置输出参数
type UpdateConfigOutput struct {
	// Config 更新后的配置
	Config *config.Config `json:"config"`
}

// ReloadConfigInput 重新加载配置输入参数
type ReloadConfigInput struct {
	// ConfigPath 配置文件路径，可选，不提供则使用默认路径
	ConfigPath string `json:"configPath,omitempty"`
}

// ReloadConfigOutput 重新加载配置输出参数
type ReloadConfigOutput struct {
	// Config 重新加载后的配置
	Config *config.Config `json:"config"`
}

// handleGetConfig 处理获取配置请求
//
// 参数:
//   - ctx: 上下文
//   - req: MCP 工具调用请求
//   - input: 获取配置输入参数（无参数）
//
// 返回:
//   - *mcpsdk.CallToolResult: MCP 工具调用结果（可以为 nil，SDK 自动处理）
//   - GetConfigOutput: 获取配置输出参数，包含当前配置
//   - error: 获取失败时返回错误
//
// 注意:
//   - 此工具不需要任何参数
//   - 直接返回全局配置对象
func (c *MCPController) handleGetConfig(_ context.Context, _ *mcpsdk.CallToolRequest, input GetConfigInput) (
	*mcpsdk.CallToolResult,
	GetConfigOutput,
	error,
) {
	return nil, GetConfigOutput{Config: config.Get()}, nil
}

// handleUpdateConfig 处理更新配置请求
//
// 参数:
//   - ctx: 上下文
//   - req: MCP 工具调用请求
//   - input: 更新配置输入参数，包含新配置
//
// 返回:
//   - *mcpsdk.CallToolResult: MCP 工具调用结果（可以为 nil，SDK 自动处理）
//   - UpdateConfigOutput: 更新配置输出参数，包含更新后的配置
//   - error: 更新失败时返回错误
//
// 注意:
//   - 先验证新配置的有效性
//   - 验证通过后保存配置到文件
//   - 最后更新全局配置对象
func (c *MCPController) handleUpdateConfig(_ context.Context, _ *mcpsdk.CallToolRequest, input UpdateConfigInput) (
	*mcpsdk.CallToolResult,
	UpdateConfigOutput,
	error,
) {
	if input.Config == nil {
		return nil, UpdateConfigOutput{}, fmt.Errorf("配置参数不能为空")
	}

	// 验证配置
	if err := config.Validate(input.Config); err != nil {
		return nil, UpdateConfigOutput{}, fmt.Errorf("配置验证失败: %w", err)
	}

	// 保存配置并更新全局配置
	if err := config.SaveAndUpdate(input.Config); err != nil {
		return nil, UpdateConfigOutput{}, fmt.Errorf("保存配置失败: %w", err)
	}

	return nil, UpdateConfigOutput{Config: config.Get()}, nil
}

// handleReloadConfig 处理重新加载配置请求
//
// 参数:
//   - ctx: 上下文
//   - req: MCP 工具调用请求
//   - input: 重新加载配置输入参数，包含可选的配置文件路径
//
// 返回:
//   - *mcpsdk.CallToolResult: MCP 工具调用结果（可以为 nil，SDK 自动处理）
//   - ReloadConfigOutput: 重新加载配置输出参数，包含重新加载后的配置
//   - error: 重新加载失败时返回错误
//
// 注意:
//   - 如果未提供 ConfigPath，使用当前配置文件路径
//   - 重新加载会从文件读取并初始化全局配置
func (c *MCPController) handleReloadConfig(_ context.Context, _ *mcpsdk.CallToolRequest, input ReloadConfigInput) (
	*mcpsdk.CallToolResult,
	ReloadConfigOutput,
	error,
) {
	configPath := input.ConfigPath
	if configPath == "" {
		configPath = c.configPath
	}

	if configPath == "" {
		return nil, ReloadConfigOutput{}, fmt.Errorf("配置文件路径未指定")
	}

	// 重新加载配置
	if err := config.LoadAndInit(configPath); err != nil {
		return nil, ReloadConfigOutput{}, fmt.Errorf("重新加载配置失败: %w", err)
	}

	return nil, ReloadConfigOutput{Config: config.Get()}, nil
}

// registerConfigTools 注册配置管理工具
//
// 注意:
//   - 在 NewMCPController 中调用此函数注册所有配置管理工具
//   - 注册后客户端可以通过 MCP 协议调用这些工具
func (c *MCPController) registerConfigTools() {
	// 注册 get_config 工具
	mcpsdk.AddTool(c.server, &mcpsdk.Tool{
		Name:        "get_config",
		Description: "获取当前系统配置",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"config": map[string]any{
					"description": "当前系统配置",
					"type":        "object",
				},
			},
		},
	}, c.handleGetConfig)

	// 注册 update_config 工具
	mcpsdk.AddTool(c.server, &mcpsdk.Tool{
		Name:        "update_config",
		Description: "更新系统配置",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"config": map[string]any{
					"description": "新的配置对象",
					"type":        "object",
				},
			},
			"required": []string{"config"},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"config": map[string]any{
					"description": "更新后的配置对象",
					"type":        "object",
				},
			},
		},
	}, c.handleUpdateConfig)

	// 注册 reload_config 工具
	mcpsdk.AddTool(c.server, &mcpsdk.Tool{
		Name:        "reload_config",
		Description: "重新加载配置文件",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"configPath": map[string]any{
					"description": "配置文件路径（可选，不提供则使用默认路径）",
					"type":        "string",
				},
			},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"config": map[string]any{
					"description": "重新加载后的配置对象",
					"type":        "object",
				},
			},
		},
	}, c.handleReloadConfig)
}
