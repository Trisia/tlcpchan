package cert

import (
	"crypto/tls"
	"fmt"
	"math/big"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"gitee.com/Trisia/gotlcp/tlcp"
)

// Manager 证书管理器，负责证书的加载、缓存和生命周期管理
type Manager struct {
	loader     *Loader
	certs      sync.Map
	certList   []*CertInfo
	certListMu sync.RWMutex
	certDir    atomic.Value
	trustedDir atomic.Value
}

// CertInfo 证书信息，包含证书的基本属性和过期时间
type CertInfo struct {
	Name         string
	SerialNumber *big.Int
	Type         CertType
	Cert         *Certificate
	NotAfter     time.Time
}

// NewManager 创建新的证书管理器
func NewManager() *Manager {
	m := &Manager{
		loader:   NewLoader(),
		certList: make([]*CertInfo, 0),
	}
	m.certDir.Store("")
	m.trustedDir.Store("")
	return m
}

// NewManagerWithCertDir 创建带证书目录的证书管理器
func NewManagerWithCertDir(certDir string) *Manager {
	m := NewManager()
	m.certDir.Store(certDir)
	return m
}

// NewManagerWithDirs 创建带证书目录和受信任证书目录的证书管理器
func NewManagerWithDirs(certDir, trustedDir string) *Manager {
	m := NewManager()
	m.certDir.Store(certDir)
	m.trustedDir.Store(trustedDir)
	return m
}

func (m *Manager) SetCertDir(dir string) {
	m.certDir.Store(dir)
}

func (m *Manager) GetCertDir() string {
	return m.certDir.Load().(string)
}

func (m *Manager) resolvePath(path string) string {
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) {
		return path
	}
	certDir := m.certDir.Load().(string)
	trustedDir := m.trustedDir.Load().(string)

	ext := filepath.Ext(path)
	isCertFile := ext == ".crt" || ext == ".cer" || ext == ".pem" || ext == ".key"

	if isCertFile && certDir != "" {
		return filepath.Join(certDir, path)
	}
	if trustedDir != "" {
		return filepath.Join(trustedDir, path)
	}
	return path
}

// Load 加载证书并缓存
func (m *Manager) Load(name, certPath, keyPath string, certType CertType) (*CertInfo, error) {
	certPath = m.resolvePath(certPath)
	keyPath = m.resolvePath(keyPath)

	var cert *Certificate
	var err error

	if certType == CertTypeTLCP {
		cert, err = m.loader.LoadTLCP(certPath, keyPath)
	} else {
		cert, err = m.loader.LoadTLS(certPath, keyPath)
	}

	if err != nil {
		return nil, fmt.Errorf("加载证书失败: %w", err)
	}

	leaf := cert.Leaf()
	serialNumber := big.NewInt(0)
	if leaf != nil {
		serialNumber = leaf.SerialNumber
	}

	info := &CertInfo{
		Name:         name,
		SerialNumber: serialNumber,
		Type:         certType,
		Cert:         cert,
		NotAfter:     cert.ExpiresAt(),
	}

	m.certs.Store(name, info)
	m.loader.Store(name, cert)
	m.appendToCertList(info)

	return info, nil
}

func (m *Manager) appendToCertList(info *CertInfo) {
	m.certListMu.Lock()
	defer m.certListMu.Unlock()
	m.certList = append(m.certList, info)
}

func (m *Manager) LoadTLCP(name, certPath, keyPath string) (*CertInfo, error) {
	return m.Load(name, certPath, keyPath, CertTypeTLCP)
}

func (m *Manager) LoadTLS(name, certPath, keyPath string) (*CertInfo, error) {
	return m.Load(name, certPath, keyPath, CertTypeTLS)
}

func (m *Manager) Get(name string) *CertInfo {
	if v, ok := m.certs.Load(name); ok {
		return v.(*CertInfo)
	}
	return nil
}

func (m *Manager) GetTLSCertificate(name string) (*tls.Certificate, error) {
	info := m.Get(name)
	if info == nil {
		return nil, fmt.Errorf("证书 %s 不存在", name)
	}

	if info.Type != CertTypeTLS {
		return nil, fmt.Errorf("证书 %s 不是TLS类型", name)
	}

	cert := info.Cert.TLSCertificate()
	return &cert, nil
}

func (m *Manager) GetTLCPCertificate(name string) (*tlcp.Certificate, error) {
	info := m.Get(name)
	if info == nil {
		return nil, fmt.Errorf("证书 %s 不存在", name)
	}

	if info.Type != CertTypeTLCP {
		return nil, fmt.Errorf("证书 %s 不是TLCP类型", name)
	}

	cert := info.Cert.TLCPCertificate()
	return &cert, nil
}

func (m *Manager) List() []*CertInfo {
	m.certListMu.RLock()
	defer m.certListMu.RUnlock()
	result := make([]*CertInfo, len(m.certList))
	copy(result, m.certList)
	return result
}

// GetBySerialNumber 通过序列号获取证书
func (m *Manager) GetBySerialNumber(serialNumber *big.Int) *CertInfo {
	m.certListMu.RLock()
	defer m.certListMu.RUnlock()
	for _, info := range m.certList {
		if info.SerialNumber != nil && info.SerialNumber.Cmp(serialNumber) == 0 {
			return info
		}
	}
	return nil
}

// GetBySerialNumberString 通过序列号字符串获取证书
func (m *Manager) GetBySerialNumberString(serialNumberStr string) *CertInfo {
	m.certListMu.RLock()
	defer m.certListMu.RUnlock()
	for _, info := range m.certList {
		if info.SerialNumber != nil && info.SerialNumber.String() == serialNumberStr {
			return info
		}
	}
	return nil
}

func (m *Manager) Delete(name string) error {
	info := m.Get(name)
	if info == nil {
		return fmt.Errorf("证书 %s 不存在", name)
	}

	m.certs.Delete(name)
	m.loader.Delete(name)
	m.removeFromCertList(name)

	return nil
}

func (m *Manager) removeFromCertList(name string) {
	m.certListMu.Lock()
	defer m.certListMu.Unlock()
	var newList []*CertInfo
	for _, item := range m.certList {
		if item.Name != name {
			newList = append(newList, item)
		}
	}
	m.certList = newList
}

// Reload 重新加载指定证书
func (m *Manager) Reload(name string) error {
	info := m.Get(name)
	if info == nil {
		return fmt.Errorf("证书 %s 不存在", name)
	}

	if err := info.Cert.ReloadFromPath(); err != nil {
		return fmt.Errorf("重载证书失败: %w", err)
	}

	info.NotAfter = info.Cert.ExpiresAt()
	return nil
}

// ReloadAll 重新加载所有证书
func (m *Manager) ReloadAll() []error {
	var names []string
	m.certs.Range(func(key, value interface{}) bool {
		names = append(names, key.(string))
		return true
	})

	var errs []error
	for _, name := range names {
		if err := m.Reload(name); err != nil {
			errs = append(errs, fmt.Errorf("重载证书 %s 失败: %w", name, err))
		}
	}
	return errs
}

func (m *Manager) Loader() *Loader {
	return m.loader
}

func (i *CertInfo) IsExpired() bool {
	return time.Now().After(i.NotAfter)
}

func (i *CertInfo) ExpiresIn() time.Duration {
	return time.Until(i.NotAfter)
}

func (i *CertInfo) Subject() string {
	return i.Cert.Subject()
}

func (i *CertInfo) Issuer() string {
	return i.Cert.Issuer()
}
