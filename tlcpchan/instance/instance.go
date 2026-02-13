package instance

import (
	"fmt"
	"sync"
	"time"

	"github.com/Trisia/tlcpchan/cert"
	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/logger"
	"github.com/Trisia/tlcpchan/proxy"
	"github.com/Trisia/tlcpchan/stats"
)

type Instance interface {
	Name() string
	Type() InstanceType
	Protocol() string
	Start() error
	Stop() error
	Reload(cfg *config.InstanceConfig) error
	Status() Status
	Stats() *stats.Stats
	Config() *config.InstanceConfig
}

type baseInstance struct {
	cfg          *config.InstanceConfig
	instanceType InstanceType
	status       Status
	stats        *stats.Stats
	certManager  *cert.Manager
	logger       *logger.Logger
	startTime    time.Time
	mu           sync.RWMutex
}

type serverInstance struct {
	*baseInstance
	proxy *proxy.ServerProxy
}

type clientInstance struct {
	*baseInstance
	proxy *proxy.ClientProxy
}

type httpServerInstance struct {
	*baseInstance
	proxy *proxy.HTTPServerProxy
}

type httpClientInstance struct {
	*baseInstance
	proxy *proxy.HTTPClientProxy
}

func NewInstance(cfg *config.InstanceConfig, certManager *cert.Manager, log *logger.Logger) (Instance, error) {
	base := &baseInstance{
		cfg:          cfg,
		instanceType: ParseInstanceType(cfg.Type),
		status:       StatusCreated,
		stats:        &stats.Stats{},
		certManager:  certManager,
		logger:       log,
	}

	switch base.instanceType {
	case TypeServer:
		p, err := proxy.NewServerProxy(cfg, certManager)
		if err != nil {
			return nil, err
		}
		return &serverInstance{baseInstance: base, proxy: p}, nil
	case TypeClient:
		p, err := proxy.NewClientProxy(cfg, certManager)
		if err != nil {
			return nil, err
		}
		return &clientInstance{baseInstance: base, proxy: p}, nil
	case TypeHTTPServer:
		p, err := proxy.NewHTTPServerProxy(cfg, certManager)
		if err != nil {
			return nil, err
		}
		return &httpServerInstance{baseInstance: base, proxy: p}, nil
	case TypeHTTPClient:
		p, err := proxy.NewHTTPClientProxy(cfg, certManager)
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
	if err := i.proxy.Stop(); err != nil {
		return err
	}
	newProxy, err := proxy.NewServerProxy(cfg, i.certManager)
	if err != nil {
		i.setStatus(StatusError)
		return err
	}
	i.proxy = newProxy
	i.cfg = cfg
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
	if err := i.proxy.Stop(); err != nil {
		return err
	}
	newProxy, err := proxy.NewClientProxy(cfg, i.certManager)
	if err != nil {
		i.setStatus(StatusError)
		return err
	}
	i.proxy = newProxy
	i.cfg = cfg
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
	if err := i.proxy.Stop(); err != nil {
		return err
	}
	newProxy, err := proxy.NewHTTPServerProxy(cfg, i.certManager)
	if err != nil {
		i.setStatus(StatusError)
		return err
	}
	i.proxy = newProxy
	i.cfg = cfg
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
	if err := i.proxy.Stop(); err != nil {
		return err
	}
	newProxy, err := proxy.NewHTTPClientProxy(cfg, i.certManager)
	if err != nil {
		i.setStatus(StatusError)
		return err
	}
	i.proxy = newProxy
	i.cfg = cfg
	return i.proxy.Start()
}

func (i *httpClientInstance) Stats() *stats.Stats {
	i.updateStats(i.proxy.Stats())
	return i.baseInstance.Stats()
}
