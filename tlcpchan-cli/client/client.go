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

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) Get(path string) ([]byte, error) {
	return c.doRequest(http.MethodGet, path, nil)
}

func (c *Client) Post(path string, data interface{}) ([]byte, error) {
	return c.doRequest(http.MethodPost, path, data)
}

func (c *Client) Put(path string, data interface{}) ([]byte, error) {
	return c.doRequest(http.MethodPut, path, data)
}

func (c *Client) Delete(path string) error {
	_, err := c.doRequest(http.MethodDelete, path, nil)
	return err
}

func (c *Client) doRequest(method, path string, data interface{}) ([]byte, error) {
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}

	u, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return nil, fmt.Errorf("构建URL失败: %w", err)
	}

	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API错误(%d): %s", resp.StatusCode, string(respBody))
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

type Certificate struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

func (c *Client) ListCertificates() ([]Certificate, error) {
	data, err := c.Get("/api/v1/certificates")
	if err != nil {
		return nil, err
	}
	var certs []Certificate
	if err := json.Unmarshal(data, &certs); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return certs, nil
}

func (c *Client) DeleteCertificate(name string) error {
	u := fmt.Sprintf("/api/v1/certificates?name=%s", url.QueryEscape(name))
	return c.Delete(u)
}

func (c *Client) ReloadCertificates() error {
	_, err := c.Post("/api/v1/certificates/reload", nil)
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
