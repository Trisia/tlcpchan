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

// Instance 代理实例接口，定义实例的基本操作
type Instance interface {
	// Name 返回实例名称
	Name() string
	// Type 返回实例类型
	Type() InstanceType
	// Protocol 返回协议类型
	Protocol() string
	// Start 启动实例
	Start() error
	// Stop 停止实例
	Stop() error
	// Reload 热重载配置
	Reload(cfg *config.InstanceConfig) error
	// Status 返回当前状态
	Status() Status
	// Stats 返回统计信息
	Stats() *stats.Stats
	// Config 返回配置
	Config() *config.InstanceConfig
}

// baseInstance 实例基类，包含所有实例类型的公共属性
type baseInstance struct {
	// cfg 实例配置
	cfg *config.InstanceConfig
	// instanceType 实例类型
	instanceType InstanceType
	// status 运行状态
	status Status
	// stats 统计信息
	stats *stats.Stats
	// certManager 证书管理器
	certManager *cert.Manager
	logger      *logger.Logger
	// startTime 启动时间
	startTime time.Time
	mu        sync.RWMutex
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
// 参数:
//   - cfg: 实例配置
//   - certManager: 证书管理器
//   - log: 日志记录器
//
// 返回:
//   - Instance: 代理实例
//   - error: 创建失败时返回错误
//
// 注意: 根据cfg.Type自动创建对应类型的实例
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
