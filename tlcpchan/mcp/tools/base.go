package tools

import (
	"context"
	"encoding/json"
)

// Tool MCP工具接口
type Tool interface {
	Name() string
	Description() string
	Methods() []string
	Execute(ctx context.Context, method string, params json.RawMessage) (interface{}, error)
}

// BaseTool 工具基类
type BaseTool struct {
	name        string
	description string
	methods     []string
}

// NewBaseTool 创建基类工具
func NewBaseTool(name, description string, methods []string) *BaseTool {
	return &BaseTool{
		name:        name,
		description: description,
		methods:     methods,
	}
}

// Name 获取工具名称
func (t *BaseTool) Name() string {
	return t.name
}

// Description 获取工具描述
func (t *BaseTool) Description() string {
	return t.description
}

// Methods 获取工具方法列表
func (t *BaseTool) Methods() []string {
	return t.methods
}

// ToolRegistry 工具注册表
type ToolRegistry struct {
	tools map[string]Tool
}

// NewToolRegistry 创建工具注册表
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register 注册工具
func (r *ToolRegistry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}

// Get 获取工具
func (r *ToolRegistry) Get(name string) (Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// List 列出所有工具
func (r *ToolRegistry) List() []Tool {
	result := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		result = append(result, tool)
	}
	return result
}
