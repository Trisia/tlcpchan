package keystore

import (
	"fmt"
	"sync"
	"time"
)

// Manager keystore 管理器
type Manager struct {
	baseDir      string
	store        *Store
	keyStores    map[string]KeyStore
	keyStoreInfo map[string]*KeyStoreInfo
	loaders      map[LoaderType]Loader
	mu           sync.RWMutex
}

// NewManager 创建 keystore 管理器
func NewManager(baseDir string) *Manager {
	m := &Manager{
		baseDir:      baseDir,
		store:        NewStore(baseDir),
		keyStores:    make(map[string]KeyStore),
		keyStoreInfo: make(map[string]*KeyStoreInfo),
		loaders:      make(map[LoaderType]Loader),
	}

	m.loaders[LoaderTypeFile] = NewFileLoader(baseDir)
	m.loaders[LoaderTypeNamed] = NewNamedLoader(m)

	return m
}

// Initialize 初始化管理器，加载已存储的 keystore
func (m *Manager) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	records, err := m.store.LoadAll()
	if err != nil {
		return fmt.Errorf("加载keystore记录失败: %w", err)
	}

	for _, record := range records {
		info := &KeyStoreInfo{
			Name:         record.Name,
			Type:         record.Type,
			LoaderType:   record.LoaderType,
			LoaderConfig: record.LoaderConfig,
			Protected:    record.Protected,
			CreatedAt:    record.CreatedAt,
			UpdatedAt:    record.UpdatedAt,
		}
		m.keyStoreInfo[record.Name] = info
	}

	return nil
}

// RegisterLoader 注册加载器
func (m *Manager) RegisterLoader(loaderType LoaderType, loader Loader) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loaders[loaderType] = loader
}

// Create 创建新的 keystore
func (m *Manager) Create(name string, loaderConfig LoaderConfig, protected bool) (*KeyStoreInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.keyStoreInfo[name]; exists {
		return nil, fmt.Errorf("keystore %s 已存在", name)
	}

	loader, ok := m.loaders[loaderConfig.Type]
	if !ok {
		return nil, fmt.Errorf("不支持的加载器类型: %s", loaderConfig.Type)
	}

	ks, err := loader.Load(loaderConfig)
	if err != nil {
		return nil, fmt.Errorf("加载keystore失败: %w", err)
	}

	now := time.Now()
	info := &KeyStoreInfo{
		Name:         name,
		Type:         ks.Type(),
		LoaderType:   loaderConfig.Type,
		LoaderConfig: loaderConfig,
		Protected:    protected,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	record := &StoreRecord{
		Name:         info.Name,
		Type:         info.Type,
		LoaderType:   info.LoaderType,
		LoaderConfig: info.LoaderConfig,
		Protected:    info.Protected,
		CreatedAt:    info.CreatedAt,
		UpdatedAt:    info.UpdatedAt,
	}

	if err := m.store.Save(record); err != nil {
		return nil, fmt.Errorf("保存keystore记录失败: %w", err)
	}

	m.keyStoreInfo[name] = info
	m.keyStores[name] = ks

	return info, nil
}

// Delete 删除 keystore
func (m *Manager) Delete(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	info, exists := m.keyStoreInfo[name]
	if !exists {
		return fmt.Errorf("keystore %s 不存在", name)
	}

	if info.Protected {
		return fmt.Errorf("keystore %s 受保护，无法删除", name)
	}

	if err := m.store.Delete(name); err != nil {
		return fmt.Errorf("删除keystore记录失败: %w", err)
	}

	delete(m.keyStoreInfo, name)
	delete(m.keyStores, name)

	return nil
}

// Get 获取 keystore 信息
func (m *Manager) Get(name string) (*KeyStoreInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info, exists := m.keyStoreInfo[name]
	if !exists {
		return nil, fmt.Errorf("keystore %s 不存在", name)
	}

	return info, nil
}

// GetKeyStore 获取 keystore 实例
func (m *Manager) GetKeyStore(name string) (KeyStore, error) {
	m.mu.RLock()
	if ks, exists := m.keyStores[name]; exists {
		defer m.mu.RUnlock()
		return ks, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	if ks, exists := m.keyStores[name]; exists {
		return ks, nil
	}

	info, exists := m.keyStoreInfo[name]
	if !exists {
		return nil, fmt.Errorf("keystore %s 不存在", name)
	}

	loader, ok := m.loaders[info.LoaderType]
	if !ok {
		return nil, fmt.Errorf("不支持的加载器类型: %s", info.LoaderType)
	}

	ks, err := loader.Load(info.LoaderConfig)
	if err != nil {
		return nil, fmt.Errorf("加载keystore失败: %w", err)
	}

	m.keyStores[name] = ks
	return ks, nil
}

// List 列出所有 keystore
func (m *Manager) List() []*KeyStoreInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*KeyStoreInfo, 0, len(m.keyStoreInfo))
	for _, info := range m.keyStoreInfo {
		result = append(result, info)
	}

	return result
}

// Reload 重新加载 keystore
func (m *Manager) Reload(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	info, exists := m.keyStoreInfo[name]
	if !exists {
		return fmt.Errorf("keystore %s 不存在", name)
	}

	if ks, exists := m.keyStores[name]; exists {
		if err := ks.Reload(); err != nil {
			return err
		}
	}

	delete(m.keyStores, name)
	info.UpdatedAt = time.Now()

	record := &StoreRecord{
		Name:         info.Name,
		Type:         info.Type,
		LoaderType:   info.LoaderType,
		LoaderConfig: info.LoaderConfig,
		Protected:    info.Protected,
		CreatedAt:    info.CreatedAt,
		UpdatedAt:    info.UpdatedAt,
	}

	return m.store.Save(record)
}

// ReloadAll 重新加载所有 keystore
func (m *Manager) ReloadAll() []error {
	m.mu.RLock()
	names := make([]string, 0, len(m.keyStoreInfo))
	for name := range m.keyStoreInfo {
		names = append(names, name)
	}
	m.mu.RUnlock()

	var errs []error
	for _, name := range names {
		if err := m.Reload(name); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

// LoadFromConfig 从简化配置加载 keystore
// config 可以是:
// - string: keystore 名称
// - map[string]string: 文件路径配置
// - *LoaderConfig: 完整配置
func (m *Manager) LoadFromConfig(config interface{}) (KeyStore, error) {
	switch v := config.(type) {
	case string:
		return m.GetKeyStore(v)
	case map[string]string:
		loaderConfig := LoaderConfig{
			Type:   LoaderTypeFile,
			Params: make(map[string]string),
		}
		if signCert, ok := v["sign-cert"]; ok {
			loaderConfig.Params["sign-cert"] = signCert
		}
		if signKey, ok := v["sign-key"]; ok {
			loaderConfig.Params["sign-key"] = signKey
		}
		if encCert, ok := v["enc-cert"]; ok {
			loaderConfig.Params["enc-cert"] = encCert
		}
		if encKey, ok := v["enc-key"]; ok {
			loaderConfig.Params["enc-key"] = encKey
		}
		loader, ok := m.loaders[LoaderTypeFile]
		if !ok {
			return nil, fmt.Errorf("文件加载器未注册")
		}
		return loader.Load(loaderConfig)
	case *LoaderConfig:
		loader, ok := m.loaders[v.Type]
		if !ok {
			return nil, fmt.Errorf("不支持的加载器类型: %s", v.Type)
		}
		return loader.Load(*v)
	default:
		return nil, fmt.Errorf("不支持的配置类型: %T", config)
	}
}
