package proxy

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/Trisia/tlcpchan/cert"
	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/logger"
	"github.com/Trisia/tlcpchan/stats"
)

type HTTPClientProxy struct {
	cfg          *config.InstanceConfig
	adapter      *TLCPAdapter
	handler      *ConnHandler
	listener     net.Listener
	certManager  *cert.Manager
	stats        *stats.Collector
	logger       *logger.Logger
	shutdownChan chan struct{}
	wg           sync.WaitGroup
	mu           sync.Mutex
	running      bool

	protocolCache map[string]protocolCacheEntry
	cacheMu       sync.RWMutex
	cacheTTL      time.Duration
}

func NewHTTPClientProxy(cfg *config.InstanceConfig, certManager *cert.Manager) (*HTTPClientProxy, error) {
	adapter, err := NewTLCPAdapter(cfg, certManager)
	if err != nil {
		return nil, fmt.Errorf("创建协议适配器失败: %w", err)
	}

	return &HTTPClientProxy{
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

func (p *HTTPClientProxy) Start() error {
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

	p.logger.Info("HTTP客户端代理启动: %s -> %s, 协议: %s", p.cfg.Listen, p.cfg.Target, p.cfg.Protocol)

	p.wg.Add(1)
	go p.acceptLoop()

	return nil
}

func (p *HTTPClientProxy) acceptLoop() {
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

func (p *HTTPClientProxy) handleConnection(clientConn net.Conn) {
	defer p.wg.Done()
	defer p.stats.DecrementConnections()
	defer clientConn.Close()

	reader := bufio.NewReader(clientConn)
	for {
		req, err := http.ReadRequest(reader)
		if err != nil {
			if err != io.EOF {
				p.logger.Debug("读取请求失败: %v", err)
			}
			return
		}

		p.stats.IncrementRequests()

		vars := ExtractVariables(clientConn.RemoteAddr().String(), clientConn.LocalAddr().String(), p.cfg.Target, p.cfg.Protocol, p.cfg.Name)

		if err := p.processRequest(req, vars); err != nil {
			p.logger.Error("处理请求失败: %v", err)
			p.stats.IncrementErrors()
			return
		}

		resp, err := p.forwardRequest(req, vars)
		if err != nil {
			p.logger.Error("转发请求失败: %v", err)
			p.stats.IncrementErrors()
			return
		}

		p.processResponse(resp, vars)

		if err := resp.Write(clientConn); err != nil {
			p.logger.Debug("写入响应失败: %v", err)
			return
		}

		if req.Close {
			return
		}

		if resp.Close {
			return
		}
	}
}

func (p *HTTPClientProxy) processRequest(req *http.Request, vars *Variables) error {
	if p.cfg.HTTP == nil {
		return nil
	}

	for key, value := range p.cfg.HTTP.RequestHeaders.Add {
		req.Header.Add(key, vars.Replace(value))
	}

	for _, key := range p.cfg.HTTP.RequestHeaders.Remove {
		req.Header.Del(key)
	}

	for key, value := range p.cfg.HTTP.RequestHeaders.Set {
		req.Header.Set(key, vars.Replace(value))
	}

	return nil
}

func (p *HTTPClientProxy) forwardRequest(req *http.Request, vars *Variables) (*http.Response, error) {
	scheme := "https"
	protocol := p.getProtocol()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("读取请求体失败: %w", err)
	}
	req.Body = io.NopCloser(bytes.NewReader(body))

	targetURL := fmt.Sprintf("%s://%s%s", scheme, p.cfg.Target, req.URL.Path)
	if req.URL.RawQuery != "" {
		targetURL += "?" + req.URL.RawQuery
	}

	newReq, err := http.NewRequest(req.Method, targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	for key, values := range req.Header {
		for _, value := range values {
			newReq.Header.Add(key, value)
		}
	}

	newReq.Header.Set("Host", req.Host)

	var client *http.Client
	if protocol == ProtocolTLCP {
		client = p.createTLCPClient()
	} else {
		client = p.createTLSClient()
	}

	resp, err := client.Do(newReq)
	if err != nil {
		if protocol == ProtocolAuto || protocol == ProtocolTLCP {
			p.logger.Debug("TLCP请求失败，尝试TLS: %v", err)
			client = p.createTLSClient()
			resp, err = client.Do(newReq)
			if err == nil {
				p.cacheProtocol(ProtocolTLS)
			}
		}
		if err != nil {
			return nil, fmt.Errorf("发送请求失败: %w", err)
		}
	} else if protocol == ProtocolAuto {
		p.cacheProtocol(ProtocolTLCP)
	}

	return resp, nil
}

func (p *HTTPClientProxy) processResponse(resp *http.Response, vars *Variables) {
	if p.cfg.HTTP == nil {
		return
	}

	for key, value := range p.cfg.HTTP.ResponseHeaders.Add {
		resp.Header.Add(key, vars.Replace(value))
	}

	for _, key := range p.cfg.HTTP.ResponseHeaders.Remove {
		resp.Header.Del(key)
	}

	for key, value := range p.cfg.HTTP.ResponseHeaders.Set {
		resp.Header.Set(key, vars.Replace(value))
	}
}

func (p *HTTPClientProxy) getProtocol() ProtocolType {
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

func (p *HTTPClientProxy) cacheProtocol(protocol ProtocolType) {
	p.cacheMu.Lock()
	defer p.cacheMu.Unlock()
	p.protocolCache[p.cfg.Target] = protocolCacheEntry{
		protocol: protocol,
		detected: time.Now(),
	}
}

func (p *HTTPClientProxy) createTLCPClient() *http.Client {
	transport := &http.Transport{
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dialer := &tlcp.Dialer{
				NetDialer: &net.Dialer{
					Timeout: 10 * time.Second,
				},
				Config: p.adapter.TLCPConfig(),
			}
			return dialer.Dial(network, p.cfg.Target)
		},
	}

	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}
}

func (p *HTTPClientProxy) createTLSClient() *http.Client {
	transport := &http.Transport{
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dialer := &tls.Dialer{
				Config: p.adapter.TLSConfig(),
			}
			return dialer.DialContext(ctx, network, p.cfg.Target)
		},
	}

	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}
}

func (p *HTTPClientProxy) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return nil
	}

	p.logger.Info("停止HTTP客户端代理: %s", p.cfg.Name)

	close(p.shutdownChan)

	if p.listener != nil {
		p.listener.Close()
	}

	p.wg.Wait()

	p.running = false
	p.shutdownChan = make(chan struct{})

	return nil
}

func (p *HTTPClientProxy) Restart() error {
	if err := p.Stop(); err != nil {
		return err
	}
	return p.Start()
}

func (p *HTTPClientProxy) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}

func (p *HTTPClientProxy) Name() string {
	return p.cfg.Name
}

func (p *HTTPClientProxy) Config() *config.InstanceConfig {
	return p.cfg
}

func (p *HTTPClientProxy) Stats() stats.Stats {
	return p.stats.GetStats()
}

func (p *HTTPClientProxy) ClearProtocolCache() {
	p.cacheMu.Lock()
	defer p.cacheMu.Unlock()
	p.protocolCache = make(map[string]protocolCacheEntry)
}
