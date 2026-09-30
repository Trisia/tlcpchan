package mcp

import (
	"context"
	"fmt"

	"github.com/Trisia/tlcpchan/security/ibcparams"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListIBCParamsInput 列出 IBC 信任池输入（无参数）
type ListIBCParamsInput struct{}

// ListIBCParamsOutput 列出 IBC 信任池输出
type ListIBCParamsOutput struct {
	// Params KGC 公共参数列表
	Params []*ibcparams.IBCParam `json:"params"`
}

// AddIBCParamsInput 添加 KGC 公共参数输入
type AddIBCParamsInput struct {
	// Filename 保存的文件名，唯一标识符
	Filename string `json:"filename"`
	// Params 公共参数数据，支持 PEM / DER / HEX / Base64 文本
	Params string `json:"params"`
}

// AddIBCParamsOutput 添加 KGC 公共参数输出
type AddIBCParamsOutput struct {
	// Param 添加成功的条目元信息
	Param *ibcparams.IBCParam `json:"param"`
}

// RemoveIBCParamsInput 删除 KGC 公共参数输入
type RemoveIBCParamsInput struct {
	// Filename 文件名，唯一标识符
	Filename string `json:"filename"`
}

// RemoveIBCParamsOutput 删除 KGC 公共参数输出
type RemoveIBCParamsOutput struct {
	// Success 是否成功删除
	Success bool `json:"success"`
}

// ReloadIBCParamsInput 重载 IBC 信任池输入（无参数）
type ReloadIBCParamsInput struct{}

// ReloadIBCParamsOutput 重载 IBC 信任池输出
type ReloadIBCParamsOutput struct {
	// Success 是否成功重载
	Success bool `json:"success"`
	// Count 重载后信任池中的条目数
	Count int `json:"count"`
}

/**
 * handleListIBCParams 处理列出 IBC 信任池请求
 *
 * 参数:
 *   - ctx: 上下文
 *   - req: MCP 工具调用请求
 *   - input: 列出 IBC 信任池输入参数（无参数）
 *
 * 返回:
 *   - *mcpsdk.CallToolResult: MCP 工具调用结果（可以为 nil，SDK 自动处理）
 *   - ListIBCParamsOutput: KGC 公共参数列表
 *   - error: 列出失败时返回错误
 */
func (c *MCPController) handleListIBCParams(_ context.Context, _ *mcpsdk.CallToolRequest, input ListIBCParamsInput) (
	*mcpsdk.CallToolResult,
	ListIBCParamsOutput,
	error,
) {
	if c.ibcParamMgr == nil {
		return nil, ListIBCParamsOutput{}, fmt.Errorf("IBC 信任池管理器未初始化")
	}
	return nil, ListIBCParamsOutput{Params: c.ibcParamMgr.List()}, nil
}

/**
 * handleAddIBCParams 处理添加 KGC 公共参数请求
 *
 * 参数:
 *   - ctx: 上下文
 *   - req: MCP 工具调用请求
 *   - input: 添加输入参数，包含文件名与公共参数文本
 *
 * 返回:
 *   - *mcpsdk.CallToolResult: MCP 工具调用结果（可以为 nil，SDK 自动处理）
 *   - AddIBCParamsOutput: 添加成功的条目元信息
 *   - error: 添加失败时返回错误
 *
 * 注意:
 *   - 必须通过可信渠道获取 KGC 公共参数后再入库，绝不直接信任对端下发的参数
 */
func (c *MCPController) handleAddIBCParams(_ context.Context, _ *mcpsdk.CallToolRequest, input AddIBCParamsInput) (
	*mcpsdk.CallToolResult,
	AddIBCParamsOutput,
	error,
) {
	if c.ibcParamMgr == nil {
		return nil, AddIBCParamsOutput{}, fmt.Errorf("IBC 信任池管理器未初始化")
	}
	if input.Filename == "" {
		return nil, AddIBCParamsOutput{}, fmt.Errorf("文件名不能为空")
	}
	if input.Params == "" {
		return nil, AddIBCParamsOutput{}, fmt.Errorf("公共参数不能为空")
	}

	param, err := c.ibcParamMgr.Add(input.Filename, []byte(input.Params))
	if err != nil {
		return nil, AddIBCParamsOutput{}, fmt.Errorf("添加失败: %w", err)
	}

	c.log.Info("添加 IBC 公共参数: %s", input.Filename)
	return nil, AddIBCParamsOutput{Param: param}, nil
}

/**
 * handleRemoveIBCParams 处理删除 KGC 公共参数请求
 *
 * 参数:
 *   - ctx: 上下文
 *   - req: MCP 工具调用请求
 *   - input: 删除输入参数，包含文件名
 *
 * 返回:
 *   - *mcpsdk.CallToolResult: MCP 工具调用结果（可以为 nil，SDK 自动处理）
 *   - RemoveIBCParamsOutput: 删除结果
 *   - error: 删除失败时返回错误
 */
func (c *MCPController) handleRemoveIBCParams(_ context.Context, _ *mcpsdk.CallToolRequest, input RemoveIBCParamsInput) (
	*mcpsdk.CallToolResult,
	RemoveIBCParamsOutput,
	error,
) {
	if c.ibcParamMgr == nil {
		return nil, RemoveIBCParamsOutput{}, fmt.Errorf("IBC 信任池管理器未初始化")
	}
	if input.Filename == "" {
		return nil, RemoveIBCParamsOutput{}, fmt.Errorf("文件名不能为空")
	}

	if err := c.ibcParamMgr.Delete(input.Filename); err != nil {
		return nil, RemoveIBCParamsOutput{}, fmt.Errorf("删除失败: %w", err)
	}

	c.log.Info("删除 IBC 公共参数: %s", input.Filename)
	return nil, RemoveIBCParamsOutput{Success: true}, nil
}

/**
 * handleReloadIBCParams 处理重载 IBC 信任池请求
 *
 * 参数:
 *   - ctx: 上下文
 *   - req: MCP 工具调用请求
 *   - input: 重载输入参数（无参数）
 *
 * 返回:
 *   - *mcpsdk.CallToolResult: MCP 工具调用结果（可以为 nil，SDK 自动处理）
 *   - ReloadIBCParamsOutput: 重载结果与当前条目数
 *   - error: 重载失败时返回错误
 *
 * 注意:
 *   - 重载后需要重启或重载相关实例才会生效
 */
func (c *MCPController) handleReloadIBCParams(_ context.Context, _ *mcpsdk.CallToolRequest, input ReloadIBCParamsInput) (
	*mcpsdk.CallToolResult,
	ReloadIBCParamsOutput,
	error,
) {
	if c.ibcParamMgr == nil {
		return nil, ReloadIBCParamsOutput{}, fmt.Errorf("IBC 信任池管理器未初始化")
	}

	if err := c.ibcParamMgr.Reload(); err != nil {
		return nil, ReloadIBCParamsOutput{}, fmt.Errorf("重载失败: %w", err)
	}

	c.log.Info("重载 IBC 信任池")
	return nil, ReloadIBCParamsOutput{Success: true, Count: len(c.ibcParamMgr.List())}, nil
}

/**
 * registerIBCParamTools 注册 IBC 信任池管理工具
 *
 * 注意:
 *   - 在 NewMCPController 中调用此函数注册所有 IBC 信任池管理工具
 *   - 注册后客户端可以通过 MCP 协议调用这些工具
 */
func (c *MCPController) registerIBCParamTools() {
	// 注册 list_ibc_params 工具
	mcpsdk.AddTool(c.server, &mcpsdk.Tool{
		Name:        "list_ibc_params",
		Description: "获取 IBC 信任池（KGC 公共参数）列表，客户端与服务端共用同一列表",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"params": map[string]any{
					"description": "KGC 公共参数列表",
					"type":        "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"filename":           map[string]any{"type": "string", "description": "文件名"},
							"districtName":       map[string]any{"type": "string", "description": "KGC 属地区域名"},
							"districtSerial":     map[string]any{"type": "integer", "description": "KGC 序号"},
							"notBefore":          map[string]any{"type": "string", "description": "生效时间"},
							"notAfter":           map[string]any{"type": "string", "description": "失效时间"},
							"issuerIdentity":     map[string]any{"type": "string", "description": "颁发者标识"},
							"signKeyFingerprint": map[string]any{"type": "string", "description": "签名主公钥 SM3 指纹"},
							"encKeyFingerprint":  map[string]any{"type": "string", "description": "加密主公钥 SM3 指纹"},
						},
					},
				},
			},
		},
	}, c.handleListIBCParams)

	// 注册 add_ibc_params 工具
	mcpsdk.AddTool(c.server, &mcpsdk.Tool{
		Name:        "add_ibc_params",
		Description: "添加 KGC 公共参数到 IBC 信任池（支持 PEM / DER / HEX / Base64 文本）。必须通过可信渠道获取参数",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"filename": map[string]any{"type": "string", "description": "保存的文件名，唯一标识符"},
				"params":   map[string]any{"type": "string", "description": "公共参数内容（PEM / DER / HEX / Base64）"},
			},
			"required": []string{"filename", "params"},
		},
	}, c.handleAddIBCParams)

	// 注册 remove_ibc_params 工具
	mcpsdk.AddTool(c.server, &mcpsdk.Tool{
		Name:        "remove_ibc_params",
		Description: "从 IBC 信任池中删除指定的 KGC 公共参数",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"filename": map[string]any{"type": "string", "description": "文件名，唯一标识符"},
			},
			"required": []string{"filename"},
		},
	}, c.handleRemoveIBCParams)

	// 注册 reload_ibc_params 工具
	mcpsdk.AddTool(c.server, &mcpsdk.Tool{
		Name:        "reload_ibc_params",
		Description: "重新扫描信任池目录并重建 IBC 信任池，重载后需要重启或重载相关实例才会生效",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"success": map[string]any{"type": "boolean", "description": "是否成功重载"},
				"count":   map[string]any{"type": "integer", "description": "重载后信任池条目数"},
			},
		},
	}, c.handleReloadIBCParams)
}
