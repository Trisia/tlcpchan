package rootcert

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/emmansun/gmsm/smx509"
)

// Manager 根证书管理器
type Manager struct {
	baseDir    string
	store      *Store
	certs      map[string]*RootCert
	certPool   *x509.CertPool
	smCertPool *smx509.CertPool
	mu         sync.RWMutex
}

// NewManager 创建根证书管理器
func NewManager(baseDir string) *Manager {
	return &Manager{
		baseDir:    baseDir,
		store:      NewStore(baseDir),
		certs:      make(map[string]*RootCert),
		certPool:   x509.NewCertPool(),
		smCertPool: smx509.NewCertPool(),
	}
}

// Initialize 初始化管理器，预加载根证书
func (m *Manager) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.store.ensureBaseDir(); err != nil {
		return err
	}

	records, err := m.store.LoadAll()
	if err != nil {
		return fmt.Errorf("加载根证书记录失败: %w", err)
	}

	for _, record := range records {
		certPath := m.store.getCertPath(record.Name)
		data, err := os.ReadFile(certPath)
		if err != nil {
			continue
		}

		rootCert, err := m.parseCert(data, record.Name, record.AddedAt)
		if err != nil {
			continue
		}

		m.certs[record.Name] = rootCert
		m.addToPools(data)
	}

	return nil
}

// Add 添加根证书
func (m *Manager) Add(name string, certData []byte) (*RootCert, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.certs[name]; exists {
		return nil, fmt.Errorf("根证书 %s 已存在", name)
	}

	now := time.Now()
	rootCert, err := m.parseCert(certData, name, now)
	if err != nil {
		return nil, err
	}

	if err := m.store.Save(name, certData, now); err != nil {
		return nil, fmt.Errorf("保存根证书失败: %w", err)
	}

	m.certs[name] = rootCert
	m.addToPools(certData)

	return rootCert, nil
}

// Delete 删除根证书
func (m *Manager) Delete(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.certs[name]; !exists {
		return fmt.Errorf("根证书 %s 不存在", name)
	}

	if err := m.store.Delete(name); err != nil {
		return fmt.Errorf("删除根证书失败: %w", err)
	}

	delete(m.certs, name)
	m.rebuildPools()

	return nil
}

// Get 获取根证书
func (m *Manager) Get(name string) (*RootCert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cert, exists := m.certs[name]
	if !exists {
		return nil, fmt.Errorf("根证书 %s 不存在", name)
	}

	return cert, nil
}

// List 列出所有根证书
func (m *Manager) List() []*RootCert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*RootCert, 0, len(m.certs))
	for _, cert := range m.certs {
		result = append(result, cert)
	}

	return result
}

// GetPool 获取根证书池（包含所有根证书）
func (m *Manager) GetPool() RootCertPool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return &rootCertPool{
		certPool:   m.certPool,
		smCertPool: m.smCertPool,
		certs:      m.certs,
	}
}

// GetPoolForNames 获取指定名称的根证书池
func (m *Manager) GetPoolForNames(names []string) (RootCertPool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	certPool := x509.NewCertPool()
	smCertPool := smx509.NewCertPool()
	certs := make(map[string]*RootCert)

	for _, name := range names {
		cert, exists := m.certs[name]
		if !exists {
			certPath := m.store.getCertPath(name)
			data, err := os.ReadFile(certPath)
			if err != nil {
				return nil, fmt.Errorf("根证书 %s 不存在", name)
			}
			block, _ := pem.Decode(data)
			if block == nil {
				continue
			}
			certPool.AppendCertsFromPEM(data)
			smCertPool.AppendCertsFromPEM(data)
			continue
		}

		certs[name] = cert
		certPath := m.store.getCertPath(name)
		data, _ := os.ReadFile(certPath)
		certPool.AppendCertsFromPEM(data)
		smCertPool.AppendCertsFromPEM(data)
	}

	return &rootCertPool{
		certPool:   certPool,
		smCertPool: smCertPool,
		certs:      certs,
	}, nil
}

// Reload 重新加载所有根证书
func (m *Manager) Reload() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.certs = make(map[string]*RootCert)
	m.certPool = x509.NewCertPool()
	m.smCertPool = smx509.NewCertPool()

	records, err := m.store.LoadAll()
	if err != nil {
		return fmt.Errorf("加载根证书记录失败: %w", err)
	}

	for _, record := range records {
		certPath := m.store.getCertPath(record.Name)
		data, err := os.ReadFile(certPath)
		if err != nil {
			continue
		}

		rootCert, err := m.parseCert(data, record.Name, record.AddedAt)
		if err != nil {
			continue
		}

		m.certs[record.Name] = rootCert
		m.addToPools(data)
	}

	return nil
}

func (m *Manager) parseCert(data []byte, name string, addedAt time.Time) (*RootCert, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("无法解析证书PEM块")
	}

	var cert *x509.Certificate
	if smCert, err := smx509.ParseCertificate(block.Bytes); err == nil {
		cert = smCert.ToX509()
	} else if stdCert, err := x509.ParseCertificate(block.Bytes); err == nil {
		cert = stdCert
	} else {
		return nil, fmt.Errorf("解析证书失败: %w", err)
	}

	return &RootCert{
		Name:     name,
		Cert:     cert,
		NotAfter: cert.NotAfter,
		Subject:  cert.Subject.String(),
		Issuer:   cert.Issuer.String(),
		AddedAt:  addedAt,
	}, nil
}

func (m *Manager) addToPools(data []byte) {
	m.certPool.AppendCertsFromPEM(data)
	m.smCertPool.AppendCertsFromPEM(data)
}

func (m *Manager) rebuildPools() {
	m.certPool = x509.NewCertPool()
	m.smCertPool = smx509.NewCertPool()

	for name := range m.certs {
		certPath := m.store.getCertPath(name)
		data, _ := os.ReadFile(certPath)
		m.addToPools(data)
	}
}

type rootCertPool struct {
	certPool   *x509.CertPool
	smCertPool *smx509.CertPool
	certs      map[string]*RootCert
}

func (p *rootCertPool) GetCertPool() *x509.CertPool {
	return p.certPool
}

func (p *rootCertPool) GetSMCertPool() *smx509.CertPool {
	return p.smCertPool
}

func (p *rootCertPool) GetCerts() []*RootCert {
	result := make([]*RootCert, 0, len(p.certs))
	for _, cert := range p.certs {
		result = append(result, cert)
	}
	return result
}
