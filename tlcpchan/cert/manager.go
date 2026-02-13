package cert

import (
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"gitee.com/Trisia/gotlcp/tlcp"
)

type Manager struct {
	loader *Loader
	certs  map[string]*CertInfo
	mu     sync.RWMutex
}

type CertInfo struct {
	Name     string
	Type     CertType
	Cert     *Certificate
	NotAfter time.Time
}

func NewManager() *Manager {
	return &Manager{
		loader: NewLoader(),
		certs:  make(map[string]*CertInfo),
	}
}

func (m *Manager) Load(name, certPath, keyPath string, certType CertType) (*CertInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

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
