package rootcert

import (
	"crypto/x509"
	"time"

	"github.com/emmansun/gmsm/smx509"
)

// RootCert 根证书信息
type RootCert struct {
	Filename     string            `json:"filename" yaml:"filename"`         // 证书文件名
	Cert         *x509.Certificate `json:"-" yaml:"-"`                       // 解析后的证书对象
	NotBefore    time.Time         `json:"notBefore" yaml:"notBefore"`       // 证书生效时间
	NotAfter     time.Time         `json:"notAfter" yaml:"notAfter"`         // 证书过期时间
	Subject      string            `json:"subject" yaml:"subject"`           // 证书主题
	Issuer       string            `json:"issuer" yaml:"issuer"`             // 证书颁发者
	KeyType      string            `json:"keyType" yaml:"keyType"`           // 密钥类型（如 "SM2", "RSA-2048", "ECDSA-P256"）
	SerialNumber string            `json:"serialNumber" yaml:"serialNumber"` // 证书序列号（十六进制）
	Version      int               `json:"version" yaml:"version"`           // 证书版本
	IsCA         bool              `json:"isCA" yaml:"isCA"`                 // 是否为 CA 证书
	KeyUsage     []string          `json:"keyUsage" yaml:"keyUsage"`         // 密钥用途
}

// RootCertPool 根证书池接口
type RootCertPool interface {
	GetCertPool() *x509.CertPool
	GetSMCertPool() *smx509.CertPool
	GetCerts() []*RootCert
}
