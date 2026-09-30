package keystore

import (
	"crypto/tls"
	"time"

	"gitee.com/Trisia/gotlcp/tlcp"
)

// KeyStoreType keystore 类型
type KeyStoreType string

const (
	KeyStoreTypeTLCP KeyStoreType = "tlcp"
	KeyStoreTypeTLS  KeyStoreType = "tls"
	// KeyStoreTypeIBC IBC（SM9 标识密码）身份，不使用 X.509 证书
	KeyStoreTypeIBC KeyStoreType = "ibc"
)

// KeyType 密钥类型
type KeyType string

const (
	KeyTypeSign KeyType = "sign"
	KeyTypeEnc  KeyType = "enc"
)

// KeyStore 抽象 keystore 接口
type KeyStore interface {
	Type() KeyStoreType
	TLCPCertificate() ([]*tlcp.Certificate, error)
	TLSCertificate() (*tls.Certificate, error)
}

// IBCKeyStore IBC（SM9）身份密钥存储的可选能力接口。
//
// 仅 ibc 类型的 keystore 实现；调用方通过类型断言判断本端是否具备 IBC 能力，
// 断言失败即视为无 IBC 能力（该实例的 IBC/IBSDH 套件不会参与协商）。
type IBCKeyStore interface {
	KeyStore
	// IBCIdentity 返回装载完成的 IBC 身份（标识 + KGC 公共参数 + 用户私钥）
	// 返回值：
	//   - *tlcp.IBCIdentity：装载完成的身份凭据，材料不全时仍返回非 nil（缺哪部分即为 nil 字段）
	//   - error：材料读取、解析或签名私钥自检失败时返回非 nil
	IBCIdentity() (*tlcp.IBCIdentity, error)
	// IBCInfo 返回 IBC 身份的只读元信息，供 API 与 UI 展示；装载失败时返回 nil
	IBCInfo() *IBCInfo
}

// IBCInfo IBC（SM9）keystore 的只读元信息，用于 API 与 UI 展示。
//
// 该结构不包含任何私钥内容。
type IBCInfo struct {
	// Identity 本端标识可读形式，如 "server@tlcpchan.local"；未配置时为空串
	Identity string `json:"identity" yaml:"identity"`
	// HasParams 是否提供本端 KGC 公共参数
	HasParams bool `json:"hasParams" yaml:"hasParams"`
	// DistrictName KGC 属地区域名；未提供公共参数时为空串
	DistrictName string `json:"districtName" yaml:"districtName"`
	// DistrictSerial 同一区域下的 KGC 序号；未提供公共参数时为 0
	DistrictSerial int `json:"districtSerial" yaml:"districtSerial"`
	// NotBefore 公共参数生效时间（零值表示未提供/不限）
	NotBefore time.Time `json:"notBefore" yaml:"notBefore"`
	// NotAfter 公共参数失效时间（零值表示未提供/不限）
	NotAfter time.Time `json:"notAfter" yaml:"notAfter"`
	// HasSignKey 是否提供签名私钥（hid=0x01）
	HasSignKey bool `json:"hasSignKey" yaml:"hasSignKey"`
	// HasEncryptKey 是否提供加密私钥（hid=0x03）
	HasEncryptKey bool `json:"hasEncryptKey" yaml:"hasEncryptKey"`
	// HasKeyExchangeKey 是否提供密钥交换私钥（hid=0x02）
	HasKeyExchangeKey bool `json:"hasKeyExchangeKey" yaml:"hasKeyExchangeKey"`
}

// LoaderType 加载器类型
type LoaderType string

const (
	LoaderTypeFile  LoaderType = "file"
	LoaderTypeNamed LoaderType = "named"
	LoaderTypeSKF   LoaderType = "skf"
	LoaderTypeSDF   LoaderType = "sdf"
	// LoaderTypeIBCFile IBC（SM9）身份文件加载器
	LoaderTypeIBCFile LoaderType = "ibc-file"
)

// KeyStoreInfo keystore 信息
type KeyStoreInfo struct {
	Name       string            `json:"name" yaml:"name"`
	Type       KeyStoreType      `json:"type" yaml:"type"`
	LoaderType LoaderType        `json:"loaderType" yaml:"loaderType"`
	Params     map[string]string `json:"params" yaml:"params"`
	Protected  bool              `json:"protected" yaml:"protected"`
	CreatedAt  time.Time         `json:"createdAt" yaml:"createdAt"`
	UpdatedAt  time.Time         `json:"updatedAt" yaml:"updatedAt"`
	// IBC IBC 身份只读元信息，仅 type=ibc 时非空
	IBC *IBCInfo `json:"ibc,omitempty" yaml:"ibc,omitempty"`
}
