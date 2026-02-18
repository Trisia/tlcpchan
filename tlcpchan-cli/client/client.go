package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
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
	Name       string      `json:"name"`
	Type       string      `json:"type"`
	Listen     string      `json:"listen"`
	Target     string      `json:"target"`
	Protocol   string      `json:"protocol"`
	Auth       string      `json:"auth,omitempty"`
	Enabled    bool        `json:"enabled"`
	ClientCA   []string    `json:"clientCa,omitempty"`
	ServerCA   []string    `json:"serverCa,omitempty"`
	TLCP       *TLCPConfig `json:"tlcp,omitempty"`
	TLS        *TLSConfig  `json:"tls,omitempty"`
	HTTP       *HTTPConfig `json:"http,omitempty"`
	SNI        string      `json:"sni,omitempty"`
	BufferSize int         `json:"bufferSize,omitempty"`
}

type TLCPConfig struct {
	Auth               string   `json:"auth,omitempty"`
	MinVersion         string   `json:"minVersion,omitempty"`
	MaxVersion         string   `json:"maxVersion,omitempty"`
	CipherSuites       []string `json:"cipherSuites,omitempty"`
	CurvePreferences   []string `json:"curvePreferences,omitempty"`
	SessionTickets     bool     `json:"sessionTickets,omitempty"`
	SessionCache       bool     `json:"sessionCache,omitempty"`
	InsecureSkipVerify bool     `json:"insecureSkipVerify,omitempty"`
}

type TLSConfig struct {
	Auth               string   `json:"auth,omitempty"`
	MinVersion         string   `json:"minVersion,omitempty"`
	MaxVersion         string   `json:"maxVersion,omitempty"`
	CipherSuites       []string `json:"cipherSuites,omitempty"`
	CurvePreferences   []string `json:"curvePreferences,omitempty"`
	SessionTickets     bool     `json:"sessionTickets,omitempty"`
	SessionCache       bool     `json:"sessionCache,omitempty"`
	InsecureSkipVerify bool     `json:"insecureSkipVerify,omitempty"`
}

type HTTPConfig struct {
	RequestHeaders  HeadersConfig `json:"requestHeaders,omitempty"`
	ResponseHeaders HeadersConfig `json:"responseHeaders,omitempty"`
}

type HeadersConfig struct {
	Add    map[string]string `json:"add,omitempty"`
	Remove []string          `json:"remove,omitempty"`
	Set    map[string]string `json:"set,omitempty"`
}

func (c *Client) ListInstances() ([]Instance, error) {
	data, err := c.Get("/api/instances")
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
	data, err := c.Get("/api/instances/" + url.PathEscape(name))
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
	_, err := c.Post("/api/instances", cfg)
	return err
}

func (c *Client) UpdateInstance(name string, cfg *InstanceConfig) error {
	_, err := c.Put("/api/instances/"+url.PathEscape(name), cfg)
	return err
}

func (c *Client) DeleteInstance(name string) error {
	return c.Delete("/api/instances/" + url.PathEscape(name))
}

func (c *Client) StartInstance(name string) error {
	_, err := c.Post("/api/instances/"+url.PathEscape(name)+"/start", nil)
	return err
}

func (c *Client) StopInstance(name string) error {
	_, err := c.Post("/api/instances/"+url.PathEscape(name)+"/stop", nil)
	return err
}

func (c *Client) ReloadInstance(name string) error {
	_, err := c.Post("/api/instances/"+url.PathEscape(name)+"/reload", nil)
	return err
}

func (c *Client) RestartInstance(name string) error {
	_, err := c.Post("/api/instances/"+url.PathEscape(name)+"/restart", nil)
	return err
}

func (c *Client) InstanceStats(name string) (map[string]interface{}, error) {
	data, err := c.Get("/api/instances/" + url.PathEscape(name) + "/stats")
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
	data, err := c.Get("/api/instances/" + url.PathEscape(name) + "/logs")
	if err != nil {
		return nil, err
	}
	var logs []map[string]interface{}
	if err := json.Unmarshal(data, &logs); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return logs, nil
}

type KeyStoreInfo struct {
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	LoaderType string            `json:"loaderType"`
	Params     map[string]string `json:"params"`
	Protected  bool              `json:"protected"`
	CreatedAt  string            `json:"createdAt"`
	UpdatedAt  string            `json:"updatedAt"`
}

type GenerateKeyStoreRequest struct {
	Name           string                     `json:"name"`
	Type           string                     `json:"type"`
	Protected      bool                       `json:"protected"`
	CertConfig     GenerateKeyStoreCertConfig `json:"certConfig"`
	SignerKeyStore string                     `json:"signerKeyStore,omitempty"`
}

type GenerateKeyStoreCertConfig struct {
	CommonName      string   `json:"commonName"`
	Country         string   `json:"country,omitempty"`
	StateOrProvince string   `json:"stateOrProvince,omitempty"`
	Locality        string   `json:"locality,omitempty"`
	Org             string   `json:"org,omitempty"`
	OrgUnit         string   `json:"orgUnit,omitempty"`
	EmailAddress    string   `json:"emailAddress,omitempty"`
	Years           int      `json:"years,omitempty"`
	Days            int      `json:"days,omitempty"`
	KeyAlgorithm    string   `json:"keyAlgorithm,omitempty"`
	KeyBits         int      `json:"keyBits,omitempty"`
	DNSNames        []string `json:"dnsNames,omitempty"`
	IPAddresses     []string `json:"ipAddresses,omitempty"`
}

func (c *Client) ListKeyStores() ([]KeyStoreInfo, error) {
	data, err := c.Get("/api/security/keystores")
	if err != nil {
		return nil, err
	}
	var keyStores []KeyStoreInfo
	if err := json.Unmarshal(data, &keyStores); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return keyStores, nil
}

func (c *Client) GetKeyStore(name string) (*KeyStoreInfo, error) {
	data, err := c.Get("/api/security/keystores/" + url.PathEscape(name))
	if err != nil {
		return nil, err
	}
	var ks KeyStoreInfo
	if err := json.Unmarshal(data, &ks); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return &ks, nil
}

func (c *Client) CreateKeyStore(name string, loaderType string, params map[string]string, protected bool) (*KeyStoreInfo, error) {
	req := struct {
		Name       string            `json:"name"`
		LoaderType string            `json:"loaderType"`
		Params     map[string]string `json:"params"`
		Protected  bool              `json:"protected"`
	}{
		Name:       name,
		LoaderType: loaderType,
		Params:     params,
		Protected:  protected,
	}

	data, err := c.Post("/api/security/keystores", req)
	if err != nil {
		return nil, err
	}

	var ks KeyStoreInfo
	if err := json.Unmarshal(data, &ks); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return &ks, nil
}

func (c *Client) CreateKeyStoreWithFiles(name string, loaderType string, files map[string][]byte, protected bool) (*KeyStoreInfo, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	_ = writer.WriteField("name", name)
	_ = writer.WriteField("loaderType", loaderType)
	_ = writer.WriteField("protected", fmt.Sprintf("%t", protected))

	for fieldName, fileData := range files {
		part, err := writer.CreateFormFile(fieldName, fieldName)
		if err != nil {
			return nil, fmt.Errorf("创建表单字段失败: %w", err)
		}
		_, _ = part.Write(fileData)
	}

	_ = writer.Close()

	url, err := url.JoinPath(c.baseURL, "/api/security/keystores")
	if err != nil {
		return nil, fmt.Errorf("构建URL失败: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, &body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

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

	var ks KeyStoreInfo
	if err := json.Unmarshal(respBody, &ks); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return &ks, nil
}

func (c *Client) GenerateKeyStore(req GenerateKeyStoreRequest) (*KeyStoreInfo, error) {
	data, err := c.Post("/api/security/keystores/generate", req)
	if err != nil {
		return nil, err
	}
	var ks KeyStoreInfo
	if err := json.Unmarshal(data, &ks); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return &ks, nil
}

func (c *Client) DeleteKeyStore(name string) error {
	return c.Delete("/api/security/keystores/" + url.PathEscape(name))
}

func (c *Client) ReloadKeyStore(name string) error {
	_, err := c.Post("/api/security/keystores/"+url.PathEscape(name)+"/reload", nil)
	return err
}

type RootCertInfo struct {
	Filename string `json:"filename"`
	Subject  string `json:"subject"`
	Issuer   string `json:"issuer"`
	NotAfter string `json:"notAfter"`
}

type GenerateRootCARequest struct {
	CommonName      string `json:"commonName"`
	Country         string `json:"country,omitempty"`
	StateOrProvince string `json:"stateOrProvince,omitempty"`
	Locality        string `json:"locality,omitempty"`
	Org             string `json:"org,omitempty"`
	OrgUnit         string `json:"orgUnit,omitempty"`
	EmailAddress    string `json:"emailAddress,omitempty"`
	Years           int    `json:"years,omitempty"`
	Days            int    `json:"days,omitempty"`
}

func (c *Client) ListRootCerts() ([]RootCertInfo, error) {
	data, err := c.Get("/api/security/rootcerts")
	if err != nil {
		return nil, err
	}
	var certs []RootCertInfo
	if err := json.Unmarshal(data, &certs); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return certs, nil
}

func (c *Client) GetRootCert(filename string) (*RootCertInfo, error) {
	data, err := c.Get("/api/security/rootcerts/" + url.PathEscape(filename))
	if err != nil {
		return nil, err
	}
	var cert RootCertInfo
	if err := json.Unmarshal(data, &cert); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return &cert, nil
}

func (c *Client) AddRootCert(filename string, certData []byte) (*RootCertInfo, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	_ = writer.WriteField("filename", filename)

	part, err := writer.CreateFormFile("cert", filename)
	if err != nil {
		return nil, fmt.Errorf("创建表单字段失败: %w", err)
	}
	_, _ = part.Write(certData)

	_ = writer.Close()

	url, err := url.JoinPath(c.baseURL, "/api/security/rootcerts")
	if err != nil {
		return nil, fmt.Errorf("构建URL失败: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, &body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

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

	var cert RootCertInfo
	if err := json.Unmarshal(respBody, &cert); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return &cert, nil
}

func (c *Client) GenerateRootCA(req GenerateRootCARequest) (*RootCertInfo, error) {
	data, err := c.Post("/api/security/rootcerts/generate", req)
	if err != nil {
		return nil, err
	}
	var cert RootCertInfo
	if err := json.Unmarshal(data, &cert); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return &cert, nil
}

func (c *Client) DeleteRootCert(filename string) error {
	return c.Delete("/api/security/rootcerts/" + url.PathEscape(filename))
}

func (c *Client) ReloadRootCerts() error {
	_, err := c.Post("/api/security/rootcerts/reload", nil)
	return err
}

func (c *Client) GetConfig() (map[string]interface{}, error) {
	data, err := c.Get("/api/config")
	if err != nil {
		return nil, err
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return cfg, nil
}

func (c *Client) UpdateConfig(cfg map[string]interface{}) (map[string]interface{}, error) {
	data, err := c.Post("/api/config", cfg)
	if err != nil {
		return nil, err
	}
	var newCfg map[string]interface{}
	if err := json.Unmarshal(data, &newCfg); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return newCfg, nil
}

func (c *Client) ReloadConfig() (map[string]interface{}, error) {
	data, err := c.Post("/api/config/reload", nil)
	if err != nil {
		return nil, err
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return cfg, nil
}

type SystemInfo struct {
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
	data, err := c.Get("/api/system/info")
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
	data, err := c.Get("/api/system/health")
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
	Version string `json:"version"`
}

func (c *Client) GetVersion() (*VersionInfo, error) {
	data, err := c.Get("/api/system/version")
	if err != nil {
		return nil, err
	}
	var info VersionInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return &info, nil
}

type HealthCheckResult struct {
	Protocol string `json:"protocol"`
	Success  bool   `json:"success"`
	Latency  int64  `json:"latency_ms"`
	Error    string `json:"error,omitempty"`
}

type InstanceHealthResponse struct {
	Instance string              `json:"instance"`
	Results  []HealthCheckResult `json:"results"`
}

func (c *Client) InstanceHealth(name string, timeout *int) (*InstanceHealthResponse, error) {
	path := "/api/instances/" + url.PathEscape(name) + "/health"
	if timeout != nil {
		path += fmt.Sprintf("?timeout=%d", *timeout)
	}
	data, err := c.Get(path)
	if err != nil {
		return nil, err
	}
	var resp InstanceHealthResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return &resp, nil
}
