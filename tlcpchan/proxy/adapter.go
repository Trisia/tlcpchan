package proxy

import (
	"crypto/tls"
	"net"
	"os"
	"sync"
	"time"

	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/Trisia/tlcpchan/cert"
	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/logger"
	"github.com/emmansun/gmsm/smx509"
)

type ProtocolType int

const (
	ProtocolAuto ProtocolType = iota
	ProtocolTLCP
	ProtocolTLS
)

func (p ProtocolType) String() string {
	switch p {
	case ProtocolTLCP:
		return "tlcp"
	case ProtocolTLS:
		return "tls"
	default:
		return "auto"
	}
}

func ParseProtocolType(s string) ProtocolType {
	switch s {
	case "tlcp":
		return ProtocolTLCP
	case "tls":
		return ProtocolTLS
	default:
		return ProtocolAuto
	}
}

type TLCPAdapter struct {
	tlcpConfig *tlcp.Config
	tlsConfig  *tls.Config
	protocol   ProtocolType
	logger     *logger.Logger
	mu         sync.RWMutex
}

func NewTLCPAdapter(cfg *config.InstanceConfig, certManager *cert.Manager) (*TLCPAdapter, error) {
	adapter := &TLCPAdapter{
		protocol: ParseProtocolType(cfg.Protocol),
		logger:   logger.Default(),
	}

	if cfg.Type == "server" || cfg.Type == "http-server" {
		if err := adapter.initServerConfig(cfg, certManager); err != nil {
			return nil, err
		}
	} else {
		if err := adapter.initClientConfig(cfg, certManager); err != nil {
			return nil, err
		}
	}

	return adapter, nil
}

func loadSMCertPool(paths []string) (*smx509.CertPool, error) {
	pool := smx509.NewCertPool()

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		if !pool.AppendCertsFromPEM(data) {
			return nil, err
		}
	}

	return pool, nil
}

func (a *TLCPAdapter) initServerConfig(cfg *config.InstanceConfig, certManager *cert.Manager) error {
	if cfg.Auth == "one-way" || cfg.Auth == "mutual" {
		if cfg.Certs.TLCP.Cert != "" && cfg.Certs.TLCP.Key != "" {
			tlcpCert, err := certManager.LoadTLCP(cfg.Name+"-tlcp", cfg.Certs.TLCP.Cert, cfg.Certs.TLCP.Key)
			if err != nil {
				return err
			}
			a.tlcpConfig = &tlcp.Config{
				Certificates: []tlcp.Certificate{tlcpCert.Cert.TLCPCertificate()},
			}
		}

		if cfg.Certs.TLS.Cert != "" && cfg.Certs.TLS.Key != "" {
			tlsCert, err := certManager.LoadTLS(cfg.Name+"-tls", cfg.Certs.TLS.Cert, cfg.Certs.TLS.Key)
			if err != nil {
				return err
			}
			a.tlsConfig = &tls.Config{
				Certificates: []tls.Certificate{tlsCert.Cert.TLSCertificate()},
			}
		}
	}

	if cfg.Auth == "mutual" {
		if len(cfg.ClientCA) > 0 && a.tlcpConfig != nil {
			pool, err := loadSMCertPool(cfg.ClientCA)
			if err != nil {
				return err
			}
			a.tlcpConfig.ClientCAs = pool
			a.tlcpConfig.ClientAuth = tlcp.RequireAndVerifyClientCert
		}

		if len(cfg.ClientCA) > 0 && a.tlsConfig != nil {
			pool, err := certManager.Loader().LoadClientCA(cfg.ClientCA)
			if err != nil {
				return err
			}
			a.tlsConfig.ClientCAs = pool
			a.tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}
	}

	a.applyTLCPSettings(&cfg.TLCP)
	a.applyTLSSettings(&cfg.TLS)

	return nil
}

func (a *TLCPAdapter) initClientConfig(cfg *config.InstanceConfig, certManager *cert.Manager) error {
	a.tlcpConfig = &tlcp.Config{
		InsecureSkipVerify: cfg.TLCP.InsecureSkipVerify,
	}
	a.tlsConfig = &tls.Config{
		InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
	}

	if cfg.SNI != "" {
		a.tlcpConfig.ServerName = cfg.SNI
		a.tlsConfig.ServerName = cfg.SNI
	}

	if cfg.Auth == "mutual" {
		if cfg.Certs.TLCP.Cert != "" && cfg.Certs.TLCP.Key != "" {
			tlcpCert, err := certManager.LoadTLCP(cfg.Name+"-tlcp", cfg.Certs.TLCP.Cert, cfg.Certs.TLCP.Key)
			if err != nil {
				return err
			}
			a.tlcpConfig.Certificates = []tlcp.Certificate{tlcpCert.Cert.TLCPCertificate()}
		}

		if cfg.Certs.TLS.Cert != "" && cfg.Certs.TLS.Key != "" {
			tlsCert, err := certManager.LoadTLS(cfg.Name+"-tls", cfg.Certs.TLS.Cert, cfg.Certs.TLS.Key)
			if err != nil {
				return err
			}
			a.tlsConfig.Certificates = []tls.Certificate{tlsCert.Cert.TLSCertificate()}
		}
	}

	if len(cfg.ServerCA) > 0 {
		smPool, err := loadSMCertPool(cfg.ServerCA)
		if err != nil {
			return err
		}
		a.tlcpConfig.RootCAs = smPool

		pool, err := certManager.Loader().LoadServerCA(cfg.ServerCA)
		if err != nil {
			return err
		}
		a.tlsConfig.RootCAs = pool
	}

	a.applyTLCPSettings(&cfg.TLCP)
	a.applyTLSSettings(&cfg.TLS)

	return nil
}

func (a *TLCPAdapter) applyTLCPSettings(cfg *config.TLCPConfig) {
	if a.tlcpConfig == nil {
		return
	}

	if len(cfg.CipherSuites) > 0 {
		suites, _ := config.ParseCipherSuites(cfg.CipherSuites, true)
		a.tlcpConfig.CipherSuites = suites
	}

	if cfg.MinVersion != "" {
		v, _ := config.ParseTLSVersion(cfg.MinVersion, true)
		a.tlcpConfig.MinVersion = v
	}

	if cfg.MaxVersion != "" {
		v, _ := config.ParseTLSVersion(cfg.MaxVersion, true)
		a.tlcpConfig.MaxVersion = v
	}

	if cfg.SessionCache {
		a.tlcpConfig.SessionCache = tlcp.NewLRUSessionCache(100)
	}
}

func (a *TLCPAdapter) applyTLSSettings(cfg *config.TLSConfig) {
	if a.tlsConfig == nil {
		return
	}

	if len(cfg.CipherSuites) > 0 {
		suites, _ := config.ParseCipherSuites(cfg.CipherSuites, false)
		a.tlsConfig.CipherSuites = suites
	}

	if cfg.MinVersion != "" {
		v, _ := config.ParseTLSVersion(cfg.MinVersion, false)
		a.tlsConfig.MinVersion = v
	}

	if cfg.MaxVersion != "" {
		v, _ := config.ParseTLSVersion(cfg.MaxVersion, false)
		a.tlsConfig.MaxVersion = v
	}

	if cfg.SessionCache {
		a.tlsConfig.ClientSessionCache = tls.NewLRUClientSessionCache(100)
	}
}

func (a *TLCPAdapter) TLCPListener(l net.Listener) net.Listener {
	return tlcp.NewListener(l, a.tlcpConfig)
}

func (a *TLCPAdapter) TLSListener(l net.Listener) net.Listener {
	return tls.NewListener(l, a.tlsConfig)
}

func (a *TLCPAdapter) AutoListener(l net.Listener) net.Listener {
	return NewAutoProtocolListener(l, a.tlcpConfig, a.tlsConfig)
}

func (a *TLCPAdapter) WrapServerListener(l net.Listener) net.Listener {
	a.mu.RLock()
	defer a.mu.RUnlock()

	switch a.protocol {
	case ProtocolTLCP:
		return a.TLCPListener(l)
	case ProtocolTLS:
		return a.TLSListener(l)
	default:
		return a.AutoListener(l)
	}
}

func (a *TLCPAdapter) DialTLCP(network, addr string) (net.Conn, error) {
	return tlcp.Dial(network, addr, a.tlcpConfig)
}

func (a *TLCPAdapter) DialTLS(network, addr string) (net.Conn, error) {
	return tls.Dial(network, addr, a.tlsConfig)
}

func (a *TLCPAdapter) DialWithProtocol(network, addr string, protocol ProtocolType) (net.Conn, error) {
	switch protocol {
	case ProtocolTLCP:
		return a.DialTLCP(network, addr)
	case ProtocolTLS:
		return a.DialTLS(network, addr)
	default:
		return a.autoDial(network, addr)
	}
}

func (a *TLCPAdapter) autoDial(network, addr string) (net.Conn, error) {
	conn, err := a.DialTLCP(network, addr)
	if err == nil {
		return conn, nil
	}

	a.logger.Debug("TLCP连接失败，尝试TLS: %v", err)
	return a.DialTLS(network, addr)
}

func (a *TLCPAdapter) Protocol() ProtocolType {
	return a.protocol
}

func (a *TLCPAdapter) TLCPConfig() *tlcp.Config {
	return a.tlcpConfig
}

func (a *TLCPAdapter) TLSConfig() *tls.Config {
	return a.tlsConfig
}

type AutoProtocolListener struct {
	net.Listener
	tlcpConfig *tlcp.Config
	tlsConfig  *tls.Config
}

func NewAutoProtocolListener(l net.Listener, tlcpCfg *tlcp.Config, tlsCfg *tls.Config) *AutoProtocolListener {
	return &AutoProtocolListener{
		Listener:   l,
		tlcpConfig: tlcpCfg,
		tlsConfig:  tlsCfg,
	}
}

func (l *AutoProtocolListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return newAutoProtocolConn(conn, l.tlcpConfig, l.tlsConfig), nil
}

type autoProtocolConn struct {
	net.Conn
	tlcpConfig *tlcp.Config
	tlsConfig  *tls.Config
	peeked     []byte
	handshaked bool
	conn       net.Conn
	mu         sync.Mutex
}

func newAutoProtocolConn(conn net.Conn, tlcpCfg *tlcp.Config, tlsCfg *tls.Config) *autoProtocolConn {
	return &autoProtocolConn{
		Conn:       conn,
		tlcpConfig: tlcpCfg,
		tlsConfig:  tlsCfg,
	}
}

func (c *autoProtocolConn) Read(b []byte) (n int, err error) {
	c.mu.Lock()
	if c.handshaked {
		c.mu.Unlock()
		if c.conn != nil {
			return c.conn.Read(b)
		}
		return c.Conn.Read(b)
	}

	if len(c.peeked) > 0 {
		n = copy(b, c.peeked)
		c.peeked = c.peeked[n:]
		c.mu.Unlock()
		return n, nil
	}

	c.mu.Unlock()
	return c.Conn.Read(b)
}

func (c *autoProtocolConn) Handshake() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.handshaked {
		return nil
	}

	peekBuf := make([]byte, 6)
	c.Conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := ioReadFull(c.Conn, peekBuf)
	c.Conn.SetReadDeadline(time.Time{})

	if err != nil {
		return err
	}

	c.peeked = peekBuf[:n]

	protocol := detectProtocol(peekBuf[:n])

	if protocol == ProtocolTLCP && c.tlcpConfig != nil {
		tlcpConn := tlcp.Server(c.Conn, c.tlcpConfig)
		if err := tlcpConn.Handshake(); err != nil {
			return err
		}
		c.conn = tlcpConn
	} else if c.tlsConfig != nil {
		tlsConn := tls.Server(c.Conn, c.tlsConfig)
		if err := tlsConn.Handshake(); err != nil {
			return err
		}
		c.conn = tlsConn
	}

	c.handshaked = true
	return nil
}

func (c *autoProtocolConn) Write(b []byte) (n int, err error) {
	c.mu.Lock()
	handshaked := c.handshaked
	conn := c.conn
	c.mu.Unlock()

	if handshaked && conn != nil {
		return conn.Write(b)
	}
	return c.Conn.Write(b)
}

func (c *autoProtocolConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return c.conn.Close()
	}
	return c.Conn.Close()
}

func ioReadFull(r net.Conn, buf []byte) (n int, err error) {
	for n < len(buf) {
		nn, err := r.Read(buf[n:])
		n += nn
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

func detectProtocol(data []byte) ProtocolType {
	if len(data) < 5 {
		return ProtocolTLS
	}

	recordType := data[0]
	version := uint16(data[3])<<8 | uint16(data[4])

	if recordType == 22 {
		if version == 0x0101 {
			return ProtocolTLCP
		}
		if version == 0x0301 || version == 0x0302 || version == 0x0303 || version == 0x0304 {
			return ProtocolTLS
		}
	}

	return ProtocolTLS
}
