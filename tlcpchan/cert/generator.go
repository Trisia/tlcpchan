package cert

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
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

type Generator struct {
	tlcpDir string
	tlsDir  string
	caCerts map[string]*caCertInfo
}

type caCertInfo struct {
	cert     *x509.Certificate
	priv     interface{}
	certPath string
	keyPath  string
}

func NewGenerator(tlcpDir, tlsDir string) *Generator {
	return &Generator{
		tlcpDir: tlcpDir,
		tlsDir:  tlsDir,
		caCerts: make(map[string]*caCertInfo),
	}
}

func (g *Generator) GenerateSM2RootCA(name string, days int) (*CertInfo, error) {
	if err := os.MkdirAll(g.tlcpDir, 0755); err != nil {
		return nil, fmt.Errorf("创建TLCP证书目录失败: %w", err)
	}

	priv, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("生成SM2密钥对失败: %w", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("生成序列号失败: %w", err)
	}

	now := time.Now()
	template := &smx509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   name,
			Organization: []string{"TLCP Channel"},
		},
		NotBefore:             now,
		NotAfter:              now.AddDate(0, 0, days),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
	}

	certDER, err := smx509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		return nil, fmt.Errorf("创建SM2根CA证书失败: %w", err)
	}

	certPath := filepath.Join(g.tlcpDir, name+".crt")
	keyPath := filepath.Join(g.tlcpDir, name+".key")

	if err := g.writeCertAndKey(certPath, keyPath, certDER, priv); err != nil {
		return nil, err
	}

	cert, err := smx509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("解析生成的证书失败: %w", err)
	}

	g.caCerts[name] = &caCertInfo{
		cert:     cert.ToX509(),
		priv:     priv,
		certPath: certPath,
		keyPath:  keyPath,
	}

	return &CertInfo{
		Name:     name,
		Type:     CertTypeTLCP,
		NotAfter: now.AddDate(0, 0, days),
	}, nil
}

func (g *Generator) GenerateSM2ServerCert(name, caName string, hosts []string, days int) (*CertInfo, error) {
	ca, err := g.getOrCreateSM2CA(caName)
	if err != nil {
		return nil, err
	}

	priv, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("生成SM2密钥对失败: %w", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("生成序列号失败: %w", err)
	}

	now := time.Now()
	template := &smx509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   name,
			Organization: []string{"TLCP Channel"},
		},
		NotBefore:   now,
		NotAfter:    now.AddDate(0, 0, days),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    hosts,
	}

	caCert, err := smx509.ParseCertificate(ca.cert.Raw)
	if err != nil {
		return nil, fmt.Errorf("解析CA证书失败: %w", err)
	}

	certDER, err := smx509.CreateCertificate(rand.Reader, template, caCert, &priv.PublicKey, ca.priv)
	if err != nil {
		return nil, fmt.Errorf("创建SM2服务端证书失败: %w", err)
	}

	certPath := filepath.Join(g.tlcpDir, name+".crt")
	keyPath := filepath.Join(g.tlcpDir, name+".key")

	if err := g.writeCertAndKey(certPath, keyPath, certDER, priv); err != nil {
		return nil, err
	}

	return &CertInfo{
		Name:     name,
		Type:     CertTypeTLCP,
		NotAfter: now.AddDate(0, 0, days),
	}, nil
}

func (g *Generator) GenerateSM2ClientCert(name, caName string, days int) (*CertInfo, error) {
	ca, err := g.getOrCreateSM2CA(caName)
	if err != nil {
		return nil, err
	}

	priv, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("生成SM2密钥对失败: %w", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("生成序列号失败: %w", err)
	}

	now := time.Now()
	template := &smx509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   name,
			Organization: []string{"TLCP Channel"},
		},
		NotBefore:   now,
		NotAfter:    now.AddDate(0, 0, days),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	caCert, err := smx509.ParseCertificate(ca.cert.Raw)
	if err != nil {
		return nil, fmt.Errorf("解析CA证书失败: %w", err)
	}

	certDER, err := smx509.CreateCertificate(rand.Reader, template, caCert, &priv.PublicKey, ca.priv)
	if err != nil {
		return nil, fmt.Errorf("创建SM2客户端证书失败: %w", err)
	}

	certPath := filepath.Join(g.tlcpDir, name+".crt")
	keyPath := filepath.Join(g.tlcpDir, name+".key")

	if err := g.writeCertAndKey(certPath, keyPath, certDER, priv); err != nil {
		return nil, err
	}

	return &CertInfo{
		Name:     name,
		Type:     CertTypeTLCP,
		NotAfter: now.AddDate(0, 0, days),
	}, nil
}

func (g *Generator) getOrCreateSM2CA(name string) (*caCertInfo, error) {
	if ca, ok := g.caCerts[name]; ok {
		return ca, nil
	}

	certPath := filepath.Join(g.tlcpDir, name+".crt")
	keyPath := filepath.Join(g.tlcpDir, name+".key")

	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("读取CA证书失败: %w", err)
	}

	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("读取CA私钥失败: %w", err)
	}

	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, fmt.Errorf("解析CA证书PEM失败")
	}

	smCert, err := smx509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析CA证书失败: %w", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, fmt.Errorf("解析CA私钥PEM失败")
	}

	priv, err := smx509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		priv, err = smx509.ParseECPrivateKey(keyBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("解析CA私钥失败: %w", err)
		}
	}

	ca := &caCertInfo{
		cert:     smCert.ToX509(),
		priv:     priv,
		certPath: certPath,
		keyPath:  keyPath,
	}
	g.caCerts[name] = ca
	return ca, nil
}

func (g *Generator) GenerateRSARootCA(name string, days int) (*CertInfo, error) {
	if err := os.MkdirAll(g.tlsDir, 0755); err != nil {
		return nil, fmt.Errorf("创建TLS证书目录失败: %w", err)
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("生成RSA密钥对失败: %w", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("生成序列号失败: %w", err)
	}

	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   name,
			Organization: []string{"TLCP Channel"},
		},
		NotBefore:             now,
		NotAfter:              now.AddDate(0, 0, days),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		return nil, fmt.Errorf("创建RSA根CA证书失败: %w", err)
	}

	certPath := filepath.Join(g.tlsDir, name+".crt")
	keyPath := filepath.Join(g.tlsDir, name+".key")

	if err := g.writeCertAndKey(certPath, keyPath, certDER, priv); err != nil {
		return nil, err
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("解析生成的证书失败: %w", err)
	}

	g.caCerts[name] = &caCertInfo{
		cert:     cert,
		priv:     priv,
		certPath: certPath,
		keyPath:  keyPath,
	}

	return &CertInfo{
		Name:     name,
		Type:     CertTypeTLS,
		NotAfter: now.AddDate(0, 0, days),
	}, nil
}

func (g *Generator) GenerateRSAServerCert(name, caName string, hosts []string, days int) (*CertInfo, error) {
	ca, err := g.getOrCreateRSACA(caName)
	if err != nil {
		return nil, err
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("生成RSA密钥对失败: %w", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("生成序列号失败: %w", err)
	}

	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   name,
			Organization: []string{"TLCP Channel"},
		},
		NotBefore:   now,
		NotAfter:    now.AddDate(0, 0, days),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    hosts,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, ca.cert, &priv.PublicKey, ca.priv)
	if err != nil {
		return nil, fmt.Errorf("创建RSA服务端证书失败: %w", err)
	}

	certPath := filepath.Join(g.tlsDir, name+".crt")
	keyPath := filepath.Join(g.tlsDir, name+".key")

	if err := g.writeCertAndKey(certPath, keyPath, certDER, priv); err != nil {
		return nil, err
	}

	return &CertInfo{
		Name:     name,
		Type:     CertTypeTLS,
		NotAfter: now.AddDate(0, 0, days),
	}, nil
}

func (g *Generator) GenerateRSAClientCert(name, caName string, days int) (*CertInfo, error) {
	ca, err := g.getOrCreateRSACA(caName)
	if err != nil {
		return nil, err
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("生成RSA密钥对失败: %w", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("生成序列号失败: %w", err)
	}

	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   name,
			Organization: []string{"TLCP Channel"},
		},
		NotBefore:   now,
		NotAfter:    now.AddDate(0, 0, days),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, ca.cert, &priv.PublicKey, ca.priv)
	if err != nil {
		return nil, fmt.Errorf("创建RSA客户端证书失败: %w", err)
	}

	certPath := filepath.Join(g.tlsDir, name+".crt")
	keyPath := filepath.Join(g.tlsDir, name+".key")

	if err := g.writeCertAndKey(certPath, keyPath, certDER, priv); err != nil {
		return nil, err
	}

	return &CertInfo{
		Name:     name,
		Type:     CertTypeTLS,
		NotAfter: now.AddDate(0, 0, days),
	}, nil
}

func (g *Generator) getOrCreateRSACA(name string) (*caCertInfo, error) {
	if ca, ok := g.caCerts[name]; ok {
		return ca, nil
	}

	certPath := filepath.Join(g.tlsDir, name+".crt")
	keyPath := filepath.Join(g.tlsDir, name+".key")

	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("读取CA证书失败: %w", err)
	}

	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("读取CA私钥失败: %w", err)
	}

	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, fmt.Errorf("解析CA证书PEM失败")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析CA证书失败: %w", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, fmt.Errorf("解析CA私钥PEM失败")
	}

	var priv interface{}
	switch keyBlock.Type {
	case "RSA PRIVATE KEY":
		priv, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("解析RSA私钥失败: %w", err)
		}
	case "EC PRIVATE KEY":
		priv, err = x509.ParseECPrivateKey(keyBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("解析EC私钥失败: %w", err)
		}
	case "PRIVATE KEY":
		priv, err = x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("解析PKCS8私钥失败: %w", err)
		}
	default:
		return nil, fmt.Errorf("不支持的私钥类型: %s", keyBlock.Type)
	}

	ca := &caCertInfo{
		cert:     cert,
		priv:     priv,
		certPath: certPath,
		keyPath:  keyPath,
	}
	g.caCerts[name] = ca
	return ca, nil
}

func (g *Generator) writeCertAndKey(certPath, keyPath string, certDER []byte, priv interface{}) error {
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	var keyBytes []byte
	var err error
	switch p := priv.(type) {
	case *sm2.PrivateKey:
		keyBytes, err = smx509.MarshalPKCS8PrivateKey(p)
		if err != nil {
			return fmt.Errorf("序列化SM2私钥失败: %w", err)
		}
	case *rsa.PrivateKey:
		keyBytes = x509.MarshalPKCS1PrivateKey(p)
	case *ecdsa.PrivateKey:
		keyBytes, err = x509.MarshalECPrivateKey(p)
		if err != nil {
			return fmt.Errorf("序列化EC私钥失败: %w", err)
		}
	default:
		keyBytes, err = x509.MarshalPKCS8PrivateKey(p)
		if err != nil {
			return fmt.Errorf("序列化私钥失败: %w", err)
		}
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyBytes,
	})

	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		return fmt.Errorf("写入证书文件失败: %w", err)
	}

	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		os.Remove(certPath)
		return fmt.Errorf("写入私钥文件失败: %w", err)
	}

	return nil
}

func (g *Generator) InitDefaultCerts() error {
	tlcpCAPath := filepath.Join(g.tlcpDir, "ca-sm2.crt")
	if _, err := os.Stat(tlcpCAPath); os.IsNotExist(err) {
		if _, err := g.GenerateSM2RootCA("ca-sm2", 3650); err != nil {
			return fmt.Errorf("生成TLCP根CA失败: %w", err)
		}
		if _, err := g.GenerateSM2ServerCert("server-sm2", "ca-sm2", []string{"localhost", "127.0.0.1"}, 365); err != nil {
			return fmt.Errorf("生成TLCP服务端证书失败: %w", err)
		}
		if _, err := g.GenerateSM2ClientCert("client-sm2", "ca-sm2", 365); err != nil {
			return fmt.Errorf("生成TLCP客户端证书失败: %w", err)
		}
	}

	tlsCAPath := filepath.Join(g.tlsDir, "ca-rsa.crt")
	if _, err := os.Stat(tlsCAPath); os.IsNotExist(err) {
		if _, err := g.GenerateRSARootCA("ca-rsa", 3650); err != nil {
			return fmt.Errorf("生成TLS根CA失败: %w", err)
		}
		if _, err := g.GenerateRSAServerCert("server-rsa", "ca-rsa", []string{"localhost", "127.0.0.1"}, 365); err != nil {
			return fmt.Errorf("生成TLS服务端证书失败: %w", err)
		}
		if _, err := g.GenerateRSAClientCert("client-rsa", "ca-rsa", 365); err != nil {
			return fmt.Errorf("生成TLS客户端证书失败: %w", err)
		}
	}

	return nil
}
