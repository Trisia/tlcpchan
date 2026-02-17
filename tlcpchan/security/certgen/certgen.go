package certgen

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/emmansun/gmsm/sm2"
	"github.com/emmansun/gmsm/smx509"
)

// CertType 证书类型
type CertType string

const (
	// CertTypeRootCA 根 CA 证书
	CertTypeRootCA CertType = "root-ca"
	// CertTypeTLCPSign TLCP 签名证书
	CertTypeTLCPSign CertType = "tlcp-sign"
	// CertTypeTLCPEnc TLCP 加密证书
	CertTypeTLCPEnc CertType = "tlcp-enc"
	// CertTypeTLS TLS 证书
	CertTypeTLS CertType = "tls"
)

// KeyAlgorithm 密钥算法类型
type KeyAlgorithm string

const (
	// KeyAlgorithmSM2 SM2 国密算法
	KeyAlgorithmSM2 KeyAlgorithm = "sm2"
	// KeyAlgorithmRSA RSA 算法
	KeyAlgorithmRSA KeyAlgorithm = "rsa"
	// KeyAlgorithmECDSA ECDSA 算法
	KeyAlgorithmECDSA KeyAlgorithm = "ecdsa"
)

// CertGenConfig 证书生成配置
type CertGenConfig struct {
	// Type 证书类型
	Type CertType
	// CommonName Common Name (CN)
	CommonName string
	// Org Organization (O)
	Org string
	// OrgUnit Organizational Unit (OU)
	OrgUnit string
	// Years 有效期（年）
	Years int
	// KeyAlgorithm 密钥算法
	KeyAlgorithm KeyAlgorithm
}

// GeneratedCert 生成的证书结果
type GeneratedCert struct {
	// CertPEM 证书 PEM 数据
	CertPEM []byte
	// KeyPEM 私钥 PEM 数据
	KeyPEM []byte
}

// GenerateRootCA 生成自签名根 CA 证书（SM2）
//
// 功能：
//
//	生成一个自签名的 SM2 根 CA 证书，用于签发其他证书
//
// 参数：
//
//	cfg - 证书生成配置，包含 CommonName、Org、OrgUnit、Years 等信息
//	      - 如果 CommonName 为空，默认使用 "tlcpchan-root-ca"
//	      - 如果 Org 为空，默认使用 "tlcpchan"
//	      - 如果 Years <= 0，默认使用 10 年
//
// 返回值：
//
//	*GeneratedCert - 包含证书 PEM 和私钥 PEM 的结果
//	error - 错误信息，包括密钥生成失败、证书创建失败、私钥序列化失败等
//
// 注意事项：
//   - 生成的根 CA 证书使用 SM2 算法
//   - 证书具有 KeyUsageCertSign 和 KeyUsageCRLSign 权限
//   - IsCA 设置为 true，表示这是一个 CA 证书
func GenerateRootCA(cfg CertGenConfig) (*GeneratedCert, error) {
	if cfg.CommonName == "" {
		cfg.CommonName = "tlcpchan-root-ca"
	}
	if cfg.Org == "" {
		cfg.Org = "tlcpchan"
	}
	if cfg.Years <= 0 {
		cfg.Years = 10
	}

	priv, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("生成SM2密钥失败: %w", err)
	}

	now := time.Now()
	template := &smx509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:         cfg.CommonName,
			Organization:       []string{cfg.Org},
			OrganizationalUnit: []string{cfg.OrgUnit},
		},
		NotBefore:             now,
		NotAfter:              now.AddDate(cfg.Years, 0, 0),
		KeyUsage:              smx509.KeyUsageCertSign | smx509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certBytes, err := smx509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		return nil, fmt.Errorf("创建证书失败: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	privBytes, err := smx509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("序列化私钥失败: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	})

	return &GeneratedCert{
		CertPEM: certPEM,
		KeyPEM:  keyPEM,
	}, nil
}

// GenerateTLCPPair 生成 TLCP 证书对（签名证书+加密证书，由根证书签发）
//
// 功能：
//
//	生成一对 TLCP 证书，包括签名证书和加密证书，均由指定的根证书签发
//
// 参数：
//
//	signerCert - 签发者（根 CA）证书
//	signerKey - 签发者（根 CA）私钥
//	signCfg - 签名证书配置
//	        - 如果 CommonName 为空，默认使用 "tlcp-sign"
//	        - 如果 Org 为空，默认使用 "tlcpchan"
//	        - 如果 Years <= 0，默认使用 5 年
//	encCfg - 加密证书配置
//	        - 如果 CommonName 为空，默认使用 "tlcp-enc"
//	        - 如果 Org 为空，默认使用 "tlcpchan"
//	        - 如果 Years <= 0，默认使用 5 年
//
// 返回值：
//
//	signCert - 签名证书结果
//	encCert - 加密证书结果
//	error - 错误信息，包括签名证书生成失败或加密证书生成失败
//
// 注意事项：
//   - 两个证书都使用 SM2 算法
//   - 签名证书具有 KeyUsageDigitalSignature 权限
//   - 加密证书具有 KeyUsageKeyEncipherment | KeyUsageDataEncipherment 权限
func GenerateTLCPPair(signerCert *x509.Certificate, signerKey crypto.PrivateKey, signCfg, encCfg CertGenConfig) (signCert, encCert *GeneratedCert, err error) {
	if signCfg.CommonName == "" {
		signCfg.CommonName = "tlcp-sign"
	}
	if signCfg.Org == "" {
		signCfg.Org = "tlcpchan"
	}
	if signCfg.Years <= 0 {
		signCfg.Years = 5
	}

	if encCfg.CommonName == "" {
		encCfg.CommonName = "tlcp-enc"
	}
	if encCfg.Org == "" {
		encCfg.Org = "tlcpchan"
	}
	if encCfg.Years <= 0 {
		encCfg.Years = 5
	}

	signCert, err = generateTLCPCert(signerCert, signerKey, signCfg, smx509.KeyUsageDigitalSignature)
	if err != nil {
		return nil, nil, fmt.Errorf("生成签名证书失败: %w", err)
	}

	encCert, err = generateTLCPCert(signerCert, signerKey, encCfg, smx509.KeyUsageKeyEncipherment|smx509.KeyUsageDataEncipherment)
	if err != nil {
		return nil, nil, fmt.Errorf("生成加密证书失败: %w", err)
	}

	return signCert, encCert, nil
}

// generateTLCPCert 生成单个 TLCP 证书
//
// 功能：
//
//	生成单个由根证书签发的 TLCP 证书
//
// 参数：
//
//	signerCert - 签发者（根 CA）证书（标准库 x509.Certificate 类型）
//	signerKey - 签发者（根 CA）私钥
//	cfg - 证书配置
//	keyUsage - 密钥用途
//
// 返回值：
//
//	*GeneratedCert - 包含证书 PEM 和私钥 PEM 的结果
//	error - 错误信息，包括密钥生成失败、签发者证书解析失败、证书创建失败、私钥序列化失败等
//
// 注意事项：
//   - 这是一个内部辅助函数
//   - 使用 SM2 算法生成密钥对
//   - IsCA 设置为 false，表示这不是一个 CA 证书
func generateTLCPCert(signerCert *x509.Certificate, signerKey crypto.PrivateKey, cfg CertGenConfig, keyUsage smx509.KeyUsage) (*GeneratedCert, error) {
	priv, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("生成SM2密钥失败: %w", err)
	}

	smSignerCert, err := smx509.ParseCertificate(signerCert.Raw)
	if err != nil {
		return nil, fmt.Errorf("解析签发者证书失败: %w", err)
	}

	now := time.Now()
	template := &smx509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName:         cfg.CommonName,
			Organization:       []string{cfg.Org},
			OrganizationalUnit: []string{cfg.OrgUnit},
		},
		NotBefore:             now,
		NotAfter:              now.AddDate(cfg.Years, 0, 0),
		KeyUsage:              keyUsage,
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	certBytes, err := smx509.CreateCertificate(rand.Reader, template, smSignerCert, &priv.PublicKey, signerKey)
	if err != nil {
		return nil, fmt.Errorf("创建证书失败: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	privBytes, err := smx509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("序列化私钥失败: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	})

	return &GeneratedCert{
		CertPEM: certPEM,
		KeyPEM:  keyPEM,
	}, nil
}

// GenerateTLSCert 生成 TLS 证书（由根证书签发）
//
// 功能：
//
//	生成一个由根证书签发的 TLS 证书
//
// 参数：
//
//	signerCert - 签发者（根 CA）证书
//	signerKey - 签发者（根 CA）私钥
//	cfg - 证书配置
//	    - 如果 CommonName 为空，默认使用 "tls-cert"
//	    - 如果 Org 为空，默认使用 "tlcpchan"
//	    - 如果 Years <= 0，默认使用 5 年
//
// 返回值：
//
//	*GeneratedCert - 包含证书 PEM 和私钥 PEM 的结果
//	error - 错误信息，包括密钥生成失败、证书创建失败、私钥序列化失败等
//
// 注意事项：
//   - 使用 ECDSA P-256 算法生成密钥对
//   - 证书具有 KeyUsageDigitalSignature | KeyUsageKeyEncipherment 权限
//   - 支持 ExtKeyUsageServerAuth 和 ExtKeyUsageClientAuth 扩展密钥用途
//   - IsCA 设置为 false，表示这不是一个 CA 证书
func GenerateTLSCert(signerCert *x509.Certificate, signerKey crypto.PrivateKey, cfg CertGenConfig) (*GeneratedCert, error) {
	if cfg.CommonName == "" {
		cfg.CommonName = "tls-cert"
	}
	if cfg.Org == "" {
		cfg.Org = "tlcpchan"
	}
	if cfg.Years <= 0 {
		cfg.Years = 5
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("生成ECDSA密钥失败: %w", err)
	}

	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName:         cfg.CommonName,
			Organization:       []string{cfg.Org},
			OrganizationalUnit: []string{cfg.OrgUnit},
		},
		NotBefore:             now,
		NotAfter:              now.AddDate(cfg.Years, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, signerCert, &priv.PublicKey, signerKey)
	if err != nil {
		return nil, fmt.Errorf("创建证书失败: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("序列化私钥失败: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	})

	return &GeneratedCert{
		CertPEM: certPEM,
		KeyPEM:  keyPEM,
	}, nil
}

// SaveCertToFile 保存证书和密钥到文件
//
// 功能：
//
//	将证书 PEM 数据和私钥 PEM 数据分别保存到指定的文件路径
//
// 参数：
//
//	certPEM - 证书 PEM 数据
//	keyPEM - 私钥 PEM 数据
//	certPath - 证书保存路径
//	keyPath - 私钥保存路径
//
// 返回值：
//
//	error - 错误信息，包括目录创建失败、文件写入失败等
//
// 注意事项：
//   - 会自动创建证书和密钥文件所在的目录（权限 0755）
//   - 证书文件权限设置为 0644
//   - 私钥文件权限设置为 0600（仅所有者可读写）
func SaveCertToFile(certPEM, keyPEM []byte, certPath, keyPath string) error {
	if err := os.MkdirAll(filepath.Dir(certPath), 0755); err != nil {
		return fmt.Errorf("创建证书目录失败: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(keyPath), 0755); err != nil {
		return fmt.Errorf("创建密钥目录失败: %w", err)
	}

	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		return fmt.Errorf("写入证书文件失败: %w", err)
	}

	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		return fmt.Errorf("写入密钥文件失败: %w", err)
	}

	return nil
}

// LoadCertFromFile 从文件加载证书和密钥
//
// 功能：
//
//	从指定的文件路径加载证书和私钥
//
// 参数：
//
//	certPath - 证书文件路径
//	keyPath - 私钥文件路径
//
// 返回值：
//
//	*x509.Certificate - 解析后的证书对象
//	crypto.PrivateKey - 解析后的私钥对象（可能是 *ecdsa.PrivateKey 或其他类型）
//	error - 错误信息，包括文件读取失败、PEM 解析失败、证书/私钥解析失败等
//
// 注意事项：
//   - 支持 "EC PRIVATE KEY" 和 "PRIVATE KEY" (PKCS8) 格式的私钥
//   - 不支持其他类型的私钥格式
func LoadCertFromFile(certPath, keyPath string) (*x509.Certificate, crypto.PrivateKey, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, nil, fmt.Errorf("读取证书文件失败: %w", err)
	}

	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("读取密钥文件失败: %w", err)
	}

	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, nil, fmt.Errorf("无法解析证书PEM")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("解析证书失败: %w", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, nil, fmt.Errorf("无法解析密钥PEM")
	}

	var priv crypto.PrivateKey
	switch keyBlock.Type {
	case "EC PRIVATE KEY":
		priv, err = x509.ParseECPrivateKey(keyBlock.Bytes)
	case "PRIVATE KEY":
		priv, err = x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	default:
		return nil, nil, fmt.Errorf("不支持的密钥类型: %s", keyBlock.Type)
	}

	if err != nil {
		return nil, nil, fmt.Errorf("解析私钥失败: %w", err)
	}

	return cert, priv, nil
}
