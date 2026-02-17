package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/controller"
	"github.com/Trisia/tlcpchan/initialization"
	"github.com/Trisia/tlcpchan/instance"
	"github.com/Trisia/tlcpchan/logger"
	"github.com/Trisia/tlcpchan/security"
	"github.com/Trisia/tlcpchan/security/keystore"
)

var (
	configFile  = flag.String("config", "", "配置文件路径")
	workDirFlag = flag.String("workdir", "", "工作目录路径")
	showVersion = flag.Bool("version", false, "显示版本信息")
	startUI     = flag.Bool("ui", false, "启动 UI 服务")
	version     = "1.0.0"
)

func init() {
	flag.StringVar(configFile, "c", "", "配置文件路径(缩写)")
	flag.StringVar(workDirFlag, "w", "", "工作目录路径(缩写)")
	flag.BoolVar(showVersion, "v", false, "显示版本信息(缩写)")
}

// getWorkDir 获取工作目录
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
func ensureWorkDir(dir string) string {
	dirs := []string{
		dir,
		filepath.Join(dir, "logs"),
		filepath.Join(dir, "keystores"),
		filepath.Join(dir, "rootcerts"),
	}
	for _, d := range dirs {
		os.MkdirAll(d, 0755)
	}
	return dir
}

// findUIExecutable 查找 UI 可执行文件
// 首先尝试在当前可执行文件同目录查找，然后尝试系统 PATH
func findUIExecutable() (string, error) {
	uiName := getUIExecutableName()

	exe, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exe)
		uiPath := filepath.Join(exeDir, uiName)
		if _, err := os.Stat(uiPath); err == nil {
			return uiPath, nil
		}
	}

	return exec.LookPath(uiName)
}

// startUIServer 启动 UI 服务子进程
// 参数:
//   - uiPath: UI 可执行文件路径
//   - workDir: 工作目录
// 返回值:
//   - *exec.Cmd: 子进程对象
//   - error: 错误信息
func startUIServer(uiPath, workDir string) (*exec.Cmd, error) {
	cmd := exec.Command(uiPath, "-api", "http://localhost:30080")
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("启动 UI 服务失败: %w", err)
	}

	return cmd, nil
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("tlcpchan version %s\n", version)
		os.Exit(0)
	}

	wd := getWorkDir(*workDirFlag)
	wd = ensureWorkDir(wd)

	configPath := *configFile
	if configPath == "" {
		configPath = filepath.Join(wd, "config.yaml")
	}

	var uiCmd *exec.Cmd

	cfg, err := config.Load(configPath)
	if err != nil {
		cfg = config.Default()
		cfg.WorkDir = wd
		logger.Info("使用默认配置: %v", err)
	} else {
		cfg.WorkDir = wd
	}
	config.Init(cfg)

	// 检查并执行初始化
	initMgr := initialization.NewManager(cfg, configPath, wd)
	if !initMgr.CheckInitialized() {
		logger.Info("检测到首次启动，开始初始化...")
		if err := initMgr.Initialize(); err != nil {
			logger.Fatal("初始化失败: %v", err)
		}
		logger.Info("初始化完成")
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

	keyStoreMgr := security.NewKeyStoreManager()

	// 从配置加载 keystores
	ksEntries := make([]keystore.ConfigEntry, 0, len(cfg.KeyStores))
	for _, ksCfg := range cfg.KeyStores {
		ksEntries = append(ksEntries, keystore.ConfigEntry{
			Name:   ksCfg.Name,
			Type:   ksCfg.Type,
			Params: ksCfg.Params,
		})
	}
	if err := keyStoreMgr.LoadFromConfigs(ksEntries); err != nil {
		logger.Warn("加载 keystores 失败: %v", err)
	} else {
		logger.Info("已加载 %d 个 keystores", len(cfg.KeyStores))
	}

	rootCertMgr := security.NewRootCertManager(cfg.GetRootCertDir())
	if err := rootCertMgr.Initialize(); err != nil {
		logger.Warn("初始化根证书管理器失败: %v", err)
	}

	instMgr := instance.NewManager(logger.Default(), keyStoreMgr, rootCertMgr)

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
		Config:          cfg,
		ConfigPath:      configPath,
		Version:         version,
		KeyStoreManager: keyStoreMgr,
		RootCertManager: rootCertMgr,
	}
	apiServer := controller.NewServer(opts)

	go func() {
		if err := apiServer.Start(cfg.Server.API.Address); err != nil {
			logger.Fatal("API服务启动失败: %v", err)
		}
	}()

	if *startUI {
		uiPath, err := findUIExecutable()
		if err != nil {
			logger.Error("未找到 UI 程序，跳过 UI 启动: %v", err)
			logger.Error("UI 程序应命名为 %q，可放在当前程序同目录或系统 PATH 中", getUIExecutableName())
		} else {
			uiCmd, err = startUIServer(uiPath, wd)
			if err != nil {
				logger.Error("启动 UI 服务失败: %v", err)
			} else {
				logger.Info("UI 服务已启动 (PID: %d)", uiCmd.Process.Pid)
			}
		}
	}

	logger.Info("tlcpchan %s 启动完成", version)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Info("收到信号 %v，开始关闭...", sig)

	if uiCmd != nil && uiCmd.Process != nil {
		logger.Info("正在关闭 UI 服务...")
		if err := uiCmd.Process.Signal(syscall.SIGTERM); err != nil {
			logger.Warn("发送 SIGTERM 到 UI 服务失败: %v，尝试强制关闭", err)
			uiCmd.Process.Kill()
		} else {
			done := make(chan error, 1)
			go func() {
				done <- uiCmd.Wait()
			}()
			select {
			case <-done:
				logger.Info("UI 服务已正常关闭")
			case <-time.After(5 * time.Second):
				logger.Warn("UI 服务关闭超时，强制关闭")
				uiCmd.Process.Kill()
				uiCmd.Wait()
				logger.Info("UI 服务已强制关闭")
			}
		}
	}

	instMgr.StopAll()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	apiServer.Stop(ctx)

	logger.Info("tlcpchan 已关闭")
}
