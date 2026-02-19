package initialization

import (
	"os"
	"path/filepath"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/logger"
	"github.com/Trisia/tlcpchan/security/certgen"
	"github.com/Trisia/tlcpchan/security/keystore"
)

// Manager 初始化管理器
type Manager struct {
	cfg        *config.Config
	configPath string
	workDir    string
}

// NewManager 创建初始化管理器
func NewManager(cfg *config.Config, configPath, workDir string) *Manager {
	return &Manager{
		cfg:        cfg,
		configPath: configPath,
		workDir:    workDir,
	}
}

// CheckInitialized 检查是否已经初始化
func (m *Manager) CheckInitialized() bool {
	// 1. 检查配置文件是否存在
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return false
	}

	// 2. 读取配置
	cfg, err := config.Load(m.configPath)
	if err != nil {
		return false
	}

	// 3. 检查必要的 keystores
	hasTLCPRootCA := false
	hasTLSRootCA := false
	hasDefaultTLCP := false
	hasDefaultTLS := false
	for _, ks := range cfg.KeyStores {
		if ks.Name == "tlcpchan-tlcp-root-ca" {
			hasTLCPRootCA = true
		}
		if ks.Name == "tlcpchan-tls-root-ca" {
			hasTLSRootCA = true
		}
		if ks.Name == "default-tlcp" {
			hasDefaultTLCP = true
		}
		if ks.Name == "default-tls" {
			hasDefaultTLS = true
		}
	}
	if !hasTLCPRootCA || !hasTLSRootCA || !hasDefaultTLCP || !hasDefaultTLS {
		return false
	}

	// 4. 检查必要的 instances
	hasAutoProxy := false
	for _, inst := range cfg.Instances {
		if inst.Name == "auto-proxy" {
			hasAutoProxy = true
			break
		}
	}
	if !hasAutoProxy {
		return false
	}

	// 5. 检查关键文件是否存在
	keyFiles := []string{
		filepath.Join(m.workDir, "keystores", "tlcpchan-tlcp-root-ca.crt"),
		filepath.Join(m.workDir, "keystores", "tlcpchan-tlcp-root-ca.key"),
		filepath.Join(m.workDir, "keystores", "tlcpchan-tls-root-ca.crt"),
		filepath.Join(m.workDir, "keystores", "tlcpchan-tls-root-ca.key"),
		filepath.Join(m.workDir, "keystores", "default-tlcp-sign.crt"),
		filepath.Join(m.workDir, "keystores", "default-tlcp-sign.key"),
		filepath.Join(m.workDir, "keystores", "default-tlcp-enc.crt"),
		filepath.Join(m.workDir, "keystores", "default-tlcp-enc.key"),
		filepath.Join(m.workDir, "keystores", "default-tls.crt"),
		filepath.Join(m.workDir, "keystores", "default-tls.key"),
	}
	for _, f := range keyFiles {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			return false
		}
	}

	return true
}

// Initialize 执行完整初始化流程
func (m *Manager) Initialize() error {
	logger.Info("开始初始化...")

	// 1. 生成 TLCP 根 CA（SM2）
	logger.Info("生成 TLCP 根 CA 证书（SM2）...")
	tlcpRootCA, err := certgen.GenerateTLCPRootCA(certgen.CertGenConfig{
		Type:       certgen.CertTypeRootCA,
		CommonName: "tlcpchan-tlcp-root-ca",
		Org:        "tlcpchan",
		OrgUnit:    "tlcpchan",
		Years:      10,
		Days:       0,
	})
	if err != nil {
		return err
	}

	// 保存 TLCP 根证书到 keystores
	tlcpRootCACertPath := filepath.Join(m.workDir, "keystores", "tlcpchan-tlcp-root-ca.crt")
	tlcpRootCAKeyPath := filepath.Join(m.workDir, "keystores", "tlcpchan-tlcp-root-ca.key")
	if err := certgen.SaveCertToFile(tlcpRootCA.CertPEM, tlcpRootCA.KeyPEM, tlcpRootCACertPath, tlcpRootCAKeyPath); err != nil {
		return err
	}

	// 同时复制到 rootcerts 目录
	tlcpRootCertPath := filepath.Join(m.workDir, "rootcerts", "tlcpchan-tlcp-root-ca.crt")
	if err := os.MkdirAll(filepath.Dir(tlcpRootCertPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(tlcpRootCertPath, tlcpRootCA.CertPEM, 0644); err != nil {
		return err
	}
	logger.Info("TLCP 根 CA 证书生成完成")

	// 2. 生成 TLS 根 CA（RSA 2048）
	logger.Info("生成 TLS 根 CA 证书（RSA 2048）...")
	tlsRootCA, err := certgen.GenerateTLSRootCA(certgen.CertGenConfig{
		Type:       certgen.CertTypeRootCA,
		CommonName: "tlcpchan-tls-root-ca",
		Org:        "tlcpchan",
		OrgUnit:    "tlcpchan",
		Years:      10,
		Days:       0,
		KeyBits:    2048,
	})
	if err != nil {
		return err
	}

	// 保存 TLS 根证书到 keystores
	tlsRootCACertPath := filepath.Join(m.workDir, "keystores", "tlcpchan-tls-root-ca.crt")
	tlsRootCAKeyPath := filepath.Join(m.workDir, "keystores", "tlcpchan-tls-root-ca.key")
	if err := certgen.SaveCertToFile(tlsRootCA.CertPEM, tlsRootCA.KeyPEM, tlsRootCACertPath, tlsRootCAKeyPath); err != nil {
		return err
	}

	// 同时复制到 rootcerts 目录
	tlsRootCertPath := filepath.Join(m.workDir, "rootcerts", "tlcpchan-tls-root-ca.crt")
	if err := os.WriteFile(tlsRootCertPath, tlsRootCA.CertPEM, 0644); err != nil {
		return err
	}
	logger.Info("TLS 根 CA 证书生成完成")

	// 3. 加载 TLCP 根证书用于签发 TLCP 证书
	tlcpSignerCert, tlcpSignerKey, err := certgen.LoadTLCPCertFromFile(tlcpRootCACertPath, tlcpRootCAKeyPath)
	if err != nil {
		return err
	}

	// 4. 生成 TLCP 证书对（用 TLCP 根 CA 签发）
	logger.Info("生成 TLCP 证书对...")
	tlcpSignCfg := certgen.CertGenConfig{
		Type:       certgen.CertTypeTLCPSign,
		CommonName: "tlcpchan-default-tlcp-sign",
		Org:        "tlcpchan",
		OrgUnit:    "tlcpchan",
		Years:      5,
		Days:       0,
	}
	tlcpEncCfg := certgen.CertGenConfig{
		Type:       certgen.CertTypeTLCPEnc,
		CommonName: "tlcpchan-default-tlcp-enc",
		Org:        "tlcpchan",
		OrgUnit:    "tlcpchan",
		Years:      5,
		Days:       0,
	}
	tlcpSignCert, tlcpEncCert, err := certgen.GenerateTLCPPair(tlcpSignerCert, tlcpSignerKey, tlcpSignCfg, tlcpEncCfg)
	if err != nil {
		return err
	}

	// 保存 TLCP 证书
	tlcpSignCertPath := filepath.Join(m.workDir, "keystores", "default-tlcp-sign.crt")
	tlcpSignKeyPath := filepath.Join(m.workDir, "keystores", "default-tlcp-sign.key")
	tlcpEncCertPath := filepath.Join(m.workDir, "keystores", "default-tlcp-enc.crt")
	tlcpEncKeyPath := filepath.Join(m.workDir, "keystores", "default-tlcp-enc.key")

	if err := certgen.SaveCertToFile(tlcpSignCert.CertPEM, tlcpSignCert.KeyPEM, tlcpSignCertPath, tlcpSignKeyPath); err != nil {
		return err
	}
	if err := certgen.SaveCertToFile(tlcpEncCert.CertPEM, tlcpEncCert.KeyPEM, tlcpEncCertPath, tlcpEncKeyPath); err != nil {
		return err
	}
	logger.Info("TLCP 证书对生成完成")

	// 5. 加载 TLS 根证书用于签发 TLS 证书
	tlsSignerCert, tlsSignerKey, err := certgen.LoadTLSCertFromFile(tlsRootCACertPath, tlsRootCAKeyPath)
	if err != nil {
		return err
	}

	// 6. 生成 TLS 证书（用 TLS 根 CA 签发，RSA 2048）
	logger.Info("生成 TLS 证书（RSA 2048）...")
	tlsCfg := certgen.CertGenConfig{
		Type:         certgen.CertTypeTLS,
		CommonName:   "tlcpchan-default-tls",
		Org:          "tlcpchan",
		OrgUnit:      "tlcpchan",
		Years:        5,
		Days:         0,
		KeyAlgorithm: certgen.KeyAlgorithmRSA,
		KeyBits:      2048,
	}
	tlsCert, err := certgen.GenerateTLSCert(tlsSignerCert, tlsSignerKey, tlsCfg)
	if err != nil {
		return err
	}

	tlsCertPath := filepath.Join(m.workDir, "keystores", "default-tls.crt")
	tlsKeyPath := filepath.Join(m.workDir, "keystores", "default-tls.key")
	if err := certgen.SaveCertToFile(tlsCert.CertPEM, tlsCert.KeyPEM, tlsCertPath, tlsKeyPath); err != nil {
		return err
	}
	logger.Info("TLS 证书生成完成")

	// 7. 配置 keystores
	m.cfg.KeyStores = []config.KeyStoreConfig{
		{
			Name: "tlcpchan-tlcp-root-ca",
			Type: keystore.LoaderTypeFile,
			Params: map[string]string{
				"sign-cert": "./keystores/tlcpchan-tlcp-root-ca.crt",
				"sign-key":  "./keystores/tlcpchan-tlcp-root-ca.key",
			},
		},
		{
			Name: "tlcpchan-tls-root-ca",
			Type: keystore.LoaderTypeFile,
			Params: map[string]string{
				"sign-cert": "./keystores/tlcpchan-tls-root-ca.crt",
				"sign-key":  "./keystores/tlcpchan-tls-root-ca.key",
			},
		},
		{
			Name: "default-tlcp",
			Type: keystore.LoaderTypeFile,
			Params: map[string]string{
				"sign-cert": "./keystores/default-tlcp-sign.crt",
				"sign-key":  "./keystores/default-tlcp-sign.key",
				"enc-cert":  "./keystores/default-tlcp-enc.crt",
				"enc-key":   "./keystores/default-tlcp-enc.key",
			},
		},
		{
			Name: "default-tls",
			Type: keystore.LoaderTypeFile,
			Params: map[string]string{
				"sign-cert": "./keystores/default-tls.crt",
				"sign-key":  "./keystores/default-tls.key",
			},
		},
	}

	// 6. 配置 auto-proxy 实例
	m.cfg.Instances = []config.InstanceConfig{
		{
			Name:     "auto-proxy",
			Type:     "server",
			Listen:   ":20443",
			Target:   "127.0.0.1:20080",
			Protocol: "auto",
			Enabled:  true,
			TLCP: config.TLCPConfig{
				Auth: "one-way",
				Keystore: &config.KeyStoreConfig{
					Type: keystore.LoaderTypeFile,
					Params: map[string]string{
						"sign-cert": "./keystores/default-tlcp-sign.crt",
						"sign-key":  "./keystores/default-tlcp-sign.key",
						"enc-cert":  "./keystores/default-tlcp-enc.crt",
						"enc-key":   "./keystores/default-tlcp-enc.key",
					},
				},
			},
			TLS: config.TLSConfig{
				Auth: "one-way",
				Keystore: &config.KeyStoreConfig{
					Type: keystore.LoaderTypeFile,
					Params: map[string]string{
						"sign-cert": "./keystores/default-tls.crt",
						"sign-key":  "./keystores/default-tls.key",
					},
				},
			},
		},
	}

	// 7. 保存配置文件
	logger.Info("保存配置文件...")
	if err := config.Save(m.configPath, m.cfg); err != nil {
		return err
	}

	// 8. 创建初始化标志文件
	initializedFile := filepath.Join(m.workDir, ".tlcpchan-initialized")
	if err := os.WriteFile(initializedFile, []byte("initialized"), 0644); err != nil {
		return err
	}

	logger.Info("初始化完成！")
	return nil
}
