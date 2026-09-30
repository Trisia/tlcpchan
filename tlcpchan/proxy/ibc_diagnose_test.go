package proxy

import (
	"crypto/tls"
	"strings"
	"testing"
	"time"

	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/Trisia/tlcpchan/security"
	"github.com/Trisia/tlcpchan/security/certgen"
	"github.com/Trisia/tlcpchan/security/keystore"
)

// ibcDiagnoseFixture 诊断测试材料。
type ibcDiagnoseFixture struct {
	kgc       *certgen.GeneratedIBCParams // 测试 KGC
	fullIdent *tlcp.IBCIdentity           // 材料齐备的身份
	identity  []byte                      // 标识字节
}

// newIBCDiagnoseFixture 构造诊断测试所需的 KGC 与用户身份。
//
// 参数：
//   - t: 测试上下文
//
// 返回值：
//   - *ibcDiagnoseFixture: 测试材料
func newIBCDiagnoseFixture(t *testing.T) *ibcDiagnoseFixture {
	t.Helper()
	kgc, err := certgen.GenerateIBCParams("tlcpchan.local", 1, tlcp.ValidityPeriod{
		NotBefore: time.Now().Add(-time.Hour).Truncate(time.Second),
		NotAfter:  time.Now().AddDate(10, 0, 0),
	})
	if err != nil {
		t.Fatalf("GenerateIBCParams() error = %v", err)
	}
	identity := []byte("server@tlcpchan.local")
	generated, err := certgen.GenerateIBCIdentity(kgc.Master, identity)
	if err != nil {
		t.Fatalf("GenerateIBCIdentity() error = %v", err)
	}

	ident, err := keystore.LoadIBCIdentityFromData(generated.Identity, kgc.ParamsPEM,
		generated.SignKeyPEM, generated.EncKeyPEM, generated.KexKeyPEM)
	if err != nil {
		t.Fatalf("LoadIBCIdentityFromData() error = %v", err)
	}

	return &ibcDiagnoseFixture{kgc: kgc, fullIdent: ident, identity: identity}
}

// cloneIBCIdentity 复制一份身份，便于按场景裁剪材料。
//
// 返回值：
//   - *tlcp.IBCIdentity: 身份副本
func (f *ibcDiagnoseFixture) cloneIBCIdentity() *tlcp.IBCIdentity {
	clone := *f.fullIdent
	return &clone
}

// ibcKeyStoreStub 仅用于诊断测试的 IBC keystore 替身。
type ibcKeyStoreStub struct{}

// Type 返回 ibc 类型。
func (ibcKeyStoreStub) Type() security.KeyStoreType { return security.KeyStoreTypeIBC }

// TLCPCertificate 返回空，诊断逻辑不使用证书。
func (ibcKeyStoreStub) TLCPCertificate() ([]*tlcp.Certificate, error) { return nil, nil }

// TLSCertificate 返回空，诊断逻辑不使用证书。
func (ibcKeyStoreStub) TLSCertificate() (*tls.Certificate, error) { return nil, nil }

// findDiagnosis 在诊断结论中查找包含指定关键字且级别匹配的条目。
//
// 参数：
//   - diagnoses: 诊断结论列表
//   - level: 期望级别
//   - keyword: 期望包含的关键字
//
// 返回值：
//   - bool: 命中返回 true
func findDiagnosis(diagnoses []ibcDiagnosis, level, keyword string) bool {
	for _, diagnosis := range diagnoses {
		if diagnosis.Level == level && strings.Contains(diagnosis.Message, keyword) {
			return true
		}
	}
	return false
}

// TestDiagnoseIBCSuites 覆盖 IBC 能力诊断的各分支。
func TestDiagnoseIBCSuites(t *testing.T) {
	fixture := newIBCDiagnoseFixture(t)
	ibcKS := ibcKeyStoreStub{}
	ibcSuite := []uint16{0xE057}   // IBC_SM4_GCM_SM3
	ibsdhSuite := []uint16{0xE055} // IBSDH_SM4_GCM_SM3
	certSuite := []uint16{0xE011}  // ECC_SM4_CBC_SM3

	t.Run("未配置 IBC 套件时不产生诊断", func(t *testing.T) {
		if got := diagnoseIBCSuites("inst", true, certSuite, nil, nil, 0); len(got) != 0 {
			t.Fatalf("诊断结论 = %+v, want 空", got)
		}
		if got := diagnoseIBCSuites("inst", true, nil, ibcKS, fixture.fullIdent, 1); len(got) != 0 {
			t.Fatalf("诊断结论 = %+v, want 空", got)
		}
	})

	t.Run("配置套件但缺少 IBC keystore", func(t *testing.T) {
		got := diagnoseIBCSuites("inst", true, ibcSuite, nil, nil, 1)
		if !findDiagnosis(got, ibcDiagLevelError, "未配置可用的 IBC 身份 keystore") {
			t.Fatalf("诊断结论 = %+v, want 缺 keystore 的 error", got)
		}
	})

	t.Run("IBC 身份装载失败", func(t *testing.T) {
		got := diagnoseIBCSuites("inst", true, ibcSuite, ibcKS, nil, 1)
		if !findDiagnosis(got, ibcDiagLevelError, "IBC 身份装载失败") {
			t.Fatalf("诊断结论 = %+v, want 装载失败的 error", got)
		}
	})

	t.Run("IBSDH 且标识为空", func(t *testing.T) {
		ident := fixture.cloneIBCIdentity()
		ident.Identity = nil
		got := diagnoseIBCSuites("inst", true, ibsdhSuite, ibcKS, ident, 1)
		if !findDiagnosis(got, ibcDiagLevelWarn, "identity_need(205)") {
			t.Fatalf("诊断结论 = %+v, want identity_need 的 warn", got)
		}
	})

	t.Run("仅 IBC 套件且标识为空", func(t *testing.T) {
		ident := fixture.cloneIBCIdentity()
		ident.Identity = nil
		got := diagnoseIBCSuites("inst", true, ibcSuite, ibcKS, ident, 1)
		if !findDiagnosis(got, ibcDiagLevelWarn, "标识为空") {
			t.Fatalf("诊断结论 = %+v, want 标识为空的 warn", got)
		}
		if findDiagnosis(got, ibcDiagLevelWarn, "identity_need(205)") {
			t.Fatalf("IBC 套件不应提示 identity_need: %+v", got)
		}
	})

	t.Run("服务端缺少本端公共参数", func(t *testing.T) {
		ident := fixture.cloneIBCIdentity()
		ident.Parameters = nil
		got := diagnoseIBCSuites("inst", true, ibcSuite, ibcKS, ident, 1)
		if !findDiagnosis(got, ibcDiagLevelError, "bad_ibcparam(203)") {
			t.Fatalf("诊断结论 = %+v, want bad_ibcparam 的 error", got)
		}
	})

	t.Run("客户端缺少公共参数且无私钥", func(t *testing.T) {
		ident := fixture.cloneIBCIdentity()
		ident.Parameters = nil
		ident.SignPrivateKey = nil
		got := diagnoseIBCSuites("inst", false, ibcSuite, ibcKS, ident, 1)
		if !findDiagnosis(got, ibcDiagLevelWarn, "单向认证场景可省略") {
			t.Fatalf("诊断结论 = %+v, want 公共参数可省略的 warn", got)
		}
		if findDiagnosis(got, ibcDiagLevelError, "bad_ibcparam(203)") {
			t.Fatalf("客户端单向认证场景不应报 bad_ibcparam: %+v", got)
		}
	})

	t.Run("服务端缺少签名私钥", func(t *testing.T) {
		ident := fixture.cloneIBCIdentity()
		ident.SignPrivateKey = nil
		got := diagnoseIBCSuites("inst", true, ibcSuite, ibcKS, ident, 1)
		if !findDiagnosis(got, ibcDiagLevelError, "缺少签名私钥") {
			t.Fatalf("诊断结论 = %+v, want 缺签名私钥的 error", got)
		}
	})

	t.Run("客户端缺少签名私钥", func(t *testing.T) {
		ident := fixture.cloneIBCIdentity()
		ident.SignPrivateKey = nil
		got := diagnoseIBCSuites("inst", false, ibcSuite, ibcKS, ident, 1)
		if !findDiagnosis(got, ibcDiagLevelWarn, "若服务端要求客户端认证将失败") {
			t.Fatalf("诊断结论 = %+v, want 客户端缺签名私钥的 warn", got)
		}
	})

	t.Run("服务端 IBC 套件缺少加密私钥", func(t *testing.T) {
		ident := fixture.cloneIBCIdentity()
		ident.EncryptPrivateKey = nil
		got := diagnoseIBCSuites("inst", true, ibcSuite, ibcKS, ident, 1)
		if !findDiagnosis(got, ibcDiagLevelError, "缺少加密私钥") {
			t.Fatalf("诊断结论 = %+v, want 缺加密私钥的 error", got)
		}
	})

	t.Run("IBSDH 套件缺少密钥交换私钥", func(t *testing.T) {
		ident := fixture.cloneIBCIdentity()
		ident.KeyExchangePrivateKey = nil
		got := diagnoseIBCSuites("inst", true, ibsdhSuite, ibcKS, ident, 1)
		if !findDiagnosis(got, ibcDiagLevelError, "缺少密钥交换私钥") {
			t.Fatalf("诊断结论 = %+v, want 缺密钥交换私钥的 error", got)
		}
	})

	t.Run("信任池为空", func(t *testing.T) {
		got := diagnoseIBCSuites("inst", true, ibcSuite, ibcKS, fixture.fullIdent, 0)
		if !findDiagnosis(got, ibcDiagLevelError, "信任池为空") {
			t.Fatalf("诊断结论 = %+v, want 信任池为空的 error", got)
		}
	})

	t.Run("材料齐备且信任池非空", func(t *testing.T) {
		got := diagnoseIBCSuites("inst", true, ibcSuite, ibcKS, fixture.fullIdent, 1)
		if len(got) != 0 {
			t.Fatalf("诊断结论 = %+v, want 空", got)
		}
	})
}
