package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/security"
	"github.com/Trisia/tlcpchan/security/certgen"
)

// CertificateManagerTool 证书管理工具
type CertificateManagerTool struct {
	*BaseTool
	rootCertMgr *security.RootCertManager
	cfg         *config.Config
	configPath  string
}

// NewCertificateManagerTool 创建证书管理工具
func NewCertificateManagerTool(rootCertMgr *security.RootCertManager, cfg *config.Config, configPath string) *CertificateManagerTool {
	return &CertificateManagerTool{
		BaseTool: NewBaseTool(
			"certificate_manager",
			"管理国密/TLS证书的操作，包括生成、导入、验证等",
			[]string{
				"generate_root_ca",
				"import_certificate",
				"list_certificates",
				"get_certificate",
				"delete_certificate",
				"validate_certificate",
			},
		),
		rootCertMgr: rootCertMgr,
		cfg:         cfg,
		configPath:  configPath,
	}
}

// Execute 执行工具方法
func (t *CertificateManagerTool) Execute(ctx context.Context, method string, params json.RawMessage) (interface{}, error) {
	switch method {
	case "generate_root_ca":
		var req GenerateRootCARequest
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("解析参数失败: %w", err)
		}
		return t.generateRootCA(ctx, req)
	case "import_certificate":
		var req ImportCertificateRequest
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("解析参数失败: %w", err)
		}
		return t.importCertificate(ctx, req)
	case "list_certificates":
		return t.listCertificates(ctx)
	case "get_certificate":
		var req struct {
			Filename string `json:"filename"`
		}
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("解析参数失败: %w", err)
		}
		return t.getCertificate(ctx, req.Filename)
	case "delete_certificate":
		var req struct {
			Filename string `json:"filename"`
		}
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("解析参数失败: %w", err)
		}
		return t.deleteCertificate(ctx, req.Filename)
	case "validate_certificate":
		var req struct {
			CertData string `json:"certData"`
		}
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("解析参数失败: %w", err)
		}
		return t.validateCertificate(ctx, req.CertData)
	default:
		return nil, fmt.Errorf("未知方法: %s", method)
	}
}

// GenerateRootCARequest 生成根CA请求
type GenerateRootCARequest struct {
	Type            string `json:"type"`
	CommonName      string `json:"commonName"`
	Country         string `json:"country,omitempty"`
	StateOrProvince string `json:"stateOrProvince,omitempty"`
	Locality        string `json:"locality,omitempty"`
	Org             string `json:"org,omitempty"`
	OrgUnit         string `json:"orgUnit,omitempty"`
	Years           int    `json:"years,omitempty"`
}

// generateRootCA 生成根CA
func (t *CertificateManagerTool) generateRootCA(ctx context.Context, req GenerateRootCARequest) (interface{}, error) {
	if req.CommonName == "" {
		return nil, fmt.Errorf("通用名称(CN)不能为空")
	}
	if req.Org == "" {
		req.Org = "tlcpchan"
	}
	if req.Years <= 0 {
		req.Years = 10
	}

	var rootCA *certgen.GeneratedCert
	var err error

	cfg := certgen.CertGenConfig{
		Type:            certgen.CertTypeRootCA,
		CommonName:      req.CommonName,
		Country:         req.Country,
		StateOrProvince: req.StateOrProvince,
		Locality:        req.Locality,
		Org:             req.Org,
		OrgUnit:         req.OrgUnit,
		Years:           req.Years,
	}

	if req.Type == "tls" {
		rootCA, err = certgen.GenerateTLSRootCA(cfg)
	} else {
		rootCA, err = certgen.GenerateTLCPRootCA(cfg)
	}

	if err != nil {
		return nil, fmt.Errorf("生成根CA失败: %w", err)
	}

	filename := req.CommonName + ".crt"
	cert, err := t.rootCertMgr.Add(filename, rootCA.CertPEM)
	if err != nil {
		return nil, fmt.Errorf("添加根证书失败: %w", err)
	}

	keystoreDir := t.cfg.GetKeyStoreStoreDir()
	certPath := fmt.Sprintf("%s/%s.crt", keystoreDir, req.CommonName)
	keyPath := fmt.Sprintf("%s/%s.key", keystoreDir, req.CommonName)

	if err := certgen.SaveCertToFile(rootCA.CertPEM, rootCA.KeyPEM, certPath, keyPath); err != nil {
		return nil, fmt.Errorf("保存证书和密钥失败: %w", err)
	}

	return cert, nil
}

// ImportCertificateRequest 导入证书请求
type ImportCertificateRequest struct {
	Filename string `json:"filename"`
	CertData string `json:"certData"`
	Format   string `json:"format,omitempty"`
}

// importCertificate 导入证书
func (t *CertificateManagerTool) importCertificate(ctx context.Context, req ImportCertificateRequest) (interface{}, error) {
	if req.Filename == "" {
		return nil, fmt.Errorf("文件名不能为空")
	}
	if req.CertData == "" {
		return nil, fmt.Errorf("证书数据不能为空")
	}

	certData, err := t.parseCertificateData(req.CertData, req.Format)
	if err != nil {
		return nil, fmt.Errorf("解析证书失败: %w", err)
	}

	cert, err := t.rootCertMgr.Add(req.Filename, certData)
	if err != nil {
		return nil, fmt.Errorf("添加证书失败: %w", err)
	}

	return cert, nil
}

// parseCertificateData 解析证书数据
func (t *CertificateManagerTool) parseCertificateData(data string, format string) ([]byte, error) {
	data = strings.TrimSpace(data)

	switch strings.ToLower(format) {
	case "base64":
		decoded, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			return nil, fmt.Errorf("Base64解码失败: %w", err)
		}
		return decoded, nil
	case "pem", "":
		if strings.HasPrefix(data, "-----") {
			return []byte(data), nil
		}
		decoded, err := base64.StdEncoding.DecodeString(data)
		if err == nil {
			return decoded, nil
		}
		return []byte(data), nil
	case "der":
		decoded, err := base64.StdEncoding.DecodeString(data)
		if err == nil {
			return decoded, nil
		}
		return []byte(data), nil
	default:
		if strings.HasPrefix(data, "-----") {
			return []byte(data), nil
		}
		decoded, err := base64.StdEncoding.DecodeString(data)
		if err == nil {
			return decoded, nil
		}
		return []byte(data), nil
	}
}

// listCertificates 列出所有证书
func (t *CertificateManagerTool) listCertificates(ctx context.Context) (interface{}, error) {
	certs := t.rootCertMgr.List()
	return certs, nil
}

// getCertificate 获取证书
func (t *CertificateManagerTool) getCertificate(ctx context.Context, filename string) (interface{}, error) {
	cert, err := t.rootCertMgr.Get(filename)
	if err != nil {
		return nil, fmt.Errorf("证书不存在: %s, %w", filename, err)
	}
	return cert, nil
}

// deleteCertificate 删除证书
func (t *CertificateManagerTool) deleteCertificate(ctx context.Context, filename string) (interface{}, error) {
	if err := t.rootCertMgr.Delete(filename); err != nil {
		return nil, fmt.Errorf("删除失败: %w", err)
	}
	return nil, nil
}

// validateCertificate 验证证书
func (t *CertificateManagerTool) validateCertificate(ctx context.Context, certData string) (interface{}, error) {
	var certBytes []byte
	var err error

	if strings.HasPrefix(certData, "-----") {
		certBytes = []byte(certData)
	} else {
		certBytes, err = base64.StdEncoding.DecodeString(certData)
		if err != nil {
			return nil, fmt.Errorf("Base64解码失败: %w", err)
		}
	}

	block, _ := pem.Decode(certBytes)
	if block == nil {
		return map[string]interface{}{
			"valid":   false,
			"message": "无法解析PEM格式证书",
		}, nil
	}

	return map[string]interface{}{
		"valid":   true,
		"message": "证书格式有效",
		"type":    block.Type,
	}, nil
}
