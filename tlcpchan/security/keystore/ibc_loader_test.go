package keystore

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/Trisia/tlcpchan/security/certgen"
)

// ibcTestIdentity 生成一份测试用 KGC 公共参数与用户身份。
//
// 参数：
//   - t: 测试上下文
//   - identity: 用户标识，如 "server@tlcpchan.local"
//
// 返回值：
//   - *certgen.GeneratedIBCParams: KGC 公共参数与主密钥
//   - *certgen.GeneratedIBCIdentity: 该标识对应的用户私钥
func ibcTestIdentity(t *testing.T, identity string) (*certgen.GeneratedIBCParams, *certgen.GeneratedIBCIdentity) {
	t.Helper()
	kgc, err := certgen.GenerateIBCParams("tlcpchan.local", 1, tlcp.ValidityPeriod{
		NotBefore: time.Now().Add(-time.Hour).Truncate(time.Second),
		NotAfter:  time.Now().AddDate(10, 0, 0),
	})
	if err != nil {
		t.Fatalf("GenerateIBCParams() error = %v", err)
	}
	generated, err := certgen.GenerateIBCIdentity(kgc.Master, []byte(identity))
	if err != nil {
		t.Fatalf("GenerateIBCIdentity() error = %v", err)
	}
	return kgc, generated
}

// writeIBCMaterialFiles 将一份 IBC 身份写为材料文件。
//
// 参数：
//   - t: 测试上下文
//   - dir: 目标目录
//   - name: 文件名前缀
//   - kgc: KGC 公共参数（PEM）
//   - generated: 标识与三把用户私钥（PEM）
//
// 返回值：
//   - map[string]string: ibc-file 加载器参数
func writeIBCMaterialFiles(t *testing.T, dir, name string, kgc *certgen.GeneratedIBCParams,
	generated *certgen.GeneratedIBCIdentity) map[string]string {
	t.Helper()
	paths := map[string][]byte{
		"identity.txt": generated.Identity,
		"params.pem":   kgc.ParamsPEM,
		"sign.key":     generated.SignKeyPEM,
		"enc.key":      generated.EncKeyPEM,
		"kex.key":      generated.KexKeyPEM,
	}
	params := make(map[string]string, len(paths))
	for suffix, data := range paths {
		path := filepath.Join(dir, name+"-"+suffix)
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatalf("写入材料文件 %s 失败: %v", path, err)
		}
	}
	params[IBCParamIdentity] = filepath.Join(dir, name+"-identity.txt")
	params[IBCParamParams] = filepath.Join(dir, name+"-params.pem")
	params[IBCParamSignKey] = filepath.Join(dir, name+"-sign.key")
	params[IBCParamEncKey] = filepath.Join(dir, name+"-enc.key")
	params[IBCParamKexKey] = filepath.Join(dir, name+"-kex.key")
	return params
}

// TestIBCFileLoaderLoad 验证 IBC 身份材料可从文件完整装载并生成正确的元信息。
func TestIBCFileLoaderLoad(t *testing.T) {
	dir := t.TempDir()
	kgc, generated := ibcTestIdentity(t, "server@tlcpchan.local")
	params := writeIBCMaterialFiles(t, dir, "ibc-server", kgc, generated)

	loader := NewIBCFileLoader("")
	ks, err := loader.Load(LoaderTypeIBCFile, params)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if ks.Type() != KeyStoreTypeIBC {
		t.Fatalf("Type() = %q, want %q", ks.Type(), KeyStoreTypeIBC)
	}

	ibcKS, ok := ks.(IBCKeyStore)
	if !ok {
		t.Fatal("装载结果未实现 IBCKeyStore 接口")
	}

	ident, err := ibcKS.IBCIdentity()
	if err != nil {
		t.Fatalf("IBCIdentity() error = %v", err)
	}
	if string(ident.Identity) != "server@tlcpchan.local" {
		t.Fatalf("Identity = %q, want %q", ident.Identity, "server@tlcpchan.local")
	}
	if ident.Parameters == nil {
		t.Fatal("Parameters 为空，期望已解析 KGC 公共参数")
	}
	if ident.SignPrivateKey == nil || ident.EncryptPrivateKey == nil || ident.KeyExchangePrivateKey == nil {
		t.Fatalf("三把用户私钥未齐备: sign=%v enc=%v kex=%v",
			ident.SignPrivateKey != nil, ident.EncryptPrivateKey != nil, ident.KeyExchangePrivateKey != nil)
	}

	info := ibcKS.IBCInfo()
	if info == nil {
		t.Fatal("IBCInfo() = nil")
	}
	if info.Identity != "server@tlcpchan.local" || !info.HasParams || !info.HasSignKey ||
		!info.HasEncryptKey || !info.HasKeyExchangeKey {
		t.Fatalf("IBCInfo() = %+v, 字段不符合预期", info)
	}
	if info.DistrictName != "tlcpchan.local" || info.DistrictSerial != 1 {
		t.Fatalf("IBCInfo() KGC 信息 = %s#%d, want tlcpchan.local#1", info.DistrictName, info.DistrictSerial)
	}
	if info.NotAfter.IsZero() {
		t.Fatal("IBCInfo().NotAfter 为零值，期望解析出有效期")
	}
}

// TestIBCFileLoaderPartialMaterials 验证仅提供标识时也能装载（无参数、无私钥）。
func TestIBCFileLoaderPartialMaterials(t *testing.T) {
	dir := t.TempDir()
	identityPath := filepath.Join(dir, "identity.txt")
	if err := os.WriteFile(identityPath, []byte("client@tlcpchan.local"), 0644); err != nil {
		t.Fatalf("写入标识文件失败: %v", err)
	}

	loader := NewIBCFileLoader("")
	ks, err := loader.Load(LoaderTypeIBCFile, map[string]string{IBCParamIdentity: identityPath})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	info := ks.(IBCKeyStore).IBCInfo()
	if info == nil || info.Identity != "client@tlcpchan.local" {
		t.Fatalf("IBCInfo() = %+v, want identity=client@tlcpchan.local", info)
	}
	if info.HasParams || info.HasSignKey || info.HasEncryptKey || info.HasKeyExchangeKey {
		t.Fatalf("IBCInfo() = %+v, 期望不含参数与私钥", info)
	}
}

// TestIBCFileLoaderErrors 验证各类材料错误都能在装载阶段被发现。
func TestIBCFileLoaderErrors(t *testing.T) {
	dir := t.TempDir()
	kgc, generated := ibcTestIdentity(t, "server@tlcpchan.local")
	validParams := writeIBCMaterialFiles(t, dir, "ibc-server", kgc, generated)

	otherKGC, otherGenerated := ibcTestIdentity(t, "client@tlcpchan.local")
	otherParams := writeIBCMaterialFiles(t, dir, "ibc-client", otherKGC, otherGenerated)

	tests := []struct {
		name    string
		params  map[string]string
		wantErr bool
	}{
		{
			name:    "材料全部为空",
			params:  map[string]string{},
			wantErr: true,
		},
		{
			name:    "材料路径全部为空字符串",
			params:  map[string]string{IBCParamIdentity: "", IBCParamParams: ""},
			wantErr: true,
		},
		{
			name:    "标识文件不存在",
			params:  map[string]string{IBCParamIdentity: filepath.Join(dir, "missing.txt")},
			wantErr: true,
		},
		{
			name: "私钥与标识不匹配",
			params: map[string]string{
				IBCParamIdentity: validParams[IBCParamIdentity],
				IBCParamParams:   validParams[IBCParamParams],
				// 使用另一个标识的签名私钥，自检应当失败
				IBCParamSignKey: otherParams[IBCParamSignKey],
			},
			wantErr: true,
		},
		{
			name:    "签名私钥内容非法",
			params:  map[string]string{IBCParamSignKey: writeFile(t, dir, "bad-sign.key", []byte("not-a-key"))},
			wantErr: true,
		},
		{
			name:    "完整有效材料",
			params:  validParams,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewIBCFileLoader("")
			_, err := loader.Load(LoaderTypeIBCFile, tt.params)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Load() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

// TestIBCKeyStoreCertificateErrors 验证 IBC keystore 不提供 X.509 证书。
func TestIBCKeyStoreCertificateErrors(t *testing.T) {
	dir := t.TempDir()
	kgc, generated := ibcTestIdentity(t, "server@tlcpchan.local")
	params := writeIBCMaterialFiles(t, dir, "ibc-server", kgc, generated)

	ks, err := NewIBCFileLoader("").Load(LoaderTypeIBCFile, params)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if _, err := ks.TLCPCertificate(); err == nil {
		t.Fatal("TLCPCertificate() 应当返回错误")
	}
	if _, err := ks.TLSCertificate(); err == nil {
		t.Fatal("TLSCertificate() 应当返回错误")
	}
}

// writeFile 将内容写入目录下的指定文件并返回完整路径。
//
// 参数：
//   - t: 测试上下文
//   - dir: 目标目录
//   - name: 文件名
//   - data: 文件内容
//
// 返回值：
//   - string: 写入文件的完整路径
func writeFile(t *testing.T, dir, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("写入文件 %s 失败: %v", path, err)
	}
	return path
}
