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

type CertType int

const (
	CertTypeTLS CertType = iota
	CertTypeTLCP
)

type Certificate struct {
	Certificate []*x509.Certificate
	PrivateKey  crypto.PrivateKey
	certType    CertType
	certPEM     []byte
	keyPEM      []byte
	certPath    string
	keyPath     string
	mu          sync.RWMutex
}

type Loader struct {
	certs     map[string]*Certificate
	clientCAs map[string]*x509.CertPool
	serverCAs map[string]*x509.CertPool
	mu        sync.RWMutex
}

func NewLoader() *Loader {
	return &Loader{
		certs:     make(map[string]*Certificate),
		clientCAs: make(map[string]*x509.CertPool),
		serverCAs: make(map[string]*x509.CertPool),
	}
}

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

type HotReloader struct {
	loader    *Loader
	interval  time.Duration
	stopCh    chan struct{}
	stoppedCh chan struct{}
}

func NewHotReloader(loader *Loader, interval time.Duration) *HotReloader {
	return &HotReloader{
		loader:    loader,
		interval:  interval,
		stopCh:    make(chan struct{}),
		stoppedCh: make(chan struct{}),
	}
}

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
