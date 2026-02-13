package proxy

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Trisia/tlcpchan/cert"
	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/logger"
	"github.com/Trisia/tlcpchan/stats"
)

type HTTPServerProxy struct {
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
}

func NewHTTPServerProxy(cfg *config.InstanceConfig, certManager *cert.Manager) (*HTTPServerProxy, error) {
	adapter, err := NewTLCPAdapter(cfg, certManager)
	if err != nil {
		return nil, fmt.Errorf("创建协议适配器失败: %w", err)
	}

	return &HTTPServerProxy{
		cfg:          cfg,
		adapter:      adapter,
		handler:      NewConnHandler(),
		certManager:  certManager,
		stats:        stats.DefaultCollector(),
		logger:       logger.Default(),
		shutdownChan: make(chan struct{}),
	}, nil
}

func (p *HTTPServerProxy) Start() error {
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

	p.logger.Info("HTTP服务端代理启动: %s -> %s, 协议: %s", p.cfg.Listen, p.cfg.Target, p.cfg.Protocol)

	p.wg.Add(1)
	go p.acceptLoop()

	return nil
}

func (p *HTTPServerProxy) acceptLoop() {
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

func (p *HTTPServerProxy) handleConnection(clientConn net.Conn) {
	defer p.wg.Done()
	defer p.stats.DecrementConnections()
	defer clientConn.Close()

	if autoConn, ok := clientConn.(*autoProtocolConn); ok {
		if err := autoConn.Handshake(); err != nil {
			p.logger.Error("协议握手失败: %v", err)
			p.stats.IncrementErrors()
			return
		}
	}

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

		if err := p.processRequest(req); err != nil {
			p.logger.Error("处理请求失败: %v", err)
			p.stats.IncrementErrors()
			return
		}

		resp, err := p.forwardRequest(req)
		if err != nil {
			p.logger.Error("转发请求失败: %v", err)
			p.stats.IncrementErrors()
			return
		}

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

func (p *HTTPServerProxy) processRequest(req *http.Request) error {
	if p.cfg.HTTP == nil {
		return nil
	}

	vars := ExtractVariables("", "", p.cfg.Target, p.cfg.Protocol, p.cfg.Name)

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

func (p *HTTPServerProxy) forwardRequest(req *http.Request) (*http.Response, error) {
	targetURL := fmt.Sprintf("http://%s%s", p.cfg.Target, req.URL.Path)
	if req.URL.RawQuery != "" {
		targetURL += "?" + req.URL.RawQuery
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("读取请求体失败: %w", err)
	}
	req.Body = io.NopCloser(bytes.NewReader(body))

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
	newReq.Header.Set("X-Forwarded-For", req.RemoteAddr)

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				dialer := &net.Dialer{
					Timeout: 10 * time.Second,
				}
				return dialer.DialContext(ctx, network, p.cfg.Target)
			},
		},
	}

	resp, err := client.Do(newReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}

	p.processResponse(resp)

	return resp, nil
}

func (p *HTTPServerProxy) processResponse(resp *http.Response) {
	if p.cfg.HTTP == nil {
		return
	}

	vars := ExtractVariables("", "", p.cfg.Target, p.cfg.Protocol, p.cfg.Name)

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

func (p *HTTPServerProxy) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return nil
	}

	p.logger.Info("停止HTTP服务端代理: %s", p.cfg.Name)

	close(p.shutdownChan)

	if p.listener != nil {
		p.listener.Close()
	}

	p.wg.Wait()

	p.running = false
	p.shutdownChan = make(chan struct{})

	return nil
}

func (p *HTTPServerProxy) Restart() error {
	if err := p.Stop(); err != nil {
		return err
	}
	return p.Start()
}

func (p *HTTPServerProxy) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}

func (p *HTTPServerProxy) Name() string {
	return p.cfg.Name
}

func (p *HTTPServerProxy) Config() *config.InstanceConfig {
	return p.cfg
}

func (p *HTTPServerProxy) Stats() stats.Stats {
	return p.stats.GetStats()
}

func getRemoteAddrFromRequest(req *http.Request) string {
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	return req.RemoteAddr
}
