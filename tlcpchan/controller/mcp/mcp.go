package mcp

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/instance"
	"github.com/Trisia/tlcpchan/logger"
	"github.com/Trisia/tlcpchan/security"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ServerOptions MCP 控制器配置选项
type ServerOptions struct {
	// Config 系统配置
	Config *config.Config
	// ConfigPath 配置文件路径
	ConfigPath string
	// KeyStoreManager 密钥存储管理器
	KeyStoreManager *security.KeyStoreManager
	// RootCertManager 根证书管理器
	RootCertManager *security.RootCertManager
	// InstanceManager 实例管理器
	InstanceManager *instance.Manager
	// StaticDir 静态文件目录
	StaticDir string
}

// MCPController MCP 控制器
type MCPController struct {
	config      *config.Config
	instanceMgr *instance.Manager
	keyStoreMgr *security.KeyStoreManager
	rootCertMgr *security.RootCertManager
	configPath  string
	server      *mcpsdk.Server
	sseHandler  *mcpsdk.SSEHandler
	log         *logger.Logger
	mu          sync.RWMutex
	started     bool
}

// NewMCPController 创建新的 MCP 控制器
//
// 参数:
//   - opts: 控制器配置选项
//
// 返回:
//   - *MCPController: MCP 控制器实例
//   - error: 创建失败时返回错误
//
// 注意:
//   - 如果 config.MCP.Enabled 为 false，控制器不会启动任何服务
//   - config.MCP.APIKey 为空时，跳过认证（开放访问）
func NewMCPController(opts *ServerOptions) (*MCPController, error) {
	c := &MCPController{
		config:      opts.Config,
		instanceMgr: opts.InstanceManager,
		keyStoreMgr: opts.KeyStoreManager,
		rootCertMgr: opts.RootCertManager,
		configPath:  opts.ConfigPath,
		log:         logger.Default(),
		started:     false,
	}

	// 检查是否启用 MCP 服务
	if !opts.Config.MCP.Enabled {
		c.log.Info("MCP 服务未启用")
		return c, nil
	}

	// 设置默认服务器信息
	serverName := opts.Config.MCP.ServerInfo.Name
	if serverName == "" {
		serverName = "tlcpchan-mcp"
	}
	serverVersion := opts.Config.MCP.ServerInfo.Version
	if serverVersion == "" {
		serverVersion = "1.0.0"
	}

	// 创建 MCP SDK 服务器
	server := mcpsdk.NewServer(
		&mcpsdk.Implementation{
			Name:    serverName,
			Version: serverVersion,
		},
		nil, // server capabilities
	)

	c.server = server

	// 创建 SSE Handler
	c.sseHandler = mcpsdk.NewSSEHandler(func(r *http.Request) *mcpsdk.Server {
		return server
	}, nil)

	// 注册配置管理工具
	c.registerConfigTools()

	// 注册密钥存储管理工具
	c.registerKeystoreTools()

	// 注册日志管理工具
	c.registerLogTools()

	// 注册系统信息工具
	c.registerSystemTools()

	// 注册实例管理工具
	c.registerInstanceTools()

	c.log.Info("MCP 控制器创建成功: %s v%s", serverName, serverVersion)

	return c, nil
}

// RegisterRoutes 注册 MCP 路由到外部路由器
//
// 参数:
//   - registerFunc: 路由注册函数，签名为 func(pattern string, handler http.HandlerFunc)
//
// 注意:
//   - 如果 MCP 服务未启用，不注册任何路由
func (c *MCPController) RegisterRoutes(registerFunc func(pattern string, handler http.HandlerFunc)) {
	if c == nil || !c.config.MCP.Enabled {
		return
	}
	registerFunc("/api/mcp/sse", c.handleSSE)
}

// Start 启动 MCP 服务
//
// 参数:
//   - ctx: 上下文，用于取消启动
//
// 返回:
//   - error: 启动失败时返回错误
//
// 注意:
//   - 如果 MCP 服务未启用，立即返回 nil
func (c *MCPController) Start(ctx context.Context) error {
	if c == nil || !c.config.MCP.Enabled {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started {
		return fmt.Errorf("MCP 服务已启动")
	}

	c.started = true
	c.log.Info("MCP 服务启动")

	// 监听上下文取消
	go func() {
		<-ctx.Done()
		c.Stop()
	}()

	return nil
}

// Stop 停止 MCP 服务
//
// 返回:
//   - error: 停止失败时返回错误
//
// 注意:
//   - 如果 MCP 服务未启动，立即返回 nil
func (c *MCPController) Stop() error {
	if c == nil || !c.config.MCP.Enabled {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.started {
		return nil
	}

	c.started = false
	c.log.Info("MCP 服务停止")

	return nil
}

// handleSSE 处理 SSE 连接请求
//
// 参数:
//   - w: HTTP 响应写入器
//   - r: HTTP 请求
//
// 注意:
//   - 验证请求方法和认证
//   - 委托给 SSE Handler 处理
func (c *MCPController) handleSSE(w http.ResponseWriter, r *http.Request) {
	// 验证请求方法
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 验证认证
	if err := mcpAuthenticate(c.config.MCP.APIKey, r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 委托给 SSE Handler 处理
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.started {
		http.Error(w, "MCP 服务未启动", http.StatusServiceUnavailable)
		return
	}

	if c.sseHandler != nil {
		c.sseHandler.ServeHTTP(w, r)
	} else {
		http.Error(w, "SSE Handler 未初始化", http.StatusServiceUnavailable)
	}
}

// mcpAuthenticate 验证请求是否通过 MCP 认证
//
// 参数:
//   - apiKey: 配置的 API Key，为空表示开放访问
//   - r: HTTP 请求
//
// 返回:
//   - error: 认证失败时返回错误，通过时返回 nil
//
// 注意:
//   - 如果 apiKey 为空字符串，跳过认证（开放访问）
//   - 如果 apiKey 非空，验证请求头中的 Authorization
//   - Authorization 格式: Bearer <api_key>
func mcpAuthenticate(apiKey string, r *http.Request) error {
	// 如果 API Key 为空，跳过认证（开放访问）
	if apiKey == "" {
		return nil
	}

	// 检查 Authorization 请求头
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return fmt.Errorf("认证失败：缺少 Authorization 头")
	}

	// 验证 Bearer Token 格式
	if !strings.HasPrefix(auth, "Bearer ") {
		return fmt.Errorf("认证失败：无效的 Authorization 格式")
	}

	// 提取 Token
	token := strings.TrimPrefix(auth, "Bearer ")
	if token == "" {
		return fmt.Errorf("认证失败：Token 为空")
	}

	// 验证 Token 是否匹配
	if token != apiKey {
		return fmt.Errorf("认证失败：Token 不匹配")
	}

	return nil
}
