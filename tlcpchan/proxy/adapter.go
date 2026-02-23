package proxy

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"gitee.com/Trisia/gotlcp/pa"
	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/logger"
	"github.com/Trisia/tlcpchan/security"
	"github.com/Trisia/tlcpchan/security/keystore"
	"github.com/Trisia/tlcpchan/stats"
)

func detectProtocol(data []byte) ProtocolType {
	if len(data) < 5 {
		return ProtocolTLS
	}

	if data[0] == 0x16 && data[1] == 0x01 {
		return ProtocolTLCP
	}

	return ProtocolTLS
}

func needEncKeypair(suites []uint16) bool {
	for _, suite := range suites {
		if suite == 0xC011 || suite == 0xC012 || suite == 0xC013 ||
			suite == 0xC014 || suite == 0xC019 || suite == 0xC01A {
			return true
		}
	}
	return false
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
	mu               sync.RWMutex
	protocol         ProtocolType
	tlcpConfig       *tlcp.Config
	tlsConfig        *tls.Config
	outerTLCPConfig  *tlcp.Config
	outerTLSConfig   *tls.Config
	atomicTLCPConfig atomic.Value
	atomicTLSConfig  atomic.Value
	tlcpKeyStore     security.KeyStore
	tlsKeyStore      security.KeyStore
	keyStoreManager  *security.KeyStoreManager
	rootCertManager  *security.RootCertManager
	stats            *stats.Collector
	logger           *logger.Logger
}

func NewTLCPAdapter(
	keyStoreMgr *security.KeyStoreManager,
	rootCertMgr *security.RootCertManager,
) (*TLCPAdapter, error) {
	return &TLCPAdapter{
		keyStoreManager: keyStoreMgr,
		rootCertManager: rootCertMgr,
		stats:           stats.DefaultCollector(),
		logger:          logger.Default(),
	}, nil
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

	if string(ksConfig.Type) == string(keystore.LoaderTypeNamed) {
		return a.keyStoreManager.LoadFromConfig(ksConfig.Params["name"])
	}

	return a.keyStoreManager.LoadAndRegister(ksConfig.Name, suggestedName, string(ksConfig.Type), ksConfig.Params)
}

func (a *TLCPAdapter) TLCPListener(l net.Listener) net.Listener {
	return tlcp.NewListener(l, a.outerTLCPConfig)
}

func (a *TLCPAdapter) TLSListener(l net.Listener) net.Listener {
	return tls.NewListener(l, a.outerTLSConfig)
}

func (a *TLCPAdapter) AutoListener(l net.Listener) net.Listener {
	return pa.NewListener(l, a.outerTLCPConfig, a.outerTLSConfig)
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

func (a *TLCPAdapter) getTimeoutConfig(cfg *config.InstanceConfig) *config.TimeoutConfig {
	if cfg.Timeout != nil {
		return cfg.Timeout
	}
	return config.DefaultTimeout()
}

func (a *TLCPAdapter) DialTLCP(network, addr string, cfg *config.InstanceConfig) (net.Conn, error) {
	timeout := a.getTimeoutConfig(cfg).Dial
	dialer := &net.Dialer{
		Timeout: timeout,
	}
	tlcpConfig := a.atomicTLCPConfig.Load().(*tlcp.Config)
	return tlcp.DialWithDialer(dialer, network, addr, tlcpConfig)
}

func (a *TLCPAdapter) DialTLS(network, addr string, cfg *config.InstanceConfig) (net.Conn, error) {
	timeout := a.getTimeoutConfig(cfg).Dial
	dialer := &net.Dialer{
		Timeout: timeout,
	}
	tlsConfig := a.atomicTLSConfig.Load().(*tls.Config)
	return tls.DialWithDialer(dialer, network, addr, tlsConfig)
}

func (a *TLCPAdapter) DialWithProtocol(network, addr string, protocol ProtocolType, cfg *config.InstanceConfig) (net.Conn, error) {
	switch protocol {
	case ProtocolTLCP:
		return a.DialTLCP(network, addr, cfg)
	case ProtocolTLS:
		return a.DialTLS(network, addr, cfg)
	default:
		return a.autoDial(network, addr, cfg)
	}
}

func (a *TLCPAdapter) autoDial(network, addr string, cfg *config.InstanceConfig) (net.Conn, error) {
	conn, err := a.DialTLCP(network, addr, cfg)
	if err == nil {
		return conn, nil
	}

	a.logger.Debug("TLCP连接失败，尝试TLS: %v", err)
	return a.DialTLS(network, addr, cfg)
}

func (a *TLCPAdapter) Protocol() ProtocolType {
	a.mu.RLock()
	defer a.mu.RUnlock()
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
	a.protocol = ParseProtocolType(cfg.Protocol)
	a.mu.Unlock()

	if cfg.Type == TypeServer || cfg.Type == TypeHTTPServer {
		return a.reloadServerConfig(cfg)
	}
	return a.reloadClientConfig(cfg)
}

func (a *TLCPAdapter) reloadServerConfig(cfg *config.InstanceConfig) error {
	var tlcpConfig *tlcp.Config
	var tlsConfig *tls.Config
	var rootCertPool security.RootCertPool

	if newTLCPKS, err := a.loadKeyStoreFromConfig(cfg.TLCP.Keystore, cfg.Name+"-tlcp"); err == nil {
		a.tlcpKeyStore = newTLCPKS
	} else {
		a.tlcpKeyStore = nil
	}

	if newTLSKS, err := a.loadKeyStoreFromConfig(cfg.TLS.Keystore, cfg.Name+"-tls"); err == nil {
		a.tlsKeyStore = newTLSKS
	} else {
		a.tlsKeyStore = nil
	}

	if len(cfg.ClientCA) > 0 {
		rootCertPool = a.rootCertManager.GetPool()
	}

	if a.tlcpKeyStore != nil {
		tlcpConfig = &tlcp.Config{}
		certs, err := a.tlcpKeyStore.TLCPCertificate()
		if err != nil {
			return err
		}
		if len(certs) == 0 {
			return fmt.Errorf("TLCP证书不能为空")
		}
		tlcpConfig.Certificates = make([]tlcp.Certificate, len(certs))
		for i, cert := range certs {
			tlcpConfig.Certificates[i] = *cert
		}

		tlcpConfig.ClientAuth, _ = config.ParseTLCPClientAuth(cfg.TLCP.ClientAuthType)
		if rootCertPool != nil {
			tlcpConfig.ClientCAs = rootCertPool.GetSMCertPool()
		}

		if len(cfg.TLCP.CipherSuites) > 0 {
			suites, _ := config.ParseCipherSuites(cfg.TLCP.CipherSuites, true)
			tlcpConfig.CipherSuites = suites
		}

		if cfg.TLCP.MinVersion != "" {
			v, _ := config.ParseTLSVersion(cfg.TLCP.MinVersion, true)
			tlcpConfig.MinVersion = v
		}

		if cfg.TLCP.MaxVersion != "" {
			v, _ := config.ParseTLSVersion(cfg.TLCP.MaxVersion, true)
			tlcpConfig.MaxVersion = v
		}

		if cfg.TLCP.SessionCache {
			tlcpConfig.SessionCache = tlcp.NewLRUSessionCache(100)
		}
	}

	if a.tlsKeyStore != nil {
		tlsConfig = &tls.Config{}
		cert, err := a.tlsKeyStore.TLSCertificate()
		if err != nil {
			return err
		}
		tlsConfig.Certificates = []tls.Certificate{*cert}

		tlsConfig.ClientAuth, _ = config.ParseTLSClientAuth(cfg.TLS.ClientAuthType)
		if rootCertPool != nil {
			tlsConfig.ClientCAs = rootCertPool.GetCertPool()
		}

		if len(cfg.TLS.CipherSuites) > 0 {
			suites, _ := config.ParseCipherSuites(cfg.TLS.CipherSuites, false)
			tlsConfig.CipherSuites = suites
		}

		if cfg.TLS.MinVersion != "" {
			v, _ := config.ParseTLSVersion(cfg.TLS.MinVersion, false)
			tlsConfig.MinVersion = v
		}

		if cfg.TLS.MaxVersion != "" {
			v, _ := config.ParseTLSVersion(cfg.TLS.MaxVersion, false)
			tlsConfig.MaxVersion = v
		}

		if cfg.TLS.SessionCache {
			tlsConfig.ClientSessionCache = tls.NewLRUClientSessionCache(100)
		}
	}

	a.mu.Lock()

	a.tlcpConfig = tlcpConfig
	a.tlsConfig = tlsConfig

	if a.outerTLCPConfig == nil && tlcpConfig != nil {
		a.outerTLCPConfig = &tlcp.Config{
			GetConfigForClient: func(hello *tlcp.ClientHelloInfo) (*tlcp.Config, error) {
				return a.atomicTLCPConfig.Load().(*tlcp.Config), nil
			},
		}
	}

	if a.outerTLSConfig == nil && tlsConfig != nil {
		a.outerTLSConfig = &tls.Config{
			GetConfigForClient: func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
				return a.atomicTLSConfig.Load().(*tls.Config), nil
			},
		}
	}

	a.atomicTLCPConfig.Store(tlcpConfig)
	a.atomicTLSConfig.Store(tlsConfig)

	a.mu.Unlock()

	return nil
}

func (a *TLCPAdapter) reloadClientConfig(cfg *config.InstanceConfig) error {
	var tlcpConfig *tlcp.Config
	var tlsConfig *tls.Config
	var rootCertPool security.RootCertPool

	if newTLCPKS, err := a.loadKeyStoreFromConfig(cfg.TLCP.Keystore, cfg.Name+"-tlcp"); err == nil {
		a.tlcpKeyStore = newTLCPKS
	} else {
		a.tlcpKeyStore = nil
	}

	if newTLSKS, err := a.loadKeyStoreFromConfig(cfg.TLS.Keystore, cfg.Name+"-tls"); err == nil {
		a.tlsKeyStore = newTLSKS
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
		certs, err := a.tlcpKeyStore.TLCPCertificate()
		if err != nil {
			return err
		}
		if len(certs) == 0 {
			return fmt.Errorf("TLCP证书不能为空")
		}
		tlcpConfig.Certificates = make([]tlcp.Certificate, len(certs))
		for i, cert := range certs {
			tlcpConfig.Certificates[i] = *cert
		}
	}
	if rootCertPool != nil {
		tlcpConfig.RootCAs = rootCertPool.GetSMCertPool()
	}
	if cfg.SNI != "" {
		tlcpConfig.ServerName = cfg.SNI
	}
	if a.tlcpKeyStore != nil {
		certs, err := a.tlcpKeyStore.TLCPCertificate()
		if err != nil {
			return err
		}
		if len(certs) == 0 {
			return fmt.Errorf("TLCP证书不能为空")
		}
		tlcpConfig.GetClientCertificate = func(*tlcp.CertificateRequestInfo) (*tlcp.Certificate, error) {
			return certs[0], nil
		}
		if len(certs) > 1 {
			tlcpConfig.GetClientKECertificate = func(*tlcp.CertificateRequestInfo) (*tlcp.Certificate, error) {
				return certs[1], nil
			}
		}
	}
	if rootCertPool != nil {
		tlcpConfig.RootCAs = rootCertPool.GetSMCertPool()
	}

	if len(cfg.TLCP.CipherSuites) > 0 {
		suites, _ := config.ParseCipherSuites(cfg.TLCP.CipherSuites, true)
		tlcpConfig.CipherSuites = suites
	}

	if cfg.TLCP.MinVersion != "" {
		v, _ := config.ParseTLSVersion(cfg.TLCP.MinVersion, true)
		tlcpConfig.MinVersion = v
	}

	if cfg.TLCP.MaxVersion != "" {
		v, _ := config.ParseTLSVersion(cfg.TLCP.MaxVersion, true)
		tlcpConfig.MaxVersion = v
	}

	if cfg.TLCP.SessionCache {
		tlcpConfig.SessionCache = tlcp.NewLRUSessionCache(100)
	}

	tlsConfig = &tls.Config{
		InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
	}
	if cfg.SNI != "" {
		tlsConfig.ServerName = cfg.SNI
	}
	if a.tlsKeyStore != nil {
		cert, err := a.tlsKeyStore.TLSCertificate()
		if err != nil {
			return err
		}
		tlsConfig.Certificates = []tls.Certificate{*cert}
	}
	if rootCertPool != nil {
		tlsConfig.RootCAs = rootCertPool.GetCertPool()
	}

	if len(cfg.TLS.CipherSuites) > 0 {
		suites, _ := config.ParseCipherSuites(cfg.TLS.CipherSuites, false)
		tlsConfig.CipherSuites = suites
	}

	if cfg.TLS.MinVersion != "" {
		v, _ := config.ParseTLSVersion(cfg.TLS.MinVersion, false)
		tlsConfig.MinVersion = v
	}

	if cfg.TLS.MaxVersion != "" {
		v, _ := config.ParseTLSVersion(cfg.TLS.MaxVersion, false)
		tlsConfig.MaxVersion = v
	}

	if cfg.TLS.SessionCache {
		tlsConfig.ClientSessionCache = tls.NewLRUClientSessionCache(100)
	}

	a.mu.Lock()

	a.tlcpConfig = tlcpConfig
	a.tlsConfig = tlsConfig

	a.atomicTLCPConfig.Store(tlcpConfig)
	a.atomicTLSConfig.Store(tlsConfig)

	a.mu.Unlock()

	return nil
}

type HealthCheckResult struct {
	Protocol string `json:"protocol"`
	Success  bool   `json:"success"`
	Latency  int64  `json:"latencyMs"`
	Error    string `json:"error,omitempty"`
}

func (a *TLCPAdapter) CheckHealth(protocol ProtocolType, timeout time.Duration, targetAddr string) *HealthCheckResult {
	start := time.Now()
	result := &HealthCheckResult{
		Protocol: protocol.String(),
		Success:  false,
	}

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
	case ProtocolAuto:
		conn, err = a.checkTLCPHealth(dialer, targetAddr)
		if err != nil {
			a.logger.Debug("TLCP健康检查失败，尝试TLS: %v", err)
			conn, err = a.checkTLSHealth(dialer, targetAddr)
		}
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
	baseConfig := a.atomicTLCPConfig.Load().(*tlcp.Config)
	if baseConfig == nil {
		return nil, fmt.Errorf("TLCP配置未初始化")
	}

	healthConfig := baseConfig.Clone()
	healthConfig.InsecureSkipVerify = true

	return tlcp.DialWithDialer(dialer, "tcp", targetAddr, healthConfig)
}

func (a *TLCPAdapter) checkTLSHealth(dialer *net.Dialer, targetAddr string) (net.Conn, error) {
	baseConfig := a.atomicTLSConfig.Load().(*tls.Config)
	if baseConfig == nil {
		return nil, fmt.Errorf("TLS配置未初始化")
	}

	healthConfig := baseConfig.Clone()
	healthConfig.InsecureSkipVerify = true

	return tls.DialWithDialer(dialer, "tcp", targetAddr, healthConfig)
}
