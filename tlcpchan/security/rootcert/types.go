package rootcert

import (
	"crypto/x509"
	"time"

	"github.com/emmansun/gmsm/smx509"
)

// RootCert 根证书信息
type RootCert struct {
	Filename string            `json:"filename" yaml:"filename"` // 证书文件名
	Cert     *x509.Certificate `json:"-" yaml:"-"`               // 解析后的证书对象
	NotAfter time.Time         `json:"notAfter" yaml:"notAfter"` // 证书过期时间
	Subject  string            `json:"subject" yaml:"subject"`   // 证书主题
	Issuer   string            `json:"issuer" yaml:"issuer"`     // 证书颁发者
}

// RootCertPool 根证书池接口
type RootCertPool interface {
	GetCertPool() *x509.CertPool
	GetSMCertPool() *smx509.CertPool
	GetCerts() []*RootCert
}
