package keystore

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gitee.com/Trisia/gotlcp/tlcp"
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
	tlcpCerts    []*tlcp.Certificate
	mu           sync.RWMutex
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
	return f.tlsCert, nil
}

func (f *FileKeyStore) TLCPCertificate() ([]*tlcp.Certificate, error) {
	f.mu.RLock()
	if len(f.tlcpCerts) > 0 {
		defer f.mu.RUnlock()
		return f.tlcpCerts, nil
	}
	f.mu.RUnlock()

	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.tlcpCerts) > 0 {
		return f.tlcpCerts, nil
	}

	certs := make([]*tlcp.Certificate, 0, 2)

	signCert, err := f.loadTLCPKeyPair(f.signCertPath, f.signKeyPath)
	if err != nil {
		return nil, fmt.Errorf("加载签名证书失败: %w", err)
	}
	certs = append(certs, signCert)

	if f.encCertPath != "" && f.encKeyPath != "" {
		encCert, err := f.loadTLCPKeyPair(f.encCertPath, f.encKeyPath)
		if err != nil {
			return nil, fmt.Errorf("加载加密证书失败: %w", err)
		}
		certs = append(certs, encCert)
	}

	f.tlcpCerts = certs
	return f.tlcpCerts, nil
}

func (f *FileKeyStore) loadTLCPKeyPair(certPath, keyPath string) (*tlcp.Certificate, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("读取证书文件失败: %w", err)
	}

	keyPEM, err := os.ReadFile(keyPath)
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

	return tlcpCert, nil
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

func (n *NamedKeyStore) TLCPCertificate() ([]*tlcp.Certificate, error) {
	return n.delegate.TLCPCertificate()
}

func (n *NamedKeyStore) TLSCertificate() (*tls.Certificate, error) {
	return n.delegate.TLSCertificate()
}
