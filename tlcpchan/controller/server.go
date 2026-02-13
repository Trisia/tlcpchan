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

type Server struct {
	httpServer *http.Server
	router     *Router
	cfg        *config.Config
	log        *logger.Logger
}

type ServerOptions struct {
	Config     *config.Config
	ConfigPath string
	CertDir    string
	Version    string
}

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

	instanceCtrl.RegisterRoutes(router)
	configCtrl.RegisterRoutes(router)
	certCtrl.RegisterRoutes(router)
	systemCtrl.RegisterRoutes(router)

	return &Server{
		router: router,
		cfg:    opts.Config,
		log:    log,
	}
}

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
