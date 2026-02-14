package cert

import (
	"crypto/tls"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"gitee.com/Trisia/gotlcp/tlcp"
)

// Manager 证书管理器，负责证书的加载、缓存和生命周期管理
type Manager struct {
	// loader 证书加载器
	loader *Loader
	// certs 证书信息缓存，key为证书名称
	certs map[string]*CertInfo
	// certDir 证书密钥目录路径，用于存储证书和私钥文件
	certDir string
	// trustedDir 受信任证书目录路径，用于存储根证书和CA证书
	trustedDir string
	mu         sync.RWMutex
}

// CertInfo 证书信息，包含证书的基本属性和过期时间
type CertInfo struct {
	// Name 证书名称，用于标识和检索
	Name string
	// Type 证书类型，可选值: CertTypeTLCP, CertTypeTLS
	Type CertType
	// Cert 证书对象
	Cert *Certificate
	// NotAfter 证书过期时间
	NotAfter time.Time
}

// NewManager 创建新的证书管理器
// 返回:
//   - *Manager: 证书管理器实例
func NewManager() *Manager {
	return &Manager{
		loader: NewLoader(),
		certs:  make(map[string]*CertInfo),
	}
}

// NewManagerWithCertDir 创建带证书目录的证书管理器
// 参数:
//   - certDir: 证书目录路径，相对路径将基于此目录解析
//
// 返回:
//   - *Manager: 证书管理器实例
func NewManagerWithCertDir(certDir string) *Manager {
	return &Manager{
		loader:  NewLoader(),
		certs:   make(map[string]*CertInfo),
		certDir: certDir,
	}
}

// NewManagerWithDirs 创建带证书目录和受信任证书目录的证书管理器
// 参数:
//   - certDir: 证书密钥目录路径，用于存储证书和私钥文件
//   - trustedDir: 受信任证书目录路径，用于存储根证书和CA证书
//
// 返回:
//   - *Manager: 证书管理器实例
func NewManagerWithDirs(certDir, trustedDir string) *Manager {
	return &Manager{
		loader:     NewLoader(),
		certs:      make(map[string]*CertInfo),
		certDir:    certDir,
		trustedDir: trustedDir,
	}
}

func (m *Manager) SetCertDir(dir string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.certDir = dir
}

func (m *Manager) GetCertDir() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.certDir
}

func (m *Manager) resolvePath(path string) string {
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) {
		return path
	}
	// 根据文件扩展名判断证书类型
	ext := filepath.Ext(path)
	isCertFile := ext == ".crt" || ext == ".cer" || ext == ".pem" || ext == ".key"

	// 如果是证书或密钥文件，使用certDir
	if isCertFile && m.certDir != "" {
		return filepath.Join(m.certDir, path)
	}
	// 否则使用trustedDir（用于CA证书等）
	if m.trustedDir != "" {
		return filepath.Join(m.trustedDir, path)
	}
	return path
}

// Load 加载证书并缓存
// 参数:
//   - name: 证书名称，用于后续检索
//   - certPath: 证书文件路径，相对路径将基于certDir解析
//   - keyPath: 私钥文件路径，相对路径将基于certDir解析
//   - certType: 证书类型（CertTypeTLCP或CertTypeTLS）
//
// 返回:
//   - *CertInfo: 证书信息
//   - error: 加载失败时返回错误
func (m *Manager) Load(name, certPath, keyPath string, certType CertType) (*CertInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

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

	info := &CertInfo{
		Name:     name,
		Type:     certType,
		Cert:     cert,
		NotAfter: cert.ExpiresAt(),
	}
	m.certs[name] = info
	m.loader.Store(name, cert)

	return info, nil
}

func (m *Manager) LoadTLCP(name, certPath, keyPath string) (*CertInfo, error) {
	return m.Load(name, certPath, keyPath, CertTypeTLCP)
}

func (m *Manager) LoadTLS(name, certPath, keyPath string) (*CertInfo, error) {
	return m.Load(name, certPath, keyPath, CertTypeTLS)
}

func (m *Manager) Get(name string) *CertInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.certs[name]
}

func (m *Manager) GetTLSCertificate(name string) (*tls.Certificate, error) {
	m.mu.RLock()
	info, ok := m.certs[name]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("证书 %s 不存在", name)
	}

	if info.Type != CertTypeTLS {
		return nil, fmt.Errorf("证书 %s 不是TLS类型", name)
	}

	cert := info.Cert.TLSCertificate()
	return &cert, nil
}

func (m *Manager) GetTLCPCertificate(name string) (*tlcp.Certificate, error) {
	m.mu.RLock()
	info, ok := m.certs[name]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("证书 %s 不存在", name)
	}

	if info.Type != CertTypeTLCP {
		return nil, fmt.Errorf("证书 %s 不是TLCP类型", name)
	}

	cert := info.Cert.TLCPCertificate()
	return &cert, nil
}

func (m *Manager) List() []*CertInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*CertInfo, 0, len(m.certs))
	for _, info := range m.certs {
		list = append(list, info)
	}
	return list
}

func (m *Manager) Delete(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.certs[name]; !ok {
		return fmt.Errorf("证书 %s 不存在", name)
	}

	delete(m.certs, name)
	m.loader.Delete(name)
	return nil
}

// Reload 重新加载指定证书
// 参数:
//   - name: 证书名称
//
// 返回:
//   - error: 证书不存在或重载失败时返回错误
func (m *Manager) Reload(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	info, ok := m.certs[name]
	if !ok {
		return fmt.Errorf("证书 %s 不存在", name)
	}

	if err := info.Cert.ReloadFromPath(); err != nil {
		return fmt.Errorf("重载证书失败: %w", err)
	}

	info.NotAfter = info.Cert.ExpiresAt()
	return nil
}

// ReloadAll 重新加载所有证书
// 返回:
//   - []error: 重载失败的错误列表，为空表示全部成功
func (m *Manager) ReloadAll() []error {
	m.mu.RLock()
	names := make([]string, 0, len(m.certs))
	for name := range m.certs {
		names = append(names, name)
	}
	m.mu.RUnlock()

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
