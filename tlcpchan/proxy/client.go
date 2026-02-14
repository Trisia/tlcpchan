package proxy

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Trisia/tlcpchan/cert"
	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/logger"
	"github.com/Trisia/tlcpchan/stats"
)

// protocolCacheEntry 协议缓存条目，记录目标地址的协议检测结果
type protocolCacheEntry struct {
	// protocol 检测到的协议类型
	protocol ProtocolType
	// detected 检测时间
	detected time.Time
}

// ClientProxy 客户端代理，接收普通TCP连接并以TLCP/TLS连接目标服务
type ClientProxy struct {
	// cfg 实例配置
	cfg *config.InstanceConfig
	// adapter 协议适配器
	adapter *TLCPAdapter
	// handler 连接处理器
	handler *ConnHandler
	// listener TCP监听器
	listener net.Listener
	// certManager 证书管理器
	certManager *cert.Manager
	// stats 统计收集器
	stats        *stats.Collector
	logger       *logger.Logger
	shutdownChan chan struct{}
	wg           sync.WaitGroup
	mu           sync.Mutex
	running      bool

	// protocolCache 协议检测缓存，避免重复检测同一目标
	protocolCache map[string]protocolCacheEntry
	cacheMu       sync.RWMutex
	// cacheTTL 缓存有效期，默认5分钟
	cacheTTL time.Duration
}

// NewClientProxy 创建新的客户端代理实例
// 参数:
//   - cfg: 实例配置，不能为 nil
//   - certManager: 证书管理器，用于加载TLCP/TLS证书
//
// 返回:
//   - *ClientProxy: 客户端代理实例
//   - error: 创建协议适配器失败时返回错误
func NewClientProxy(cfg *config.InstanceConfig, certManager *cert.Manager) (*ClientProxy, error) {
	// 验证客户端配置
	if err := ValidateClientConfig(cfg); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	adapter, err := NewTLCPAdapter(cfg, certManager)
	if err != nil {
		return nil, fmt.Errorf("创建协议适配器失败: %w", err)
	}

	return &ClientProxy{
		cfg:           cfg,
		adapter:       adapter,
		handler:       NewConnHandler(),
		certManager:   certManager,
		stats:         stats.DefaultCollector(),
		logger:        logger.Default(),
		shutdownChan:  make(chan struct{}),
		protocolCache: make(map[string]protocolCacheEntry),
		cacheTTL:      5 * time.Minute,
	}, nil
}

// Start 启动客户端代理
// 返回:
//   - error: 代理已在运行或监听端口失败时返回错误
//
// 注意: 该方法会启动后台goroutine接受连接，调用Stop()停止
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

// acceptLoop 接受连接循环，持续监听新连接并异步处理
// 注意: 该方法在独立goroutine中运行，通过shutdownChan控制退出
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

// handleConnection 处理单个客户端连接
// 参数:
//   - clientConn: 客户端连接（普通TCP连接）
//
// 注意: 该方法负责协议检测、连接目标服务、双向数据转发
func (p *ClientProxy) handleConnection(clientConn net.Conn) {
	defer p.wg.Done()
	defer p.stats.DecrementConnections()
	defer clientConn.Close()

	start := time.Now()

	protocol := p.getProtocol()
	if protocol == ProtocolAuto {
		protocol = p.detectAndCacheProtocol()
	}

	targetConn, err := p.adapter.DialWithProtocol("tcp", p.cfg.Target, protocol)
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

// getProtocol 获取当前使用的协议类型
// 返回:
//   - ProtocolType: 协议类型，若配置为auto且缓存有效则返回缓存值，否则返回ProtocolAuto
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

// detectAndCacheProtocol 检测目标服务支持的协议并缓存结果
// 返回:
//   - ProtocolType: 检测到的协议类型（ProtocolTLCP或ProtocolTLS）
//
// 注意: 检测结果会缓存5分钟，避免重复检测开销
func (p *ClientProxy) detectAndCacheProtocol() ProtocolType {
	conn, err := p.adapter.DialTLCP("tcp", p.cfg.Target)
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

// Stop 停止客户端代理
// 返回:
//   - error: 停止失败时返回错误（当前实现始终返回nil）
//
// 注意: 该方法会等待所有连接处理完成后再返回
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

// Restart 重启客户端代理
// 返回:
//   - error: 停止或启动失败时返回错误
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

// ClearProtocolCache 清除协议检测缓存
// 注意: 当目标服务协议变更时应调用此方法
func (p *ClientProxy) ClearProtocolCache() {
	p.cacheMu.Lock()
	defer p.cacheMu.Unlock()
	p.protocolCache = make(map[string]protocolCacheEntry)
}
