package proxy

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/logger"
	"github.com/Trisia/tlcpchan/security"
	"github.com/Trisia/tlcpchan/security/keystore"
)

func needEncKeypair(suites []uint16) bool {
	for _, suite := range suites {
		if suite == tlcp.ECC_SM4_CBC_SM3 || suite == tlcp.ECDHE_SM4_CBC_SM3 {
			return true
		}
	}
	return false
}

func ValidateClientConfig(cfg *config.InstanceConfig) error {
	tlcpAuth := cfg.TLCP.Auth
	if tlcpAuth == "" {
		tlcpAuth = string(config.AuthNone)
	}

	if cfg.Protocol == string(config.ProtocolTLCP) || cfg.Protocol == string(config.ProtocolAuto) {
		suites, _ := config.ParseCipherSuites(cfg.TLCP.CipherSuites, true)
		if len(suites) > 0 && needEncKeypair(suites) && tlcpAuth != string(config.AuthMutual) {
			return fmt.Errorf("[%v] 密码套件只能在双向身份认证下使用，当前认证模式: %s", suites, tlcpAuth)
		}
	}

	return nil
}

type ProtocolType int

const (
	ProtocolAuto ProtocolType = iota
	ProtocolTLCP
	ProtocolTLS
)

const (
	TypeServer     = "server"
	TypeClient     = "client"
	TypeHTTPServer = "http-server"
	TypeHTTPClient = "http-client"
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
	mu              sync.RWMutex
	protocol        ProtocolType
	tlcpConfig      *tlcp.Config
	tlsConfig       *tls.Config
	rootCertPool    security.RootCertPool
	tlcpKeyStore    security.KeyStore
	tlsKeyStore     security.KeyStore
	cfg             *config.InstanceConfig
	keyStoreManager *security.KeyStoreManager
	rootCertManager *security.RootCertManager
	logger          *logger.Logger
}

func NewTLCPAdapter(cfg *config.InstanceConfig,
	keyStoreMgr *security.KeyStoreManager,
	rootCertMgr *security.RootCertManager) (*TLCPAdapter, error) {
	adapter := &TLCPAdapter{
		protocol:        ParseProtocolType(cfg.Protocol),
		cfg:             cfg,
		keyStoreManager: keyStoreMgr,
		rootCertManager: rootCertMgr,
		logger:          logger.Default(),
	}

	if cfg.Type == TypeServer || cfg.Type == TypeHTTPServer {
		if err := adapter.initServerConfig(cfg); err != nil {
			return nil, err
		}
	} else {
		if err := adapter.initClientConfig(cfg); err != nil {
			return nil, err
		}
	}

	return adapter, nil
}

func (a *TLCPAdapter) updateConfig(tlcpCfg *tlcp.Config, tlsCfg *tls.Config, rootCertPool security.RootCertPool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.tlcpConfig = tlcpCfg
	a.tlsConfig = tlsCfg
	a.rootCertPool = rootCertPool
}

func (a *TLCPAdapter) getTLCPConfig() *tlcp.Config {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.tlcpConfig
}

func (a *TLCPAdapter) getTLSConfig() *tls.Config {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.tlsConfig
}

func (a *TLCPAdapter) loadKeyStoreFromConfig(ksConfig *config.KeyStoreConfig, suggestedName string) (security.KeyStore, error) {
	if ksConfig == nil {
		return nil, nil
	}

	// 直接通过名称加载
	if string(ksConfig.Type) == string(keystore.LoaderTypeNamed) {
		return a.keyStoreManager.LoadFromConfig(ksConfig.Params["name"])
	}

	// 对于其他类型，使用 LoadAndRegister 方法加载并注册
	return a.keyStoreManager.LoadAndRegister(ksConfig.Name, suggestedName, string(ksConfig.Type), ksConfig.Params)
}

func (a *TLCPAdapter) initServerConfig(cfg *config.InstanceConfig) error {
	var tlcpConfig *tlcp.Config
	var tlsConfig *tls.Config
	var rootCertPool security.RootCertPool

	tlcpAuth := cfg.TLCP.Auth
	if tlcpAuth == "" {
		tlcpAuth = string(config.AuthNone)
	}
	tlsAuth := cfg.TLS.Auth
	if tlsAuth == "" {
		tlsAuth = string(config.AuthNone)
	}

	if tlcpAuth == string(config.AuthOneWay) || tlcpAuth == string(config.AuthMutual) {
		if ks, err := a.loadKeyStoreFromConfig(cfg.TLCP.Keystore, cfg.Name+"-tlcp"); err == nil && ks != nil {
			a.tlcpKeyStore = ks
		}
	}

	if tlsAuth == string(config.AuthOneWay) || tlsAuth == string(config.AuthMutual) {
		if ks, err := a.loadKeyStoreFromConfig(cfg.TLS.Keystore, cfg.Name+"-tls"); err == nil && ks != nil {
			a.tlsKeyStore = ks
		}
	}

	if len(cfg.ClientCA) > 0 {
		rootCertPool = a.rootCertManager.GetPool()
	}

	if a.tlcpKeyStore != nil {
		tlcpConfig = &tlcp.Config{}
		tlcpConfig.GetConfigForClient = func(chi *tlcp.ClientHelloInfo) (*tlcp.Config, error) {
			cfgCopy := a.buildTLCPServerConfig(rootCertPool, tlcpAuth)
			return cfgCopy, nil
		}
	}

	if a.tlsKeyStore != nil {
		tlsConfig = &tls.Config{}
		tlsConfig.GetConfigForClient = func(chi *tls.ClientHelloInfo) (*tls.Config, error) {
			cfgCopy := a.buildTLSServerConfig(rootCertPool, tlsAuth)
			return cfgCopy, nil
		}
	}

	a.updateConfig(tlcpConfig, tlsConfig, rootCertPool)

	return nil
}

func (a *TLCPAdapter) buildTLCPServerConfig(rootCertPool security.RootCertPool, auth string) *tlcp.Config {
	cfg := &tlcp.Config{}

	if a.tlcpKeyStore != nil {
		cfg.GetCertificate = func(chi *tlcp.ClientHelloInfo) (*tlcp.Certificate, error) {
			return a.tlcpKeyStore.TLCPCertificate()
		}
	}

	if auth == string(config.AuthMutual) && rootCertPool != nil {
		cfg.ClientCAs = rootCertPool.GetSMCertPool()
		cfg.ClientAuth = tlcp.RequireAndVerifyClientCert
	} else if auth == string(config.AuthOneWay) {
		cfg.ClientAuth = tlcp.NoClientCert
	}

	a.applyTLCPSettingsToConfig(cfg, &a.cfg.TLCP)
	return cfg
}

func (a *TLCPAdapter) buildTLSServerConfig(rootCertPool security.RootCertPool, auth string) *tls.Config {
	cfg := &tls.Config{}

	if a.tlsKeyStore != nil {
		cfg.GetCertificate = func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return a.tlsKeyStore.TLSCertificate()
		}
	}

	if auth == string(config.AuthMutual) && rootCertPool != nil {
		cfg.ClientCAs = rootCertPool.GetCertPool()
		cfg.ClientAuth = tls.RequireAndVerifyClientCert
	} else if auth == string(config.AuthOneWay) {
		cfg.ClientAuth = tls.NoClientCert
	}

	a.applyTLSSettingsToConfig(cfg, &a.cfg.TLS)
	return cfg
}

func (a *TLCPAdapter) initClientConfig(cfg *config.InstanceConfig) error {
	var tlcpConfig *tlcp.Config
	var tlsConfig *tls.Config
	var rootCertPool security.RootCertPool

	tlcpAuth := cfg.TLCP.Auth
	if tlcpAuth == "" {
		tlcpAuth = string(config.AuthNone)
	}
	tlsAuth := cfg.TLS.Auth
	if tlsAuth == "" {
		tlsAuth = string(config.AuthNone)
	}

	tlcpConfig = &tlcp.Config{
		InsecureSkipVerify: cfg.TLCP.InsecureSkipVerify,
	}

	if cfg.SNI != "" {
		tlcpConfig.ServerName = cfg.SNI
	}

	if tlcpAuth == string(config.AuthMutual) {
		if ks, err := a.loadKeyStoreFromConfig(cfg.TLCP.Keystore, cfg.Name+"-tlcp"); err == nil && ks != nil {
			a.tlcpKeyStore = ks
			tlcpConfig.GetClientCertificate = func(*tlcp.CertificateRequestInfo) (*tlcp.Certificate, error) {
				return a.tlcpKeyStore.TLCPCertificate()
			}
		}
	}

	tlsConfig = &tls.Config{
		InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
	}

	if cfg.SNI != "" {
		tlsConfig.ServerName = cfg.SNI
	}

	if tlsAuth == string(config.AuthMutual) {
		if ks, err := a.loadKeyStoreFromConfig(cfg.TLS.Keystore, cfg.Name+"-tls"); err == nil && ks != nil {
			a.tlsKeyStore = ks
			tlsConfig.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
				return a.tlsKeyStore.TLSCertificate()
			}
		}
	}

	if len(cfg.ServerCA) > 0 {
		rootCertPool = a.rootCertManager.GetPool()
		tlcpConfig.RootCAs = rootCertPool.GetSMCertPool()
		tlsConfig.RootCAs = rootCertPool.GetCertPool()
	}

	a.applyTLCPSettingsToConfig(tlcpConfig, &cfg.TLCP)
	a.applyTLSSettingsToConfig(tlsConfig, &cfg.TLS)

	a.updateConfig(tlcpConfig, tlsConfig, rootCertPool)

	return nil
}

func (a *TLCPAdapter) applyTLCPSettingsToConfig(cfg *tlcp.Config, tlsCfg *config.TLCPConfig) {
	if cfg == nil {
		return
	}

	if len(tlsCfg.CipherSuites) > 0 {
		suites, _ := config.ParseCipherSuites(tlsCfg.CipherSuites, true)
		cfg.CipherSuites = suites
	}

	if tlsCfg.MinVersion != "" {
		v, _ := config.ParseTLSVersion(tlsCfg.MinVersion, true)
		cfg.MinVersion = v
	}

	if tlsCfg.MaxVersion != "" {
		v, _ := config.ParseTLSVersion(tlsCfg.MaxVersion, true)
		cfg.MaxVersion = v
	}

	if tlsCfg.SessionCache {
		cfg.SessionCache = tlcp.NewLRUSessionCache(100)
	}
}

func (a *TLCPAdapter) applyTLSSettingsToConfig(cfg *tls.Config, tlsCfg *config.TLSConfig) {
	if cfg == nil {
		return
	}

	if len(tlsCfg.CipherSuites) > 0 {
		suites, _ := config.ParseCipherSuites(tlsCfg.CipherSuites, false)
		cfg.CipherSuites = suites
	}

	if tlsCfg.MinVersion != "" {
		v, _ := config.ParseTLSVersion(tlsCfg.MinVersion, false)
		cfg.MinVersion = v
	}

	if tlsCfg.MaxVersion != "" {
		v, _ := config.ParseTLSVersion(tlsCfg.MaxVersion, false)
		cfg.MaxVersion = v
	}

	if tlsCfg.SessionCache {
		cfg.ClientSessionCache = tls.NewLRUClientSessionCache(100)
	}
}

func (a *TLCPAdapter) TLCPListener(l net.Listener) net.Listener {
	return tlcp.NewListener(l, a.getTLCPConfig())
}

func (a *TLCPAdapter) TLSListener(l net.Listener) net.Listener {
	return tls.NewListener(l, a.getTLSConfig())
}

func (a *TLCPAdapter) AutoListener(l net.Listener) net.Listener {
	timeout := a.getTimeoutConfig().Handshake
	return NewAutoProtocolListener(l, a.getTLCPConfig(), a.getTLSConfig(), timeout)
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

func (a *TLCPAdapter) getTimeoutConfig() *config.TimeoutConfig {
	if a.cfg.Timeout != nil {
		return a.cfg.Timeout
	}
	return config.DefaultTimeout()
}

func (a *TLCPAdapter) DialTLCP(network, addr string) (net.Conn, error) {
	timeout := a.getTimeoutConfig().Dial
	dialer := &net.Dialer{
		Timeout: timeout,
	}
	return tlcp.DialWithDialer(dialer, network, addr, a.getTLCPConfig())
}

func (a *TLCPAdapter) DialTLS(network, addr string) (net.Conn, error) {
	timeout := a.getTimeoutConfig().Dial
	dialer := &net.Dialer{
		Timeout: timeout,
	}
	return tls.DialWithDialer(dialer, network, addr, a.getTLSConfig())
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
	return a.getTLCPConfig()
}

func (a *TLCPAdapter) TLSConfig() *tls.Config {
	return a.getTLSConfig()
}

func (a *TLCPAdapter) ReloadConfig(cfg *config.InstanceConfig) error {
	a.mu.Lock()
	oldCfg := a.cfg
	a.cfg = cfg
	a.mu.Unlock()

	if cfg.Type == TypeServer || cfg.Type == TypeHTTPServer {
		return a.reloadServerConfig(cfg, oldCfg)
	}
	return a.reloadClientConfig(cfg, oldCfg)
}

func (a *TLCPAdapter) reloadServerConfig(cfg, oldCfg *config.InstanceConfig) error {
	var tlcpConfig *tlcp.Config
	var tlsConfig *tls.Config
	var rootCertPool security.RootCertPool

	tlcpAuth := cfg.TLCP.Auth
	if tlcpAuth == "" {
		tlcpAuth = string(config.AuthNone)
	}
	tlsAuth := cfg.TLS.Auth
	if tlsAuth == "" {
		tlsAuth = string(config.AuthNone)
	}

	if newTLCPKS, err := a.loadKeyStoreFromConfig(cfg.TLCP.Keystore, cfg.Name+"-tlcp"); err == nil {
		if a.tlcpKeyStore == nil || !a.tlcpKeyStore.Equals(newTLCPKS) {
			a.tlcpKeyStore = newTLCPKS
		}
	} else {
		a.tlcpKeyStore = nil
	}

	if newTLSKS, err := a.loadKeyStoreFromConfig(cfg.TLS.Keystore, cfg.Name+"-tls"); err == nil {
		if a.tlsKeyStore == nil || !a.tlsKeyStore.Equals(newTLSKS) {
			a.tlsKeyStore = newTLSKS
		}
	} else {
		a.tlsKeyStore = nil
	}

	if len(cfg.ClientCA) > 0 {
		rootCertPool = a.rootCertManager.GetPool()
	}

	if a.tlcpKeyStore != nil {
		tlcpConfig = &tlcp.Config{}
		tlcpConfig.GetConfigForClient = func(chi *tlcp.ClientHelloInfo) (*tlcp.Config, error) {
			cfgCopy := a.buildTLCPServerConfig(rootCertPool, tlcpAuth)
			return cfgCopy, nil
		}
	}

	if a.tlsKeyStore != nil {
		tlsConfig = &tls.Config{}
		tlsConfig.GetConfigForClient = func(chi *tls.ClientHelloInfo) (*tls.Config, error) {
			cfgCopy := a.buildTLSServerConfig(rootCertPool, tlsAuth)
			return cfgCopy, nil
		}
	}

	a.updateConfig(tlcpConfig, tlsConfig, rootCertPool)
	return nil
}

func (a *TLCPAdapter) reloadClientConfig(cfg, oldCfg *config.InstanceConfig) error {
	var tlcpConfig *tlcp.Config
	var tlsConfig *tls.Config
	var rootCertPool security.RootCertPool

	tlcpAuth := cfg.TLCP.Auth
	if tlcpAuth == "" {
		tlcpAuth = string(config.AuthNone)
	}
	tlsAuth := cfg.TLS.Auth
	if tlsAuth == "" {
		tlsAuth = string(config.AuthNone)
	}

	if newTLCPKS, err := a.loadKeyStoreFromConfig(cfg.TLCP.Keystore, cfg.Name+"-tlcp"); err == nil {
		if a.tlcpKeyStore == nil || !a.tlcpKeyStore.Equals(newTLCPKS) {
			a.tlcpKeyStore = newTLCPKS
		}
	} else {
		a.tlcpKeyStore = nil
	}

	if newTLSKS, err := a.loadKeyStoreFromConfig(cfg.TLS.Keystore, cfg.Name+"-tls"); err == nil {
		if a.tlsKeyStore == nil || !a.tlsKeyStore.Equals(newTLSKS) {
			a.tlsKeyStore = newTLSKS
		}
	} else {
		a.tlsKeyStore = nil
	}

	if len(cfg.ServerCA) > 0 {
		rootCertPool = a.rootCertManager.GetPool()
	}

	tlcpConfig = &tlcp.Config{
		InsecureSkipVerify: cfg.TLCP.InsecureSkipVerify,
	}
	if cfg.SNI != "" {
		tlcpConfig.ServerName = cfg.SNI
	}
	if a.tlcpKeyStore != nil {
		tlcpConfig.GetClientCertificate = func(*tlcp.CertificateRequestInfo) (*tlcp.Certificate, error) {
			return a.tlcpKeyStore.TLCPCertificate()
		}
	}
	if rootCertPool != nil {
		tlcpConfig.RootCAs = rootCertPool.GetSMCertPool()
	}
	a.applyTLCPSettingsToConfig(tlcpConfig, &cfg.TLCP)

	tlsConfig = &tls.Config{
		InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
	}
	if cfg.SNI != "" {
		tlsConfig.ServerName = cfg.SNI
	}
	if a.tlsKeyStore != nil {
		tlsConfig.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			return a.tlsKeyStore.TLSCertificate()
		}
	}
	if rootCertPool != nil {
		tlsConfig.RootCAs = rootCertPool.GetCertPool()
	}
	a.applyTLSSettingsToConfig(tlsConfig, &cfg.TLS)

	a.updateConfig(tlcpConfig, tlsConfig, rootCertPool)
	return nil
}

type AutoProtocolListener struct {
	net.Listener
	tlcpConfig       *tlcp.Config
	tlsConfig        *tls.Config
	handshakeTimeout time.Duration
}

func NewAutoProtocolListener(l net.Listener, tlcpCfg *tlcp.Config, tlsCfg *tls.Config, handshakeTimeout time.Duration) *AutoProtocolListener {
	return &AutoProtocolListener{
		Listener:         l,
		tlcpConfig:       tlcpCfg,
		tlsConfig:        tlsCfg,
		handshakeTimeout: handshakeTimeout,
	}
}

func (l *AutoProtocolListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return newAutoProtocolConn(conn, l.tlcpConfig, l.tlsConfig, l.handshakeTimeout), nil
}

type autoProtocolConn struct {
	net.Conn
	tlcpConfig       *tlcp.Config
	tlsConfig        *tls.Config
	handshakeTimeout time.Duration
	peeked           []byte
	handshaked       bool
	conn             net.Conn
	mu               sync.Mutex
}

func newAutoProtocolConn(conn net.Conn, tlcpCfg *tlcp.Config, tlsCfg *tls.Config, handshakeTimeout time.Duration) *autoProtocolConn {
	return &autoProtocolConn{
		Conn:             conn,
		tlcpConfig:       tlcpCfg,
		tlsConfig:        tlsCfg,
		handshakeTimeout: handshakeTimeout,
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
	timeout := c.handshakeTimeout
	if timeout == 0 {
		timeout = 15 * time.Second
	}
	c.Conn.SetReadDeadline(time.Now().Add(timeout))
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
