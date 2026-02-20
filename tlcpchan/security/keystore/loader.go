package keystore

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"

	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/emmansun/gmsm/sm3"
	"github.com/emmansun/gmsm/smx509"
)

// Loader 加载器接口
type Loader interface {
	Load(loaderType LoaderType, params map[string]string) (KeyStore, error)
}

// FileKeyStore 基于文件的 keystore 实现
type FileKeyStore struct {
	keyStoreType KeyStoreType
	signCertPath string
	signKeyPath  string
	encCertPath  string
	encKeyPath   string
	tlsCert      *tls.Certificate
	tlcpCert     *tlcp.Certificate
	sm3Hash      []byte
	mu           sync.RWMutex
}

// fileKeyStoreJSON 用于序列化 FileKeyStore 配置的结构体
type fileKeyStoreJSON struct {
	KeyStoreType KeyStoreType `json:"keyStoreType"`
	SignCertPath string       `json:"signCertPath"`
	SignKeyPath  string       `json:"signKeyPath"`
	EncCertPath  string       `json:"encCertPath"`
	EncKeyPath   string       `json:"encKeyPath"`
}

func (f *FileKeyStore) Type() KeyStoreType {
	return f.keyStoreType
}

func (f *FileKeyStore) TLSCertificate() (*tls.Certificate, error) {
	f.mu.RLock()
	if f.tlsCert != nil {
		defer f.mu.RUnlock()
		return f.tlsCert, nil
	}
	f.mu.RUnlock()

	f.mu.Lock()
	defer f.mu.Unlock()

	if f.tlsCert != nil {
		return f.tlsCert, nil
	}

	cert, err := tls.LoadX509KeyPair(f.signCertPath, f.signKeyPath)
	if err != nil {
		return nil, fmt.Errorf("加载TLS证书失败: %w", err)
	}
	f.tlsCert = &cert
	f.computeSM3Hash()
	return f.tlsCert, nil
}

func (f *FileKeyStore) TLCPCertificate() (*tlcp.Certificate, error) {
	f.mu.RLock()
	if f.tlcpCert != nil {
		defer f.mu.RUnlock()
		return f.tlcpCert, nil
	}
	f.mu.RUnlock()

	f.mu.Lock()
	defer f.mu.Unlock()

	if f.tlcpCert != nil {
		return f.tlcpCert, nil
	}

	certPEM, err := os.ReadFile(f.signCertPath)
	if err != nil {
		return nil, fmt.Errorf("读取证书文件失败: %w", err)
	}

	keyPEM, err := os.ReadFile(f.signKeyPath)
	if err != nil {
		return nil, fmt.Errorf("读取私钥文件失败: %w", err)
	}

	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, fmt.Errorf("无法解析证书PEM块")
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, fmt.Errorf("无法解析私钥PEM块")
	}

	var certs []*x509.Certificate
	var privateKey crypto.PrivateKey

	if smCerts, err := smx509.ParseCertificates(certBlock.Bytes); err == nil && len(smCerts) > 0 {
		certs = make([]*x509.Certificate, len(smCerts))
		for i, smCert := range smCerts {
			certs[i] = smCert.ToX509()
		}
		switch keyBlock.Type {
		case "PRIVATE KEY", "EC PRIVATE KEY":
			privateKey, err = smx509.ParsePKCS8PrivateKey(keyBlock.Bytes)
			if err != nil {
				privateKey, err = smx509.ParseECPrivateKey(keyBlock.Bytes)
				if err != nil {
					return nil, fmt.Errorf("解析SM2私钥失败: %w", err)
				}
			}
		default:
			privateKey, err = smx509.ParsePKCS8PrivateKey(keyBlock.Bytes)
			if err != nil {
				return nil, fmt.Errorf("解析私钥失败: %w", err)
			}
		}
	} else {
		stdCerts, err := x509.ParseCertificates(certBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("解析证书失败: %w", err)
		}
		certs = stdCerts
		switch keyBlock.Type {
		case "RSA PRIVATE KEY":
			privateKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
			if err != nil {
				return nil, fmt.Errorf("解析RSA私钥失败: %w", err)
			}
		case "EC PRIVATE KEY":
			privateKey, err = x509.ParseECPrivateKey(keyBlock.Bytes)
			if err != nil {
				return nil, fmt.Errorf("解析EC私钥失败: %w", err)
			}
		case "PRIVATE KEY":
			privateKey, err = x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
			if err != nil {
				return nil, fmt.Errorf("解析PKCS8私钥失败: %w", err)
			}
		default:
			return nil, fmt.Errorf("不支持的私钥类型: %s", keyBlock.Type)
		}
	}

	raw := make([][]byte, len(certs))
	for i, c := range certs {
		raw[i] = c.Raw
	}

	tlcpCert := &tlcp.Certificate{
		Certificate: raw,
		PrivateKey:  privateKey,
	}

	f.tlcpCert = tlcpCert
	f.computeSM3Hash()
	return f.tlcpCert, nil
}

func (f *FileKeyStore) Reload() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.tlsCert = nil
	f.tlcpCert = nil
	f.sm3Hash = nil
	return nil
}

func (f *FileKeyStore) Equals(other KeyStore) bool {
	if other == nil {
		return false
	}
	otherFileKS, ok := other.(*FileKeyStore)
	if !ok {
		return false
	}
	f.ensureHashComputed()
	otherFileKS.ensureHashComputed()
	f.mu.RLock()
	otherFileKS.mu.RLock()
	defer f.mu.RUnlock()
	defer otherFileKS.mu.RUnlock()
	return bytes.Equal(f.sm3Hash, otherFileKS.sm3Hash)
}

// computeSM3Hash 计算 FileKeyStore 的 SM3 Hash 值
// Hash 构成：JSON 配置 + TLS 证书 DER（如果有） + TLCP 证书 DER（如果有）
// 注意：此函数必须在持有 f.mu 写锁的情况下调用
func (f *FileKeyStore) computeSM3Hash() {
	jsonData := fileKeyStoreJSON{
		KeyStoreType: f.keyStoreType,
		SignCertPath: f.signCertPath,
		SignKeyPath:  f.signKeyPath,
		EncCertPath:  f.encCertPath,
		EncKeyPath:   f.encKeyPath,
	}

	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		f.sm3Hash = nil
		return
	}

	hasher := sm3.New()
	hasher.Write(jsonBytes)

	if f.tlsCert != nil {
		for _, der := range f.tlsCert.Certificate {
			hasher.Write(der)
		}
	}

	if f.tlcpCert != nil {
		for _, der := range f.tlcpCert.Certificate {
			hasher.Write(der)
		}
	}

	f.sm3Hash = hasher.Sum(nil)
}

func (f *FileKeyStore) ensureHashComputed() {
	f.mu.RLock()
	hashEmpty := len(f.sm3Hash) == 0
	f.mu.RUnlock()
	if hashEmpty {
		f.mu.Lock()
		defer f.mu.Unlock()
		if len(f.sm3Hash) == 0 {
			if f.tlsCert == nil {
				_, _ = f.TLSCertificate()
			}
			if f.tlcpCert == nil {
				_, _ = f.TLCPCertificate()
			}
			if f.tlsCert != nil || f.tlcpCert != nil {
				f.computeSM3Hash()
			}
		}
	}
}

// FileLoader 文件加载器实现
type FileLoader struct {
	baseDir string
}

func NewFileLoader(baseDir string) *FileLoader {
	return &FileLoader{baseDir: baseDir}
}

func (fl *FileLoader) resolvePath(path string) string {
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) {
		return path
	}
	if fl.baseDir != "" {
		return filepath.Join(fl.baseDir, path)
	}
	return path
}

func (fl *FileLoader) Load(loaderType LoaderType, params map[string]string) (KeyStore, error) {
	signCertPath := fl.resolvePath(params["sign-cert"])
	signKeyPath := fl.resolvePath(params["sign-key"])
	encCertPath := fl.resolvePath(params["enc-cert"])
	encKeyPath := fl.resolvePath(params["enc-key"])

	if signCertPath == "" || signKeyPath == "" {
		return nil, fmt.Errorf("签名证书和密钥路径不能为空")
	}

	ksType := KeyStoreTypeTLS
	if encCertPath != "" && encKeyPath != "" {
		ksType = KeyStoreTypeTLCP
	}

	return &FileKeyStore{
		keyStoreType: ksType,
		signCertPath: signCertPath,
		signKeyPath:  signKeyPath,
		encCertPath:  encCertPath,
		encKeyPath:   encKeyPath,
	}, nil
}

// NamedLoader 命名加载器实现（通过名称引用已存在的 keystore）
type NamedLoader struct {
	manager *Manager
}

func NewNamedLoader(manager *Manager) *NamedLoader {
	return &NamedLoader{manager: manager}
}

func (nl *NamedLoader) Load(loaderType LoaderType, params map[string]string) (KeyStore, error) {
	name := params["name"]
	if name == "" {
		return nil, fmt.Errorf("keystore名称不能为空")
	}
	ks, err := nl.manager.GetKeyStore(name)
	if err != nil {
		return nil, err
	}
	return &NamedKeyStore{
		name:     name,
		manager:  nl.manager,
		delegate: ks,
	}, nil
}

// NamedKeyStore 命名加载器的keystore包装器
type NamedKeyStore struct {
	name     string
	manager  *Manager
	delegate KeyStore
}

func (n *NamedKeyStore) Type() KeyStoreType {
	return n.delegate.Type()
}

func (n *NamedKeyStore) TLCPCertificate() (*tlcp.Certificate, error) {
	return n.delegate.TLCPCertificate()
}

func (n *NamedKeyStore) TLSCertificate() (*tls.Certificate, error) {
	return n.delegate.TLSCertificate()
}

func (n *NamedKeyStore) Reload() error {
	return n.delegate.Reload()
}

func (n *NamedKeyStore) Equals(other KeyStore) bool {
	return n.delegate.Equals(other)
}

func (n *NamedKeyStore) GenerateCSR(keyType KeyType, params CSRParams) ([]byte, error) {
	return n.delegate.GenerateCSR(keyType, params)
}

// loadPrivateKey 从文件加载私钥
func (f *FileKeyStore) loadPrivateKey(keyPath string) (crypto.PrivateKey, error) {
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("读取私钥文件失败: %w", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, fmt.Errorf("无法解析私钥PEM块")
	}

	var privateKey crypto.PrivateKey
	if f.keyStoreType == KeyStoreTypeTLCP {
		switch keyBlock.Type {
		case "PRIVATE KEY", "EC PRIVATE KEY":
			privateKey, err = smx509.ParsePKCS8PrivateKey(keyBlock.Bytes)
			if err != nil {
				privateKey, err = smx509.ParseECPrivateKey(keyBlock.Bytes)
				if err != nil {
					return nil, fmt.Errorf("解析SM2私钥失败: %w", err)
				}
			}
		default:
			privateKey, err = smx509.ParsePKCS8PrivateKey(keyBlock.Bytes)
			if err != nil {
				return nil, fmt.Errorf("解析私钥失败: %w", err)
			}
		}
	} else {
		switch keyBlock.Type {
		case "RSA PRIVATE KEY":
			privateKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
			if err != nil {
				return nil, fmt.Errorf("解析RSA私钥失败: %w", err)
			}
		case "EC PRIVATE KEY":
			privateKey, err = x509.ParseECPrivateKey(keyBlock.Bytes)
			if err != nil {
				return nil, fmt.Errorf("解析EC私钥失败: %w", err)
			}
		case "PRIVATE KEY":
			privateKey, err = x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
			if err != nil {
				return nil, fmt.Errorf("解析PKCS8私钥失败: %w", err)
			}
		default:
			return nil, fmt.Errorf("不支持的私钥类型: %s", keyBlock.Type)
		}
	}

	return privateKey, nil
}

// GenerateCSR 生成证书请求
func (f *FileKeyStore) GenerateCSR(keyType KeyType, params CSRParams) ([]byte, error) {
	var keyPath string
	if f.keyStoreType == KeyStoreTypeTLCP {
		if keyType == KeyTypeSign {
			keyPath = f.signKeyPath
		} else {
			keyPath = f.encKeyPath
			if keyPath == "" {
				return nil, fmt.Errorf("未配置加密密钥")
			}
		}
	} else {
		keyPath = f.signKeyPath
	}

	privateKey, err := f.loadPrivateKey(keyPath)
	if err != nil {
		return nil, err
	}

	subject := pkix.Name{
		CommonName:         params.CommonName,
		Country:            []string{params.Country},
		Province:           []string{params.StateOrProvince},
		Locality:           []string{params.Locality},
		Organization:       []string{params.Org},
		OrganizationalUnit: []string{params.OrgUnit},
	}

	template := &x509.CertificateRequest{
		Subject:            subject,
		DNSNames:           params.DNSNames,
		IPAddresses:        make([]net.IP, 0, len(params.IPAddresses)),
		SignatureAlgorithm: x509.UnknownSignatureAlgorithm,
	}

	for _, ipStr := range params.IPAddresses {
		if ip := net.ParseIP(ipStr); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		}
	}

	if f.keyStoreType == KeyStoreTypeTLCP {
		derBytes, err := smx509.CreateCertificateRequest(rand.Reader, template, privateKey)
		if err != nil {
			return nil, fmt.Errorf("生成SM2证书请求失败: %w", err)
		}
		return derBytes, nil
	}

	derBytes, err := x509.CreateCertificateRequest(rand.Reader, template, privateKey)
	if err != nil {
		return nil, fmt.Errorf("生成证书请求失败: %w", err)
	}
	return derBytes, nil
}
