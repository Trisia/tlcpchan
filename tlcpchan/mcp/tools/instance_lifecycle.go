package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/instance"
	"github.com/Trisia/tlcpchan/logger"
)

// InstanceLifecycleTool 实例生命周期管理工具
type InstanceLifecycleTool struct {
	*BaseTool
	manager *instance.Manager
	log     *logger.Logger
}

// NewInstanceLifecycleTool 创建实例生命周期管理工具
func NewInstanceLifecycleTool(manager *instance.Manager) *InstanceLifecycleTool {
	return &InstanceLifecycleTool{
		BaseTool: NewBaseTool(
			"instance_lifecycle",
			"管理TLCP/TLS代理实例的生命周期，包括创建、启动、停止、删除等操作",
			[]string{"list", "get", "create", "delete", "start", "stop", "restart", "reload", "get_stats", "get_health"},
		),
		manager: manager,
		log:     logger.Default(),
	}
}

// listParams 列出实例参数
type listParams struct {
}

// listResponse 列出实例响应
type listResponse struct {
	Instances []instanceInfo `json:"instances"`
}

// instanceInfo 实例信息
type instanceInfo struct {
	Name    string                 `json:"name"`
	Status  string                 `json:"status"`
	Config  *config.InstanceConfig `json:"config"`
	Enabled bool                   `json:"enabled"`
}

// getParams 获取实例参数
type getParams struct {
	Name string `json:"name"`
}

// getResponse 获取实例响应
type getResponse struct {
	Name   string                 `json:"name"`
	Status string                 `json:"status"`
	Config *config.InstanceConfig `json:"config"`
}

// createParams 创建实例参数
type createParams struct {
	Name               string             `json:"name"`
	Type               string             `json:"type"`
	Protocol           string             `json:"protocol"`
	Listen             string             `json:"listen"`
	Target             string             `json:"target"`
	Enabled            *bool              `json:"enabled,omitempty"`
	TLCPClientAuthType string             `veron:"tlcp_client_auth_type,omitempty"`
	TLSClientAuthType  string             `veron:"tls_client_auth_type,omitempty"`
	TLCP               *config.TLCPConfig `json:"tlcp,omitempty"`
	TLS                *config.TLSConfig  `json:"tls,omitempty"`
	ClientCA           []string           `json:"client_ca,omitempty"`
}

// createResponse 创建实例响应
type createResponse struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// deleteParams 删除实例参数
type deleteParams struct {
	Name string `json:"name"`
}

// startParams 启动实例参数
type startParams struct {
	Name string `json:"name"`
}

// startResponse 启动实例响应
type startResponse struct {
	Status string `json:"status"`
}

// stopParams 停止实例参数
type stopParams struct {
	Name string `json:"name"`
}

// stopResponse 停止实例响应
type stopResponse struct {
	Status string `json:"status"`
}

// restartParams 重启实例参数
type restartParams struct {
	Name string `json:"name"`
}

// restartResponse 重启实例响应
type restartResponse struct {
	Status string `json:"status"`
}

// reloadParams 重载实例参数
type reloadParams struct {
	Name string `json:"name"`
}

// reloadResponse 重载实例响应
type reloadResponse struct {
	Status string `json:"status"`
}

// getStatsParams 获取统计参数
type getStatsParams struct {
	Name string `json:"name"`
}

// getHealthParams 获取健康状态参数
type getHealthParams struct {
	Name     string `json:"name"`
	Timeout  *int   `json:"timeout,omitempty"`
	Protocol string `json:"protocol,omitempty"`
}

// getHealthResponse 获取健康状态响应
type getHealthResponse struct {
	Instance string      `json:"instance"`
	Results  interface{} `json:"results"`
}

// Execute 执行工具方法
func (t *InstanceLifecycleTool) Execute(ctx context.Context, method string, params json.RawMessage) (interface{}, error) {
	switch method {
	case "list":
		return t.list()
	case "get":
		var p getParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("无效的参数: %w", err)
		}
		return t.get(p.Name)
	case "create":
		var p createParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("无效的参数: %w", err)
		}
		return t.create(p)
	case "delete":
		var p deleteParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("无效的参数: %w", err)
		}
		return t.delete(p.Name)
	case "start":
		var p startParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("无效的参数: %w", err)
		}
		return t.start(p.Name)
	case "stop":
		var p stopParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("无效的参数: %w", err)
		}
		return t.stop(p.Name)
	case "restart":
		var p restartParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("无效的参数: %w", err)
		}
		return t.restart(p.Name)
	case "reload":
		var p reloadParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("无效的参数: %w", err)
		}
		return t.reload(p.Name)
	case "get_stats":
		var p getStatsParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("无效的参数: %w", err)
		}
		return t.getStats(p.Name)
	case "get_health":
		var p getHealthParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("无效的参数: %w", err)
		}
		return t.getHealth(p)
	default:
		return nil, fmt.Errorf("未知方法: %s", method)
	}
}

// list 列出所有实例
func (t *InstanceLifecycleTool) list() (*listResponse, error) {
	instances := t.manager.List()
	result := make([]instanceInfo, len(instances))
	for i, inst := range instances {
		cfg := inst.Config()
		result[i] = instanceInfo{
			Name:    inst.Name(),
			Status:  string(inst.Status()),
			Config:  cfg,
			Enabled: cfg.Enabled,
		}
	}
	return &listResponse{Instances: result}, nil
}

// get 获取实例详情
func (t *InstanceLifecycleTool) get(name string) (*getResponse, error) {
	inst, ok := t.manager.Get(name)
	if !ok {
		return nil, fmt.Errorf("实例不存在: %s", name)
	}
	return &getResponse{
		Name:   inst.Name(),
		Status: string(inst.Status()),
		Config: inst.Config(),
	}, nil
}

// create 创建实例
func (t *InstanceLifecycleTool) create(params createParams) (*createResponse, error) {
	cfg := &config.InstanceConfig{
		Name:     params.Name,
		Type:     params.Type,
		Protocol: params.Protocol,
		Listen:   params.Listen,
		Target:   params.Target,
		Enabled:  true,
	}

	if params.Enabled != nil {
		cfg.Enabled = *params.Enabled
	}
	if params.TLCP != nil {
		cfg.TLCP = *params.TLCP
		if params.TLCPClientAuthType != "" {
			cfg.TLCP.ClientAuthType = params.TLCPClientAuthType
		}
	}
	if params.TLS != nil {
		cfg.TLS = *params.TLS
		if params.TLSClientAuthType != "" {
			cfg.TLS.ClientAuthType = params.TLSClientAuthType
		}
	}
	if params.ClientCA != nil {
		cfg.ClientCA = params.ClientCA
	}

	inst, err := t.manager.Create(cfg)
	if err != nil {
		return nil, fmt.Errorf("创建实例失败: %w", err)
	}

	t.log.Info("创建实例成功: %s", params.Name)
	return &createResponse{
		Name:   inst.Name(),
		Status: string(inst.Status()),
	}, nil
}

// delete 删除实例
func (t *InstanceLifecycleTool) delete(name string) (map[string]interface{}, error) {
	if err := t.manager.Delete(name); err != nil {
		return nil, fmt.Errorf("删除实例失败: %w", err)
	}
	t.log.Info("删除实例成功: %s", name)
	return map[string]interface{}{"message": fmt.Sprintf("实例 %s 删除成功", name)}, nil
}

// start 启动实例
func (t *InstanceLifecycleTool) start(name string) (*startResponse, error) {
	inst, ok := t.manager.Get(name)
	if !ok {
		return nil, fmt.Errorf("实例不存在: %s", name)
	}
	if err := inst.Start(); err != nil {
		return nil, fmt.Errorf("启动实例失败: %w", err)
	}
	t.log.Info("启动实例成功: %s", name)
	return &startResponse{Status: string(inst.Status())}, nil
}

// stop 停止实例
func (t *InstanceLifecycleTool) stop(name string) (*stopResponse, error) {
	inst, ok := t.manager.Get(name)
	if !ok {
		return nil, fmt.Errorf("实例不存在: %s", name)
	}
	if err := inst.Stop(); err != nil {
		return nil, fmt.Errorf("停止实例失败: %w", err)
	}
	t.log.Info("停止实例成功: %s", name)
	return &stopResponse{Status: string(inst.Status())}, nil
}

// restart 重启实例
func (t *InstanceLifecycleTool) restart(name string) (*restartResponse, error) {
	inst, ok := t.manager.Get(name)
	if !ok {
		return nil, fmt.Errorf("实例不存在: %s", name)
	}
	cfg := inst.Config()
	if err := inst.Restart(cfg); err != nil {
		return nil, fmt.Errorf("重启实例失败: %w", err)
	}
	t.log.Info("重启实例成功: %s", name)
	return &restartResponse{Status: string(inst.Status())}, nil
}

// reload 重载实例
func (t *InstanceLifecycleTool) reload(name string) (*reloadResponse, error) {
	inst, ok := t.manager.Get(name)
	if !ok {
		return nil, fmt.Errorf("实例不存在: %s", name)
	}
	cfg := inst.Config()
	if err := inst.Reload(cfg); err != nil {
		return nil, fmt.Errorf("重载实例失败: %w", err)
	}
	t.log.Info("重载实例成功: %s", name)
	return &reloadResponse{Status: string(inst.Status())}, nil
}

// getStats 获取实例统计
func (t *InstanceLifecycleTool) getStats(name string) (interface{}, error) {
	inst, ok := t.manager.Get(name)
	if !ok {
		return nil, fmt.Errorf("实例不存在: %s", name)
	}
	return inst.Stats(), nil
}

// getHealth 获取实例健康状态
func (t *InstanceLifecycleTool) getHealth(params getHealthParams) (*getHealthResponse, error) {
	inst, ok := t.manager.Get(params.Name)
	if !ok {
		return nil, fmt.Errorf("实例不存在: %s", params.Name)
	}

	results := make([]interface{}, 0)
	t.log.Debug("获取实例健康状态: %s", params.Name)
	_ = inst
	return &getHealthResponse{
		Instance: params.Name,
		Results:  results,
	}, nil
}
