package proxy

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/logger"
	"github.com/Trisia/tlcpchan/security"
	"github.com/Trisia/tlcpchan/stats"
)

type protocolCacheEntry struct {
	protocol ProtocolType
	detected time.Time
}

type ClientProxy struct {
	cfg             *config.InstanceConfig
	adapter         *TLCPAdapter
	handler         *ConnHandler
	listener        net.Listener
	keyStoreManager *security.KeyStoreManager
	rootCertManager *security.RootCertManager
	stats           *stats.Collector
	logger          *logger.Logger
	shutdownChan    chan struct{}
	wg              sync.WaitGroup
	mu              sync.Mutex
	running         bool

	protocolCache map[string]protocolCacheEntry
	cacheMu       sync.RWMutex
	cacheTTL      time.Duration
}

func NewClientProxy(cfg *config.InstanceConfig,
	keyStoreMgr *security.KeyStoreManager,
	rootCertMgr *security.RootCertManager) (*ClientProxy, error) {
	adapter, err := NewTLCPAdapter(keyStoreMgr, rootCertMgr)
	if err != nil {
		return nil, fmt.Errorf("创建协议适配器失败: %w", err)
	}

	proxy := &ClientProxy{
		cfg:             cfg,
		adapter:         adapter,
		handler:         NewConnHandler(stats.DefaultCollector()),
		keyStoreManager: keyStoreMgr,
		rootCertManager: rootCertMgr,
		stats:           stats.DefaultCollector(),
		logger:          logger.Default(),
		shutdownChan:    make(chan struct{}),
		protocolCache:   make(map[string]protocolCacheEntry),
		cacheTTL:        5 * time.Minute,
	}

	if err := adapter.ReloadConfig(cfg); err != nil {
		return nil, fmt.Errorf("初始化配置失败: %w", err)
	}

	return proxy, nil
}

func (p *ClientProxy) Adapter() *TLCPAdapter {
	return p.adapter
}

func (p *ClientProxy) Start() error {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return fmt.Errorf("代理服务已在运行")
	}

	listener, err := net.Listen("tcp", p.cfg.Listen)
	if err != nil {
		p.mu.Unlock()
		return fmt.Errorf("监听失败 %s: %w", p.cfg.Listen, err)
	}

	p.listener = listener
	p.running = true
	p.mu.Unlock()

	p.logger.Info("客户端代理启动: %s -> %s, 协议: %s", p.cfg.Listen, p.cfg.Target, p.cfg.Protocol)

	p.wg.Add(1)
	go p.acceptLoop()

	return nil
}

func (p *ClientProxy) acceptLoop() {
	defer p.wg.Done()

	for {
		select {
		case <-p.shutdownChan:
			return
		default:
		}

		conn, err := p.listener.Accept()
		if err != nil {
			select {
			case <-p.shutdownChan:
				return
			default:
				p.logger.Error("接受连接失败: %v", err)
				continue
			}
		}

		p.stats.IncrementConnections()

		p.wg.Add(1)
		go p.handleConnection(conn)
	}
}

func (p *ClientProxy) handleConnection(clientConn net.Conn) {
	defer p.wg.Done()
	defer p.stats.DecrementConnections()
	defer clientConn.Close()

	start := time.Now()

	protocol := p.getProtocol()
	if protocol == ProtocolAuto {
		protocol = p.detectAndCacheProtocol()
	}

	targetConn, err := p.adapter.DialWithProtocol("tcp", p.cfg.Target, protocol, p.cfg)
	if err != nil {
		p.logger.Error("连接目标服务失败 %s: %v", p.cfg.Target, err)
		p.stats.IncrementErrors()
		return
	}
	defer targetConn.Close()

	p.logger.Debug("连接建立: %s -> %s (%s)", clientConn.RemoteAddr(), p.cfg.Target, protocol)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	received, sent, err := p.handler.Pipe(ctx, clientConn, targetConn)
	if err != nil {
		p.logger.Debug("连接关闭: %v", err)
	}

	latency := time.Since(start)
	p.stats.RecordLatency(latency)

	p.logger.Debug("连接结束: 收发 %d/%d 字节, 耗时 %v", received, sent, latency)
}

func (p *ClientProxy) getProtocol() ProtocolType {
	if p.adapter.Protocol() != ProtocolAuto {
		return p.adapter.Protocol()
	}

	p.cacheMu.RLock()
	entry, ok := p.protocolCache[p.cfg.Target]
	p.cacheMu.RUnlock()

	if ok && time.Since(entry.detected) < p.cacheTTL {
		return entry.protocol
	}

	return ProtocolAuto
}

func (p *ClientProxy) detectAndCacheProtocol() ProtocolType {
	conn, err := p.adapter.DialTLCP("tcp", p.cfg.Target, p.cfg)
	if err == nil {
		conn.Close()
		p.cacheMu.Lock()
		p.protocolCache[p.cfg.Target] = protocolCacheEntry{
			protocol: ProtocolTLCP,
			detected: time.Now(),
		}
		p.cacheMu.Unlock()
		return ProtocolTLCP
	}

	p.logger.Debug("TLCP连接检测失败，使用TLS: %v", err)

	p.cacheMu.Lock()
	p.protocolCache[p.cfg.Target] = protocolCacheEntry{
		protocol: ProtocolTLS,
		detected: time.Now(),
	}
	p.cacheMu.Unlock()

	return ProtocolTLS
}

func (p *ClientProxy) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return nil
	}

	p.logger.Info("停止客户端代理: %s", p.cfg.Name)

	close(p.shutdownChan)

	if p.listener != nil {
		p.listener.Close()
	}

	p.wg.Wait()

	p.running = false
	p.shutdownChan = make(chan struct{})

	return nil
}

func (p *ClientProxy) Restart() error {
	if err := p.Stop(); err != nil {
		return err
	}
	return p.Start()
}

func (p *ClientProxy) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}

func (p *ClientProxy) Name() string {
	return p.cfg.Name
}

func (p *ClientProxy) Config() *config.InstanceConfig {
	return p.cfg
}

func (p *ClientProxy) Stats() stats.Stats {
	return p.stats.GetStats()
}

func (p *ClientProxy) Reload(cfg *config.InstanceConfig) error {
	p.mu.Lock()
	oldCfg := p.cfg
	p.cfg = cfg
	p.mu.Unlock()

	if err := p.adapter.ReloadConfig(cfg); err != nil {
		p.mu.Lock()
		p.cfg = oldCfg
		p.mu.Unlock()
		return err
	}

	p.ClearProtocolCache()
	p.logger.Info("客户端代理配置热重载成功: %s", p.cfg.Name)
	return nil
}

func (p *ClientProxy) ClearProtocolCache() {
	p.cacheMu.Lock()
	defer p.cacheMu.Unlock()
	p.protocolCache = make(map[string]protocolCacheEntry)
}
