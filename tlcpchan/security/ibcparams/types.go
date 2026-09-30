package ibcparams

import (
	"time"

	"gitee.com/Trisia/gotlcp/tlcp"
)

// IBCParam KGC 公共参数条目。
//
// 每个条目对应信任池目录下的一个文件，由 (DistrictName, DistrictSerial) 唯一标识一个 KGC。
type IBCParam struct {
	// Filename 文件名，条目在信任池中的唯一标识
	Filename string `json:"filename" yaml:"filename"`
	// DistrictName KGC 属地区域名（URI 或 IRI 编码的 IA5 字符串）
	DistrictName string `json:"districtName" yaml:"districtName"`
	// DistrictSerial 同一区域下的序号，与 DistrictName 共同唯一标识 KGC
	DistrictSerial int `json:"districtSerial" yaml:"districtSerial"`
	// NotBefore 公共参数生效时间（零值表示不限）
	NotBefore time.Time `json:"notBefore" yaml:"notBefore"`
	// NotAfter 公共参数失效时间（零值表示不限）
	NotAfter time.Time `json:"notAfter" yaml:"notAfter"`
	// IssuerIdentity 公共参数颁发者标识（KGC 标识）的可读形式
	IssuerIdentity string `json:"issuerIdentity" yaml:"issuerIdentity"`
	// SignKeyFingerprint 签名主公钥的 SM3 指纹（HEX 小写）
	SignKeyFingerprint string `json:"signKeyFingerprint" yaml:"signKeyFingerprint"`
	// EncKeyFingerprint 加密主公钥的 SM3 指纹（HEX 小写）
	EncKeyFingerprint string `json:"encKeyFingerprint" yaml:"encKeyFingerprint"`
	// Params 解析后的公共参数对象，仅供内部装载与比对使用，不对外序列化
	Params *tlcp.IBCSysParams `json:"-" yaml:"-"`
}
