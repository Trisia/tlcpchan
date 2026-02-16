package proxy

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/Trisia/tlcpchan/cert"
	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/logger"
)

// ECDHE密码套件ID列表
var ecdheCipherSuites = map[uint16]bool{
	0xC013: true, // ECDHE_SM4_CBC_SM3
	0xC014: true, // ECDHE_SM4_GCM_SM3
	0xC01A: true, // ECDHE_SM4_CCM_SM3
}

// isECDHECipherSuite 检查是否为ECDHE密码套件
// 参数:
//   - suites: 密码套件列表
//
// 返回:
//   - bool: 如果列表中包含任何ECDHE密码套件则返回true
func isECDHECipherSuite(suites []uint16) bool {
	for _, s := range suites {
		if ecdheCipherSuites[s] {
			return true
		}
	}
	return false
}

// ValidateClientConfig 验证客户端配置
// 参数:
//   - cfg: 实例配置
//
// 返回:
//   - error: 配置无效时返回错误
//
// 注意: ECDHE密码套件只能在双向身份认证下使用
func ValidateClientConfig(cfg *config.InstanceConfig) error {
	// 获取TLCP认证模式
	tlcpAuth := cfg.TLCP.Auth
	if tlcpAuth == "" {
		tlcpAuth = string(config.AuthNone)
	}

	// TLCP ECDHE密码套件必须使用双向认证
	if cfg.Protocol == string(config.ProtocolTLCP) || cfg.Protocol == string(config.ProtocolAuto) {
		suites, _ := config.ParseCipherSuites(cfg.TLCP.CipherSuites, true)
		if len(suites) > 0 && isECDHECipherSuite(suites) && tlcpAuth != string(config.AuthMutual) {
			return fmt.Errorf("ECDHE密码套件只能在双向身份认证下使用，当前认证模式: %s", tlcpAuth)
		}

		// 双向认证必须提供签名证书
		if tlcpAuth == string(config.AuthMutual) {
			if cfg.Certs.TLCP.Cert == "" || cfg.Certs.TLCP.Key == "" {
				return fmt.Errorf("双向认证模式下必须配置TLCP签名证书和密钥")
			}
		}
	}

	return nil
}

// ProtocolType 协议类型
type ProtocolType int

const (
	// ProtocolAuto 自动检测协议类型，同时支持TLCP和TLS
	ProtocolAuto ProtocolType = iota
	// ProtocolTLCP 国密TLCP协议
	ProtocolTLCP
	// ProtocolTLS 标准TLS协议
	ProtocolTLS
)

// InstanceType 实例类型（避免循环导入）
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

// TLCPAdapter 协议适配器，负责TLCP/TLS协议的配置和连接处理
// 支持配置热加载，无需重启服务即可更新证书和配置
type TLCPAdapter struct {
	mu           sync.RWMutex
	protocol     ProtocolType
	tlcpConfig   *tlcp.Config
	tlsConfig    *tls.Config
	clientCAPool *cert.HotCertPool
	serverCAPool *cert.HotCertPool
	tlcpCertRef  *cert.Certificate
	tlsCertRef   *cert.Certificate
	cfg          *config.InstanceConfig
	certManager  *cert.Manager
	logger       *logger.Logger
}

// NewTLCPAdapter 创建新的协议适配器
// 参数:
//   - cfg: 实例配置，包含协议类型、证书配置等
//   - certManager: 证书管理器
//
// 返回:
//   - *TLCPAdapter: 协议适配器实例
//   - error: 初始化配置失败时返回错误
//
// 注意: 根据实例类型(server/client)自动初始化相应配置
func NewTLCPAdapter(cfg *config.InstanceConfig, certManager *cert.Manager) (*TLCPAdapter, error) {
	adapter := &TLCPAdapter{
		protocol:    ParseProtocolType(cfg.Protocol),
		cfg:         cfg,
		certManager: certManager,
		logger:      logger.Default(),
	}

	if cfg.Type == TypeServer || cfg.Type == TypeHTTPServer {
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

// updateConfig 更新当前配置
func (a *TLCPAdapter) updateConfig(tlcpCfg *tlcp.Config, tlsCfg *tls.Config, clientCAPool *cert.HotCertPool, serverCAPool *cert.HotCertPool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.tlcpConfig = tlcpCfg
	a.tlsConfig = tlsCfg
	a.clientCAPool = clientCAPool
	a.serverCAPool = serverCAPool
}

// getTLCPConfig 获取当前TLCP配置
func (a *TLCPAdapter) getTLCPConfig() *tlcp.Config {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.tlcpConfig
}

// getTLSConfig 获取当前TLS配置
func (a *TLCPAdapter) getTLSConfig() *tls.Config {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.tlsConfig
}

// initServerConfig 初始化服务端配置
func (a *TLCPAdapter) initServerConfig(cfg *config.InstanceConfig, certManager *cert.Manager) error {
	var tlcpConfig *tlcp.Config
	var tlsConfig *tls.Config
	var clientCAPool *cert.HotCertPool

	// 获取认证模式，优先使用协议特定配置
	tlcpAuth := cfg.TLCP.Auth
	if tlcpAuth == "" {
		tlcpAuth = string(config.AuthNone)
	}
	tlsAuth := cfg.TLS.Auth
	if tlsAuth == "" {
		tlsAuth = string(config.AuthNone)
	}

	// TLCP服务端配置
	if tlcpAuth == string(config.AuthOneWay) || tlcpAuth == string(config.AuthMutual) {
		if cfg.Certs.TLCP.Cert != "" && cfg.Certs.TLCP.Key != "" {
			tlcpCert, err := certManager.LoadTLCP(cfg.Name+"-tlcp", cfg.Certs.TLCP.Cert, cfg.Certs.TLCP.Key)
			if err != nil {
				return err
			}
			a.tlcpCertRef = tlcpCert.Cert
		}
	}

	// TLS服务端配置
	if tlsAuth == string(config.AuthOneWay) || tlsAuth == string(config.AuthMutual) {
		if cfg.Certs.TLS.Cert != "" && cfg.Certs.TLS.Key != "" {
			tlsCert, err := certManager.LoadTLS(cfg.Name+"-tls", cfg.Certs.TLS.Cert, cfg.Certs.TLS.Key)
			if err != nil {
				return err
			}
			a.tlsCertRef = tlsCert.Cert
		}
	}

	// 创建可热重载的ClientCA池
	if len(cfg.ClientCA) > 0 {
		clientCAPool = cert.NewHotCertPool(cfg.ClientCA)
		if err := clientCAPool.Load(); err != nil {
			return err
		}
	}

	// 构建TLCP配置
	if a.tlcpCertRef != nil {
		tlcpConfig = &tlcp.Config{}
		tlcpConfig.GetConfigForClient = func(chi *tlcp.ClientHelloInfo) (*tlcp.Config, error) {
			cfgCopy := a.buildTLCPServerConfig(clientCAPool, tlcpAuth)
			return cfgCopy, nil
		}
	}

	// 构建TLS配置
	if a.tlsCertRef != nil {
		tlsConfig = &tls.Config{}
		tlsConfig.GetConfigForClient = func(chi *tls.ClientHelloInfo) (*tls.Config, error) {
			cfgCopy := a.buildTLSServerConfig(clientCAPool, tlsAuth)
			return cfgCopy, nil
		}
	}

	// 更新当前配置
	a.updateConfig(tlcpConfig, tlsConfig, clientCAPool, nil)

	return nil
}

// buildTLCPServerConfig 构建TLCP服务端配置
func (a *TLCPAdapter) buildTLCPServerConfig(clientCAPool *cert.HotCertPool, auth string) *tlcp.Config {
	cfg := &tlcp.Config{}

	if a.tlcpCertRef != nil {
		cfg.GetCertificate = a.tlcpCertRef.GetTLCPCertificate
	}

	if auth == string(config.AuthMutual) && clientCAPool != nil {
		cfg.ClientCAs = clientCAPool.SMPool()
		cfg.ClientAuth = tlcp.RequireAndVerifyClientCert
	} else if auth == string(config.AuthOneWay) {
		cfg.ClientAuth = tlcp.NoClientCert
	}

	a.applyTLCPSettingsToConfig(cfg, &a.cfg.TLCP)
	return cfg
}

// buildTLSServerConfig 构建TLS服务端配置
func (a *TLCPAdapter) buildTLSServerConfig(clientCAPool *cert.HotCertPool, auth string) *tls.Config {
	cfg := &tls.Config{}

	if a.tlsCertRef != nil {
		cfg.GetCertificate = a.tlsCertRef.GetCertificate
	}

	if auth == string(config.AuthMutual) && clientCAPool != nil {
		cfg.ClientCAs = clientCAPool.Pool()
		cfg.ClientAuth = tls.RequireAndVerifyClientCert
	} else if auth == string(config.AuthOneWay) {
		cfg.ClientAuth = tls.NoClientCert
	}

	a.applyTLSSettingsToConfig(cfg, &a.cfg.TLS)
	return cfg
}

// initClientConfig 初始化客户端配置
func (a *TLCPAdapter) initClientConfig(cfg *config.InstanceConfig, certManager *cert.Manager) error {
	var tlcpConfig *tlcp.Config
	var tlsConfig *tls.Config
	var serverCAPool *cert.HotCertPool

	// 获取认证模式，优先使用协议特定配置
	tlcpAuth := cfg.TLCP.Auth
	if tlcpAuth == "" {
		tlcpAuth = string(config.AuthNone)
	}
	tlsAuth := cfg.TLS.Auth
	if tlsAuth == "" {
		tlsAuth = string(config.AuthNone)
	}

	// TLCP客户端配置
	tlcpConfig = &tlcp.Config{
		InsecureSkipVerify: cfg.TLCP.InsecureSkipVerify,
	}

	if cfg.SNI != "" {
		tlcpConfig.ServerName = cfg.SNI
	}

	// 使用GetClientCertificate支持证书热加载
	if tlcpAuth == string(config.AuthMutual) && cfg.Certs.TLCP.Cert != "" && cfg.Certs.TLCP.Key != "" {
		tlcpCert, err := certManager.LoadTLCP(cfg.Name+"-tlcp", cfg.Certs.TLCP.Cert, cfg.Certs.TLCP.Key)
		if err != nil {
			return err
		}
		a.tlcpCertRef = tlcpCert.Cert
		tlcpConfig.GetClientCertificate = func(*tlcp.CertificateRequestInfo) (*tlcp.Certificate, error) {
			cert := a.tlcpCertRef.TLCPCertificate()
			return &cert, nil
		}
	}

	// TLS客户端配置
	tlsConfig = &tls.Config{
		InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
	}

	if cfg.SNI != "" {
		tlsConfig.ServerName = cfg.SNI
	}

	// 使用GetClientCertificate支持证书热加载
	if tlsAuth == string(config.AuthMutual) && cfg.Certs.TLS.Cert != "" && cfg.Certs.TLS.Key != "" {
		tlsCert, err := certManager.LoadTLS(cfg.Name+"-tls", cfg.Certs.TLS.Cert, cfg.Certs.TLS.Key)
		if err != nil {
			return err
		}
		a.tlsCertRef = tlsCert.Cert
		tlsConfig.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			cert := a.tlsCertRef.TLSCertificate()
			return &cert, nil
		}
	}

	// 创建可热重载的ServerCA池
	if len(cfg.ServerCA) > 0 {
		serverCAPool = cert.NewHotCertPool(cfg.ServerCA)
		if err := serverCAPool.Load(); err != nil {
			return err
		}
		tlcpConfig.RootCAs = serverCAPool.SMPool()
		tlsConfig.RootCAs = serverCAPool.Pool()
	}

	a.applyTLCPSettingsToConfig(tlcpConfig, &cfg.TLCP)
	a.applyTLSSettingsToConfig(tlsConfig, &cfg.TLS)

	// 更新当前配置
	a.updateConfig(tlcpConfig, tlsConfig, nil, serverCAPool)

	return nil
}

// applyTLCPSettingsToConfig 应用TLCP设置到指定配置
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

// applyTLSSettingsToConfig 应用TLS设置到指定配置
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

// TLCPListener 创建TLCP监听器
func (a *TLCPAdapter) TLCPListener(l net.Listener) net.Listener {
	return tlcp.NewListener(l, a.getTLCPConfig())
}

// TLSListener 创建TLS监听器
func (a *TLCPAdapter) TLSListener(l net.Listener) net.Listener {
	return tls.NewListener(l, a.getTLSConfig())
}

// AutoListener 创建自动协议检测监听器
func (a *TLCPAdapter) AutoListener(l net.Listener) net.Listener {
	return NewAutoProtocolListener(l, a.getTLCPConfig(), a.getTLSConfig())
}

// WrapServerListener 包装服务端监听器，根据协议类型返回对应监听器
// 参数:
//   - l: 原始TCP监听器
//
// 返回:
//   - net.Listener: 包装后的监听器（TLCP/TLS/Auto）
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

// DialTLCP 使用TLCP协议连接
func (a *TLCPAdapter) DialTLCP(network, addr string) (net.Conn, error) {
	return tlcp.Dial(network, addr, a.getTLCPConfig())
}

// DialTLS 使用TLS协议连接
func (a *TLCPAdapter) DialTLS(network, addr string) (net.Conn, error) {
	return tls.Dial(network, addr, a.getTLSConfig())
}

// DialWithProtocol 使用指定协议连接目标服务
// 参数:
//   - network: 网络类型，如 "tcp"
//   - addr: 目标地址，格式 "host:port"
//   - protocol: 协议类型
//
// 返回:
//   - net.Conn: 已建立协议连接
//   - error: 连接失败时返回错误
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

// ReloadCertificates 重载所有证书（包括端证书和CA证书池）
// 返回:
//   - error: 重载失败时返回错误
func (a *TLCPAdapter) ReloadCertificates() error {
	a.mu.RLock()
	clientCAPool := a.clientCAPool
	serverCAPool := a.serverCAPool
	a.mu.RUnlock()

	var errs []error

	// 重载端证书
	if a.tlcpCertRef != nil {
		if err := a.tlcpCertRef.ReloadFromPath(); err != nil {
			errs = append(errs, fmt.Errorf("重载TLCP证书失败: %w", err))
		}
	}
	if a.tlsCertRef != nil {
		if err := a.tlsCertRef.ReloadFromPath(); err != nil {
			errs = append(errs, fmt.Errorf("重载TLS证书失败: %w", err))
		}
	}

	// 重载CA证书池
	if clientCAPool != nil {
		if err := clientCAPool.Reload(); err != nil {
			errs = append(errs, fmt.Errorf("重载ClientCA池失败: %w", err))
		}
	}
	if serverCAPool != nil {
		if err := serverCAPool.Reload(); err != nil {
			errs = append(errs, fmt.Errorf("重载ServerCA池失败: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("部分证书重载失败: %v", errs)
	}

	return nil
}

// ReloadConfig 完全重载配置（用于配置变更时）
// 参数:
//   - cfg: 新的实例配置
//
// 返回:
//   - error: 重载失败时返回错误
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
	var clientCAPool *cert.HotCertPool

	tlcpAuth := cfg.TLCP.Auth
	if tlcpAuth == "" {
		tlcpAuth = string(config.AuthNone)
	}
	tlsAuth := cfg.TLS.Auth
	if tlsAuth == "" {
		tlsAuth = string(config.AuthNone)
	}

	// 如果证书路径变化，重新加载证书
	if cfg.Certs.TLCP.Cert != oldCfg.Certs.TLCP.Cert || cfg.Certs.TLCP.Key != oldCfg.Certs.TLCP.Key {
		if cfg.Certs.TLCP.Cert != "" && cfg.Certs.TLCP.Key != "" {
			tlcpCert, err := a.certManager.LoadTLCP(cfg.Name+"-tlcp", cfg.Certs.TLCP.Cert, cfg.Certs.TLCP.Key)
			if err != nil {
				return err
			}
			a.tlcpCertRef = tlcpCert.Cert
		} else {
			a.tlcpCertRef = nil
		}
	}

	if cfg.Certs.TLS.Cert != oldCfg.Certs.TLS.Cert || cfg.Certs.TLS.Key != oldCfg.Certs.TLS.Key {
		if cfg.Certs.TLS.Cert != "" && cfg.Certs.TLS.Key != "" {
			tlsCert, err := a.certManager.LoadTLS(cfg.Name+"-tls", cfg.Certs.TLS.Cert, cfg.Certs.TLS.Key)
			if err != nil {
				return err
			}
			a.tlsCertRef = tlsCert.Cert
		} else {
			a.tlsCertRef = nil
		}
	}

	// 重新创建ClientCA池
	if len(cfg.ClientCA) > 0 {
		clientCAPool = cert.NewHotCertPool(cfg.ClientCA)
		if err := clientCAPool.Load(); err != nil {
			return err
		}
	}

	// 构建新的配置
	if a.tlcpCertRef != nil {
		tlcpConfig = &tlcp.Config{}
		tlcpConfig.GetConfigForClient = func(chi *tlcp.ClientHelloInfo) (*tlcp.Config, error) {
			cfgCopy := a.buildTLCPServerConfig(clientCAPool, tlcpAuth)
			return cfgCopy, nil
		}
	}

	if a.tlsCertRef != nil {
		tlsConfig = &tls.Config{}
		tlsConfig.GetConfigForClient = func(chi *tls.ClientHelloInfo) (*tls.Config, error) {
			cfgCopy := a.buildTLSServerConfig(clientCAPool, tlsAuth)
			return cfgCopy, nil
		}
	}

	a.updateConfig(tlcpConfig, tlsConfig, clientCAPool, nil)
	return nil
}

func (a *TLCPAdapter) reloadClientConfig(cfg, oldCfg *config.InstanceConfig) error {
	var tlcpConfig *tlcp.Config
	var tlsConfig *tls.Config
	var serverCAPool *cert.HotCertPool

	tlcpAuth := cfg.TLCP.Auth
	if tlcpAuth == "" {
		tlcpAuth = string(config.AuthNone)
	}
	tlsAuth := cfg.TLS.Auth
	if tlsAuth == "" {
		tlsAuth = string(config.AuthNone)
	}

	// 如果证书路径变化，重新加载证书
	if cfg.Certs.TLCP.Cert != oldCfg.Certs.TLCP.Cert || cfg.Certs.TLCP.Key != oldCfg.Certs.TLCP.Key {
		if cfg.Certs.TLCP.Cert != "" && cfg.Certs.TLCP.Key != "" {
			tlcpCert, err := a.certManager.LoadTLCP(cfg.Name+"-tlcp", cfg.Certs.TLCP.Cert, cfg.Certs.TLCP.Key)
			if err != nil {
				return err
			}
			a.tlcpCertRef = tlcpCert.Cert
		} else {
			a.tlcpCertRef = nil
		}
	}

	if cfg.Certs.TLS.Cert != oldCfg.Certs.TLS.Cert || cfg.Certs.TLS.Key != oldCfg.Certs.TLS.Key {
		if cfg.Certs.TLS.Cert != "" && cfg.Certs.TLS.Key != "" {
			tlsCert, err := a.certManager.LoadTLS(cfg.Name+"-tls", cfg.Certs.TLS.Cert, cfg.Certs.TLS.Key)
			if err != nil {
				return err
			}
			a.tlsCertRef = tlsCert.Cert
		} else {
			a.tlsCertRef = nil
		}
	}

	// 重新创建ServerCA池
	if len(cfg.ServerCA) > 0 {
		serverCAPool = cert.NewHotCertPool(cfg.ServerCA)
		if err := serverCAPool.Load(); err != nil {
			return err
		}
	}

	// 构建新的TLCP配置
	tlcpConfig = &tlcp.Config{
		InsecureSkipVerify: cfg.TLCP.InsecureSkipVerify,
	}
	if cfg.SNI != "" {
		tlcpConfig.ServerName = cfg.SNI
	}
	if a.tlcpCertRef != nil {
		tlcpConfig.GetClientCertificate = func(*tlcp.CertificateRequestInfo) (*tlcp.Certificate, error) {
			cert := a.tlcpCertRef.TLCPCertificate()
			return &cert, nil
		}
	}
	if serverCAPool != nil {
		tlcpConfig.RootCAs = serverCAPool.SMPool()
	}
	a.applyTLCPSettingsToConfig(tlcpConfig, &cfg.TLCP)

	// 构建新的TLS配置
	tlsConfig = &tls.Config{
		InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
	}
	if cfg.SNI != "" {
		tlsConfig.ServerName = cfg.SNI
	}
	if a.tlsCertRef != nil {
		tlsConfig.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			cert := a.tlsCertRef.TLSCertificate()
			return &cert, nil
		}
	}
	if serverCAPool != nil {
		tlsConfig.RootCAs = serverCAPool.Pool()
	}
	a.applyTLSSettingsToConfig(tlsConfig, &cfg.TLS)

	a.updateConfig(tlcpConfig, tlsConfig, nil, serverCAPool)
	return nil
}

// AutoProtocolListener 自动协议检测监听器，根据客户端握手自动识别TLCP或TLS协议
type AutoProtocolListener struct {
	net.Listener
	// tlcpConfig TLCP协议配置
	tlcpConfig *tlcp.Config
	// tlsConfig TLS协议配置
	tlsConfig *tls.Config
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

// autoProtocolConn 自动协议检测连接，支持延迟握手和协议自动识别
type autoProtocolConn struct {
	net.Conn
	// tlcpConfig TLCP协议配置
	tlcpConfig *tlcp.Config
	// tlsConfig TLS协议配置
	tlsConfig *tls.Config
	// peeked 预读的字节数据，用于协议检测
	peeked []byte
	// handshaked 是否已完成握手
	handshaked bool
	// conn 握手后的实际连接（TLCP或TLS）
	conn net.Conn
	mu   sync.Mutex
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

// detectProtocol 根据握手数据检测协议类型
// 参数:
//   - data: 握手数据（至少5字节）
//
// 返回:
//   - ProtocolType: 检测到的协议类型
//
// 注意: TLCP版本号为0x0101，TLS为0x0301~0x0304
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
