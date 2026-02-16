package rootcert

import (
	"crypto/x509"
	"time"

	"github.com/emmansun/gmsm/smx509"
)

// RootCert 根证书信息
type RootCert struct {
	Name     string            `json:"name" yaml:"name"`
	Cert     *x509.Certificate `json:"-" yaml:"-"`
	NotAfter time.Time         `json:"notAfter" yaml:"notAfter"`
	Subject  string            `json:"subject" yaml:"subject"`
	Issuer   string            `json:"issuer" yaml:"issuer"`
	AddedAt  time.Time         `json:"addedAt" yaml:"addedAt"`
}

// RootCertPool 根证书池接口
type RootCertPool interface {
	GetCertPool() *x509.CertPool
	GetSMCertPool() *smx509.CertPool
	GetCerts() []*RootCert
}
