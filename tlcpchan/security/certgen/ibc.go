package certgen

import (
	"crypto"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/emmansun/gmsm/sm9"
	"github.com/emmansun/gmsm/smx509"
)

// IBC 材料的 PEM 类型标识，遵循 GM/T 0081-2020 对 IBCSysParams 的命名。
const (
	// IBCParamsPEMType KGC 公共参数 PEM 类型
	IBCParamsPEMType = "IBC PARAMETERS"
	// IBCSignMasterPEMType SM9 签名主密钥 PEM 类型
	IBCSignMasterPEMType = "SM9 SIGN MASTER KEY"
	// IBCEncryptMasterPEMType SM9 加密主密钥 PEM 类型
	IBCEncryptMasterPEMType = "SM9 ENCRYPT MASTER KEY"
)

// SM9 用户私钥的派生用途标识（hid），与 GM/T 0044 定义一致。
const (
	// IBCSignHid 签名私钥用途标识
	IBCSignHid byte = 0x01
	// IBCKeyExchangeHid 密钥交换私钥用途标识
	IBCKeyExchangeHid byte = 0x02
	// IBCEncryptHid 加密私钥用途标识
	IBCEncryptHid byte = 0x03
)

// IBCMaster 一套 SM9 KGC 主密钥对，用于派生用户私钥与生成公共参数。
//
// 主密钥是测试 KGC 的核心秘密，只允许保存在本地（0600），不得对外提供下载。
type IBCMaster struct {
	// SignMaster 签名主私钥，用于派生签名用户私钥（hid=0x01）
	SignMaster *sm9.SignMasterPrivateKey
	// EncryptMaster 加密主私钥，用于派生加密用户私钥（hid=0x03）与密钥交换用户私钥（hid=0x02）
	EncryptMaster *sm9.EncryptMasterPrivateKey
}

// GeneratedIBCParams 生成的 KGC 公共参数及其 PEM 编码。
type GeneratedIBCParams struct {
	// Params 解析完成的公共参数对象
	Params *tlcp.IBCSysParams
	// Master 生成的主密钥对，用于随后派生用户私钥
	Master *IBCMaster
	// ParamsPEM 公共参数的 PEM 编码，用于写入信任池目录
	ParamsPEM []byte
	// MasterPEM 主密钥的 PEM 编码（两个 PEM 块），必须以 0600 权限保存
	MasterPEM []byte
}

// GeneratedIBCIdentity 由 KGC 主密钥派生出的标识与三把用户私钥。
type GeneratedIBCIdentity struct {
	// Identity 本端标识字节串
	Identity []byte
	// SignKeyPEM 签名用户私钥（hid=0x01）的 PKCS#8 PEM
	SignKeyPEM []byte
	// EncKeyPEM 加密用户私钥（hid=0x03）的 PKCS#8 PEM
	EncKeyPEM []byte
	// KexKeyPEM 密钥交换用户私钥（hid=0x02）的 PKCS#8 PEM
	KexKeyPEM []byte
}

// GenerateIBCParams 生成一套 SM9 KGC 主密钥并导出公共参数。
//
// 参数：
//   - districtName: KGC 属地区域名，同时作为参数中的 pkgID 与 issuerID 的 identityData
//   - districtSerial: KGC 序号，与 districtName 共同唯一标识该 KGC
//   - validity: 公共参数有效期
//
// 返回值：
//   - *GeneratedIBCParams: 公共参数对象与其 PEM 编码、主密钥 PEM 编码
//   - error: 主密钥生成、公共参数构造或 PEM 编码失败时返回错误
//
// 注意事项：
//   - 该能力仅供内置测试 KGC 使用，生产环境应由外部 KGC 派生后导入
func GenerateIBCParams(districtName string, districtSerial int, validity tlcp.ValidityPeriod) (*GeneratedIBCParams, error) {
	signMaster, err := sm9.GenerateSignMasterKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("生成 SM9 签名主密钥失败: %w", err)
	}
	encMaster, err := sm9.GenerateEncryptMasterKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("生成 SM9 加密主密钥失败: %w", err)
	}

	params, err := tlcp.NewIBCSysParamsFromMaster(districtName, districtSerial, validity, signMaster, encMaster)
	if err != nil {
		return nil, fmt.Errorf("生成 IBC 公共参数失败: %w", err)
	}

	paramsPEM := pem.EncodeToMemory(&pem.Block{Type: IBCParamsPEMType, Bytes: params.Raw})

	signMasterPEM, err := marshalIBCMasterPEM(signMaster, IBCSignMasterPEMType)
	if err != nil {
		return nil, err
	}
	encMasterPEM, err := marshalIBCMasterPEM(encMaster, IBCEncryptMasterPEMType)
	if err != nil {
		return nil, err
	}

	return &GeneratedIBCParams{
		Params:    params,
		Master:    &IBCMaster{SignMaster: signMaster, EncryptMaster: encMaster},
		ParamsPEM: paramsPEM,
		MasterPEM: append(signMasterPEM, encMasterPEM...),
	}, nil
}

// LoadIBCMasterFromPEM 从 PEM 数据装载 KGC 主密钥对。
//
// 参数：
//   - data: 主密钥 PEM 数据，需包含签名主密钥与加密主密钥两个 PEM 块
//
// 返回值：
//   - *IBCMaster: 装载完成的主密钥对
//   - error: 缺少任一主密钥块或解码失败时返回错误
func LoadIBCMasterFromPEM(data []byte) (*IBCMaster, error) {
	master := &IBCMaster{}
	rest := data
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		switch block.Type {
		case IBCSignMasterPEMType:
			key, err := smx509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("解析 SM9 签名主密钥失败: %w", err)
			}
			signMaster, ok := key.(*sm9.SignMasterPrivateKey)
			if !ok {
				return nil, fmt.Errorf("SM9 签名主密钥类型不符: %T", key)
			}
			master.SignMaster = signMaster
		case IBCEncryptMasterPEMType:
			key, err := smx509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("解析 SM9 加密主密钥失败: %w", err)
			}
			encMaster, ok := key.(*sm9.EncryptMasterPrivateKey)
			if !ok {
				return nil, fmt.Errorf("SM9 加密主密钥类型不符: %T", key)
			}
			master.EncryptMaster = encMaster
		}
	}

	if master.SignMaster == nil || master.EncryptMaster == nil {
		return nil, fmt.Errorf("主密钥文件缺少签名主密钥或加密主密钥")
	}
	return master, nil
}

// GenerateIBCIdentity 由 KGC 主密钥按标识派生三把用户私钥。
//
// 参数：
//   - master: KGC 主密钥对，不能为 nil
//   - identity: 本端标识字节串，如 "server@tlcpchan.local"，不能为空
//
// 返回值：
//   - *GeneratedIBCIdentity: 标识与三把用户私钥的 PKCS#8 PEM
//   - error: 主密钥为空、标识为空或派生/编码失败时返回错误
//
// 注意事项：
//   - 一次性派生签名（hid=0x01）、加密（hid=0x03）、密钥交换（hid=0x02）三把私钥
func GenerateIBCIdentity(master *IBCMaster, identity []byte) (*GeneratedIBCIdentity, error) {
	if master == nil || master.SignMaster == nil || master.EncryptMaster == nil {
		return nil, fmt.Errorf("KGC 主密钥不完整")
	}
	if len(identity) == 0 {
		return nil, fmt.Errorf("IBC 标识不能为空")
	}

	signKey, err := master.SignMaster.GenerateUserKey(identity, IBCSignHid)
	if err != nil {
		return nil, fmt.Errorf("派生签名私钥失败: %w", err)
	}
	encKey, err := master.EncryptMaster.GenerateUserKey(identity, IBCEncryptHid)
	if err != nil {
		return nil, fmt.Errorf("派生加密私钥失败: %w", err)
	}
	kexKey, err := master.EncryptMaster.GenerateUserKey(identity, IBCKeyExchangeHid)
	if err != nil {
		return nil, fmt.Errorf("派生密钥交换私钥失败: %w", err)
	}

	signPEM, err := MarshalIBCPrivateKeyPEM(signKey)
	if err != nil {
		return nil, fmt.Errorf("编码签名私钥失败: %w", err)
	}
	encPEM, err := MarshalIBCPrivateKeyPEM(encKey)
	if err != nil {
		return nil, fmt.Errorf("编码加密私钥失败: %w", err)
	}
	kexPEM, err := MarshalIBCPrivateKeyPEM(kexKey)
	if err != nil {
		return nil, fmt.Errorf("编码密钥交换私钥失败: %w", err)
	}

	return &GeneratedIBCIdentity{
		Identity:   identity,
		SignKeyPEM: signPEM,
		EncKeyPEM:  encPEM,
		KexKeyPEM:  kexPEM,
	}, nil
}

// MarshalIBCPrivateKeyPEM 将 SM9 用户私钥编码为 PKCS#8 PEM。
//
// 参数：
//   - key: SM9 用户私钥（*sm9.SignPrivateKey 或 *sm9.EncryptPrivateKey）
//
// 返回值：
//   - []byte: "PRIVATE KEY" 类型的 PEM 数据
//   - error: 私钥类型不支持或编码失败时返回错误
func MarshalIBCPrivateKeyPEM(key crypto.PrivateKey) ([]byte, error) {
	derData, err := smx509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: derData}), nil
}

// SaveIBCParamsToFile 将 KGC 公共参数写入指定文件（0644）。
//
// 参数：
//   - paramsPEM: 公共参数 PEM 数据
//   - path: 目标文件路径
//
// 返回值：
//   - error: 目录创建或写入失败时返回错误
func SaveIBCParamsToFile(paramsPEM []byte, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	if err := os.WriteFile(path, paramsPEM, 0644); err != nil {
		return fmt.Errorf("写入公共参数失败: %w", err)
	}
	return nil
}

// SaveIBCMasterToFile 将 KGC 主密钥写入指定文件（0600）。
//
// 参数：
//   - masterPEM: 主密钥 PEM 数据
//   - path: 目标文件路径
//
// 返回值：
//   - error: 目录创建或写入失败时返回错误
//
// 注意事项：
//   - 主密钥文件权限固定 0600，且不通过任何 API 提供下载
func SaveIBCMasterToFile(masterPEM []byte, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	if err := os.WriteFile(path, masterPEM, 0600); err != nil {
		return fmt.Errorf("写入主密钥失败: %w", err)
	}
	return nil
}

// marshalIBCMasterPEM 将 SM9 主私钥编码为指定类型的 PEM 块。
//
// 参数：
//   - key: SM9 主私钥
//   - pemType: PEM 类型标识
//
// 返回值：
//   - []byte: PEM 数据
//   - error: PKCS#8 编码失败时返回错误
func marshalIBCMasterPEM(key crypto.PrivateKey, pemType string) ([]byte, error) {
	derData, err := smx509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("编码 SM9 主密钥失败: %w", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: pemType, Bytes: derData}), nil
}
