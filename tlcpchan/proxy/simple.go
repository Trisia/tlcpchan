package proxy

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/logger"
	"github.com/Trisia/tlcpchan/stats"
)

// ServerProxy 服务端代理（简化版）
type ServerProxy struct {
	cfg          *config.InstanceConfig
	listener     net.Listener
	stats        *stats.Collector
	logger       *logger.Logger
	shutdownChan chan struct{}
	wg           sync.WaitGroup
	mu           sync.Mutex
	running      bool
}

// NewServerProxy 创建服务端代理
func NewServerProxy(cfg *config.InstanceConfig) (*ServerProxy, error) {
	return &ServerProxy{
		cfg:          cfg,
		stats:        stats.DefaultCollector(),
		logger:       logger.Default(),
		shutdownChan: make(chan struct{}),
	}, nil
}

func (p *ServerProxy) Start() error {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return fmt.Errorf("代理已在运行")
	}

	listener, err := net.Listen("tcp", p.cfg.Listen)
	if err != nil {
		p.mu.Unlock()
		return fmt.Errorf("监听失败: %w", err)
	}

	p.listener = listener
	p.running = true
	p.mu.Unlock()

	p.wg.Add(1)
	go p.acceptLoop()

	p.logger.Info("服务端代理启动: %s -> %s", p.cfg.Listen, p.cfg.Target)
	return nil
}

func (p *ServerProxy) Stop() error {
	p.mu.Lock()
	if !p.running {
		p.mu.Unlock()
		return nil
	}

	close(p.shutdownChan)
	if p.listener != nil {
		p.listener.Close()
	}
	p.running = false
	p.mu.Unlock()

	p.wg.Wait()
	p.logger.Info("服务端代理停止: %s", p.cfg.Name)
	return nil
}

func (p *ServerProxy) Reload(cfg *config.InstanceConfig) error {
	return fmt.Errorf("热重载暂不支持")
}

func (p *ServerProxy) Stats() stats.Stats {
	return p.stats.GetStats()
}

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

		p.wg.Add(1)
		go p.handleConn(conn)
	}
}

func (p *ServerProxy) handleConn(clientConn net.Conn) {
	defer p.wg.Done()
	defer clientConn.Close()

	targetConn, err := net.DialTimeout("tcp", p.cfg.Target, 10*time.Second)
	if err != nil {
		p.logger.Error("连接目标失败: %v", err)
		return
	}
	defer targetConn.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		buf := make([]byte, 32*1024)
		for {
			n, err := clientConn.Read(buf)
			if n > 0 {
				if _, werr := targetConn.Write(buf[:n]); werr != nil {
					break
				}
			}
			if err != nil {
				break
			}
		}
	}()

	go func() {
		defer wg.Done()
		buf := make([]byte, 32*1024)
		for {
			n, err := targetConn.Read(buf)
			if n > 0 {
				if _, werr := clientConn.Write(buf[:n]); werr != nil {
					break
				}
			}
			if err != nil {
				break
			}
		}
	}()

	wg.Wait()
}

// ClientProxy 客户端代理（简化版）
type ClientProxy struct {
	cfg          *config.InstanceConfig
	listener     net.Listener
	stats        *stats.Collector
	logger       *logger.Logger
	shutdownChan chan struct{}
	wg           sync.WaitGroup
	mu           sync.Mutex
	running      bool
}

// NewClientProxy 创建客户端代理
func NewClientProxy(cfg *config.InstanceConfig) (*ClientProxy, error) {
	return &ClientProxy{
		cfg:          cfg,
		stats:        stats.DefaultCollector(),
		logger:       logger.Default(),
		shutdownChan: make(chan struct{}),
	}, nil
}

func (p *ClientProxy) Start() error {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return fmt.Errorf("代理已在运行")
	}

	listener, err := net.Listen("tcp", p.cfg.Listen)
	if err != nil {
		p.mu.Unlock()
		return fmt.Errorf("监听失败: %w", err)
	}

	p.listener = listener
	p.running = true
	p.mu.Unlock()

	p.wg.Add(1)
	go p.acceptLoop()

	p.logger.Info("客户端代理启动: %s -> %s", p.cfg.Listen, p.cfg.Target)
	return nil
}

func (p *ClientProxy) Stop() error {
	p.mu.Lock()
	if !p.running {
		p.mu.Unlock()
		return nil
	}

	close(p.shutdownChan)
	if p.listener != nil {
		p.listener.Close()
	}
	p.running = false
	p.mu.Unlock()

	p.wg.Wait()
	p.logger.Info("客户端代理停止: %s", p.cfg.Name)
	return nil
}

func (p *ClientProxy) Reload(cfg *config.InstanceConfig) error {
	return fmt.Errorf("热重载暂不支持")
}

func (p *ClientProxy) Stats() stats.Stats {
	return p.stats.GetStats()
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

		p.wg.Add(1)
		go p.handleConn(conn)
	}
}

func (p *ClientProxy) handleConn(clientConn net.Conn) {
	defer p.wg.Done()
	defer clientConn.Close()

	targetConn, err := net.DialTimeout("tcp", p.cfg.Target, 10*time.Second)
	if err != nil {
		p.logger.Error("连接目标失败: %v", err)
		return
	}
	defer targetConn.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		buf := make([]byte, 32*1024)
		for {
			n, err := clientConn.Read(buf)
			if n > 0 {
				if _, werr := targetConn.Write(buf[:n]); werr != nil {
					break
				}
			}
			if err != nil {
				break
			}
		}
	}()

	go func() {
		defer wg.Done()
		buf := make([]byte, 32*1024)
		for {
			n, err := targetConn.Read(buf)
			if n > 0 {
				if _, werr := clientConn.Write(buf[:n]); werr != nil {
					break
				}
			}
			if err != nil {
				break
			}
		}
	}()

	wg.Wait()
}

// HTTPServerProxy HTTP服务端代理（简化版）
type HTTPServerProxy struct {
	*ServerProxy
}

// NewHTTPServerProxy 创建HTTP服务端代理
func NewHTTPServerProxy(cfg *config.InstanceConfig) (*HTTPServerProxy, error) {
	sp, err := NewServerProxy(cfg)
	if err != nil {
		return nil, err
	}
	return &HTTPServerProxy{ServerProxy: sp}, nil
}

// HTTPClientProxy HTTP客户端代理（简化版）
type HTTPClientProxy struct {
	*ClientProxy
}

// NewHTTPClientProxy 创建HTTP客户端代理
func NewHTTPClientProxy(cfg *config.InstanceConfig) (*HTTPClientProxy, error) {
	cp, err := NewClientProxy(cfg)
	if err != nil {
		return nil, err
	}
	return &HTTPClientProxy{ClientProxy: cp}, nil
}
