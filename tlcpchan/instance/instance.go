package instance

import (
	"fmt"
	"sync"
	"time"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/logger"
	"github.com/Trisia/tlcpchan/proxy"
	"github.com/Trisia/tlcpchan/security"
	"github.com/Trisia/tlcpchan/stats"
)

// Instance 代理实例接口，定义实例的基本操作
type Instance interface {
	Name() string
	Type() InstanceType
	Protocol() string
	Start() error
	Stop() error
	Reload(cfg *config.InstanceConfig) error
	Restart(cfg *config.InstanceConfig) error
	Status() Status
	Stats() *stats.Stats
	Config() *config.InstanceConfig
}

// baseInstance 实例基类，包含所有实例类型的公共属性
type baseInstance struct {
	cfg             *config.InstanceConfig
	instanceType    InstanceType
	status          Status
	stats           *stats.Stats
	keyStoreManager *security.KeyStoreManager
	rootCertManager *security.RootCertManager
	logger          *logger.Logger
	startTime       time.Time
	mu              sync.RWMutex
}

// serverInstance TCP服务端代理实例
type serverInstance struct {
	*baseInstance
	proxy *proxy.ServerProxy
}

// clientInstance TCP客户端代理实例
type clientInstance struct {
	*baseInstance
	proxy *proxy.ClientProxy
}

// httpServerInstance HTTP服务端代理实例
type httpServerInstance struct {
	*baseInstance
	proxy *proxy.HTTPServerProxy
}

// httpClientInstance HTTP客户端代理实例
type httpClientInstance struct {
	*baseInstance
	proxy *proxy.HTTPClientProxy
}

// NewInstance 创建新的代理实例
func NewInstance(cfg *config.InstanceConfig,
	keyStoreMgr *security.KeyStoreManager,
	rootCertMgr *security.RootCertManager,
	log *logger.Logger) (Instance, error) {
	base := &baseInstance{
		cfg:             cfg,
		instanceType:    ParseInstanceType(cfg.Type),
		status:          StatusCreated,
		stats:           &stats.Stats{},
		keyStoreManager: keyStoreMgr,
		rootCertManager: rootCertMgr,
		logger:          log,
	}

	switch base.instanceType {
	case TypeServer:
		p, err := proxy.NewServerProxy(cfg, keyStoreMgr, rootCertMgr)
		if err != nil {
			return nil, err
		}
		return &serverInstance{baseInstance: base, proxy: p}, nil
	case TypeClient:
		p, err := proxy.NewClientProxy(cfg, keyStoreMgr, rootCertMgr)
		if err != nil {
			return nil, err
		}
		return &clientInstance{baseInstance: base, proxy: p}, nil
	case TypeHTTPServer:
		p, err := proxy.NewHTTPServerProxy(cfg, keyStoreMgr, rootCertMgr)
		if err != nil {
			return nil, err
		}
		return &httpServerInstance{baseInstance: base, proxy: p}, nil
	case TypeHTTPClient:
		p, err := proxy.NewHTTPClientProxy(cfg, keyStoreMgr, rootCertMgr)
		if err != nil {
			return nil, err
		}
		return &httpClientInstance{baseInstance: base, proxy: p}, nil
	default:
		return nil, fmt.Errorf("未知的实例类型: %s", cfg.Type)
	}
}

func (i *baseInstance) Name() string {
	return i.cfg.Name
}

func (i *baseInstance) Type() InstanceType {
	return i.instanceType
}

func (i *baseInstance) Protocol() string {
	return i.cfg.Protocol
}

func (i *baseInstance) Status() Status {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.status
}

func (i *baseInstance) Stats() *stats.Stats {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.stats
}

func (i *baseInstance) Config() *config.InstanceConfig {
	return i.cfg
}

func (i *baseInstance) setStatus(status Status) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.status = status
}

func (i *baseInstance) setStartTime() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.startTime = time.Now()
}

func (i *baseInstance) updateStats(s stats.Stats) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.stats = &s
}

func (i *serverInstance) Start() error {
	if err := i.proxy.Start(); err != nil {
		i.setStatus(StatusError)
		return err
	}
	i.setStatus(StatusRunning)
	i.setStartTime()
	return nil
}

func (i *serverInstance) Stop() error {
	if err := i.proxy.Stop(); err != nil {
		return err
	}
	i.setStatus(StatusStopped)
	return nil
}

func (i *serverInstance) Reload(cfg *config.InstanceConfig) error {
	if i.Status() != StatusRunning {
		return fmt.Errorf("实例 %s 未在运行", i.Name())
	}

	if err := i.proxy.Reload(cfg); err == nil {
		i.mu.Lock()
		i.cfg = cfg
		i.mu.Unlock()
		return nil
	}
	return fmt.Errorf("热加载不支持或失败")
}

func (i *serverInstance) Restart(cfg *config.InstanceConfig) error {
	if err := i.proxy.Stop(); err != nil {
		return err
	}
	newProxy, err := proxy.NewServerProxy(cfg, i.keyStoreManager, i.rootCertManager)
	if err != nil {
		i.setStatus(StatusError)
		return err
	}
	i.proxy = newProxy
	i.mu.Lock()
	i.cfg = cfg
	i.mu.Unlock()
	return i.proxy.Start()
}

func (i *serverInstance) Stats() *stats.Stats {
	i.updateStats(i.proxy.Stats())
	return i.baseInstance.Stats()
}

func (i *clientInstance) Start() error {
	if err := i.proxy.Start(); err != nil {
		i.setStatus(StatusError)
		return err
	}
	i.setStatus(StatusRunning)
	i.setStartTime()
	return nil
}

func (i *clientInstance) Stop() error {
	if err := i.proxy.Stop(); err != nil {
		return err
	}
	i.setStatus(StatusStopped)
	return nil
}

func (i *clientInstance) Reload(cfg *config.InstanceConfig) error {
	if i.Status() != StatusRunning {
		return fmt.Errorf("实例 %s 未在运行", i.Name())
	}

	if err := i.proxy.Reload(cfg); err == nil {
		i.mu.Lock()
		i.cfg = cfg
		i.mu.Unlock()
		return nil
	}
	return fmt.Errorf("热加载不支持或失败")
}

func (i *clientInstance) Restart(cfg *config.InstanceConfig) error {
	if err := i.proxy.Stop(); err != nil {
		return err
	}
	newProxy, err := proxy.NewClientProxy(cfg, i.keyStoreManager, i.rootCertManager)
	if err != nil {
		i.setStatus(StatusError)
		return err
	}
	i.proxy = newProxy
	i.mu.Lock()
	i.cfg = cfg
	i.mu.Unlock()
	return i.proxy.Start()
}

func (i *clientInstance) Stats() *stats.Stats {
	i.updateStats(i.proxy.Stats())
	return i.baseInstance.Stats()
}

func (i *httpServerInstance) Start() error {
	if err := i.proxy.Start(); err != nil {
		i.setStatus(StatusError)
		return err
	}
	i.setStatus(StatusRunning)
	i.setStartTime()
	return nil
}

func (i *httpServerInstance) Stop() error {
	if err := i.proxy.Stop(); err != nil {
		return err
	}
	i.setStatus(StatusStopped)
	return nil
}

func (i *httpServerInstance) Reload(cfg *config.InstanceConfig) error {
	if i.Status() != StatusRunning {
		return fmt.Errorf("实例 %s 未在运行", i.Name())
	}

	if err := i.proxy.Reload(cfg); err == nil {
		i.mu.Lock()
		i.cfg = cfg
		i.mu.Unlock()
		return nil
	}
	return fmt.Errorf("热加载不支持或失败")
}

func (i *httpServerInstance) Restart(cfg *config.InstanceConfig) error {
	if err := i.proxy.Stop(); err != nil {
		return err
	}
	newProxy, err := proxy.NewHTTPServerProxy(cfg, i.keyStoreManager, i.rootCertManager)
	if err != nil {
		i.setStatus(StatusError)
		return err
	}
	i.proxy = newProxy
	i.mu.Lock()
	i.cfg = cfg
	i.mu.Unlock()
	return i.proxy.Start()
}

func (i *httpServerInstance) Stats() *stats.Stats {
	i.updateStats(i.proxy.Stats())
	return i.baseInstance.Stats()
}

func (i *httpClientInstance) Start() error {
	if err := i.proxy.Start(); err != nil {
		i.setStatus(StatusError)
		return err
	}
	i.setStatus(StatusRunning)
	i.setStartTime()
	return nil
}

func (i *httpClientInstance) Stop() error {
	if err := i.proxy.Stop(); err != nil {
		return err
	}
	i.setStatus(StatusStopped)
	return nil
}

func (i *httpClientInstance) Reload(cfg *config.InstanceConfig) error {
	if i.Status() != StatusRunning {
		return fmt.Errorf("实例 %s 未在运行", i.Name())
	}

	if err := i.proxy.Reload(cfg); err == nil {
		i.mu.Lock()
		i.cfg = cfg
		i.mu.Unlock()
		return nil
	}
	return fmt.Errorf("热加载不支持或失败")
}

func (i *httpClientInstance) Restart(cfg *config.InstanceConfig) error {
	if err := i.proxy.Stop(); err != nil {
		return err
	}
	newProxy, err := proxy.NewHTTPClientProxy(cfg, i.keyStoreManager, i.rootCertManager)
	if err != nil {
		i.setStatus(StatusError)
		return err
	}
	i.proxy = newProxy
	i.mu.Lock()
	i.cfg = cfg
	i.mu.Unlock()
	return i.proxy.Start()
}

func (i *httpClientInstance) Stats() *stats.Stats {
	i.updateStats(i.proxy.Stats())
	return i.baseInstance.Stats()
}
