package instance

import (
	"fmt"
	"sync"

	"github.com/Trisia/tlcpchan/cert"
	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/logger"
)

// Manager 实例管理器，负责代理实例的创建、启动、停止和删除
type Manager struct {
	// instances 实例映射，key为实例名称
	instances   map[string]Instance
	mu          sync.RWMutex
	logger      *logger.Logger
	certManager *cert.Manager
}

// NewManager 创建新的实例管理器
// 参数:
//   - log: 日志记录器
//   - certMgr: 证书管理器
//
// 返回:
//   - *Manager: 实例管理器实例
func NewManager(log *logger.Logger, certMgr *cert.Manager) *Manager {
	return &Manager{
		instances:   make(map[string]Instance),
		logger:      log,
		certManager: certMgr,
	}
}

// Create 创建新的代理实例
// 参数:
//   - cfg: 实例配置
//
// 返回:
//   - Instance: 创建的实例
//   - error: 实例已存在或创建失败时返回错误
func (m *Manager) Create(cfg *config.InstanceConfig) (Instance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.instances[cfg.Name]; exists {
		return nil, fmt.Errorf("实例 %s 已存在", cfg.Name)
	}

	inst, err := NewInstance(cfg, m.certManager, m.logger)
	if err != nil {
		return nil, fmt.Errorf("创建实例失败: %w", err)
	}

	m.instances[cfg.Name] = inst
	m.logger.Info("创建实例: %s, 类型: %s", cfg.Name, cfg.Type)
	return inst, nil
}

func (m *Manager) Get(name string) (Instance, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	inst, ok := m.instances[name]
	return inst, ok
}

func (m *Manager) List() []Instance {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]Instance, 0, len(m.instances))
	for _, inst := range m.instances {
		list = append(list, inst)
	}
	return list
}

// Delete 删除代理实例
// 参数:
//   - name: 实例名称
//
// 返回:
//   - error: 实例不存在或正在运行时返回错误
//
// 注意: 正在运行的实例必须先停止才能删除
func (m *Manager) Delete(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	inst, ok := m.instances[name]
	if !ok {
		return fmt.Errorf("实例 %s 不存在", name)
	}

	if inst.Status() == StatusRunning {
		return fmt.Errorf("实例 %s 正在运行，请先停止", name)
	}

	delete(m.instances, name)
	m.logger.Info("删除实例: %s", name)
	return nil
}

// StartAll 启动所有已启用的实例
// 返回:
//   - []error: 启动失败的错误列表，为空表示全部成功
//
// 注意: 未启用的实例(Enabled=false)会被跳过
func (m *Manager) StartAll() []error {
	m.mu.RLock()
	instances := make([]Instance, 0, len(m.instances))
	for _, inst := range m.instances {
		instances = append(instances, inst)
	}
	m.mu.RUnlock()

	var errors []error
	for _, inst := range instances {
		cfg := inst.Config()
		if !cfg.Enabled {
			m.logger.Debug("实例 %s 未启用，跳过启动", inst.Name())
			continue
		}
		if inst.Status() == StatusRunning {
			continue
		}
		if err := inst.Start(); err != nil {
			m.logger.Error("启动实例 %s 失败: %v", inst.Name(), err)
			errors = append(errors, fmt.Errorf("启动实例 %s 失败: %w", inst.Name(), err))
		} else {
			m.logger.Info("启动实例 %s 成功", inst.Name())
		}
	}
	return errors
}

// StopAll 停止所有运行中的实例
func (m *Manager) StopAll() {
	m.mu.RLock()
	instances := make([]Instance, 0, len(m.instances))
	for _, inst := range m.instances {
		instances = append(instances, inst)
	}
	m.mu.RUnlock()

	for _, inst := range instances {
		if inst.Status() != StatusRunning {
			continue
		}
		if err := inst.Stop(); err != nil {
			m.logger.Error("停止实例 %s 失败: %v", inst.Name(), err)
		} else {
			m.logger.Info("停止实例 %s 成功", inst.Name())
		}
	}
}
