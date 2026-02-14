package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/Trisia/tlcpchan/cert"
	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/instance"
	"github.com/Trisia/tlcpchan/logger"
)

// Server API服务器，提供RESTful API接口
type Server struct {
	httpServer *http.Server
	router     *Router
	cfg        *config.Config
	log        *logger.Logger
}

// ServerOptions API服务器配置选项
type ServerOptions struct {
	// Config 全局配置
	Config *config.Config
	// ConfigPath 配置文件路径
	ConfigPath string
	// CertDir 证书目录路径
	CertDir string
	// Version 版本号
	Version string
}

// NewServer 创建新的API服务器
// 参数:
//   - opts: 服务器配置选项
//
// 返回:
//   - *Server: API服务器实例
func NewServer(opts ServerOptions) *Server {
	router := NewRouter()
	router.Use(corsMiddleware)
	router.Use(loggingMiddleware)

	log := logger.Default()
	certMgr := cert.NewManager()
	mgr := instance.NewManager(log, certMgr)
	for i := range opts.Config.Instances {
		mgr.Create(&opts.Config.Instances[i])
	}

	instanceCtrl := NewInstanceController(mgr)
	configCtrl := NewConfigController(opts.Config, opts.ConfigPath)
	certCtrl := NewCertController(opts.CertDir)
	systemCtrl := NewSystemController(opts.Version)
	healthCtrl := NewHealthController(mgr, certMgr)

	instanceCtrl.RegisterRoutes(router)
	configCtrl.RegisterRoutes(router)
	certCtrl.RegisterRoutes(router)
	systemCtrl.RegisterRoutes(router)
	healthCtrl.RegisterRoutes(router)

	return &Server{
		router: router,
		cfg:    opts.Config,
		log:    log,
	}
}

// Start 启动API服务器
// 参数:
//   - addr: 监听地址，格式: "host:port" 或 ":port"
//
// 返回:
//   - error: 启动失败时返回错误
//
// 注意: 该方法会阻塞直到服务器停止
func (s *Server) Start(addr string) error {
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.log.Info("API服务器启动: %s", addr)
	return s.httpServer.ListenAndServe()
}

// Stop 停止API服务器
// 参数:
//   - ctx: 上下文，用于控制关闭超时
//
// 返回:
//   - error: 关闭失败时返回错误
//
// 注意: 该方法会优雅关闭，等待现有请求完成
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	s.log.Info("API服务器停止")
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) Handler() http.Handler {
	return s.router
}
