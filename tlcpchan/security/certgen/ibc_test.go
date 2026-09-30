package certgen

import (
	"crypto/rand"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/emmansun/gmsm/sm3"
	"github.com/emmansun/gmsm/sm9"
	"github.com/emmansun/gmsm/smx509"
)

// testIBCValidity 返回测试用公共参数有效期：生效时间为 1 小时前，失效时间为 10 年后。
func testIBCValidity() tlcp.ValidityPeriod {
	return tlcp.ValidityPeriod{
		NotBefore: time.Now().Add(-time.Hour).Truncate(time.Second),
		NotAfter:  time.Now().AddDate(10, 0, 0),
	}
}

// TestGenerateIBCParams 验证 KGC 公共参数与主密钥的生成、编码与解析。
func TestGenerateIBCParams(t *testing.T) {
	generated, err := GenerateIBCParams("tlcpchan.local", 1, testIBCValidity())
	if err != nil {
		t.Fatalf("GenerateIBCParams() error = %v", err)
	}

	if generated.Params == nil || generated.Master == nil {
		t.Fatal("Params 与 Master 均不应为 nil")
	}
	if generated.Params.DistrictName != "tlcpchan.local" || generated.Params.DistrictSerial != 1 {
		t.Fatalf("KGC 信息 = %s#%d", generated.Params.DistrictName, generated.Params.DistrictSerial)
	}

	// 公共参数 PEM 可解析回等价对象
	block, _ := pem.Decode(generated.ParamsPEM)
	if block == nil || block.Type != IBCParamsPEMType {
		t.Fatalf("ParamsPEM 类型 = %v, want %s", block, IBCParamsPEMType)
	}
	parsed, err := tlcp.ParseIBCSysParams(block.Bytes)
	if err != nil {
		t.Fatalf("ParseIBCSysParams() error = %v", err)
	}
	if parsed.DistrictName != generated.Params.DistrictName || parsed.DistrictSerial != generated.Params.DistrictSerial {
		t.Fatal("解析后的公共参数与生成结果不一致")
	}

	// 主密钥 PEM 应包含签名与加密两个块
	rest := generated.MasterPEM
	signBlock, rest := pem.Decode(rest)
	if signBlock == nil || signBlock.Type != IBCSignMasterPEMType {
		t.Fatalf("第一个主密钥块类型 = %v, want %s", signBlock, IBCSignMasterPEMType)
	}
	encBlock, _ := pem.Decode(rest)
	if encBlock == nil || encBlock.Type != IBCEncryptMasterPEMType {
		t.Fatalf("第二个主密钥块类型 = %v, want %s", encBlock, IBCEncryptMasterPEMType)
	}
}

// TestLoadIBCMasterFromPEM 验证主密钥 PEM 的往返装载可用于派生用户私钥。
func TestLoadIBCMasterFromPEM(t *testing.T) {
	generated, err := GenerateIBCParams("tlcpchan.local", 1, testIBCValidity())
	if err != nil {
		t.Fatalf("GenerateIBCParams() error = %v", err)
	}

	master, err := LoadIBCMasterFromPEM(generated.MasterPEM)
	if err != nil {
		t.Fatalf("LoadIBCMasterFromPEM() error = %v", err)
	}
	if master.SignMaster == nil || master.EncryptMaster == nil {
		t.Fatal("装载出的主密钥不应为 nil")
	}

	// 使用装载出的主密钥派生用户密钥并验签，验证主公钥可复原
	identity := []byte("server@tlcpchan.local")
	signPriv, err := master.SignMaster.GenerateUserKey(identity, IBCSignHid)
	if err != nil {
		t.Fatalf("GenerateUserKey(sign) error = %v", err)
	}

	digest := sm3.Sum([]byte("tlcpchan-certgen-test"))
	sig, err := sm9.SignASN1(rand.Reader, signPriv, digest[:])
	if err != nil {
		t.Fatalf("SignASN1() error = %v", err)
	}
	pub := master.SignMaster.PublicKey()
	if !sm9.VerifyASN1(pub, identity, IBCSignHid, digest[:], sig) {
		t.Fatal("使用装载出的主密钥验签失败")
	}

	// 非法 PEM 应返回错误
	if _, err := LoadIBCMasterFromPEM([]byte("not-a-pem")); err == nil {
		t.Fatal("非法主密钥 PEM 应当返回错误")
	}
}

// TestGenerateIBCIdentity 验证用户私钥派生结果的编码与用途标识。
func TestGenerateIBCIdentity(t *testing.T) {
	generated, err := GenerateIBCParams("tlcpchan.local", 1, testIBCValidity())
	if err != nil {
		t.Fatalf("GenerateIBCParams() error = %v", err)
	}

	identity := []byte("client@tlcpchan.local")
	user, err := GenerateIBCIdentity(generated.Master, identity)
	if err != nil {
		t.Fatalf("GenerateIBCIdentity() error = %v", err)
	}
	if string(user.Identity) != string(identity) {
		t.Fatalf("Identity = %q, want %q", user.Identity, identity)
	}

	tests := []struct {
		name string
		pem  []byte
		hid  byte
	}{
		{"签名私钥", user.SignKeyPEM, IBCSignHid},
		{"加密私钥", user.EncKeyPEM, IBCEncryptHid},
		{"密钥交换私钥", user.KexKeyPEM, IBCKeyExchangeHid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block, _ := pem.Decode(tt.pem)
			if block == nil {
				t.Fatal("PEM 解码失败")
			}
			key, err := smx509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				t.Fatalf("ParsePKCS8PrivateKey() error = %v", err)
			}
			switch tt.hid {
			case IBCSignHid:
				if _, ok := key.(*sm9.SignPrivateKey); !ok {
					t.Fatalf("签名私钥类型 = %T, want *sm9.SignPrivateKey", key)
				}
			default:
				if _, ok := key.(*sm9.EncryptPrivateKey); !ok {
					t.Fatalf("加密/密钥交换私钥类型 = %T, want *sm9.EncryptPrivateKey", key)
				}
			}
		})
	}

	// SM9 用户私钥由主密钥与标识确定性派生，同一标识两次派生结果应完全一致
	second, err := GenerateIBCIdentity(generated.Master, identity)
	if err != nil {
		t.Fatalf("GenerateIBCIdentity() error = %v", err)
	}
	if string(user.SignKeyPEM) != string(second.SignKeyPEM) {
		t.Fatal("两次派生结果不一致，SM9 用户私钥应为确定性派生")
	}
	if string(user.KexKeyPEM) != string(second.KexKeyPEM) {
		t.Fatal("两次派生得到的密钥交换私钥不一致")
	}
}

// TestSaveIBCFilesPermissions 验证参数文件为 0644、主密钥文件为 0600。
func TestSaveIBCFilesPermissions(t *testing.T) {
	dir := t.TempDir()
	generated, err := GenerateIBCParams("tlcpchan.local", 1, testIBCValidity())
	if err != nil {
		t.Fatalf("GenerateIBCParams() error = %v", err)
	}

	paramsPath := filepath.Join(dir, "kgc.pem")
	if err := SaveIBCParamsToFile(generated.ParamsPEM, paramsPath); err != nil {
		t.Fatalf("SaveIBCParamsToFile() error = %v", err)
	}
	masterPath := filepath.Join(dir, "master", "kgc-master.key")
	if err := SaveIBCMasterToFile(generated.MasterPEM, masterPath); err != nil {
		t.Fatalf("SaveIBCMasterToFile() error = %v", err)
	}

	paramsInfo, err := os.Stat(paramsPath)
	if err != nil {
		t.Fatalf("Stat(参数文件) error = %v", err)
	}
	if paramsInfo.Mode().Perm() != 0644 {
		t.Fatalf("参数文件权限 = %v, want 0644", paramsInfo.Mode().Perm())
	}

	masterInfo, err := os.Stat(masterPath)
	if err != nil {
		t.Fatalf("Stat(主密钥文件) error = %v", err)
	}
	if masterInfo.Mode().Perm() != 0600 {
		t.Fatalf("主密钥文件权限 = %v, want 0600", masterInfo.Mode().Perm())
	}
}

// TestMarshalIBCPrivateKeyPEM 验证通用私钥 PEM 编码对主密钥与用户密钥都可用。
func TestMarshalIBCPrivateKeyPEM(t *testing.T) {
	generated, err := GenerateIBCParams("tlcpchan.local", 1, testIBCValidity())
	if err != nil {
		t.Fatalf("GenerateIBCParams() error = %v", err)
	}

	pemData, err := MarshalIBCPrivateKeyPEM(generated.Master.SignMaster)
	if err != nil {
		t.Fatalf("MarshalIBCPrivateKeyPEM() error = %v", err)
	}
	if block, _ := pem.Decode(pemData); block == nil || block.Type != "PRIVATE KEY" {
		t.Fatalf("PEM 类型 = %v, want PRIVATE KEY", block)
	}
}
