package rootcert

import (
	"testing"

	"github.com/Trisia/tlcpchan/security/certgen"
)

// TestManagerAddSM2AndRSARootCert 验证根证书管理器统一使用 smx509 解析时，
// 既能识别国密 SM2 根证书，也能识别标准 RSA 根证书。
//
// 背景：gmsm v0.44.0 起 smx509.Certificate 与 x509.Certificate 成为相互独立的
// 结构体（ToX509 已移除），本用例覆盖改造后的解析路径，确保国密证书与标准证书
// 都能被正确加载并同步进入两个证书池。
//
// 参数：
//
//	t - Go 测试上下文
//
// 注意事项：
//   - 所有文件均写入 t.TempDir() 创建的临时目录，测试结束后由 testing 包自动清理
func TestManagerAddSM2AndRSARootCert(t *testing.T) {
	// 生成国密 SM2 根 CA（用于 TLCP）
	sm2CA, err := certgen.GenerateTLCPRootCA(certgen.CertGenConfig{
		Type:       certgen.CertTypeRootCA,
		CommonName: "test-sm2-root-ca",
		Org:        "tlcpchan-test",
		Country:    "CN",
		Years:      1,
	})
	if err != nil {
		t.Fatalf("生成 SM2 根 CA 失败: %v", err)
	}

	// 生成标准 RSA 根 CA（用于 TLS）
	rsaCA, err := certgen.GenerateTLSRootCA(certgen.CertGenConfig{
		Type:       certgen.CertTypeRootCA,
		CommonName: "test-rsa-root-ca",
		Org:        "tlcpchan-test",
		Country:    "CN",
		Years:      1,
	})
	if err != nil {
		t.Fatalf("生成 RSA 根 CA 失败: %v", err)
	}

	m := NewManager(t.TempDir())

	if _, err := m.Add("sm2-root-ca.crt", sm2CA.CertPEM); err != nil {
		t.Fatalf("添加 SM2 根证书失败: %v", err)
	}
	if _, err := m.Add("rsa-root-ca.crt", rsaCA.CertPEM); err != nil {
		t.Fatalf("添加 RSA 根证书失败: %v", err)
	}

	tests := []struct {
		name        string // 用例名称
		filename    string // 根证书文件名
		wantKeyType string // 期望的密钥类型
	}{
		{"国密 SM2 根证书", "sm2-root-ca.crt", "SM2"},
		{"标准 RSA 根证书", "rsa-root-ca.crt", "RSA-2048"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cert, err := m.Get(tt.filename)
			if err != nil {
				t.Fatalf("获取根证书失败: %v", err)
			}
			if cert.Cert == nil {
				t.Fatal("解析后的证书对象为空")
			}
			if cert.KeyType != tt.wantKeyType {
				t.Errorf("KeyType = %q, 期望 %q", cert.KeyType, tt.wantKeyType)
			}
			if !cert.IsCA {
				t.Error("IsCA = false, 期望 true")
			}
			if cert.Subject == "" || cert.Issuer == "" || cert.SerialNumber == "" {
				t.Errorf("证书元数据不完整: Subject=%q Issuer=%q SerialNumber=%q",
					cert.Subject, cert.Issuer, cert.SerialNumber)
			}
		})
	}

	// 两个证书池的加载结果
	pool := m.GetPool()
	if got := len(pool.GetCerts()); got != 2 {
		t.Errorf("根证书数量 = %d, 期望 2", got)
	}
	// 国密 TLCP 证书池使用 smx509，可解析全部两张根证书（SM2 + RSA）
	if got := len(pool.GetSMCertPool().Subjects()); got != 2 {
		t.Errorf("国密 TLCP 证书池根证书数量 = %d, 期望 2", got)
	}
	// 标准 TLS 证书池使用 crypto/x509，其不支持 SM2 曲线，
	// 无法解析国密根证书，因此只包含 RSA 根证书（既有行为，非本次改动引入）
	if got := len(pool.GetCertPool().Subjects()); got != 1 {
		t.Errorf("标准 TLS 证书池根证书数量 = %d, 期望 1", got)
	}
}
