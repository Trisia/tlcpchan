package keystore

import (
	"fmt"
	"sync"
	"time"
)

// Manager keystore 管理器
// 负责管理 keystore 的创建、加载、获取和管理
// 注意：该模块不负责持久化，持久化由控制器层通过 config.Config.KeyStores 负责
type Manager struct {
	keyStores    map[string]KeyStore      // 已加载的 keystore 实例
	keyStoreInfo map[string]*KeyStoreInfo // keystore 元信息
	loaders      map[LoaderType]Loader    // 加载器映射
	mu           sync.RWMutex             // 读写锁，保证并发安全
}

// NewManager 创建 keystore 管理器
// 返回：
//   - *Manager: 新的管理器实例
func NewManager() *Manager {
	m := &Manager{
		keyStores:    make(map[string]KeyStore),
		keyStoreInfo: make(map[string]*KeyStoreInfo),
		loaders:      make(map[LoaderType]Loader),
	}

	m.loaders[LoaderTypeFile] = NewFileLoader("")
	m.loaders[LoaderTypeNamed] = NewNamedLoader(m)

	return m
}

// ConfigEntry 用于 LoadFromConfigs 的配置条目
// 用于从外部配置加载 keystores 时使用的简单结构
type ConfigEntry struct {
	Name   string            // keystore 名称
	Type   LoaderType        // 加载器类型
	Params map[string]string // 加载器参数
}

// LoadFromConfigs 从配置列表批量加载 keystores
// 参数：
//   - configs: 配置条目列表
//
// 返回：
//   - error: 加载失败时返回错误
//
// 注意：该方法用于服务启动时从配置文件初始化 keystores
func (m *Manager) LoadFromConfigs(configs []ConfigEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, cfg := range configs {
		loader, ok := m.loaders[cfg.Type]
		if !ok {
			return fmt.Errorf("不支持的加载器类型: %s", cfg.Type)
		}

		ks, err := loader.Load(cfg.Type, cfg.Params)
		if err != nil {
			return fmt.Errorf("加载keystore %s 失败: %w", cfg.Name, err)
		}

		now := time.Now()
		info := &KeyStoreInfo{
			Name:       cfg.Name,
			Type:       ks.Type(),
			LoaderType: cfg.Type,
			Params:     cfg.Params,
			Protected:  false,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		m.keyStoreInfo[cfg.Name] = info
		m.keyStores[cfg.Name] = ks
	}

	return nil
}

// RegisterLoader 注册自定义加载器
// 参数：
//   - loaderType: 加载器类型
//   - loader: 加载器实例
func (m *Manager) RegisterLoader(loaderType LoaderType, loader Loader) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loaders[loaderType] = loader
}

// Create 创建新的 keystore
// 参数：
//   - name: keystore 名称，必须唯一
//   - loaderType: 加载器类型
//   - params: 加载器参数
//   - protected: 是否受保护（受保护的 keystore 无法删除）
//
// 返回：
//   - *KeyStoreInfo: 创建成功返回 keystore 信息
//   - error: 创建失败返回错误
//
// 注意：该方法只创建内存中的 keystore，持久化由控制器层负责
func (m *Manager) Create(name string, loaderType LoaderType, params map[string]string, protected bool) (*KeyStoreInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.keyStoreInfo[name]; exists {
		return nil, fmt.Errorf("keystore %s 已存在", name)
	}

	loader, ok := m.loaders[loaderType]
	if !ok {
		return nil, fmt.Errorf("不支持的加载器类型: %s", loaderType)
	}

	ks, err := loader.Load(loaderType, params)
	if err != nil {
		return nil, fmt.Errorf("加载keystore失败: %w", err)
	}

	now := time.Now()
	info := &KeyStoreInfo{
		Name:       name,
		Type:       ks.Type(),
		LoaderType: loaderType,
		Params:     params,
		Protected:  protected,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	m.keyStoreInfo[name] = info
	m.keyStores[name] = ks

	return info, nil
}

// Delete 删除 keystore
// 参数：
//   - name: keystore 名称
//
// 返回：
//   - error: 删除失败返回错误
//
// 注意：该方法只从内存中删除 keystore，持久化由控制器层负责
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

	delete(m.keyStoreInfo, name)
	delete(m.keyStores, name)

	return nil
}

// Get 获取 keystore 元信息
// 参数：
//   - name: keystore 名称
//
// 返回：
//   - *KeyStoreInfo: keystore 元信息
//   - error: 获取失败返回错误
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
// 参数：
//   - name: keystore 名称
//
// 返回：
//   - KeyStore: keystore 实例
//   - error: 获取失败返回错误
//
// 注意：如果 keystore 尚未加载，会自动加载
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

	ks, err := loader.Load(info.LoaderType, info.Params)
	if err != nil {
		return nil, fmt.Errorf("加载keystore失败: %w", err)
	}

	m.keyStores[name] = ks
	return ks, nil
}

// List 列出所有 keystore 元信息
// 返回：
//   - []*KeyStoreInfo: keystore 元信息列表
func (m *Manager) List() []*KeyStoreInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*KeyStoreInfo, 0, len(m.keyStoreInfo))
	for _, info := range m.keyStoreInfo {
		result = append(result, info)
	}

	return result
}

// Reload 重新加载指定的 keystore
// 参数：
//   - name: keystore 名称
//
// 返回：
//   - error: 重载失败返回错误
//
// 注意：该方法会清空缓存，下次访问时重新加载证书
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

	return nil
}

// ReloadAll 重新加载所有 keystore
// 返回：
//   - []error: 重载失败的错误列表
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
// 参数：
//   - config: 配置，可以是:
//   - string: keystore 名称
//   - map[string]string: 文件路径配置
//
// 返回：
//   - KeyStore: keystore 实例
//   - error: 加载失败返回错误
func (m *Manager) LoadFromConfig(config interface{}) (KeyStore, error) {
	switch v := config.(type) {
	case string:
		return m.GetKeyStore(v)
	case map[string]string:
		params := make(map[string]string)
		if signCert, ok := v["sign-cert"]; ok {
			params["sign-cert"] = signCert
		}
		if signKey, ok := v["sign-key"]; ok {
			params["sign-key"] = signKey
		}
		if encCert, ok := v["enc-cert"]; ok {
			params["enc-cert"] = encCert
		}
		if encKey, ok := v["enc-key"]; ok {
			params["enc-key"] = encKey
		}
		loader, ok := m.loaders[LoaderTypeFile]
		if !ok {
			return nil, fmt.Errorf("文件加载器未注册")
		}
		return loader.Load(LoaderTypeFile, params)
	default:
		return nil, fmt.Errorf("不支持的配置类型: %T", config)
	}
}

// LoadAndRegister 加载 keystore 并注册到管理器中
// 参数：
//   - name: keystore 名称，如果为空则使用 suggestedName
//   - suggestedName: 建议的名称
//   - loaderTypeStr: 加载器类型字符串
//   - params: 加载器参数
//
// 返回：
//   - KeyStore: keystore 实例
//   - error: 加载失败返回错误
//
// 注意：创建的 keystore 会被标记为受保护，无法通过 API 删除
func (m *Manager) LoadAndRegister(name string, suggestedName string, loaderTypeStr string, params map[string]string) (KeyStore, error) {
	loaderType := LoaderType(loaderTypeStr)
	// 确定使用的名称
	useName := name
	if useName == "" {
		useName = suggestedName
	}
	if useName == "" {
		return nil, fmt.Errorf("名称不能为空")
	}

	// 检查是否已存在
	m.mu.RLock()
	if ks, exists := m.keyStores[useName]; exists {
		m.mu.RUnlock()
		return ks, nil
	}
	_, exists := m.keyStoreInfo[useName]
	m.mu.RUnlock()

	if exists {
		// 已存在记录，加载它
		return m.GetKeyStore(useName)
	}

	// 创建新的 keystore，设为受保护
	_, err := m.Create(useName, loaderType, params, true)
	if err != nil {
		return nil, err
	}

	return m.GetKeyStore(useName)
}
