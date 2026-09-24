package rootcert

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/emmansun/gmsm/sm2"
	"github.com/emmansun/gmsm/smx509"
)

var certExtensions = []string{".pem", ".cer", ".crt", ".der"}

// Manager 根证书管理器
type Manager struct {
	baseDir    string
	certs      map[string]*RootCert
	certPool   *x509.CertPool
	smCertPool *smx509.CertPool
	mu         sync.RWMutex
}

// NewManager 创建根证书管理器
func NewManager(baseDir string) *Manager {
	return &Manager{
		baseDir:    baseDir,
		certs:      make(map[string]*RootCert),
		certPool:   x509.NewCertPool(),
		smCertPool: smx509.NewCertPool(),
	}
}

// Initialize 初始化管理器，加载指定目录中的所有根证书
func (m *Manager) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.loadAllCerts()
}

// Add 添加根证书（保存到目录并重新加载）
func (m *Manager) Add(filename string, certData []byte) (*RootCert, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := os.MkdirAll(m.baseDir, 0700); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	certPath := filepath.Join(m.baseDir, filename)
	if err := os.WriteFile(certPath, certData, 0600); err != nil {
		return nil, fmt.Errorf("写入证书失败: %w", err)
	}

	if err := m.loadAllCerts(); err != nil {
		return nil, err
	}

	return m.certs[filename], nil
}

// Delete 删除根证书
func (m *Manager) Delete(filename string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	certPath := filepath.Join(m.baseDir, filename)
	if err := os.Remove(certPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除证书失败: %w", err)
	}

	return m.loadAllCerts()
}

// Get 获取根证书
func (m *Manager) Get(filename string) (*RootCert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cert, exists := m.certs[filename]
	if !exists {
		return nil, fmt.Errorf("根证书 %s 不存在", filename)
	}
	return cert, nil
}

// List 列出所有根证书
func (m *Manager) List() []*RootCert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*RootCert, 0, len(m.certs))
	for _, cert := range m.certs {
		result = append(result, cert)
	}
	return result
}

// GetPool 获取根证书池（包含所有根证书）
func (m *Manager) GetPool() RootCertPool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return &rootCertPool{
		certPool:   m.certPool,
		smCertPool: m.smCertPool,
		certs:      m.certs,
	}
}

// Reload 重新加载所有根证书
func (m *Manager) Reload() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.loadAllCerts()
}

// ReadFile 读取证书文件内容
// 参数：
//   - filename: 证书文件名
//
// 返回：
//   - []byte: 证书文件内容
//   - error: 错误信息
func (m *Manager) ReadFile(filename string) ([]byte, error) {
	certPath := filepath.Join(m.baseDir, filename)
	return os.ReadFile(certPath)
}

func (m *Manager) loadAllCerts() error {
	m.certs = make(map[string]*RootCert)
	m.certPool = x509.NewCertPool()
	m.smCertPool = smx509.NewCertPool()

	if m.baseDir == "" {
		return nil
	}

	entries, err := os.ReadDir(m.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("读取目录失败: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		ext := strings.ToLower(filepath.Ext(filename))
		if !isCertExtension(ext) {
			continue
		}

		certPath := filepath.Join(m.baseDir, filename)
		data, err := os.ReadFile(certPath)
		if err != nil {
			continue
		}

		rootCert, err := m.parseCert(data, filename)
		if err != nil {
			continue
		}

		m.certs[filename] = rootCert
		m.addToPools(data)
	}

	return nil
}

// parseCert 解析证书数据并构造根证书信息
//
// 参数：
//   - data: 证书数据，支持 PEM、DER、Base64、Hex 四种编码
//   - filename: 证书文件名，用于填充 RootCert.Filename
//
// 返回：
//   - *RootCert: 解析成功后的根证书信息
//   - error: 所有编码尝试均失败时返回错误
//
// 注意事项：
//   - 统一使用 smx509 解析：它是标准库 crypto/x509 的超集（额外支持 SM2/SM4/PQC），
//     可同时解析国密证书与 RSA/ECDSA 等标准证书
//   - gmsm v0.44.0 起 smx509.Certificate 与 x509.Certificate 是相互独立的结构体，
//     不能再通过 ToX509 之类的类型转换互转
func (m *Manager) parseCert(data []byte, filename string) (*RootCert, error) {
	var cert *smx509.Certificate

	tryParse := func(raw []byte) *smx509.Certificate {
		if smCert, err := smx509.ParseCertificate(raw); err == nil {
			return smCert
		}
		return nil
	}

	block, _ := pem.Decode(data)
	if block != nil {
		cert = tryParse(block.Bytes)
	}

	if cert == nil {
		cert = tryParse(data)
	}

	if cert == nil {
		if decoded, err := decodeBase64(data); err == nil {
			cert = tryParse(decoded)
		}
	}

	if cert == nil {
		if decoded, err := decodeHex(data); err == nil {
			cert = tryParse(decoded)
		}
	}

	if cert == nil {
		return nil, fmt.Errorf("解析证书失败")
	}

	return &RootCert{
		Filename:     filename,
		Cert:         cert,
		NotBefore:    cert.NotBefore,
		NotAfter:     cert.NotAfter,
		Subject:      cert.Subject.String(),
		Issuer:       cert.Issuer.String(),
		KeyType:      getKeyType(cert),
		SerialNumber: hex.EncodeToString(cert.SerialNumber.Bytes()),
		Version:      cert.Version,
		IsCA:         cert.IsCA,
		KeyUsage:     getKeyUsage(cert),
	}, nil
}

// getKeyType 获取证书公钥的算法类型
//
// 参数：
//   - cert: smx509 解析得到的证书对象
//
// 返回：
//   - string: 密钥类型描述，取值 "SM2"、"RSA-<位长>"、"ECDSA-P<位长>" 或 "Unknown"
//
// 注意事项：
//   - SM2 公钥在 smx509 中同样以 *ecdsa.PublicKey 表示（曲线为 SM2 P-256），
//     因此需要优先用 sm2.IsSM2PublicKey 区分，避免被识别为普通 ECDSA
func getKeyType(cert *smx509.Certificate) string {
	pubKey := cert.PublicKey

	if sm2.IsSM2PublicKey(pubKey) {
		return "SM2"
	}

	switch k := pubKey.(type) {
	case *rsa.PublicKey:
		return fmt.Sprintf("RSA-%d", k.N.BitLen())
	case *ecdsa.PublicKey:
		curveBits := k.Curve.Params().BitSize
		return fmt.Sprintf("ECDSA-P%d", curveBits)
	default:
		return "Unknown"
	}
}

// getKeyUsage 获取证书的密钥用途列表
//
// 参数：
//   - cert: smx509 解析得到的证书对象
//
// 返回：
//   - []string: 密钥用途的可读名称列表，未设置任何用途时返回 nil
func getKeyUsage(cert *smx509.Certificate) []string {
	var usages []string

	usageMap := map[smx509.KeyUsage]string{
		smx509.KeyUsageDigitalSignature:  "Digital Signature",
		smx509.KeyUsageContentCommitment: "Content Commitment",
		smx509.KeyUsageKeyEncipherment:   "Key Encipherment",
		smx509.KeyUsageDataEncipherment:  "Data Encipherment",
		smx509.KeyUsageKeyAgreement:      "Key Agreement",
		smx509.KeyUsageCertSign:          "Cert Sign",
		smx509.KeyUsageCRLSign:           "CRL Sign",
		smx509.KeyUsageEncipherOnly:      "Encipher Only",
		smx509.KeyUsageDecipherOnly:      "Decipher Only",
	}

	for k, v := range usageMap {
		if cert.KeyUsage&k != 0 {
			usages = append(usages, v)
		}
	}

	return usages
}

func decodeBase64(data []byte) ([]byte, error) {
	trimmed := []byte(strings.TrimSpace(string(data)))
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(trimmed)))
	n, err := base64.StdEncoding.Decode(decoded, trimmed)
	if err == nil {
		return decoded[:n], nil
	}

	decoded = make([]byte, base64.URLEncoding.DecodedLen(len(trimmed)))
	n, err = base64.URLEncoding.Decode(decoded, trimmed)
	if err == nil {
		return decoded[:n], nil
	}

	return nil, err
}

func decodeHex(data []byte) ([]byte, error) {
	trimmed := []byte(strings.TrimSpace(string(data)))
	decoded := make([]byte, hex.DecodedLen(len(trimmed)))
	n, err := hex.Decode(decoded, trimmed)
	if err == nil {
		return decoded[:n], nil
	}
	return nil, err
}

func (m *Manager) addToPools(data []byte) {
	m.certPool.AppendCertsFromPEM(data)
	m.smCertPool.AppendCertsFromPEM(data)
}

func isCertExtension(ext string) bool {
	for _, e := range certExtensions {
		if e == ext {
			return true
		}
	}
	return false
}

type rootCertPool struct {
	certPool   *x509.CertPool
	smCertPool *smx509.CertPool
	certs      map[string]*RootCert
}

func (p *rootCertPool) GetCertPool() *x509.CertPool {
	return p.certPool
}

func (p *rootCertPool) GetSMCertPool() *smx509.CertPool {
	return p.smCertPool
}

func (p *rootCertPool) GetCerts() []*RootCert {
	result := make([]*RootCert, 0, len(p.certs))
	for _, cert := range p.certs {
		result = append(result, cert)
	}
	return result
}
