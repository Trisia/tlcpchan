package key

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

var nameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)

// Manager 密钥存储管理器
type Manager struct {
	store     *Store
	validator *Validator
	mu        sync.RWMutex
	baseDir   string
}

// NewManager 创建密钥存储管理器
func NewManager(baseDir string) *Manager {
	return &Manager{
		store:     NewStore(baseDir),
		validator: NewValidator(),
		baseDir:   baseDir,
	}
}

// EnsureBaseDir 确保基础目录存在
func (m *Manager) EnsureBaseDir() error {
	return m.store.EnsureBaseDir()
}

// Create 创建新的密钥存储
func (m *Manager) Create(name string, keyType KeyStoreType, keyParams KeyParams,
	signCert, signKey, encCert, encKey []byte) (*KeyStore, error) {

	m.mu.Lock()
	defer m.mu.Unlock()

	if !isValidName(name) {
		return nil, fmt.Errorf("无效的名称，只允许字母、数字、下划线和连字符")
	}

	if m.store.Exists(name) {
		return nil, fmt.Errorf("密钥已存在")
	}

	if keyType == KeyStoreTypeTLCP {
		if err := m.validator.ValidateCreateTLCP(signCert, signKey, encCert, encKey); err != nil {
			return nil, err
		}
	} else {
		if err := m.validator.ValidateCreateTLS(signCert, signKey); err != nil {
			return nil, err
		}
	}

	now := time.Now()
	ks := &KeyStore{
		Name:      name,
		Type:      keyType,
		KeyParams: keyParams,
		SignCert:  FileNameSignCert,
		SignKey:   FileNameSignKey,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if keyType == KeyStoreTypeTLCP {
		ks.EncCert = FileNameEncCert
		ks.EncKey = FileNameEncKey
	}

	if err := m.store.Save(ks, signCert, signKey, encCert, encKey); err != nil {
		return nil, err
	}

	return ks, nil
}

// UpdateCertificates 更新证书（密钥保持不变）
func (m *Manager) UpdateCertificates(name string, signCert, encCert []byte) (*KeyStore, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.store.Exists(name) {
		return nil, fmt.Errorf("密钥不存在")
	}

	ks, err := m.store.Load(name)
	if err != nil {
		return nil, err
	}

	if ks.Type == KeyStoreTypeTLCP {
		if err := m.validator.ValidateUpdateCertsTLCP(signCert, encCert); err != nil {
			return nil, err
		}
	} else {
		if err := m.validator.ValidateUpdateCertsTLS(signCert); err != nil {
			return nil, err
		}
	}

	if err := m.store.UpdateCertificates(name, signCert, encCert); err != nil {
		return nil, err
	}

	return m.store.Load(name)
}

// Delete 删除密钥存储
func (m *Manager) Delete(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.store.Delete(name)
}

// Get 获取密钥存储
func (m *Manager) Get(name string) (*KeyStore, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.store.Load(name)
}

// GetInfo 获取密钥存储信息
func (m *Manager) GetInfo(name string) (*KeyStoreInfo, error) {
	ks, err := m.Get(name)
	if err != nil {
		return nil, err
	}

	return m.toInfo(ks), nil
}

// List 列出所有密钥存储
func (m *Manager) List() ([]*KeyStoreInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ksList, err := m.store.List()
	if err != nil {
		return nil, err
	}

	result := make([]*KeyStoreInfo, len(ksList))
	for i, ks := range ksList {
		result[i] = m.toInfo(ks)
	}

	return result, nil
}

// GetSignCertPath 获取签名证书路径
func (m *Manager) GetSignCertPath(name string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.store.Exists(name) {
		return "", fmt.Errorf("密钥不存在")
	}

	return m.store.GetFilePath(name, FileNameSignCert), nil
}

// GetSignKeyPath 获取签名密钥路径
func (m *Manager) GetSignKeyPath(name string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.store.Exists(name) {
		return "", fmt.Errorf("密钥不存在")
	}

	return m.store.GetFilePath(name, FileNameSignKey), nil
}

// GetEncCertPath 获取加密证书路径（仅国密）
func (m *Manager) GetEncCertPath(name string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.store.Exists(name) {
		return "", fmt.Errorf("密钥不存在")
	}

	ks, err := m.store.Load(name)
	if err != nil {
		return "", err
	}

	if ks.Type != KeyStoreTypeTLCP {
		return "", fmt.Errorf("仅国密类型支持加密证书")
	}

	return m.store.GetFilePath(name, FileNameEncCert), nil
}

// GetEncKeyPath 获取加密密钥路径（仅国密）
func (m *Manager) GetEncKeyPath(name string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.store.Exists(name) {
		return "", fmt.Errorf("密钥不存在")
	}

	ks, err := m.store.Load(name)
	if err != nil {
		return "", err
	}

	if ks.Type != KeyStoreTypeTLCP {
		return "", fmt.Errorf("仅国密类型支持加密密钥")
	}

	return m.store.GetFilePath(name, FileNameEncKey), nil
}

// Exists 检查密钥是否存在
func (m *Manager) Exists(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.store.Exists(name)
}

// GetBaseDir 获取基础目录
func (m *Manager) GetBaseDir() string {
	return m.baseDir
}

func (m *Manager) toInfo(ks *KeyStore) *KeyStoreInfo {
	info := &KeyStoreInfo{
		Name:        ks.Name,
		Type:        ks.Type,
		KeyParams:   ks.KeyParams,
		HasSignCert: m.store.HasFile(ks.Name, FileNameSignCert),
		HasSignKey:  m.store.HasFile(ks.Name, FileNameSignKey),
		CreatedAt:   ks.CreatedAt,
		UpdatedAt:   ks.UpdatedAt,
	}

	if ks.Type == KeyStoreTypeTLCP {
		info.HasEncCert = m.store.HasFile(ks.Name, FileNameEncCert)
		info.HasEncKey = m.store.HasFile(ks.Name, FileNameEncKey)
	}

	return info
}

func isValidName(name string) bool {
	if name == "" {
		return false
	}
	if !nameRegex.MatchString(name) {
		return false
	}
	clean := filepath.Clean(name)
	return clean == name && len(clean) > 0 && !filepath.IsAbs(clean)
}
