package cert

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"sync"
	"time"

	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/emmansun/gmsm/smx509"
)

// HotCertPool 可热重载的证书池
type HotCertPool struct {
	mu         sync.RWMutex
	paths      []string
	pool       *x509.CertPool
	smPool     *smx509.CertPool
	lastReload time.Time
}

// NewHotCertPool 创建可热重载的证书池
// 参数:
//   - paths: CA 证书文件路径列表
//
// 返回:
//   - *HotCertPool: 可热重载的证书池实例
func NewHotCertPool(paths []string) *HotCertPool {
	return &HotCertPool{
		paths: paths,
	}
}

// Load 加载证书池
// 返回:
//   - error: 加载失败时返回错误
func (h *HotCertPool) Load() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	pool := x509.NewCertPool()
	smPool := smx509.NewCertPool()

	for _, path := range h.paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("读取CA证书失败 %s: %w", path, err)
		}

		added := false
		if smPool.AppendCertsFromPEM(data) {
			added = true
		}
		if pool.AppendCertsFromPEM(data) {
			added = true
		}
		if !added {
			return fmt.Errorf("解析CA证书失败 %s", path)
		}
	}

	h.pool = pool
	h.smPool = smPool
	h.lastReload = time.Now()
	return nil
}

// Reload 重新加载证书池
// 返回:
//   - error: 重载失败时返回错误
func (h *HotCertPool) Reload() error {
	return h.Load()
}

// Pool 获取标准 x509 证书池
// 返回:
//   - *x509.CertPool: 标准证书池
func (h *HotCertPool) Pool() *x509.CertPool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.pool
}

// SMPool 获取国密 smx509 证书池
// 返回:
//   - *smx509.CertPool: 国密证书池
func (h *HotCertPool) SMPool() *smx509.CertPool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.smPool
}

// Paths 获取证书路径列表
// 返回:
//   - []string: 证书文件路径列表
func (h *HotCertPool) Paths() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return append([]string{}, h.paths...)
}

// LastReload 获取上次重载时间
// 返回:
//   - time.Time: 上次重载时间
func (h *HotCertPool) LastReload() time.Time {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.lastReload
}

// CertType 证书类型
type CertType int

const (
	// CertTypeTLS 标准TLS证书
	CertTypeTLS CertType = iota
	// CertTypeTLCP 国密TLCP证书（SM2算法）
	CertTypeTLCP
)

// Certificate 证书对象，封装证书链和私钥
type Certificate struct {
	// Certificate 证书链，第一个元素为终端实体证书
	Certificate []*x509.Certificate
	// PrivateKey 私钥
	PrivateKey crypto.PrivateKey
	// certType 证书类型
	certType CertType
	// certPEM 证书PEM数据
	certPEM []byte
	// keyPEM 私钥PEM数据
	keyPEM []byte
	// certPath 证书文件路径
	certPath string
	// keyPath 私钥文件路径
	keyPath string
	mu      sync.RWMutex
}

// Loader 证书加载器，负责证书和CA证书池的加载与缓存
type Loader struct {
	// certs 已加载的证书缓存
	certs map[string]*Certificate
	// clientCAs 客户端CA证书池缓存
	clientCAs map[string]*x509.CertPool
	// serverCAs 服务端CA证书池缓存
	serverCAs map[string]*x509.CertPool
	mu        sync.RWMutex
}

// NewLoader 创建新的证书加载器
// 返回:
//   - *Loader: 证书加载器实例
func NewLoader() *Loader {
	return &Loader{
		certs:     make(map[string]*Certificate),
		clientCAs: make(map[string]*x509.CertPool),
		serverCAs: make(map[string]*x509.CertPool),
	}
}

// LoadTLCP 加载国密TLCP证书
// 参数:
//   - certPath: 证书文件路径，PEM格式
//   - keyPath: 私钥文件路径，PEM格式
//
// 返回:
//   - *Certificate: 证书对象
//   - error: 读取或解析失败时返回错误
func (l *Loader) LoadTLCP(certPath, keyPath string) (*Certificate, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("读取证书文件失败: %w", err)
	}

	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("读取私钥文件失败: %w", err)
	}

	cert, err := ParseTLCPCertificate(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	cert.certPath = certPath
	cert.keyPath = keyPath
	cert.certPEM = certPEM
	cert.keyPEM = keyPEM

	return cert, nil
}

// LoadTLS 加载标准TLS证书
// 参数:
//   - certPath: 证书文件路径，PEM格式
//   - keyPath: 私钥文件路径，PEM格式
//
// 返回:
//   - *Certificate: 证书对象
//   - error: 读取或解析失败时返回错误
func (l *Loader) LoadTLS(certPath, keyPath string) (*Certificate, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("读取证书文件失败: %w", err)
	}

	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("读取私钥文件失败: %w", err)
	}

	cert, err := ParseTLSCertificate(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	cert.certPath = certPath
	cert.keyPath = keyPath
	cert.certPEM = certPEM
	cert.keyPEM = keyPEM

	return cert, nil
}

// ParseTLCPCertificate 从PEM数据解析国密TLCP证书
// 参数:
//   - certPEM: 证书PEM数据
//   - keyPEM: 私钥PEM数据
//
// 返回:
//   - *Certificate: 证书对象
//   - error: 解析失败或证书私钥不匹配时返回错误
func ParseTLCPCertificate(certPEM, keyPEM []byte) (*Certificate, error) {
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, fmt.Errorf("无法解析证书PEM块")
	}

	smCerts, err := smx509.ParseCertificates(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析SM2证书失败: %w", err)
	}

	if len(smCerts) == 0 {
		return nil, fmt.Errorf("证书链为空")
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, fmt.Errorf("无法解析私钥PEM块")
	}

	var privateKey crypto.PrivateKey

	switch keyBlock.Type {
	case "PRIVATE KEY", "EC PRIVATE KEY":
		privateKey, err = smx509.ParsePKCS8PrivateKey(keyBlock.Bytes)
		if err != nil {
			privateKey, err = smx509.ParseECPrivateKey(keyBlock.Bytes)
			if err != nil {
				return nil, fmt.Errorf("解析SM2私钥失败: %w", err)
			}
		}
	default:
		privateKey, err = smx509.ParsePKCS8PrivateKey(keyBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("解析私钥失败: %w", err)
		}
	}

	certs := make([]*x509.Certificate, len(smCerts))
	for i, smCert := range smCerts {
		certs[i] = smCert.ToX509()
	}

	if err := ValidateCertKeyPair(certs, privateKey); err != nil {
		return nil, fmt.Errorf("证书私钥验证失败: %w", err)
	}

	return &Certificate{
		Certificate: certs,
		PrivateKey:  privateKey,
		certType:    CertTypeTLCP,
		certPEM:     certPEM,
		keyPEM:      keyPEM,
	}, nil
}

// ParseTLSCertificate 从PEM数据解析标准TLS证书
// 参数:
//   - certPEM: 证书PEM数据
//   - keyPEM: 私钥PEM数据
//
// 返回:
//   - *Certificate: 证书对象
//   - error: 解析失败或证书私钥不匹配时返回错误
func ParseTLSCertificate(certPEM, keyPEM []byte) (*Certificate, error) {
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, fmt.Errorf("无法解析证书PEM块")
	}

	certs, err := x509.ParseCertificates(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析证书失败: %w", err)
	}

	if len(certs) == 0 {
		return nil, fmt.Errorf("证书链为空")
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, fmt.Errorf("无法解析私钥PEM块")
	}

	var privateKey crypto.PrivateKey

	switch keyBlock.Type {
	case "RSA PRIVATE KEY":
		privateKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("解析RSA私钥失败: %w", err)
		}
	case "EC PRIVATE KEY":
		privateKey, err = x509.ParseECPrivateKey(keyBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("解析EC私钥失败: %w", err)
		}
	case "PRIVATE KEY":
		privateKey, err = x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("解析PKCS8私钥失败: %w", err)
		}
	default:
		return nil, fmt.Errorf("不支持的私钥类型: %s", keyBlock.Type)
	}

	if err := ValidateCertKeyPair(certs, privateKey); err != nil {
		return nil, fmt.Errorf("证书私钥验证失败: %w", err)
	}

	return &Certificate{
		Certificate: certs,
		PrivateKey:  privateKey,
		certType:    CertTypeTLS,
		certPEM:     certPEM,
		keyPEM:      keyPEM,
	}, nil
}

// ValidateCertKeyPair 验证证书与私钥是否匹配
// 参数:
//   - certs: 证书链
//   - privateKey: 私钥
//
// 返回:
//   - error: 证书为空或公私钥不匹配时返回错误
func ValidateCertKeyPair(certs []*x509.Certificate, privateKey crypto.PrivateKey) error {
	if len(certs) == 0 {
		return fmt.Errorf("证书链为空")
	}

	leaf := certs[0]
	pubKey := leaf.PublicKey

	switch priv := privateKey.(type) {
	case *rsa.PrivateKey:
		pub, ok := pubKey.(*rsa.PublicKey)
		if !ok {
			return fmt.Errorf("证书公钥类型与私钥不匹配")
		}
		if pub.N.Cmp(priv.N) != 0 || pub.E != priv.E {
			return fmt.Errorf("RSA证书公钥与私钥不匹配")
		}
	case *ecdsa.PrivateKey:
		pub, ok := pubKey.(*ecdsa.PublicKey)
		if !ok {
			return fmt.Errorf("证书公钥类型与私钥不匹配")
		}
		if !pub.Equal(priv.Public()) {
			return fmt.Errorf("ECDSA证书公钥与私钥不匹配")
		}
	default:
		if pubKey == nil {
			return fmt.Errorf("证书公钥为空")
		}
	}

	return nil
}

func (c *Certificate) TLSCertificate() tls.Certificate {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return tls.Certificate{
		Certificate: c.getRawCerts(),
		PrivateKey:  c.PrivateKey,
	}
}

func (c *Certificate) TLCPCertificate() tlcp.Certificate {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return tlcp.Certificate{
		Certificate: c.getRawCerts(),
		PrivateKey:  c.PrivateKey,
	}
}

func (c *Certificate) getRawCerts() [][]byte {
	raw := make([][]byte, len(c.Certificate))
	for i, cert := range c.Certificate {
		raw[i] = cert.Raw
	}
	return raw
}

func (c *Certificate) GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cert := c.TLSCertificate()
	return &cert, nil
}

func (c *Certificate) GetTLCPCertificate(*tlcp.ClientHelloInfo) (*tlcp.Certificate, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cert := c.TLCPCertificate()
	return &cert, nil
}

// Reload 从内存中的PEM数据重新加载证书
// 返回:
//   - error: 证书路径未设置或解析失败时返回错误
//
// 注意: 适用于证书文件内容已更新但路径不变的场景
func (c *Certificate) Reload() error {
	if c.certPath == "" || c.keyPath == "" {
		return fmt.Errorf("证书路径未设置，无法重载")
	}

	var newCert *Certificate
	var err error

	if c.certType == CertTypeTLCP {
		newCert, err = ParseTLCPCertificate(c.certPEM, c.keyPEM)
	} else {
		newCert, err = ParseTLSCertificate(c.certPEM, c.keyPEM)
	}

	if err != nil {
		return fmt.Errorf("重载证书失败: %w", err)
	}

	c.mu.Lock()
	c.Certificate = newCert.Certificate
	c.PrivateKey = newCert.PrivateKey
	c.mu.Unlock()

	return nil
}

// ReloadFromPath 从文件路径重新加载证书
// 返回:
//   - error: 证书路径未设置或读取解析失败时返回错误
//
// 注意: 会重新读取证书和私钥文件
func (c *Certificate) ReloadFromPath() error {
	if c.certPath == "" || c.keyPath == "" {
		return fmt.Errorf("证书路径未设置，无法重载")
	}

	var newCert *Certificate
	var err error

	if c.certType == CertTypeTLCP {
		newCert, err = ParseTLCPCertificateFromPath(c.certPath, c.keyPath)
	} else {
		newCert, err = ParseTLSCertificateFromPath(c.certPath, c.keyPath)
	}

	if err != nil {
		return fmt.Errorf("重载证书失败: %w", err)
	}

	c.mu.Lock()
	c.Certificate = newCert.Certificate
	c.PrivateKey = newCert.PrivateKey
	c.certPEM = newCert.certPEM
	c.keyPEM = newCert.keyPEM
	c.mu.Unlock()

	return nil
}

func ParseTLCPCertificateFromPath(certPath, keyPath string) (*Certificate, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("读取证书文件失败: %w", err)
	}

	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("读取私钥文件失败: %w", err)
	}

	cert, err := ParseTLCPCertificate(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	cert.certPath = certPath
	cert.keyPath = keyPath

	return cert, nil
}

func ParseTLSCertificateFromPath(certPath, keyPath string) (*Certificate, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("读取证书文件失败: %w", err)
	}

	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("读取私钥文件失败: %w", err)
	}

	cert, err := ParseTLSCertificate(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	cert.certPath = certPath
	cert.keyPath = keyPath

	return cert, nil
}

func (c *Certificate) Type() CertType {
	return c.certType
}

func (c *Certificate) Leaf() *x509.Certificate {
	if len(c.Certificate) == 0 {
		return nil
	}
	return c.Certificate[0]
}

func (c *Certificate) ExpiresAt() time.Time {
	leaf := c.Leaf()
	if leaf == nil {
		return time.Time{}
	}
	return leaf.NotAfter
}

func (c *Certificate) IsExpired() bool {
	return time.Now().After(c.ExpiresAt())
}

func (c *Certificate) Subject() string {
	leaf := c.Leaf()
	if leaf == nil {
		return ""
	}
	return leaf.Subject.String()
}

func (c *Certificate) Issuer() string {
	leaf := c.Leaf()
	if leaf == nil {
		return ""
	}
	return leaf.Issuer.String()
}

func (l *Loader) Store(name string, cert *Certificate) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.certs[name] = cert
}

func (l *Loader) Get(name string) *Certificate {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.certs[name]
}

func (l *Loader) Delete(name string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.certs, name)
}

func (l *Loader) LoadClientCA(paths []string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("读取客户端CA证书失败 %s: %w", path, err)
		}

		if !pool.AppendCertsFromPEM(data) {
			return nil, fmt.Errorf("解析客户端CA证书失败 %s", path)
		}
	}

	return pool, nil
}

func (l *Loader) LoadServerCA(paths []string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("读取服务端CA证书失败 %s: %w", path, err)
		}

		if !pool.AppendCertsFromPEM(data) {
			return nil, fmt.Errorf("解析服务端CA证书失败 %s", path)
		}
	}

	return pool, nil
}

func (l *Loader) StoreClientCA(name string, pool *x509.CertPool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.clientCAs[name] = pool
}

func (l *Loader) StoreServerCA(name string, pool *x509.CertPool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.serverCAs[name] = pool
}

func (l *Loader) GetClientCA(name string) *x509.CertPool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.clientCAs[name]
}

func (l *Loader) GetServerCA(name string) *x509.CertPool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.serverCAs[name]
}

func (l *Loader) Reload(name string) error {
	cert := l.Get(name)
	if cert == nil {
		return fmt.Errorf("证书 %s 不存在", name)
	}
	return cert.ReloadFromPath()
}

func (l *Loader) ReloadAll() error {
	l.mu.RLock()
	certs := make([]*Certificate, 0, len(l.certs))
	for _, cert := range l.certs {
		certs = append(certs, cert)
	}
	l.mu.RUnlock()

	var errs []error
	for _, cert := range certs {
		if err := cert.ReloadFromPath(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("部分证书重载失败: %v", errs)
	}
	return nil
}

// HotReloader 证书热重载器，定时检测并重新加载证书
type HotReloader struct {
	// loader 证书加载器
	loader *Loader
	// interval 检测间隔，单位: 纳秒（支持时间格式如 "1h", "30m"）
	interval  time.Duration
	stopCh    chan struct{}
	stoppedCh chan struct{}
}

// NewHotReloader 创建证书热重载器
// 参数:
//   - loader: 证书加载器
//   - interval: 检测间隔
//
// 返回:
//   - *HotReloader: 热重载器实例
func NewHotReloader(loader *Loader, interval time.Duration) *HotReloader {
	return &HotReloader{
		loader:    loader,
		interval:  interval,
		stopCh:    make(chan struct{}),
		stoppedCh: make(chan struct{}),
	}
}

// Start 启动热重载器，定时检测并重新加载证书
// 注意: 该方法启动后台goroutine，调用Stop()停止
func (h *HotReloader) Start() {
	ticker := time.NewTicker(h.interval)
	go func() {
		defer close(h.stoppedCh)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				h.loader.ReloadAll()
			case <-h.stopCh:
				return
			}
		}
	}()
}

// Stop 停止热重载器
// 注意: 该方法会等待后台goroutine退出后返回
func (h *HotReloader) Stop() {
	close(h.stopCh)
	<-h.stoppedCh
}

func LoadCertificate(certType CertType, certPath, keyPath string) (*Certificate, error) {
	switch certType {
	case CertTypeTLCP:
		return ParseTLCPCertificateFromPath(certPath, keyPath)
	case CertTypeTLS:
		return ParseTLSCertificateFromPath(certPath, keyPath)
	default:
		return nil, fmt.Errorf("未知的证书类型: %v", certType)
	}
}

// DetectCertType 检测证书类型
// 参数:
//   - certPath: 证书文件路径
//
// 返回:
//   - CertType: 证书类型（CertTypeTLCP或CertTypeTLS）
//   - error: 读取或解析失败时返回错误
func DetectCertType(certPath string) (CertType, error) {
	data, err := os.ReadFile(certPath)
	if err != nil {
		return CertTypeTLS, fmt.Errorf("读取证书文件失败: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return CertTypeTLS, fmt.Errorf("无法解析证书PEM块")
	}

	if _, err := smx509.ParseCertificate(block.Bytes); err == nil {
		return CertTypeTLCP, nil
	}

	if _, err := x509.ParseCertificate(block.Bytes); err == nil {
		return CertTypeTLS, nil
	}

	return CertTypeTLS, fmt.Errorf("无法识别证书类型")
}

// LoadBuiltinRootCAs 加载预制根证书
// 参数:
//   - dir: 预制根证书目录路径
//
// 返回:
//   - *smx509.CertPool: 国密根证书池
//   - *x509.CertPool: 标准根证书池
//   - error: 读取目录失败时返回错误
//
// 注意: 会扫描目录中所有.pem和.crt文件并加载
func LoadBuiltinRootCAs(dir string) (*smx509.CertPool, *x509.CertPool, error) {
	smPool := smx509.NewCertPool()
	stdPool := x509.NewCertPool()

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return smPool, stdPool, nil
		}
		return nil, nil, fmt.Errorf("读取预制根证书目录失败: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if len(name) < 5 {
			continue
		}

		ext := name[len(name)-4:]
		if ext != ".pem" && ext != ".crt" {
			continue
		}

		path := dir + "/" + name
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		// 尝试作为国密证书加载
		if smPool.AppendCertsFromPEM(data) {
			continue
		}

		// 尝试作为标准证书加载
		stdPool.AppendCertsFromPEM(data)
	}

	return smPool, stdPool, nil
}
