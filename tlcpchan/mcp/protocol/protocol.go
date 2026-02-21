package protocol

import (
	"encoding/json"
	"fmt"
)

// JSONRPCVersion JSON-RPC版本
const JSONRPCVersion = "2.0"

// Request JSON-RPC请求
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response JSON-RPC响应
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

// Error JSON-RPC错误
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 错误码定义
const (
	ErrParseError     = -32700
	ErrInvalidRequest = -32600
	ErrMethodNotFound = -32601
	ErrInvalidParams  = -32602
	ErrInternal       = -32603
)

// NewRequest 创建新请求
func NewRequest(id interface{}, method string, params json.RawMessage) *Request {
	return &Request{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Method:  method,
		Params:  params,
	}
}

// NewResponse 创建新响应
func NewResponse(id interface{}, result interface{}) (*Response, error) {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("序列化结果失败: %w", err)
	}
	return &Response{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Result:  resultJSON,
	}, nil
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(id interface{}, code int, message string, data interface{}) *Response {
	return &Response{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Error: &Error{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

// MCPInitializeRequest MCP初始化请求
type MCPInitializeRequest struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"clientInfo"`
}

// MCPInitializeResponse MCP初始化响应
type MCPInitializeResponse struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ServerInfo      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"serverInfo"`
}

// MCPToolsListRequest 工具列表请求
type MCPToolsListRequest struct {
	Cursor string `json:"cursor,omitempty"`
}

// MCPToolsListResponse 工具列表响应
type MCPToolsListResponse struct {
	Tools  []Tool `json:"tools"`
	Cursor string `json:"cursor,omitempty"`
}

// Tool MCP工具定义
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// MCPToolsCallRequest 工具调用请求
type MCPToolsCallRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// MCPToolsCallResponse 工具调用响应
type MCPToolsCallResponse struct {
	Content []ContentItem `json:"content"`
}

// ContentItem 内容项
type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// NewTextContent 创建文本内容
func NewTextContent(text string) ContentItem {
	return ContentItem{
		Type: "text",
		Text: text,
	}
}
