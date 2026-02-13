package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Trisia/tlcpchan/cert"
	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/controller"
	"github.com/Trisia/tlcpchan/instance"
	"github.com/Trisia/tlcpchan/logger"
)

var (
	configFile  = flag.String("config", "./config/config.yaml", "配置文件路径")
	showVersion = flag.Bool("version", false, "显示版本信息")
	version     = "1.0.0"
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("tlcpchan version %s\n", version)
		os.Exit(0)
	}

	cfg, err := config.Load(*configFile)
	if err != nil {
		cfg = config.Default()
		logger.Info("使用默认配置: %v", err)
	}

	if cfg.Server.Log != nil {
		logCfg := logger.LogConfig{
			Level:      cfg.Server.Log.Level,
			File:       cfg.Server.Log.File,
			MaxSize:    cfg.Server.Log.MaxSize,
			MaxBackups: cfg.Server.Log.MaxBackups,
			MaxAge:     cfg.Server.Log.MaxAge,
			Compress:   cfg.Server.Log.Compress,
			Enabled:    cfg.Server.Log.Enabled,
		}
		if err := logger.InitDefault(logCfg); err != nil {
			logger.Warn("初始化日志失败: %v", err)
		}
	}

	certMgr := cert.NewManager()
	instMgr := instance.NewManager(logger.Default(), certMgr)

	for i := range cfg.Instances {
		inst := &cfg.Instances[i]
		if _, err := instMgr.Create(inst); err != nil {
			logger.Error("创建实例 %s 失败: %v", inst.Name, err)
		}
	}

	errors := instMgr.StartAll()
	for _, err := range errors {
		logger.Error("启动实例失败: %v", err)
	}

	opts := controller.ServerOptions{
		Config:     cfg,
		ConfigPath: *configFile,
		CertDir:    "./certs",
		Version:    version,
	}
	apiServer := controller.NewServer(opts)

	go func() {
		if err := apiServer.Start(cfg.Server.API.Address); err != nil {
			logger.Error("API服务启动失败: %v", err)
		}
	}()

	logger.Info("tlcpchan %s 启动完成", version)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Info("收到信号 %v，开始关闭...", sig)

	instMgr.StopAll()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	apiServer.Stop(ctx)

	logger.Info("tlcpchan 已关闭")
}
