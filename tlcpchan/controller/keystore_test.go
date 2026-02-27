package controller

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/logger"
	"github.com/Trisia/tlcpchan/security"
	"github.com/Trisia/tlcpchan/security/certgen"
)

func TestSecurityController_UpdateCertificates(t *testing.T) {
	tempDir := t.TempDir()

	cfg := config.Default()
	cfg.WorkDir = tempDir

	keyStoreMgr := security.NewKeyStoreManager()
	rootCertMgr := security.NewRootCertManager(tempDir)

	ctrl := &SecurityController{
		keyStoreMgr: keyStoreMgr,
		rootCertMgr: rootCertMgr,
		cfg:         cfg,
		log:         logger.Default(),
	}

	keystoreDir := filepath.Join(tempDir, "keystores")
	if err := os.MkdirAll(keystoreDir, 0755); err != nil {
		t.Fatalf("创建 keystores 目录失败: %v", err)
	}

	caCert, err := certgen.GenerateTLCPRootCA(certgen.CertGenConfig{
		CommonName: "test-ca",
		Org:        "Test Org",
		Years:      10,
	})
	if err != nil {
		t.Fatalf("生成 CA 证书失败: %v", err)
	}

	caCertPath := filepath.Join(keystoreDir, "test-ca.crt")
	caKeyPath := filepath.Join(keystoreDir, "test-ca.key")
	if err := certgen.SaveCertToFile(caCert.CertPEM, caCert.KeyPEM, caCertPath, caKeyPath); err != nil {
		t.Fatalf("保存 CA 证书失败: %v", err)
	}

	signCfg := certgen.CertGenConfig{
		CommonName: "test-sign",
		Org:        "Test Org",
		Years:      1,
	}

	encCfg := certgen.CertGenConfig{
		CommonName: "test-enc",
		Org:        "Test Org",
		Years:      1,
	}

	signerX509Cert, signerPrivKey, err := certgen.LoadTLCPCertFromFile(caCertPath, caKeyPath)
	if err != nil {
		t.Fatalf("加载 CA 证书失败: %v", err)
	}

	signCert, _, err := certgen.GenerateTLCPPair(signerX509Cert, signerPrivKey, signCfg, signCfg)
	if err != nil {
		t.Fatalf("生成签名证书失败: %v", err)
	}

	encCert, _, err := certgen.GenerateTLCPPair(signerX509Cert, signerPrivKey, encCfg, encCfg)
	if err != nil {
		t.Fatalf("生成加密证书失败: %v", err)
	}

	signCertPath := filepath.Join(keystoreDir, "test-sign.crt")
	signKeyPath := filepath.Join(keystoreDir, "test-sign.key")
	encCertPath := filepath.Join(keystoreDir, "test-enc.crt")
	encKeyPath := filepath.Join(keystoreDir, "test-enc.key")

	if err := certgen.SaveCertToFile(signCert.CertPEM, signCert.KeyPEM, signCertPath, signKeyPath); err != nil {
		t.Fatalf("保存签名证书失败: %v", err)
	}
	if err := certgen.SaveCertToFile(encCert.CertPEM, encCert.KeyPEM, encCertPath, encKeyPath); err != nil {
		t.Fatalf("保存加密证书失败: %v", err)
	}

	_, err = keyStoreMgr.Create("test", "file", map[string]string{
		"sign-cert": "test-sign.crt",
		"sign-key":  "test-sign.key",
		"enc-cert":  "test-enc.crt",
		"enc-key":   "test-enc.key",
	}, false)
	if err != nil {
		t.Fatalf("创建 keystore 失败: %v", err)
	}

	newSignCert, _, err := certgen.GenerateTLCPPair(signerX509Cert, signerPrivKey, signCfg, signCfg)
	if err != nil {
		t.Fatalf("生成新签名证书失败: %v", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("signCert", "new-sign.crt")
	if err != nil {
		t.Fatalf("创建表单文件失败: %v", err)
	}
	if _, err := part.Write(newSignCert.CertPEM); err != nil {
		t.Fatalf("写入签名证书失败: %v", err)
	}

	part, err = writer.CreateFormFile("signKey", "new-sign.key")
	if err != nil {
		t.Fatalf("创建表单文件失败: %v", err)
	}
	if _, err := part.Write(newSignCert.KeyPEM); err != nil {
		t.Fatalf("写入签名私钥失败: %v", err)
	}

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/security/keystores/test/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	ctrl.UpdateCertificates(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("期望状态码 %d，实际 %d，响应: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	t.Log("更新证书测试通过")
}

func TestSecurityController_UpdateCertificates_Errors(t *testing.T) {
	tempDir := t.TempDir()

	cfg := config.Default()
	cfg.WorkDir = tempDir

	keyStoreMgr := security.NewKeyStoreManager()
	rootCertMgr := security.NewRootCertManager(tempDir)

	ctrl := &SecurityController{
		keyStoreMgr: keyStoreMgr,
		rootCertMgr: rootCertMgr,
		cfg:         cfg,
		log:         logger.Default(),
	}

	tests := []struct {
		name        string
		keystore    string
		body        io.Reader
		contentType string
		wantStatus  int
	}{
		{
			name:        "keystore 不存在",
			keystore:    "nonexistent",
			body:        bytes.NewBufferString(""),
			contentType: "multipart/formari-data",
			wantStatus:  http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/security/keystores/"+tt.keystore+"/upload", tt.body)
			req.Header.Set("Content-Type", tt.contentType)
			rec := httptest.NewRecorder()

			ctrl.UpdateCertificates(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("期望状态码 %d，实际 %d，响应: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}
