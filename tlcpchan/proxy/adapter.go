package proxy

import (
	"crypto/tls"
	"fmt"
	"io"
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
		if suite == 0xC011 || suite == 0xC012 || suite == 0xC013 ||
			suite == 0xC014 || suite == 0xC019 || suite == 0xC01A {
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
	recordHeader     []byte
	handshaked       bool
	conn             net.Conn
	major, minor     uint8
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
	if len(c.recordHeader) == 0 {
		return c.Conn.Read(b)
	}

	if len(b) >= len(c.recordHeader) {
		n = copy(b, c.recordHeader)
		c.recordHeader = nil
		if len(b) > n {
			var n1 = 0
			n1, err = c.Conn.Read(b[n:])
			n += n1
		}
		return n, err
	} else {
		p := c.recordHeader[:len(b)]
		n = len(b)
		copy(b, p)
		c.recordHeader = c.recordHeader[len(b):]
		if len(c.recordHeader) == 0 {
			c.recordHeader = nil
		}
		return n, nil
	}
}

func (c *autoProtocolConn) Handshake() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.handshaked {
		return nil
	}

	timeout := c.handshakeTimeout
	if timeout == 0 {
		timeout = 15 * time.Second
	}
	c.Conn.SetReadDeadline(time.Now().Add(timeout))
	err := c.readFirstHeader()
	c.Conn.SetReadDeadline(time.Time{})

	if err != nil {
		return err
	}

	protocol := detectProtocolByVersion(c.major, c.minor)

	if protocol == ProtocolTLCP && c.tlcpConfig != nil {
		tlcpConn := tlcp.Server(c, c.tlcpConfig)
		if err := tlcpConn.Handshake(); err != nil {
			return err
		}
		c.conn = tlcpConn
	} else if c.tlsConfig != nil {
		tlsConn := tls.Server(c, c.tlsConfig)
		if err := tlsConn.Handshake(); err != nil {
			return err
		}
		c.conn = tlsConn
	}

	c.handshaked = true
	return nil
}

func (c *autoProtocolConn) readFirstHeader() error {
	c.recordHeader = make([]byte, 5)
	_, err := io.ReadFull(c.Conn, c.recordHeader)
	c.major, c.minor = c.recordHeader[1], c.recordHeader[2]
	return err
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

func detectProtocolByVersion(major, minor uint8) ProtocolType {
	version := uint16(major)<<8 | uint16(minor)

	if version == 0x0101 {
		return ProtocolTLCP
	}
	if version == 0x0301 || version == 0x0302 || version == 0x0303 || version == 0x0304 {
		return ProtocolTLS
	}

	return ProtocolTLS
}

func detectProtocol(data []byte) ProtocolType {
	if len(data) < 5 {
		return ProtocolTLS
	}

	major, minor := data[1], data[2]
	return detectProtocolByVersion(major, minor)
}

type HealthCheckResult struct {
	Protocol string `json:"protocol"`
	Success  bool   `json:"success"`
	Latency  int64  `json:"latencyMs"`
	Error    string `json:"error,omitempty"`
}

func (a *TLCPAdapter) CheckHealth(protocol ProtocolType, timeout time.Duration) *HealthCheckResult {
	start := time.Now()
	result := &HealthCheckResult{
		Protocol: protocol.String(),
		Success:  false,
	}

	targetAddr := a.cfg.Target
	if targetAddr == "" {
		result.Error = "目标地址未配置"
		return result
	}

	dialer := &net.Dialer{
		Timeout: timeout,
	}

	var conn net.Conn
	var err error

	switch protocol {
	case ProtocolTLCP:
		conn, err = a.checkTLCPHealth(dialer, targetAddr)
	case ProtocolTLS:
		conn, err = a.checkTLSHealth(dialer, targetAddr)
	default:
		result.Error = "不支持的协议类型"
		return result
	}

	latency := time.Since(start).Milliseconds()
	result.Latency = latency

	if err != nil {
		result.Error = err.Error()
		return result
	}

	if conn != nil {
		conn.Close()
	}

	result.Success = true
	return result
}

func (a *TLCPAdapter) checkTLCPHealth(dialer *net.Dialer, targetAddr string) (net.Conn, error) {
	cfg := a.buildHealthCheckTLCPConfig()
	return tlcp.DialWithDialer(dialer, "tcp", targetAddr, cfg)
}

func (a *TLCPAdapter) checkTLSHealth(dialer *net.Dialer, targetAddr string) (net.Conn, error) {
	cfg := a.buildHealthCheckTLSConfig()
	return tls.DialWithDialer(dialer, "tcp", targetAddr, cfg)
}

func (a *TLCPAdapter) buildHealthCheckTLCPConfig() *tlcp.Config {
	cfg := &tlcp.Config{
		InsecureSkipVerify: a.cfg.TLCP.InsecureSkipVerify,
	}

	if a.cfg.SNI != "" {
		cfg.ServerName = a.cfg.SNI
	}

	tlcpAuth := a.cfg.TLCP.Auth
	if tlcpAuth == "" {
		tlcpAuth = string(config.AuthNone)
	}

	if tlcpAuth == string(config.AuthMutual) {
		var ks security.KeyStore
		var err error

		if a.tlcpKeyStore != nil {
			ks = a.tlcpKeyStore
		} else if a.cfg.TLCP.Keystore != nil {
			ks, err = a.loadKeyStoreFromConfig(a.cfg.TLCP.Keystore, a.cfg.Name+"-tlcp")
		}

		if ks == nil && a.tlsKeyStore != nil {
			ks = a.tlsKeyStore
		}

		if ks != nil && err == nil {
			cfg.GetClientCertificate = func(*tlcp.CertificateRequestInfo) (*tlcp.Certificate, error) {
				return ks.TLCPCertificate()
			}
		}
	}

	if len(a.cfg.ServerCA) > 0 {
		if a.rootCertPool != nil {
			cfg.RootCAs = a.rootCertPool.GetSMCertPool()
		} else {
			pool := a.rootCertManager.GetPool()
			cfg.RootCAs = pool.GetSMCertPool()
		}
	}

	a.applyTLCPSettingsToConfig(cfg, &a.cfg.TLCP)
	return cfg
}

func (a *TLCPAdapter) buildHealthCheckTLSConfig() *tls.Config {
	cfg := &tls.Config{
		InsecureSkipVerify: a.cfg.TLS.InsecureSkipVerify,
	}

	if a.cfg.SNI != "" {
		cfg.ServerName = a.cfg.SNI
	}

	tlsAuth := a.cfg.TLS.Auth
	if tlsAuth == "" {
		tlsAuth = string(config.AuthNone)
	}

	if tlsAuth == string(config.AuthMutual) {
		var ks security.KeyStore
		var err error

		if a.tlsKeyStore != nil {
			ks = a.tlsKeyStore
		} else if a.cfg.TLS.Keystore != nil {
			ks, err = a.loadKeyStoreFromConfig(a.cfg.TLS.Keystore, a.cfg.Name+"-tls")
		}

		if ks == nil && a.tlcpKeyStore != nil {
			ks = a.tlcpKeyStore
		}

		if ks != nil && err == nil {
			cfg.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
				return ks.TLSCertificate()
			}
		}
	}

	if len(a.cfg.ServerCA) > 0 {
		if a.rootCertPool != nil {
			cfg.RootCAs = a.rootCertPool.GetCertPool()
		} else {
			pool := a.rootCertManager.GetPool()
			cfg.RootCAs = pool.GetCertPool()
		}
	}

	a.applyTLSSettingsToConfig(cfg, &a.cfg.TLS)
	return cfg
}
