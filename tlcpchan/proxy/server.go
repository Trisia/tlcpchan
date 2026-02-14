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

// ServerProxy 服务端代理，接收TLCP/TLS连接并转发到目标服务
type ServerProxy struct {
	// cfg 实例配置
	cfg *config.InstanceConfig
	// adapter 协议适配器
	adapter *TLCPAdapter
	// handler 连接处理器
	handler *ConnHandler
	// listener TCP监听器（已包装协议）
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
}

// NewServerProxy 创建新的服务端代理实例
// 参数:
//   - cfg: 实例配置，不能为 nil
//   - certManager: 证书管理器，用于加载TLCP/TLS证书
//
// 返回:
//   - *ServerProxy: 服务端代理实例
//   - error: 创建协议适配器失败时返回错误
func NewServerProxy(cfg *config.InstanceConfig, certManager *cert.Manager) (*ServerProxy, error) {
	adapter, err := NewTLCPAdapter(cfg, certManager)
	if err != nil {
		return nil, fmt.Errorf("创建协议适配器失败: %w", err)
	}

	return &ServerProxy{
		cfg:          cfg,
		adapter:      adapter,
		handler:      NewConnHandler(),
		certManager:  certManager,
		stats:        stats.DefaultCollector(),
		logger:       logger.Default(),
		shutdownChan: make(chan struct{}),
	}, nil
}

// Start 启动服务端代理
// 返回:
//   - error: 代理已在运行或监听端口失败时返回错误
//
// 注意: 该方法会启动后台goroutine接受连接，调用Stop()停止
func (p *ServerProxy) Start() error {
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

	p.listener = p.adapter.WrapServerListener(listener)
	p.running = true
	p.mu.Unlock()

	p.logger.Info("服务端代理启动: %s -> %s, 协议: %s", p.cfg.Listen, p.cfg.Target, p.cfg.Protocol)

	p.wg.Add(1)
	go p.acceptLoop()

	return nil
}

// acceptLoop 接受连接循环，持续监听新连接并异步处理
// 注意: 该方法在独立goroutine中运行，通过shutdownChan控制退出
func (p *ServerProxy) acceptLoop() {
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
//   - clientConn: 客户端连接，可能已包装协议层
//
// 注意: 该方法负责协议握手、连接目标服务、双向数据转发
func (p *ServerProxy) handleConnection(clientConn net.Conn) {
	defer p.wg.Done()
	defer p.stats.DecrementConnections()
	defer clientConn.Close()
	defer func() {
		if r := recover(); r != nil {
			p.logger.Error("连接处理panic: %v", r)
			p.stats.IncrementErrors()
		}
	}()

	start := time.Now()

	if autoConn, ok := clientConn.(*autoProtocolConn); ok {
		if err := autoConn.Handshake(); err != nil {
			p.logger.Error("协议握手失败: %v", err)
			p.stats.IncrementErrors()
			return
		}
	}

	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}
	targetConn, err := dialer.Dial("tcp", p.cfg.Target)
	if err != nil {
		p.logger.Error("连接目标服务失败 %s: %v", p.cfg.Target, err)
		p.stats.IncrementErrors()
		return
	}
	defer targetConn.Close()

	p.logger.Debug("连接建立: %s <-> %s", clientConn.RemoteAddr(), p.cfg.Target)

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

// Stop 停止服务端代理
// 返回:
//   - error: 停止失败时返回错误（当前实现始终返回nil）
//
// 注意: 该方法会等待所有连接处理完成后再返回
func (p *ServerProxy) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return nil
	}

	p.logger.Info("停止服务端代理: %s", p.cfg.Name)

	close(p.shutdownChan)

	if p.listener != nil {
		p.listener.Close()
	}

	p.wg.Wait()

	p.running = false
	p.shutdownChan = make(chan struct{})

	return nil
}

// Restart 重启服务端代理
// 返回:
//   - error: 停止或启动失败时返回错误
func (p *ServerProxy) Restart() error {
	if err := p.Stop(); err != nil {
		return err
	}
	return p.Start()
}

func (p *ServerProxy) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}

func (p *ServerProxy) Name() string {
	return p.cfg.Name
}

func (p *ServerProxy) Config() *config.InstanceConfig {
	return p.cfg
}

func (p *ServerProxy) Stats() stats.Stats {
	return p.stats.GetStats()
}
