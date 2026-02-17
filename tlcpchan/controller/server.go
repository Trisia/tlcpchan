package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/instance"
	"github.com/Trisia/tlcpchan/logger"
	"github.com/Trisia/tlcpchan/security"
)

// Server API服务器，提供RESTful API接口
type Server struct {
	httpServer  *http.Server
	router      *Router
	cfg         *config.Config
	log         *logger.Logger
	keyStoreMgr *security.KeyStoreManager
	rootCertMgr *security.RootCertManager
}

// ServerOptions API服务器配置选项
type ServerOptions struct {
	Config          *config.Config
	ConfigPath      string
	Version         string
	KeyStoreManager *security.KeyStoreManager
	RootCertManager *security.RootCertManager
}

// NewServer 创建新的API服务器
func NewServer(opts ServerOptions) *Server {
	router := NewRouter()
	router.Use(corsMiddleware)
	router.Use(loggingMiddleware)

	log := logger.Default()

	var keyStoreMgr *security.KeyStoreManager
	var rootCertMgr *security.RootCertManager

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

	mgr := instance.NewManager(log, keyStoreMgr, rootCertMgr)
	for i := range opts.Config.Instances {
		mgr.Create(&opts.Config.Instances[i])
	}

	instanceCtrl := NewInstanceController(mgr)
	configCtrl := NewConfigController(opts.ConfigPath)
	securityCtrl := NewSecurityController(keyStoreMgr, rootCertMgr, opts.Config, opts.ConfigPath)
	systemCtrl := NewSystemController(opts.Version)

	instanceCtrl.RegisterRoutes(router)
	configCtrl.RegisterRoutes(router)
	securityCtrl.RegisterRoutes(router)
	systemCtrl.RegisterRoutes(router)

	return &Server{
		router:      router,
		cfg:         opts.Config,
		log:         log,
		keyStoreMgr: keyStoreMgr,
		rootCertMgr: rootCertMgr,
	}
}

// Start 启动API服务器
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
