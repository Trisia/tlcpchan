package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/mcp/connection"
	"github.com/Trisia/tlcpchan/mcp/protocol"
	"github.com/Trisia/tlcpchan/mcp/tools"
	"github.com/Trisia/tlcpchan/version"
	"github.com/gorilla/websocket"
)

// Server MCP服务器
type Server struct {
	upgrader     websocket.Upgrader
	toolRegistry *tools.ToolRegistry
	connections  sync.Map
	cfg          *config.Config
}

// NewServer 创建MCP服务器
func NewServer(cfg *config.Config) *Server {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	return &Server{
		upgrader:     upgrader,
		toolRegistry: tools.NewToolRegistry(),
		cfg:          cfg,
	}
}

// RegisterTool 注册工具
func (s *Server) RegisterTool(tool tools.Tool) {
	s.toolRegistry.Register(tool)
}

// HandleWebSocket 处理WebSocket连接
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	if s.cfg.MCP.APIKey != "" {
		clientKey := r.URL.Query().Get("api_key")
		if clientKey != s.cfg.MCP.APIKey {
			http.Error(w, "未授权", http.StatusUnauthorized)
			return
		}
	}

	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("WebSocket升级失败: %v", err), http.StatusInternalServerError)
		return
	}

	connID := fmt.Sprintf("%p", ws)
	conn := connection.NewConnection(ws, connID)
	s.connections.Store(connID, conn)

	go s.handleConnection(conn)
}

// handleConnection 处理单个连接
func (s *Server) handleConnection(conn *connection.Connection) {
	defer func() {
		conn.Close()
		s.connections.Delete(conn.ID())
	}()

	for {
		req, err := conn.ReadRequest()
		if err != nil {
			break
		}

		resp, err := s.handleRequest(req)
		if err != nil {
			resp = protocol.NewErrorResponse(req.ID, protocol.ErrInternal, err.Error(), nil)
		}

		if err := conn.WriteResponse(resp); err != nil {
			break
		}
	}
}

// handleRequest 处理请求
func (s *Server) handleRequest(req *protocol.Request) (*protocol.Response, error) {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	default:
		return protocol.NewErrorResponse(req.ID, protocol.ErrMethodNotFound, "未知方法", nil), nil
	}
}

// handleInitialize 处理初始化请求
func (s *Server) handleInitialize(req *protocol.Request) (*protocol.Response, error) {
	var initReq protocol.MCPInitializeRequest
	if err := json.Unmarshal(req.Params, &initReq); err != nil {
		return protocol.NewErrorResponse(req.ID, protocol.ErrInvalidParams, "无效参数", err.Error()), nil
	}

	resp := protocol.MCPInitializeResponse{
		ProtocolVersion: "2024-11-05",
		Capabilities: map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		ServerInfo: struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}{
			Name:    "tlcpchan-mcp",
			Version: version.Version,
		},
	}

	return protocol.NewResponse(req.ID, resp)
}

// handleToolsList 处理工具列表请求
func (s *Server) handleToolsList(req *protocol.Request) (*protocol.Response, error) {
	toolList := s.toolRegistry.List()

	toolsResult := make([]protocol.Tool, 0, len(toolList))
	for _, t := range toolList {
		tool := protocol.Tool{
			Name:        t.Name(),
			Description: t.Description(),
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"method": map[string]interface{}{
						"type":        "string",
						"description": "工具方法",
					},
					"params": map[string]interface{}{
						"type":        "object",
						"description": "方法参数",
					},
				},
				"required": []string{"method"},
			},
		}
		toolsResult = append(toolsResult, tool)
	}

	resp := protocol.MCPToolsListResponse{
		Tools: toolsResult,
	}

	return protocol.NewResponse(req.ID, resp)
}

// handleToolsCall 处理工具调用请求
func (s *Server) handleToolsCall(req *protocol.Request) (*protocol.Response, error) {
	var callReq protocol.MCPToolsCallRequest
	if err := json.Unmarshal(req.Params, &callReq); err != nil {
		return protocol.NewErrorResponse(req.ID, protocol.ErrInvalidParams, "无效参数", err.Error()), nil
	}

	tool, ok := s.toolRegistry.Get(callReq.Name)
	if !ok {
		return protocol.NewErrorResponse(req.ID, protocol.ErrMethodNotFound, "工具不存在", nil), nil
	}

	var method string
	var params json.RawMessage

	if m, ok := callReq.Arguments["method"].(string); ok {
		method = m
	}
	if p, ok := callReq.Arguments["params"]; ok {
		params, _ = json.Marshal(p)
	}

	if method == "" {
		return protocol.NewErrorResponse(req.ID, protocol.ErrInvalidParams, "方法名不能为空", nil), nil
	}

	result, err := tool.Execute(context.Background(), method, params)
	if err != nil {
		return protocol.NewErrorResponse(req.ID, protocol.ErrInternal, err.Error(), nil), nil
	}

	var content []protocol.ContentItem
	if result != nil {
		resultJSON, _ := json.Marshal(result)
		content = []protocol.ContentItem{
			protocol.NewTextContent(string(resultJSON)),
		}
	} else {
		content = []protocol.ContentItem{
			protocol.NewTextContent("操作成功"),
		}
	}

	resp := protocol.MCPToolsCallResponse{
		Content: content,
	}

	return protocol.NewResponse(req.ID, resp)
}
