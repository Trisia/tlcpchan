package cert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewLoader(t *testing.T) {
	loader := NewLoader()
	if loader == nil {
		t.Fatal("NewLoader() 返回 nil")
	}
}

func TestLoaderStoreAndGet(t *testing.T) {
	loader := NewLoader()

	certPEM, keyPEM := generateTestCertificate(t, false)
	cert, err := ParseTLSCertificate(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("ParseTLSCertificate() 失败: %v", err)
	}

	loader.Store("test-cert", cert)

	retrieved := loader.Get("test-cert")
	if retrieved == nil {
		t.Fatal("Get() 返回 nil")
	}

	if retrieved.Subject() != cert.Subject() {
		t.Errorf("检索的证书主题不匹配")
	}

	loader.Delete("test-cert")
	if loader.Get("test-cert") != nil {
		t.Error("Delete() 后证书仍存在")
	}
}

func TestParseTLSCertificateRSA(t *testing.T) {
	certPEM, keyPEM := generateTestCertificate(t, false)

	cert, err := ParseTLSCertificate(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("ParseTLSCertificate() 失败: %v", err)
	}

	if cert == nil {
		t.Fatal("ParseTLSCertificate() 返回 nil")
	}

	if cert.Type() != CertTypeTLS {
		t.Errorf("证书类型应为 CertTypeTLS, 实际为 %v", cert.Type())
	}

	if cert.Leaf() == nil {
		t.Error("Leaf() 不应返回 nil")
	}

	if cert.Subject() == "" {
		t.Error("Subject() 不应返回空字符串")
	}
}

func TestParseTLSCertificateECDSA(t *testing.T) {
	certPEM, keyPEM := generateTestECDSACertificate(t)

	cert, err := ParseTLSCertificate(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("ParseTLSCertificate() 失败: %v", err)
	}

	if cert == nil {
		t.Fatal("ParseTLSCertificate() 返回 nil")
	}

	_, ok := cert.PrivateKey.(*ecdsa.PrivateKey)
	if !ok {
		t.Error("私钥类型应为 ECDSA")
	}
}

func TestParseTLCPCertificateInvalidPEM(t *testing.T) {
	tests := []struct {
		name    string
		certPEM []byte
		keyPEM  []byte
		wantErr bool
	}{
		{
			name:    "无效证书PEM",
			certPEM: []byte("not a valid pem"),
			keyPEM:  []byte("not a valid pem"),
			wantErr: true,
		},
		{
			name:    "空证书PEM",
			certPEM: []byte(""),
			keyPEM:  []byte(""),
			wantErr: true,
		},
		{
			name:    "无效证书数据",
			certPEM: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("invalid")}),
			keyPEM:  pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("invalid")}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseTLSCertificate(tt.certPEM, tt.keyPEM)
			if !tt.wantErr {
				if err != nil {
					t.Errorf("不期望错误但返回: %v", err)
				}
			} else {
				if err == nil {
					t.Error("期望错误但未返回")
				}
			}
		})
	}
}

func TestValidateCertKeyPair(t *testing.T) {
	certPEM1, keyPEM1 := generateTestCertificate(t, false)
	certPEM2, keyPEM2 := generateTestCertificate(t, false)

	cert1, _ := ParseTLSCertificate(certPEM1, keyPEM1)
	cert2, _ := ParseTLSCertificate(certPEM2, keyPEM2)

	err := ValidateCertKeyPair(cert1.Certificate, cert1.PrivateKey)
	if err != nil {
		t.Errorf("匹配的证书私钥对验证失败: %v", err)
	}

	err = ValidateCertKeyPair(cert1.Certificate, cert2.PrivateKey)
	if err == nil {
		t.Error("不匹配的证书私钥对应返回错误")
	}
}

func TestValidateCertKeyPairEmpty(t *testing.T) {
	err := ValidateCertKeyPair(nil, nil)
	if err == nil {
		t.Error("空证书链应返回错误")
	}
}

func TestValidateCertKeyPairTypeMismatch(t *testing.T) {
	rsaCertPEM, rsaKeyPEM := generateTestCertificate(t, false)
	rsaCert, _ := ParseTLSCertificate(rsaCertPEM, rsaKeyPEM)

	ecdsaCertPEM, ecdsaKeyPEM := generateTestECDSACertificate(t)
	ecdsaCert, _ := ParseTLSCertificate(ecdsaCertPEM, ecdsaKeyPEM)

	err := ValidateCertKeyPair(rsaCert.Certificate, ecdsaCert.PrivateKey)
	if err == nil {
		t.Error("RSA证书与ECDSA私钥应返回错误")
	}

	err = ValidateCertKeyPair(ecdsaCert.Certificate, rsaCert.PrivateKey)
	if err == nil {
		t.Error("ECDSA证书与RSA私钥应返回错误")
	}
}

func TestCertificateMethods(t *testing.T) {
	certPEM, keyPEM := generateTestCertificate(t, false)
	cert, err := ParseTLSCertificate(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("ParseTLSCertificate() 失败: %v", err)
	}

	tlsCert := cert.TLSCertificate()
	if len(tlsCert.Certificate) == 0 {
		t.Error("TLSCertificate() 应返回非空证书链")
	}

	if cert.IsExpired() {
		t.Error("新创建的证书不应已过期")
	}

	if cert.ExpiresAt().IsZero() {
		t.Error("ExpiresAt() 不应返回零值")
	}

	if cert.Issuer() == "" {
		t.Error("Issuer() 不应返回空字符串")
	}
}

func TestCertificateReloadNoPath(t *testing.T) {
	certPEM, keyPEM := generateTestCertificate(t, false)
	cert, _ := ParseTLSCertificate(certPEM, keyPEM)

	err := cert.Reload()
	if err == nil {
		t.Error("无路径的证书重载应返回错误")
	}

	err = cert.ReloadFromPath()
	if err == nil {
		t.Error("无路径的证书重载应返回错误")
	}
}

func TestLoadTLS(t *testing.T) {
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	certPEM, keyPEM := generateTestCertificate(t, false)
	os.WriteFile(certPath, certPEM, 0644)
	os.WriteFile(keyPath, keyPEM, 0644)

	loader := NewLoader()
	cert, err := loader.LoadTLS(certPath, keyPath)
	if err != nil {
		t.Fatalf("LoadTLS() 失败: %v", err)
	}

	if cert.certPath != certPath {
		t.Errorf("certPath 应为 %s, 实际为 %s", certPath, cert.certPath)
	}

	if cert.keyPath != keyPath {
		t.Errorf("keyPath 应为 %s, 实际为 %s", keyPath, cert.keyPath)
	}
}

func TestLoadTLSNonExistent(t *testing.T) {
	loader := NewLoader()

	_, err := loader.LoadTLS("/nonexistent/cert.pem", "/nonexistent/key.pem")
	if err == nil {
		t.Error("加载不存在的证书应返回错误")
	}
}

func TestLoadTLCPNonExistent(t *testing.T) {
	loader := NewLoader()

	_, err := loader.LoadTLCP("/nonexistent/cert.pem", "/nonexistent/key.pem")
	if err == nil {
		t.Error("加载不存在的证书应返回错误")
	}
}

func TestLoadCertificate(t *testing.T) {
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	certPEM, keyPEM := generateTestCertificate(t, false)
	os.WriteFile(certPath, certPEM, 0644)
	os.WriteFile(keyPath, keyPEM, 0644)

	cert, err := LoadCertificate(CertTypeTLS, certPath, keyPath)
	if err != nil {
		t.Fatalf("LoadCertificate() 失败: %v", err)
	}

	if cert.Type() != CertTypeTLS {
		t.Errorf("证书类型应为 CertTypeTLS, 实际为 %v", cert.Type())
	}
}

func TestLoadCertificateInvalidType(t *testing.T) {
	_, err := LoadCertificate(CertType(99), "cert.pem", "key.pem")
	if err == nil {
		t.Error("无效证书类型应返回错误")
	}
}

func TestDetectCertType(t *testing.T) {
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")

	certPEM, _ := generateTestCertificate(t, false)
	os.WriteFile(certPath, certPEM, 0644)

	certType, err := DetectCertType(certPath)
	if err != nil {
		t.Fatalf("DetectCertType() 失败: %v", err)
	}

	if certType != CertTypeTLCP && certType != CertTypeTLS {
		t.Errorf("证书类型应为 TLCP 或 TLS, 实际为 %v", certType)
	}
}

func TestDetectCertTypeNonExistent(t *testing.T) {
	_, err := DetectCertType("/nonexistent/cert.pem")
	if err == nil {
		t.Error("检测不存在的证书类型应返回错误")
	}
}

func TestDetectCertTypeInvalidPEM(t *testing.T) {
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")

	os.WriteFile(certPath, []byte("not a pem"), 0644)

	_, err := DetectCertType(certPath)
	if err == nil {
		t.Error("检测无效PEM应返回错误")
	}
}

func TestLoadClientCA(t *testing.T) {
	tmpDir := t.TempDir()
	caPath := filepath.Join(tmpDir, "ca.pem")

	caPEM, _ := generateTestCA(t)
	os.WriteFile(caPath, caPEM, 0644)

	loader := NewLoader()
	pool, err := loader.LoadClientCA([]string{caPath})
	if err != nil {
		t.Fatalf("LoadClientCA() 失败: %v", err)
	}

	if pool == nil {
		t.Error("LoadClientCA() 不应返回 nil")
	}
}

func TestLoadClientCANonExistent(t *testing.T) {
	loader := NewLoader()

	_, err := loader.LoadClientCA([]string{"/nonexistent/ca.pem"})
	if err == nil {
		t.Error("加载不存在的CA证书应返回错误")
	}
}

func TestLoadServerCA(t *testing.T) {
	tmpDir := t.TempDir()
	caPath := filepath.Join(tmpDir, "ca.pem")

	caPEM, _ := generateTestCA(t)
	os.WriteFile(caPath, caPEM, 0644)

	loader := NewLoader()
	pool, err := loader.LoadServerCA([]string{caPath})
	if err != nil {
		t.Fatalf("LoadServerCA() 失败: %v", err)
	}

	if pool == nil {
		t.Error("LoadServerCA() 不应返回 nil")
	}
}

func TestLoaderClientCAS(t *testing.T) {
	loader := NewLoader()
	pool := x509.NewCertPool()

	loader.StoreClientCA("test-ca", pool)

	retrieved := loader.GetClientCA("test-ca")
	if retrieved == nil {
		t.Error("GetClientCA() 不应返回 nil")
	}

	if loader.GetClientCA("nonexistent") != nil {
		t.Error("GetClientCA() 对不存在的CA应返回 nil")
	}
}

func TestLoaderServerCAS(t *testing.T) {
	loader := NewLoader()
	pool := x509.NewCertPool()

	loader.StoreServerCA("test-ca", pool)

	retrieved := loader.GetServerCA("test-ca")
	if retrieved == nil {
		t.Error("GetServerCA() 不应返回 nil")
	}

	if loader.GetServerCA("nonexistent") != nil {
		t.Error("GetServerCA() 对不存在的CA应返回 nil")
	}
}

func TestLoaderReload(t *testing.T) {
	loader := NewLoader()

	err := loader.Reload("nonexistent")
	if err == nil {
		t.Error("重载不存在的证书应返回错误")
	}
}

func TestHotReloader(t *testing.T) {
	loader := NewLoader()
	reloader := NewHotReloader(loader, 100*time.Millisecond)

	reloader.Start()
	time.Sleep(150 * time.Millisecond)
	reloader.Stop()
}

func TestCertificateGetCertificate(t *testing.T) {
	certPEM, keyPEM := generateTestCertificate(t, false)
	cert, _ := ParseTLSCertificate(certPEM, keyPEM)

	tlsCert, err := cert.GetCertificate(nil)
	if err != nil {
		t.Fatalf("GetCertificate() 失败: %v", err)
	}

	if tlsCert == nil {
		t.Error("GetCertificate() 不应返回 nil")
	}
}

func TestParseTLSCertificateFromPath(t *testing.T) {
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	certPEM, keyPEM := generateTestCertificate(t, false)
	os.WriteFile(certPath, certPEM, 0644)
	os.WriteFile(keyPath, keyPEM, 0644)

	cert, err := ParseTLSCertificateFromPath(certPath, keyPath)
	if err != nil {
		t.Fatalf("ParseTLSCertificateFromPath() 失败: %v", err)
	}

	if cert.certPath != certPath {
		t.Errorf("certPath 应为 %s, 实际为 %s", certPath, cert.certPath)
	}
}

func TestCertificateLeafEmpty(t *testing.T) {
	cert := &Certificate{
		Certificate: []*x509.Certificate{},
	}

	if cert.Leaf() != nil {
		t.Error("空证书链 Leaf() 应返回 nil")
	}

	if cert.Subject() != "" {
		t.Error("空证书链 Subject() 应返回空字符串")
	}

	if cert.Issuer() != "" {
		t.Error("空证书链 Issuer() 应返回空字符串")
	}
}

func TestCertificateExpiresAtEmpty(t *testing.T) {
	cert := &Certificate{
		Certificate: []*x509.Certificate{},
	}

	if !cert.ExpiresAt().IsZero() {
		t.Error("空证书链 ExpiresAt() 应返回零值")
	}
}

func generateTestCertificate(t *testing.T, isCA bool) ([]byte, []byte) {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成RSA私钥失败: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Test Certificate",
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  isCA,
	}

	if isCA {
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("创建证书失败: %v", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})

	return certPEM, keyPEM
}

func generateTestECDSACertificate(t *testing.T) ([]byte, []byte) {
	t.Helper()

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("生成ECDSA私钥失败: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Test ECDSA Certificate",
			Organization: []string{"Test Org"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("创建证书失败: %v", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	keyBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		t.Fatalf("序列化ECDSA私钥失败: %v", err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	})

	return certPEM, keyPEM
}

func generateTestCA(t *testing.T) ([]byte, []byte) {
	t.Helper()
	return generateTestCertificate(t, true)
}
