package controller_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Trisia/tlcpchan/security/certgen"
)

// TestGenTestCertificates 生成用于集成测试的证书
// 证书保存到 ../test_certs/ 目录
// 可以独立运行：go test -v -run TestGenTestCertificates
func TestGenTestCertificates(t *testing.T) {
	certDir := filepath.Join("..", "test_certs")
	if err := os.MkdirAll(certDir, 0755); err != nil {
		t.Fatalf("创建证书目录失败: %v", err)
	}

	t.Log("开始生成测试证书...")

	generateTLCPCerts(t, certDir)
	generateTLSCerts(t, certDir)

	t.Logf("测试证书已生成到 %s/ 目录", certDir)
}

func generateTLCPCerts(t *testing.T, certDir string) {
	t.Log("生成 TLCP 根 CA 证书...")

	caCert, err := certgen.GenerateTLCPRootCA(cert(certgen.CertGenConfig{
		Type:       certgen.CertTypeRootCA,
		CommonName: "test-tlcp-ca",
		Org:        "Test Organization",
		Country:    "CN",
		Years:      10,
	}))
	if err != nil {
		t.Fatalf("生成 TLCP 根 CA 证书失败: %v", err)
	}

	caCertPath := filepath.Join(certDir, "test-tlcp-ca.crt")
	caKeyPath := filepath.Join(certDir, "test-tlcp-ca.key")
	if err := certgen.SaveCertToFile(caCert.CertPEM, caCert.KeyPEM, caCertPath, caKeyPath); err != nil {
		t.Fatalf("保存 TLCP 根 CA 证书失败: %v", err)
	}

	t.Log("加载 CA 证书用于签发...")
	signerX509Cert, signerPrivKey, err := certgen.LoadTLCPCertFromFile(caCertPath, caKeyPath)
	if err != nil {
		t.Fatalf("加载 TLCP CA 证书失败: %v", err)
	}

	t.Log("生成 TLCP 签名证书 v1 和加密证书 v1...")
	signCertV1, encCertV1, err := certgen.GenerateTLCPPair(
		signerX509Cert, signerPrivKey,
		cert(certgen.CertGenConfig{
			Type:       certgen.CertTypeTLCPSign,
			CommonName: "test-tlcp-sign-v1",
			Org:        "Test Organization",
			Country:    "CN",
			Years:      1,
		}),
		cert(certgen.CertGenConfig{
			Type:       certgen.CertTypeTLCPEnc,
			CommonName: "test-tlcp-enc-v1",
			Org:        "Test Organization",
			Country:    "CN",
			Years:      1,
		}),
	)
	if err != nil {
		t.Fatalf("生成 TLCP 证书对 v1 失败: %v", err)
	}

	if err := certgen.SaveCertToFile(
		signCertV1.CertPEM, signCertV1.KeyPEM,
		filepath.Join(certDir, "test-tlcp-sign-v1.crt"),
		filepath.Join(certDir, "test-tlcp-sign-v1.key"),
	); err != nil {
		t.Fatalf("保存 TLCP 签名证书 v1 失败: %v", err)
	}

	if err := certgen.SaveCertToFile(
		encCertV1.CertPEM, encCertV1.KeyPEM,
		filepath.Join(certDir, "test-tlcp-enc-v1.crt"),
		filepath.Join(certDir, "test-tlcp-enc-v1.key"),
	); err != nil {
		t.Fatalf("保存 TLCP 加密证书 v1 失败: %v", err)
	}

	t.Log("生成 TLCP 签名证书 v2 和加密证书 v2...")
	signCertV2, encCertV2, err := certgen.GenerateTLCPPair(
		signerX509Cert, signerPrivKey,
		cert(certgen.CertGenConfig{
			Type:       certgen.CertTypeTLCPSign,
			CommonName: "test-tlcp-sign-v2",
			Org:        "Test Organization",
			Country:    "CN",
			Years:      1,
		}),
		cert(certgen.CertGenConfig{
			Type:       certgen.CertTypeTLCPEnc,
			CommonName: "test-tlcp-enc-v2",
			Org:        "Test Organization",
			Country:    "CN",
			Years:      1,
		}),
	)
	if err != nil {
		t.Fatalf("生成 TLCP 证书对 v2 失败: %v", err)
	}

	if err := certgen.SaveCertToFile(
		signCertV2.CertPEM, signCertV2.KeyPEM,
		filepath.Join(certDir, "test-tlcp-sign-v2.crt"),
		filepath.Join(certDir, "test-tlcp-sign-v2.key"),
	); err != nil {
		t.Fatalf("保存 TLCP 签名证书 v2 失败: %v", err)
	}

	if err := certgen.SaveCertToFile(
		encCertV2.CertPEM, encCertV2.KeyPEM,
		filepath.Join(certDir, "test-tlcp-enc-v2.crt"),
		filepath.Join(certDir, "test-tlcp-enc-v2.key"),
	); err != nil {
		t.Fatalf("保存 TLCP 加密证书 v2 失败: %v", err)
	}

	t.Log("TLCP 测试证书生成完成")
}

func generateTLSCerts(t *testing.T, certDir string) {
	t.Log("生成 TLS 根 CA 证书...")

	caCert, err := certgen.GenerateTLSRootCA(cert(certgen.CertGenConfig{
		Type:       certgen.CertTypeRootCA,
		CommonName: "test-tls-ca",
		Org:        "Test Organization",
		Country:    "CN",
		Years:      10,
	}))
	if err != nil {
		t.Fatalf("生成 TLS 根 CA 证书失败: %v", err)
	}

	caCertPath := filepath.Join(certDir, "test-tls-ca.crt")
	caKeyPath := filepath.Join(certDir, "test-tls-ca.key")
	if err := certgen.SaveCertToFile(caCert.CertPEM, caCert.KeyPEM, caCertPath, caKeyPath); err != nil {
		t.Fatalf("保存 TLS 根 CA 证书失败: %v", err)
	}

	t.Log("加载 CA 证书用于签发...")
	signerX509Cert, signerPrivKey, err := certgen.LoadTLSCertFromFile(caCertPath, caKeyPath)
	if err != nil {
		t.Fatalf("加载 TLS CA 证书失败: %v", err)
	}

	t.Log("生成 TLS 证书 v1...")
	tlsCertV1, err := certgen.GenerateTLSCert(
		signerX509Cert, signerPrivKey,
		cert(certgen.CertGenConfig{
			Type:         certgen.CertTypeTLS,
			CommonName:   "test-tls-v1",
			Org:          "Test Organization",
			Country:      "CN",
			Years:        1,
			KeyAlgorithm: certgen.KeyAlgorithmECDSA,
			KeyBits:      256,
			DNSNames:     []string{"test.example.com"},
			IPAddresses:  []string{"127.0.0.1"},
		}),
	)
	if err != nil {
		t.Fatalf("生成 TLS 证书 v1 失败: %v", err)
	}

	if err := certgen.SaveCertToFile(
		tlsCertV1.CertPEM, tlsCertV1.KeyPEM,
		filepath.Join(certDir, "test-tls-v1.crt"),
		filepath.Join(certDir, "test-tls-v1.key"),
	); err != nil {
		t.Fatalf("保存 TLS 证书 v1 失败: %v", err)
	}

	t.Log("生成 TLS 证书 v2...")
	tlsCertV2, err := certgen.GenerateTLSCert(
		signerX509Cert, signerPrivKey,
		cert(certgen.CertGenConfig{
			Type:         certgen.CertTypeTLS,
			CommonName:   "test-tls-v2",
			Org:          "Test Organization",
			Country:      "CN",
			Years:        1,
			KeyAlgorithm: certgen.KeyAlgorithmECDSA,
			KeyBits:      256,
			DNSNames:     []string{"test.example.com"},
			IPAddresses:  []string{"127.0.0.1"},
		}),
	)
	if err != nil {
		t.Fatalf("生成 TLS 证书 v2 失败: %v", err)
	}

	if err := certgen.SaveCertToFile(
		tlsCertV2.CertPEM, tlsCertV2.KeyPEM,
		filepath.Join(certDir, "test-tls-v2.crt"),
		filepath.Join(certDir, "test-tls-v2.key"),
	); err != nil {
		t.Fatalf("保存 TLS 证书 v2 失败: %v", err)
	}

	t.Log("TLS 测试证书生成完成")
}

func cert(cfg certgen.CertGenConfig) certgen.CertGenConfig {
	return cfg
}
