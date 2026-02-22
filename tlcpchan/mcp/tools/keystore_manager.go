package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/security"
	"github.com/Trisia/tlcpchan/security/certgen"
	"github.com/Trisia/tlcpchan/security/keystore"
)

// KeyStoreManagerTool keystore管理工具
type KeyStoreManagerTool struct {
	*BaseTool
	keyStoreMgr *security.KeyStoreManager
	cfg         *config.Config
	configPath  string
}

// NewKeyStoreManagerTool 创建keystore管理工具
func NewKeyStoreManagerTool(keyStoreMgr *security.KeyStoreManager, cfg *config.Config, configPath string) *KeyStoreManagerTool {
	return &KeyStoreManagerTool{
		BaseTool: NewBaseTool(
			"keystore_manager",
			"管理TLCP/TLS密钥存储库的完整生命周期，包括创建、生成、删除等操作",
			[]string{
				"list",
				"get",
				"create",
				"generate",
				"delete",
			},
		),
		keyStoreMgr: keyStoreMgr,
		cfg:         cfg,
		configPath:  configPath,
	}
}

// Execute 执行工具方法
func (t *KeyStoreManagerTool) Execute(ctx context.Context, method string, params json.RawMessage) (interface{}, error) {
	switch method {
	case "list":
		return t.list(ctx)
	case "get":
		var req struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("解析参数失败: %w", err)
		}
		return t.get(ctx, req.Name)
	case "generate":
		var req GenerateKeyStoreRequest
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("解析参数失败: %w", err)
		}
		return t.generate(ctx, req)
	case "delete":
		var req struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("解析参数失败: %w", err)
		}
		return t.delete(ctx, req.Name)
	default:
		return nil, fmt.Errorf("未知方法: %s", method)
	}
}

// list 列出所有keystores
func (t *KeyStoreManagerTool) list(ctx context.Context) (interface{}, error) {
	keyStores := t.keyStoreMgr.List()
	return keyStores, nil
}

// get 获取指定keystore
func (t *KeyStoreManagerTool) get(ctx context.Context, name string) (interface{}, error) {
	ks, err := t.keyStoreMgr.Get(name)
	if err != nil {
		return nil, fmt.Errorf("keystore不存在: %s, %w", name, err)
	}
	return ks, nil
}

// GenerateKeyStoreRequest 生成keystore请求
type GenerateKeyStoreRequest struct {
	Name            string `json:"name"`
	Type            string `json:"type"`
	CommonName      string `json:"commonName"`
	Country         string `json:"country,omitempty"`
	StateOrProvince string `json:"stateOrProvince,omitempty"`
	Locality        string `json:"locality,omitempty"`
	Org             string `json:"org,omitempty"`
	OrgUnit         string `json:"orgUnit,omitempty"`
	Years           int    `json:"years,omitempty"`
}

// generate 生成包含证书的keystore
func (t *KeyStoreManagerTool) generate(ctx context.Context, req GenerateKeyStoreRequest) (interface{}, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("名称不能为空")
	}
	if req.Type == "" {
		return nil, fmt.Errorf("类型不能为空")
	}
	if req.CommonName == "" {
		return nil, fmt.Errorf("通用名称(CN)不能为空")
	}
	if req.Org == "" {
		req.Org = "tlcpchan"
	}
	if req.Years <= 0 {
		req.Years = 5
	}

	keystoreDir := t.cfg.GetKeyStoreStoreDir()

	var params map[string]string

	if req.Type == "tlcp" {
		signCertPath := fmt.Sprintf("%s/%s-sign.crt", keystoreDir, req.Name)
		signKeyPath := fmt.Sprintf("%s/%s-sign.key", keystoreDir, req.Name)
		encCertPath := fmt.Sprintf("%s/%s-enc.crt", keystoreDir, req.Name)
		encKeyPath := fmt.Sprintf("%s/%s-enc.key", keystoreDir, req.Name)

		_, err := certgen.GenerateTLCPRootCA(certgen.CertGenConfig{
			Type:            certgen.CertTypeRootCA,
			CommonName:      req.Name + "-ca",
			Country:         req.Country,
			StateOrProvince: req.StateOrProvince,
			Locality:        req.Locality,
			Org:             req.Org,
			OrgUnit:         req.OrgUnit,
			Years:           10,
		})
		if err != nil {
			return nil, fmt.Errorf("生成TLCP根CA失败: %w", err)
		}

		signCfg := certgen.CertGenConfig{
			Type:            certgen.CertTypeTLCPSign,
			CommonName:      req.CommonName,
			Country:         req.Country,
			StateOrProvince: req.StateOrProvince,
			Locality:        req.Locality,
			Org:             req.Org,
			OrgUnit:         req.OrgUnit,
			Years:           req.Years,
		}
		encCfg := certgen.CertGenConfig{
			Type:            certgen.CertTypeTLCPEnc,
			CommonName:      req.CommonName,
			Country:         req.Country,
			StateOrProvince: req.StateOrProvince,
			Locality:        req.Locality,
			Org:             req.Org,
			OrgUnit:         req.OrgUnit,
			Years:           req.Years,
		}

		signCert, encCert, err := certgen.GenerateTLCPPair(nil, nil, signCfg, encCfg)
		if err != nil {
			return nil, fmt.Errorf("生成TLCP证书失败: %w", err)
		}

		if err := certgen.SaveCertToFile(signCert.CertPEM, signCert.KeyPEM, signCertPath, signKeyPath); err != nil {
			return nil, fmt.Errorf("保存签名证书失败: %w", err)
		}
		if err := certgen.SaveCertToFile(encCert.CertPEM, encCert.KeyPEM, encCertPath, encKeyPath); err != nil {
			return nil, fmt.Errorf("保存加密证书失败: %w", err)
		}

		params = map[string]string{
			"sign-cert": fmt.Sprintf("./keystores/%s-sign.crt", req.Name),
			"sign-key":  fmt.Sprintf("./keystores/%s-sign.key", req.Name),
			"enc-cert":  fmt.Sprintf("./keystores/%s-enc.crt", req.Name),
			"enc-key":   fmt.Sprintf("./keystores/%s-enc.key", req.Name),
		}
	} else {
		certPath := fmt.Sprintf("%s/%s.crt", keystoreDir, req.Name)
		keyPath := fmt.Sprintf("%s/%s.key", keystoreDir, req.Name)

		_, err := certgen.GenerateTLSRootCA(certgen.CertGenConfig{
			Type:            certgen.CertTypeRootCA,
			CommonName:      req.Name + "-ca",
			Country:         req.Country,
			StateOrProvince: req.StateOrProvince,
			Locality:        req.Locality,
			Org:             req.Org,
			OrgUnit:         req.OrgUnit,
			Years:           10,
		})
		if err != nil {
			return nil, fmt.Errorf("生成TLS根CA失败: %w", err)
		}

		cert, err := certgen.GenerateTLSCert(nil, nil, certgen.CertGenConfig{
			Type:            certgen.CertTypeTLS,
			CommonName:      req.CommonName,
			Country:         req.Country,
			StateOrProvince: req.StateOrProvince,
			Locality:        req.Locality,
			Org:             req.Org,
			OrgUnit:         req.OrgUnit,
			Years:           req.Years,
		})
		if err != nil {
			return nil, fmt.Errorf("生成TLS证书失败: %w", err)
		}

		if err := certgen.SaveCertToFile(cert.CertPEM, cert.KeyPEM, certPath, keyPath); err != nil {
			return nil, fmt.Errorf("保存证书失败: %w", err)
		}

		params = map[string]string{
			"cert": fmt.Sprintf("./keystores/%s.crt", req.Name),
			"key":  fmt.Sprintf("./keystores/%s.key", req.Name),
		}
	}

	ks, err := t.keyStoreMgr.Create(req.Name, keystore.LoaderTypeFile, params, false)
	if err != nil {
		return nil, fmt.Errorf("创建keystore失败: %w", err)
	}

	t.cfg.KeyStores = append(t.cfg.KeyStores, config.KeyStoreConfig{
		Name:   req.Name,
		Type:   keystore.LoaderTypeFile,
		Params: params,
	})

	if err := config.Save(t.configPath, t.cfg); err != nil {
		return nil, fmt.Errorf("保存配置失败: %w", err)
	}

	return ks, nil
}

// delete 删除keystore
func (t *KeyStoreManagerTool) delete(ctx context.Context, name string) (interface{}, error) {
	if err := t.keyStoreMgr.Delete(name); err != nil {
		return nil, fmt.Errorf("删除失败: %w", err)
	}

	newKeyStores := make([]config.KeyStoreConfig, 0, len(t.cfg.KeyStores))
	for _, ks := range t.cfg.KeyStores {
		if ks.Name != name {
			newKeyStores = append(newKeyStores, ks)
		}
	}
	t.cfg.KeyStores = newKeyStores

	if err := config.Save(t.configPath, t.cfg); err != nil {
		return nil, fmt.Errorf("保存配置失败: %w", err)
	}

	return nil, nil
}
