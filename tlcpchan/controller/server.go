package controller

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/instance"
	"github.com/Trisia/tlcpchan/logger"
	"github.com/Trisia/tlcpchan/mcp"
	"github.com/Trisia/tlcpchan/mcp/tools"
	"github.com/Trisia/tlcpchan/security"
)

// Server API服务器，提供RESTful API接口和UI静态文件服务
type Server struct {
	httpServer  *http.Server
	router      *Router
	cfg         *config.Config
	log         *logger.Logger
	keyStoreMgr *security.KeyStoreManager
	rootCertMgr *security.RootCertManager
	staticDir   string
	fileServer  http.Handler
	mcpServer   *mcp.Server
}

// ServerOptions API服务器配置选项
type ServerOptions struct {
	Config          *config.Config
	ConfigPath      string
	KeyStoreManager *security.KeyStoreManager
	RootCertManager *security.RootCertManager
	InstanceManager *instance.Manager
	StaticDir       string
}

// NewServer 创建新的API服务器
func NewServer(opts ServerOptions) *Server {
	router := NewRouter()
	router.Use(corsMiddleware)
	router.Use(loggingMiddleware)

	log := logger.Default()

	var keyStoreMgr *security.KeyStoreManager
	var rootCertMgr *security.RootCertManager
	var instMgr *instance.Manager

	if opts.KeyStoreManager != nil {
		keyStoreMgr = opts.KeyStoreManager
	} else {
		keyStoreMgr = security.NewKeyStoreManager()
	}

	if opts.RootCertManager != nil {
		rootCertMgr = opts.RootCertManager
	} else {
		rootCertMgr = security.NewRootCertManager("")
	}

	if opts.InstanceManager != nil {
		instMgr = opts.InstanceManager
	} else {
		instMgr = instance.NewManager(log, keyStoreMgr, rootCertMgr)
		for i := range opts.Config.Instances {
			instMgr.Create(&opts.Config.Instances[i])
		}
	}

	instanceCtrl := NewInstanceController(instMgr)
	configCtrl := NewConfigController(opts.ConfigPath)
	securityCtrl := NewSecurityController(keyStoreMgr, rootCertMgr, opts.Config, opts.ConfigPath)
	systemCtrl := NewSystemController()

	instanceCtrl.RegisterRoutes(router)
	configCtrl.RegisterRoutes(router)
	securityCtrl.RegisterRoutes(router)
	systemCtrl.RegisterRoutes(router)

	staticDir := opts.StaticDir
	if staticDir == "" {
		staticDir = "./ui"
	}
	absStaticDir, err := filepath.Abs(staticDir)
	if err != nil {
		absStaticDir = staticDir
	}

	var mcpServer *mcp.Server
	if opts.Config.MCP.Enabled {
		mcpServer = mcp.NewServer(opts.Config)

		keyStoreTool := tools.NewKeyStoreManagerTool(keyStoreMgr, opts.Config, opts.ConfigPath)
		mcpServer.RegisterTool(keyStoreTool)

		certManagerTool := tools.NewCertificateManagerTool(rootCertMgr, opts.Config, opts.ConfigPath)
		mcpServer.RegisterTool(certManagerTool)

		instanceTool := tools.NewInstanceLifecycleTool(instMgr)
		mcpServer.RegisterTool(instanceTool)

		router.GET("/mcp/ws", mcpServer.HandleWebSocket)
		log.Info("MCP服务已启用")
	}

	return &Server{
		router:      router,
		cfg:         opts.Config,
		log:         log,
		keyStoreMgr: keyStoreMgr,
		rootCertMgr: rootCertMgr,
		staticDir:   absStaticDir,
		fileServer:  http.FileServer(http.Dir(absStaticDir)),
		mcpServer:   mcpServer,
	}
}

// ServeHTTP 实现 http.Handler 接口
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.Redirect(w, r, "./ui/", http.StatusFound)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/ui") {
		s.handleUI(w, r)
		return
	}

	s.router.ServeHTTP(w, r)
}

// handleUI 处理 UI 静态文件请求
func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	uiPath := strings.TrimPrefix(r.URL.Path, "/ui")
	if uiPath == "" {
		uiPath = "/"
	}

	originalPath := r.URL.Path
	r.URL.Path = uiPath

	filePath := filepath.Join(s.staticDir, uiPath)
	fi, err := os.Stat(filePath)
	if err == nil && !fi.IsDir() {
		s.fileServer.ServeHTTP(w, r)
		return
	}

	indexPath := filepath.Join(s.staticDir, "index.html")
	if _, err := os.Stat(indexPath); err == nil {
		r.URL.Path = originalPath
		http.ServeFile(w, r, indexPath)
		return
	}

	http.NotFound(w, r)
}

// Start 启动API服务器
func (s *Server) Start(addr string) error {
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.log.Info("API服务器启动: %s", addr)
	return s.httpServer.ListenAndServe()
}

// Stop 停止API服务器
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	s.log.Info("API服务器停止")
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) Handler() http.Handler {
	return s
}
