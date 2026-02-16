package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Trisia/tlcpchan/cert"
	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/controller"
	"github.com/Trisia/tlcpchan/instance"
	"github.com/Trisia/tlcpchan/logger"
)

var (
	configFile  = flag.String("config", "", "配置文件路径")
	workDirFlag = flag.String("workdir", "", "工作目录路径")
	showVersion = flag.Bool("version", false, "显示版本信息")
	version     = "1.0.0"
)

func init() {
	flag.StringVar(configFile, "c", "", "配置文件路径(缩写)")
	flag.StringVar(workDirFlag, "w", "", "工作目录路径(缩写)")
	flag.BoolVar(showVersion, "v", false, "显示版本信息(缩写)")
}

// getWorkDir 获取工作目录
// 如果传入了工作目录参数则使用该参数，否则使用可执行文件所在目录
// 参数:
//   - customDir: 自定义工作目录，为空则使用默认值
//
// 返回:
//   - string: 工作目录路径
func getWorkDir(customDir string) string {
	if customDir != "" {
		return customDir
	}
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

// ensureWorkDir 确保工作目录及其子目录存在
// 参数:
//   - dir: 工作目录路径
//
// 返回:
//   - string: 工作目录路径
func ensureWorkDir(dir string) string {
	dirs := []string{
		dir,
		filepath.Join(dir, "trusted"),
		filepath.Join(dir, "logs"),
		filepath.Join(dir, "config"),
		filepath.Join(dir, "keys"),
	}
	for _, d := range dirs {
		os.MkdirAll(d, 0755)
	}
	return dir
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("tlcpchan version %s\n", version)
		os.Exit(0)
	}

	// 初始化工作目录
	wd := getWorkDir(*workDirFlag)
	wd = ensureWorkDir(wd)

	// 确定配置文件路径
	configPath := *configFile
	if configPath == "" {
		configPath = filepath.Join(wd, "config", "config.yaml")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		cfg = config.Default()
		cfg.WorkDir = wd
		logger.Info("使用默认配置: %v", err)
	} else {
		cfg.WorkDir = wd
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

	certMgr := cert.NewManagerWithCertDir(cfg.GetCertDir())
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
		Config:      cfg,
		ConfigPath:  configPath,
		KeyStoreDir: cfg.GetKeyStoreDir(),
		TrustedDir:  cfg.GetTrustedDir(),
		Version:     version,
	}
	apiServer := controller.NewServer(opts)

	go func() {
		if err := apiServer.Start(cfg.Server.API.Address); err != nil {
			logger.Fatal("API服务启动失败: %v", err)
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
