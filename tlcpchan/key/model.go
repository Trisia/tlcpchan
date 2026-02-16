package key

import (
	"time"
)

// KeyStoreType 密钥存储类型
type KeyStoreType string

const (
	// KeyStoreTypeTLCP 国密类型
	KeyStoreTypeTLCP KeyStoreType = "tlcp"
	// KeyStoreTypeTLS 国际类型
	KeyStoreTypeTLS KeyStoreType = "tls"
)

// KeyParams 密钥参数
type KeyParams struct {
	// Algorithm 算法：SM2/RSA/ECDSA
	Algorithm string `yaml:"algorithm" json:"algorithm"`
	// Length 密钥长度
	Length int `yaml:"length" json:"length"`
	// Type 密钥类型
	Type string `yaml:"type" json:"type"`
}

// KeyStore 密钥存储实体
type KeyStore struct {
	// Name 名称（唯一标识）
	Name string `yaml:"name" json:"name"`
	// Type 类型：国密/国际
	Type KeyStoreType `yaml:"type" json:"type"`
	// KeyParams 密钥参数
	KeyParams KeyParams `yaml:"key_params" json:"keyParams"`
	// SignCert 签名证书文件名（相对路径）
	SignCert string `yaml:"sign_cert" json:"signCert"`
	// SignKey 签名密钥文件名（相对路径）
	SignKey string `yaml:"sign_key" json:"signKey"`
	// EncCert 加密证书文件名（仅国密，相对路径）
	EncCert string `yaml:"enc_cert,omitempty" json:"encCert,omitempty"`
	// EncKey 加密密钥文件名（仅国密，相对路径）
	EncKey string `yaml:"enc_key,omitempty" json:"encKey,omitempty"`
	// CreatedAt 创建时间
	CreatedAt time.Time `yaml:"created_at" json:"createdAt"`
	// UpdatedAt 更新时间
	UpdatedAt time.Time `yaml:"updated_at" json:"updatedAt"`
}

// KeyStoreInfo 密钥存储信息（用于API响应）
type KeyStoreInfo struct {
	// Name 名称
	Name string `json:"name"`
	// Type 类型
	Type KeyStoreType `json:"type"`
	// KeyParams 密钥参数
	KeyParams KeyParams `json:"keyParams"`
	// HasSignCert 是否有签名证书
	HasSignCert bool `json:"hasSignCert"`
	// HasSignKey 是否有签名密钥
	HasSignKey bool `json:"hasSignKey"`
	// HasEncCert 是否有加密证书（仅国密）
	HasEncCert bool `json:"hasEncCert,omitempty"`
	// HasEncKey 是否有加密密钥（仅国密）
	HasEncKey bool `json:"hasEncKey,omitempty"`
	// CreatedAt 创建时间
	CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt 更新时间
	UpdatedAt time.Time `json:"updatedAt"`
}

// CreateRequest 创建密钥请求
type CreateRequest struct {
	// Name 名称
	Name string `json:"name"`
	// Type 类型
	Type KeyStoreType `json:"type"`
	// GenerateKey 是否生成密钥（true则生成，false则使用上传的密钥）
	GenerateKey bool `json:"generateKey"`
	// KeyParams 密钥参数（GenerateKey为true时需要）
	KeyParams *KeyParams `json:"keyParams,omitempty"`
}

// UpdateCertificatesRequest 更新证书请求
type UpdateCertificatesRequest struct {
	// UpdateSignCert 是否更新签名证书
	UpdateSignCert bool `json:"updateSignCert"`
	// UpdateEncCert 是否更新加密证书（仅国密）
	UpdateEncCert bool `json:"updateEncCert,omitempty"`
}

const (
	// FileNameSignCert 签名证书文件名
	FileNameSignCert = "sign.crt"
	// FileNameSignKey 签名密钥文件名
	FileNameSignKey = "sign.key"
	// FileNameEncCert 加密证书文件名
	FileNameEncCert = "enc.crt"
	// FileNameEncKey 加密密钥文件名
	FileNameEncKey = "enc.key"
	// FileNameInfo 信息文件名
	FileNameInfo = "info.yaml"
)
