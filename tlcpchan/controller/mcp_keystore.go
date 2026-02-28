package controller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/security/keystore"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListKeystoresInput 列出密钥存储输入（无参数)
type ListKeystoresInput struct{}

// ListKeystoresOutput 列出密钥存储输出
type ListKeystoresOutput struct {
	// Keystores 密钥存储列表
	Keystores []*keystore.KeyStoreInfo `json:"keystores"`
}

// GetKeystoreInput 获取密钥存储详情输入
type GetKeystoreInput struct {
	// Name 密钥存储名称
	Name string `json:"name"`
}

// GetKeystoreOutput 获取密钥存储详情输出
type GetKeystoreOutput struct {
	// Keystore 密钥存储详细信息
	Keystore *keystore.KeyStoreInfo `json:"keystore"`
}

// CreateKeystoreInput 创建密钥存储输入
type CreateKeystoreInput struct {
	// Name 密钥存储名称
	Name string `json:"name"`
	// LoaderType 加载器类型
	LoaderType keystore.LoaderType `json:"loaderType"`
	// Params 加载器参数
	Params map[string]string `json:"params"`
	// Protected 是否受保护
	Protected bool `json:"protected"`
}

// CreateKeystoreOutput 创建密钥存储输出
type CreateKeystoreOutput struct {
	// Keystore 创建的密钥存储信息
	Keystore *keystore.KeyStoreInfo `json:"keystore"`
}

// UpdateKeystoreInput 更新密钥存储输入
type UpdateKeystoreInput struct {
	// Name 密钥存储名称
	Name string `json:"name"`
	// Params 要更新的参数
	Params map[string]string `json:"params"`
}

// UpdateKeystoreOutput 更新密钥存储输出
type UpdateKeystoreOutput struct {
	// Keystore 更新后的密钥存储信息
	Keystore *keystore.KeyStoreInfo `json:"keystore"`
}

// DeleteKeystoreInput 删除密钥存储输入
type DeleteKeystoreInput struct {
	// Name 密钥存储名称
	Name string `json:"name"`
}

// DeleteKeystoreOutput 删除密钥存储输出
type DeleteKeystoreOutput struct {
	// Success 是否成功删除
	Success bool `json:"success"`
}

/**
 * handleListKeystores 处理列出密钥存储请求
 *
 * 参数:
 *   - ctx: 上下文
 *   - req: MCP 工具调用请求
 *   - input: 列出密钥存储输入参数（无参数）
 *
 * 返回:
 *   - *mcp.CallToolResult: MCP 工具调用结果（可以为 nil，SDK 自动处理）
 *   - ListKeystoresOutput: 列出密钥存储输出参数，包含密钥存储列表
 *   - error: 列出失败时返回错误
 *
 * 注意:
 *   - 此工具不需要任何参数
 *   - 返回系统中所有配置的密钥库列表
 */
func (c *MCPController) handleListKeystores(_ context.Context, _ *mcp.CallToolRequest, input ListKeystoresInput) (
	*mcp.CallToolResult,
	ListKeystoresOutput,
	error,
) {
	keyStores := c.keyStoreMgr.List()
	return nil, ListKeystoresOutput{Keystores: keyStores}, nil
}

/**
 * handleGetKeystore 处理获取密钥存储详情请求
 *
 * 参数:
 *   - ctx: 上下文
 *   - req: MCP 工具调用请求
 *   - input: 获取密钥存储详情输入参数，包含密钥存储名称
 *
 * 返回:
 *   - *mcp.CallToolResult: MCP 工具调用结果（可以为 nil，SDK 自动处理）
 *   - GetKeystoreOutput: 获取密钥存储详情输出参数，包含密钥存储详细信息
 *   - error: 获取失败时返回错误
 *
 * 注意:
 *   - 必须提供密钥存储名称
 *   - 如果密钥存储不存在，返回错误
 */
func (c *MCPController) handleGetKeystore(_ context.Context, _ *mcp.CallToolRequest, input GetKeystoreInput) (
	*mcp.CallToolResult,
	GetKeystoreOutput,
	error,
) {
	if input.Name == "" {
		return nil, GetKeystoreOutput{}, fmt.Errorf("密钥存储名称不能为空")
	}

	info, err := c.keyStoreMgr.Get(input.Name)
	if err != nil {
		return nil, GetKeystoreOutput{}, fmt.Errorf("获取密钥存储失败: %w", err)
	}

	return nil, GetKeystoreOutput{Keystore: info}, nil
}

/**
 * handleCreateKeystore 处理创建密钥存储请求
 *
 * 参数:
 *   - ctx: 上下文
 *   - req: MCP 工具调用请求
 *   - input: 创建密钥存储输入参数，包含名称、加载器类型、参数和受保护标志
 *
 * 返回:
 *   - *mcp.CallToolResult: MCP 工具调用结果（可以为 nil，SDK 自动处理）
 *   - CreateKeystoreOutput: 创建密钥存储输出参数，包含创建的密钥存储信息
 *   - error: 创建失败时返回错误
 *
 * 注意:
 *   - 名称和加载器类型为必填参数
 *   - 如果是 file 类型，会验证文件是否存在
 *   - 创建成功后会自动更新配置文件
 */
func (c *MCPController) handleCreateKeystore(_ context.Context, _ *mcp.CallToolRequest, input CreateKeystoreInput) (
	*mcp.CallToolResult,
	CreateKeystoreOutput,
	error,
) {
	if input.Name == "" {
		return nil, CreateKeystoreOutput{}, fmt.Errorf("密钥存储名称不能为空")
	}

	if input.LoaderType == "" {
		return nil, CreateKeystoreOutput{}, fmt.Errorf("加载器类型不能为空")
	}

	if input.Params == nil {
		input.Params = make(map[string]string)
	}

	// 如果是 file 类型，验证文件是否存在
	if input.LoaderType == keystore.LoaderTypeFile {
		if err := validateMCPFileParams(c.config.WorkDir, input.Params); err != nil {
			return nil, CreateKeystoreOutput{}, fmt.Errorf("文件验证失败: %w", err)
		}
	}

	// 创建 keystore
	info, err := c.keyStoreMgr.Create(input.Name, input.LoaderType, input.Params, input.Protected)
	if err != nil {
		return nil, CreateKeystoreOutput{}, fmt.Errorf("创建密钥存储失败: %w", err)
	}

	// 更新配置文件
	c.config.KeyStores = append(c.config.KeyStores, config.KeyStoreConfig{
		Name:   input.Name,
		Type:   input.LoaderType,
		Params: input.Params,
	})

	if err := config.Save(c.config); err != nil {
		return nil, CreateKeystoreOutput{}, fmt.Errorf("保存配置失败: %w", err)
	}

	c.log.Info("创建 keystore: %s", input.Name)
	return nil, CreateKeystoreOutput{Keystore: info}, nil
}

/**
 * handleUpdateKeystore 处理更新密钥存储请求
 *
 * 参数:
 *   - ctx: 上下文
 *   - req: MCP 工具调用请求
 *   - input: 更新密钥存储输入参数，包含名称和要更新的参数
 *
 * 返回:
 *   - *mcp.CallToolResult: MCP 工具调用结果（可以为 nil，SDK 自动处理）
 *   - UpdateKeystoreOutput: 更新密钥存储输出参数，包含更新后的密钥存储信息
 *   - error: 更新失败时返回错误
 *
 * 注意:
 *   - 名称和参数为必填参数
 *   - 受保护的 keystore 不允许修改
 *   - 更新成功后会自动保存配置文件
 */
func (c *MCPController) handleUpdateKeystore(_ context.Context, _ *mcp.CallToolRequest, input UpdateKeystoreInput) (
	*mcp.CallToolResult,
	UpdateKeystoreOutput,
	error,
) {
	if input.Name == "" {
		return nil, UpdateKeystoreOutput{}, fmt.Errorf("密钥存储名称不能为空")
	}

	if input.Params == nil || len(input.Params) == 0 {
		return nil, UpdateKeystoreOutput{}, fmt.Errorf("参数不能为空")
	}

	// 检查 keystore 是否存在
	info, err := c.keyStoreMgr.Get(input.Name)
	if err != nil {
		return nil, UpdateKeystoreOutput{}, fmt.Errorf("密钥存储不存在: %w", err)
	}

	// 检查是否为 protected 状态
	if info.Protected {
		return nil, UpdateKeystoreOutput{}, fmt.Errorf("受保护的 keystore 不允许修改")
	}

	// 如果是 file 类型，验证文件是否存在
	if info.LoaderType == keystore.LoaderTypeFile {
		if err := validateMCPFileParams(c.config.WorkDir, input.Params); err != nil {
			return nil, UpdateKeystoreOutput{}, fmt.Errorf("文件验证失败: %w", err)
		}
	}

	// 更新 keystore 配置
	found := false
	for i := range c.config.KeyStores {
		if c.config.KeyStores[i].Name == input.Name {
			if c.config.KeyStores[i].Params == nil {
				c.config.KeyStores[i].Params = make(map[string]string)
			}
			// 更新参数
			for key, value := range input.Params {
				c.config.KeyStores[i].Params[key] = value
			}
			found = true
			break
		}
	}

	if !found {
		return nil, UpdateKeystoreOutput{}, fmt.Errorf("配置中未找到 keystore")
	}

	// 保存配置文件
	if err := config.Save(c.config); err != nil {
		return nil, UpdateKeystoreOutput{}, fmt.Errorf("保存配置失败: %w", err)
	}

	// 重新加载 keystore
	updatedInfo, err := c.keyStoreMgr.Get(input.Name)
	if err != nil {
		return nil, UpdateKeystoreOutput{}, fmt.Errorf("重新加载 keystore 失败: %w", err)
	}

	c.log.Info("更新 keystore 参数: %s", input.Name)
	return nil, UpdateKeystoreOutput{Keystore: updatedInfo}, nil
}

/**
 * handleDeleteKeystore 处理删除密钥存储请求
 *
 * 参数:
 *   - ctx: 上下文
 *   - req: MCP 工具调用请求
 *   - input: 删除密钥存储输入参数，包含密钥存储名称
 *
 * 返回:
 *   - *mcp.CallToolResult: MCP 工具调用结果（可以为 nil，SDK 自动处理）
 *   - DeleteKeystoreOutput: 删除密钥存储输出参数，包含删除成功标志
 *   - error: 删除失败时返回错误
 *
 * 注意:
 *   - 必须提供密钥存储名称
 *   - 受保护的 keystore 不允许删除
 *   - 删除成功后会自动更新配置文件
 */
func (c *MCPController) handleDeleteKeystore(_ context.Context, _ *mcp.CallToolRequest, input DeleteKeystoreInput) (
	*mcp.CallToolResult,
	DeleteKeystoreOutput,
	error,
) {
	if input.Name == "" {
		return nil, DeleteKeystoreOutput{}, fmt.Errorf("密钥存储名称不能为空")
	}

	// 检查 keystore 是否存在
	_, err := c.keyStoreMgr.Get(input.Name)
	if err != nil {
		return nil, DeleteKeystoreOutput{}, fmt.Errorf("密钥存储不存在: %w", err)
	}

	// 删除 keystore
	if err := c.keyStoreMgr.Delete(input.Name); err != nil {
		return nil, DeleteKeystoreOutput{}, fmt.Errorf("删除密钥存储失败: %w", err)
	}

	// 更新配置文件，删除对应的配置
	newKeyStores := make([]config.KeyStoreConfig, 0, len(c.config.KeyStores))
	for _, ks := range c.config.KeyStores {
		if ks.Name != input.Name {
			newKeyStores = append(newKeyStores, ks)
		}
	}
	c.config.KeyStores = newKeyStores

	if err := config.Save(c.config); err != nil {
		return nil, DeleteKeystoreOutput{}, fmt.Errorf("保存配置失败: %w", err)
	}

	c.log.Info("删除 keystore: %s", input.Name)
	return nil, DeleteKeystoreOutput{Success: true}, nil
}

/**
 * registerKeystoreTools 注册密钥存储管理工具
 *
 * 注意:
 *   - 在 NewMCPController 中调用此函数注册所有密钥存储管理工具
 *   - 注册后客户端可以通过 MCP 协议调用这些工具
 */
func (c *MCPController) registerKeystoreTools() {
	// 注册 list_keystores 工具
	mcp.AddTool(c.server, &mcp.Tool{
		Name:        "list_keystores",
		Description: "获取所有密钥存储（keystore）的列表信息",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"keystores": map[string]any{
					"description": "密钥存储列表",
					"type":        "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"name": map[string]any{
								"description": "密钥存储名称",
								"type":        "string",
							},
							"type": map[string]any{
								"description": "密钥存储类型（tlcp/tls）",
								"type":        "string",
							},
							"loaderType": map[string]any{
								"description": "加载器类型（file/named/skf/sdf）",
								"type":        "string",
							},
							"params": map[string]any{
								"description": "加载器参数",
								"type":        "object",
							},
							"protected": map[string]any{
								"description": "是否受保护",
								"type":        "boolean",
							},
							"createdAt": map[string]any{
								"description": "创建时间",
								"type":        "string",
							},
							"updatedAt": map[string]any{
								"description": "更新时间",
								"type":        "string",
							},
						},
					},
				},
			},
		},
	}, c.handleListKeystores)

	// 注册 get_keystore 工具
	mcp.AddTool(c.server, &mcp.Tool{
		Name:        "get_keystore",
		Description: "获取指定名称的密钥存储（keystore）的详细信息",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"description": "密钥存储名称",
					"type":        "string",
				},
			},
			"required": []string{"name"},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"keystore": map[string]any{
					"description": "密钥存储详细信息",
					"type":        "object",
				},
			},
		},
	}, c.handleGetKeystore)

	// 注册 create_keystore 工具
	mcp.AddTool(c.server, &mcp.Tool{
		Name:        "create_keystore",
		Description: "创建新的密钥存储（keystore），创建成功后会自动更新配置文件",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"description": "密钥存储名称",
					"type":        "string",
				},
				"loaderType": map[string]any{
					"description": "加载器类型（file/named/skf/sdf）",
					"type":        "string",
				},
				"params": map[string]any{
					"description": "加载器参数",
					"type":        "object",
				},
				"protected": map[string]any{
					"description": "是否受保护",
					"type":        "boolean",
				},
			},
			"required": []string{"name", "loaderType", "params"},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"keystore": map[string]any{
					"description": "创建的密钥存储信息",
					"type":        "object",
				},
			},
		},
	}, c.handleCreateKeystore)

	// 注册 update_keystore 工具
	mcp.AddTool(c.server, &mcp.Tool{
		Name:        "update_keystore",
		Description: "更新指定密钥存储（keystore）的参数（如证书和密钥路径），更新后自动保存配置",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"description": "密钥存储名称",
					"type":        "string",
				},
				"params": map[string]any{
					"description": "要更新的参数键值对",
					"type":        "object",
				},
			},
			"required": []string{"name", "params"},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"keystore": map[string]any{
					"description": "更新后的密钥存储信息",
					"type":        "object",
				},
			},
		},
	}, c.handleUpdateKeystore)

	// 注册 delete_keystore 工具
	mcp.AddTool(c.server, &mcp.Tool{
		Name:        "delete_keystore",
		Description: "删除指定的密钥存储（keystore），删除后会自动更新配置文件",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"description": "密钥存储名称",
					"type":        "string",
				},
			},
			"required": []string{"name"},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"success": map[string]any{
					"description": "是否成功删除",
					"type":        "boolean",
				},
			},
		},
	}, c.handleDeleteKeystore)
}

/**
 * validateMCPFileParams 验证 file 类型 keystore 的所有文件路径是否存在
 *
 * 参数：
 *   - workDir: 工作目录，用于解析相对路径
 *   - params: keystore 参数，包含文件路径
 *
 * 返回：
 *   - error: 如果文件不存在则返回错误信息，否则返回 nil
 */
func validateMCPFileParams(workDir string, params map[string]string) error {
	for _, filePath := range params {
		if filePath == "" {
			continue
		}

		// 解析文件路径，处理相对路径
		var fullPath string
		if strings.HasPrefix(filePath, "./") || strings.HasPrefix(filePath, "/") {
			fullPath = filePath
		} else {
			fullPath = "./" + filePath
		}

		// 如果是相对路径，相对于工作目录解析
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(workDir, fullPath)
		}

		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return fmt.Errorf("文件 %s 不存在", filePath)
		}
	}
	return nil
}
