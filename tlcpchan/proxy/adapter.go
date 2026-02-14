package proxy

import (
	"crypto/tls"
	"fmt"
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
		tlcpAuth = "none"
	}

	// TLCP ECDHE密码套件必须使用双向认证
	if cfg.Protocol == "tlcp" || cfg.Protocol == "auto" {
		suites, _ := config.ParseCipherSuites(cfg.TLCP.CipherSuites, true)
		if len(suites) > 0 && isECDHECipherSuite(suites) && tlcpAuth != "mutual" {
			return fmt.Errorf("ECDHE密码套件只能在双向身份认证下使用，当前认证模式: %s", tlcpAuth)
		}

		// 双向认证必须提供签名证书
		if tlcpAuth == "mutual" {
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
type TLCPAdapter struct {
	// tlcpConfig TLCP协议配置
	tlcpConfig *tlcp.Config
	// tlsConfig TLS协议配置
	tlsConfig *tls.Config
	// protocol 当前使用的协议类型
	protocol ProtocolType
	// tlcpCertRef TLCP证书引用，用于热加载
	tlcpCertRef *cert.Certificate
	// tlsCertRef TLS证书引用，用于热加载
	tlsCertRef *cert.Certificate
	// certLoader 证书加载器，用于加载预制证书
	certLoader *cert.EmbeddedCertLoader
	logger     *logger.Logger
	mu         sync.RWMutex
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

// loadSMCertPool 加载国密CA证书池
// 参数:
//   - paths: CA证书文件路径列表
//
// 返回:
//   - *smx509.CertPool: 国密证书池
//   - error: 读取或解析证书失败时返回错误
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
	// 获取认证模式，优先使用协议特定配置
	tlcpAuth := cfg.TLCP.Auth
	if tlcpAuth == "" {
		tlcpAuth = "none"
	}
	tlsAuth := cfg.TLS.Auth
	if tlsAuth == "" {
		tlsAuth = "none"
	}

	// TLCP服务端配置
	if tlcpAuth == "one-way" || tlcpAuth == "mutual" {
		if cfg.Certs.TLCP.Cert != "" && cfg.Certs.TLCP.Key != "" {
			tlcpCert, err := certManager.LoadTLCP(cfg.Name+"-tlcp", cfg.Certs.TLCP.Cert, cfg.Certs.TLCP.Key)
			if err != nil {
				return err
			}
			a.tlcpCertRef = tlcpCert.Cert
			a.tlcpConfig = &tlcp.Config{
				GetCertificate: a.tlcpCertRef.GetTLCPCertificate,
			}
		}

		if tlcpAuth == "mutual" && len(cfg.ClientCA) > 0 && a.tlcpConfig != nil {
			pool, err := loadSMCertPool(cfg.ClientCA)
			if err != nil {
				return err
			}
			a.tlcpConfig.ClientCAs = pool
			// 根据Auth字段自动设置ClientAuth
			a.tlcpConfig.ClientAuth = tlcp.RequireAndVerifyClientCert
		} else if tlcpAuth == "one-way" && a.tlcpConfig != nil {
			// 单向认证时不要求客户端证书
			a.tlcpConfig.ClientAuth = tlcp.NoClientCert
		}
	}

	// TLS服务端配置
	if tlsAuth == "one-way" || tlsAuth == "mutual" {
		if cfg.Certs.TLS.Cert != "" && cfg.Certs.TLS.Key != "" {
			tlsCert, err := certManager.LoadTLS(cfg.Name+"-tls", cfg.Certs.TLS.Cert, cfg.Certs.TLS.Key)
			if err != nil {
				return err
			}
			a.tlsCertRef = tlsCert.Cert
			a.tlsConfig = &tls.Config{
				GetCertificate: a.tlsCertRef.GetCertificate,
			}
		}

		if tlsAuth == "mutual" && len(cfg.ClientCA) > 0 && a.tlsConfig != nil {
			pool, err := certManager.Loader().LoadClientCA(cfg.ClientCA)
			if err != nil {
				return err
			}
			a.tlsConfig.ClientCAs = pool
			// 根据Auth字段自动设置ClientAuth
			a.tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		} else if tlsAuth == "one-way" && a.tlsConfig != nil {
			// 单向认证时不要求客户端证书
			a.tlsConfig.ClientAuth = tls.NoClientCert
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

	// 获取认证模式，优先使用协议特定配置
	tlcpAuth := cfg.TLCP.Auth
	if tlcpAuth == "" {
		tlcpAuth = "none"
	}
	tlsAuth := cfg.TLS.Auth
	if tlsAuth == "" {
		tlsAuth = "none"
	}

	// TLCP客户端双向认证
	if tlcpAuth == "mutual" {
		if cfg.Certs.TLCP.Cert != "" && cfg.Certs.TLCP.Key != "" {
			tlcpCert, err := certManager.LoadTLCP(cfg.Name+"-tlcp", cfg.Certs.TLCP.Cert, cfg.Certs.TLCP.Key)
			if err != nil {
				return err
			}
			a.tlcpConfig.Certificates = []tlcp.Certificate{tlcpCert.Cert.TLCPCertificate()}
		}
	}

	// TLS客户端双向认证
	if tlsAuth == "mutual" {
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

func (a *TLCPAdapter) DialTLCP(network, addr string) (net.Conn, error) {
	return tlcp.Dial(network, addr, a.tlcpConfig)
}

func (a *TLCPAdapter) DialTLS(network, addr string) (net.Conn, error) {
	return tls.Dial(network, addr, a.tlsConfig)
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
	return a.tlcpConfig
}

func (a *TLCPAdapter) TLSConfig() *tls.Config {
	return a.tlsConfig
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
