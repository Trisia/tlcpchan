package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client HTTP客户端，用于与TLCP Channel API服务通信
type Client struct {
	// baseURL API服务基础URL
	baseURL string
	// httpClient HTTP客户端实例
	httpClient *http.Client
}

// NewClient 创建新的API客户端
// 参数:
//   - baseURL: API服务基础URL，格式: "http://host:port"
//
// 返回:
//   - *Client: API客户端实例
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Get 发送GET请求
// 参数:
//   - path: API路径，相对于baseURL
//
// 返回:
//   - []byte: 响应体数据
//   - error: 请求失败时返回错误
func (c *Client) Get(path string) ([]byte, error) {
	return c.doRequest(http.MethodGet, path, nil)
}

// Post 发送POST请求
// 参数:
//   - path: API路径，相对于baseURL
//   - data: 请求体数据，会自动序列化为JSON
//
// 返回:
//   - []byte: 响应体数据
//   - error: 请求失败时返回错误
func (c *Client) Post(path string, data interface{}) ([]byte, error) {
	return c.doRequest(http.MethodPost, path, data)
}

// Put 发送PUT请求
// 参数:
//   - path: API路径，相对于baseURL
//   - data: 请求体数据，会自动序列化为JSON
//
// 返回:
//   - []byte: 响应体数据
//   - error: 请求失败时返回错误
func (c *Client) Put(path string, data interface{}) ([]byte, error) {
	return c.doRequest(http.MethodPut, path, data)
}

// Delete 发送DELETE请求
// 参数:
//   - path: API路径，相对于baseURL
//
// 返回:
//   - error: 请求失败时返回错误
func (c *Client) Delete(path string) error {
	_, err := c.doRequest(http.MethodDelete, path, nil)
	return err
}

// doRequest 执行HTTP请求
// 参数:
//   - method: HTTP方法，如 "GET", "POST", "PUT", "DELETE"
//   - path: API路径，相对于baseURL
//   - data: 请求体数据，为nil时发送空请求体
//
// 返回:
//   - []byte: 响应体数据
//   - error: 请求失败时返回错误
//
// 注意: 自动设置Content-Type为application/json，状态码>=400时返回错误
func (c *Client) doRequest(method, path string, data interface{}) ([]byte, error) {
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}

	url, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return nil, fmt.Errorf("构建URL失败: %w", err)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("请求失败: %s - %s", resp.Status, string(respBody))
	}

	return respBody, nil
}

type Instance struct {
	Name    string         `json:"name"`
	Status  string         `json:"status"`
	Config  InstanceConfig `json:"config"`
	Enabled bool           `json:"enabled"`
}

type InstanceConfig struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Listen   string `json:"listen"`
	Target   string `json:"target"`
	Protocol string `json:"protocol"`
	Auth     string `json:"auth"`
	Enabled  bool   `json:"enabled"`
}

func (c *Client) ListInstances() ([]Instance, error) {
	data, err := c.Get("/api/v1/instances")
	if err != nil {
		return nil, err
	}
	var instances []Instance
	if err := json.Unmarshal(data, &instances); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return instances, nil
}

func (c *Client) GetInstance(name string) (*Instance, error) {
	data, err := c.Get("/api/v1/instances/" + url.PathEscape(name))
	if err != nil {
		return nil, err
	}
	var inst Instance
	if err := json.Unmarshal(data, &inst); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return &inst, nil
}

func (c *Client) CreateInstance(cfg *InstanceConfig) error {
	_, err := c.Post("/api/v1/instances", cfg)
	return err
}

func (c *Client) UpdateInstance(name string, cfg *InstanceConfig) error {
	_, err := c.Put("/api/v1/instances/"+url.PathEscape(name), cfg)
	return err
}

func (c *Client) DeleteInstance(name string) error {
	return c.Delete("/api/v1/instances/" + url.PathEscape(name))
}

func (c *Client) StartInstance(name string) error {
	_, err := c.Post("/api/v1/instances/"+url.PathEscape(name)+"/start", nil)
	return err
}

func (c *Client) StopInstance(name string) error {
	_, err := c.Post("/api/v1/instances/"+url.PathEscape(name)+"/stop", nil)
	return err
}

func (c *Client) ReloadInstance(name string) error {
	_, err := c.Post("/api/v1/instances/"+url.PathEscape(name)+"/reload", nil)
	return err
}

func (c *Client) ReloadInstanceCertificates(name string) error {
	_, err := c.Post("/api/v1/instances/"+url.PathEscape(name)+"/reload-certs", nil)
	return err
}

func (c *Client) InstanceStats(name string) (map[string]interface{}, error) {
	data, err := c.Get("/api/v1/instances/" + url.PathEscape(name) + "/stats")
	if err != nil {
		return nil, err
	}
	var stats map[string]interface{}
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return stats, nil
}

func (c *Client) InstanceLogs(name string) ([]map[string]interface{}, error) {
	data, err := c.Get("/api/v1/instances/" + url.PathEscape(name) + "/logs")
	if err != nil {
		return nil, err
	}
	var logs []map[string]interface{}
	if err := json.Unmarshal(data, &logs); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return logs, nil
}

type TrustedCertificate struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	SerialNumber string `json:"serialNumber,omitempty"`
	Subject      string `json:"subject,omitempty"`
	Issuer       string `json:"issuer,omitempty"`
	ExpiresAt    string `json:"expiresAt,omitempty"`
	IsCA         bool   `json:"isCA,omitempty"`
}

func (c *Client) ListTrustedCertificates() ([]TrustedCertificate, error) {
	data, err := c.Get("/api/v1/trusted")
	if err != nil {
		return nil, err
	}
	var certs []TrustedCertificate
	if err := json.Unmarshal(data, &certs); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return certs, nil
}

func (c *Client) DeleteTrustedCertificate(name string) error {
	u := fmt.Sprintf("/api/v1/trusted?name=%s", url.QueryEscape(name))
	return c.Delete(u)
}

func (c *Client) ReloadTrustedCertificates() error {
	_, err := c.Post("/api/v1/trusted/reload", nil)
	return err
}

func (c *Client) GetConfig() (map[string]interface{}, error) {
	data, err := c.Get("/api/v1/config")
	if err != nil {
		return nil, err
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return cfg, nil
}

func (c *Client) ReloadConfig() error {
	_, err := c.Post("/api/v1/config/reload", nil)
	return err
}

type SystemInfo struct {
	GoVersion    string `json:"go_version"`
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	NumCPU       int    `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
	MemAllocMB   uint64 `json:"mem_alloc_mb"`
	MemTotalMB   uint64 `json:"mem_total_mb"`
	MemSysMB     uint64 `json:"mem_sys_mb"`
	StartTime    string `json:"start_time"`
	Uptime       string `json:"uptime"`
}

func (c *Client) GetSystemInfo() (*SystemInfo, error) {
	data, err := c.Get("/api/v1/system/info")
	if err != nil {
		return nil, err
	}
	var info SystemInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return &info, nil
}

type HealthStatus struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

func (c *Client) HealthCheck() (*HealthStatus, error) {
	data, err := c.Get("/api/v1/system/health")
	if err != nil {
		return nil, err
	}
	var health HealthStatus
	if err := json.Unmarshal(data, &health); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return &health, nil
}

type VersionInfo struct {
	Version   string `json:"version"`
	GoVersion string `json:"go_version"`
}

func (c *Client) GetVersion() (*VersionInfo, error) {
	data, err := c.Get("/api/v1/system/version")
	if err != nil {
		return nil, err
	}
	var info VersionInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return &info, nil
}

// InstanceHealthCheckResult 实例健康检测结果
type InstanceHealthCheckResult struct {
	Success   bool                `json:"success"`
	LatencyMs float64             `json:"latency_ms"`
	Error     string              `json:"error,omitempty"`
	TLCPInfo  *ProtocolHealthInfo `json:"tlcp_info,omitempty"`
	TLSInfo   *ProtocolHealthInfo `json:"tls_info,omitempty"`
}

// ProtocolHealthInfo 协议健康信息
type ProtocolHealthInfo struct {
	Success           bool    `json:"success"`
	LatencyMs         float64 `json:"latency_ms"`
	CertValid         bool    `json:"cert_valid"`
	CertExpiry        string  `json:"cert_expiry,omitempty"`
	CertDaysRemaining int     `json:"cert_days_remaining,omitempty"`
	Error             string  `json:"error,omitempty"`
}

// CheckInstanceHealth 执行实例健康检测
func (c *Client) CheckInstanceHealth(name string, fullHandshake bool) (*InstanceHealthCheckResult, error) {
	path := fmt.Sprintf("/api/v1/instances/%s/health", name)
	data, err := c.Post(path, map[string]bool{"full_handshake": fullHandshake})
	if err != nil {
		return nil, err
	}
	var result InstanceHealthCheckResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return &result, nil
}
